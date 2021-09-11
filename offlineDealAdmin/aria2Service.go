package offlineDealAdmin

import (
	"encoding/json"
	"fmt"
	"github.com/jasonlvhit/gocron"
	"log"
	"net/url"
	"strconv"
	"strings"
	"swan-miner/common"
	"swan-miner/common/utils"
	"swan-miner/config"
	"swan-miner/logs"
	"swan-miner/models"
	"time"
)

const MAX_DOWNLOADING_TASKS = 10

const DEAL_DOWNLOADING_STATUS = "Downloading"
const DEAL_DOWNLOADED_STATUS = "Downloaded"
const DEAL_DOWNLOAD_FAILED_STATUS = "DownloadFailed"
const DEAL_CREATED_STATUS = "Created"
const DEAL_WAITING_STATUS = "Waiting"

const ARIA2_TASK_ERROR_STATUS = "error"
const ARIA2_TASK_ACTIVE_STATUS = "active"
const ARIA2_TASK_COMPLETE_STATUS = "complete"

var logger = logs.GetLogger()

type DownloadOption struct {
	Out string   `json:"out"`
	Dir string   `json:"dir"`
}

type Aria2Service struct {
	MinerFid string
	OutDir   string
}

type Aria2GetStatusSuccess struct {
	Id 		string            `json:"id"`
	JsonRpc string            `json:"jsonrpc"`
	Result 	*Aria2StatusResult `json:"result"`
}

type Aria2GetStatusFail struct {
	Id 		string            `json:"id"`
	JsonRpc string            `json:"jsonrpc"`
	Error 	*Aria2StatusError  `json:"error"`
}

type Aria2StatusError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Aria2StatusResult struct {
	Bitfield        string                  `json:"bitfield"`
	CompletedLength int                     `json:"completedLength"`
	Connections     string                  `json:"connections"`
	Dir             string                  `json:"dir"`
	DownloadSpeed   int                     `json:"downloadSpeed"`
	ErrorCode       string                  `json:"errorCode"`
	ErrorMessage    string                  `json:"errorMessage"`
	Gid             string                  `json:"gid"`
	NumPieces       string                  `json:"numPieces"`
	PieceLength     string                  `json:"pieceLength"`
	Status          string                  `json:"status"`
	TotalLength     int                     `json:"totalLength"`
	UploadLength    string                  `json:"uploadLength"`
	UploadSpeed     string                  `json:"uploadSpeed"`
	Files           []Aria2StatusResultFile `json:"files"`
}

type Aria2StatusResultFile struct {
	CompletedLength int                        `json:"completedLength"`
	Index           string                     `json:"index"`
	Length          int                        `json:"length"`
	Path            string                     `json:"path"`
	Selected        string                     `json:"selected"`
	Uris            []Aria2StatusResultFileUri `json:"uris"`
}

type Aria2StatusResultFileUri struct {
	Status string `json:"status"`
	Uri    string `json:"uri"`
}

func GetAria2Service() (*Aria2Service){
	aria2Service := &Aria2Service{
		MinerFid: config.GetConfig().Main.MinerFid,
		OutDir: config.GetConfig().Aria2.Aria2DownloadDir,
	}

	return aria2Service
}

func isCompleted(taskState *Aria2StatusResult) (bool){
	if taskState.ErrorCode != "0" || taskState.TotalLength == 0{
		return false
	}

	if taskState.Status == ARIA2_TASK_COMPLETE_STATUS && taskState.CompletedLength == taskState.TotalLength{
		return true
	}

	return false
}

func  (self *Aria2Service) findNextDealReady2Download(swanClient *utils.SwanClient) (*models.OfflineDeal) {
	deals := swanClient.GetOfflineDeals(self.MinerFid, DEAL_CREATED_STATUS, "1")
	if len(deals) == 0 {
		deals = swanClient.GetOfflineDeals(self.MinerFid, DEAL_WAITING_STATUS, "1")
	}

	if len(deals)>0{
		offlineDeal := deals[0]
		return &offlineDeal
	}

	return nil
}

func (self *Aria2Service) CheckDownloadStatus4Deal(aria2Client *utils.Aria2Client, swanClient *utils.SwanClient, deal *models.OfflineDeal, gid string) {
	response := aria2Client.GetDownloadStatus(gid)
	aria2GetStatusSuccess := Aria2GetStatusSuccess{}
	json.Unmarshal([]byte(response), &aria2GetStatusSuccess)
	if aria2GetStatusSuccess.Result == nil {
		aria2GetStatusFail := Aria2GetStatusFail{}
		json.Unmarshal([]byte(response), &aria2GetStatusFail)
		code := aria2GetStatusFail.Error.Code
		message := aria2GetStatusFail.Error.Message
		msg := fmt.Sprintf("Get status for %s, code:%s, message:%s", gid, code, message)
		swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_DOWNLOAD_FAILED_STATUS, msg)
		logger.Error(msg)
		return
	}

	if len(aria2GetStatusSuccess.Result.Files) != 1 {
		note := "wrong file amount"
		swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_DOWNLOAD_FAILED_STATUS, note)
		logs.GetLogger().Error(note)
		return
	}

	result := aria2GetStatusSuccess.Result
	code := result.ErrorCode
	message := result.ErrorMessage
	status := result.Status
	file := aria2GetStatusSuccess.Result.Files[0]
	filePath := file.Path
	fileSize := file.Length
	completedLen := file.CompletedLength // utils.GetFieldStrFromJson(taskState, "completedLength")
	var completePercent = 0
	if fileSize > 0 {
		completePercent = completedLen / fileSize * 100
	}
	downloadSpeed := result.DownloadSpeed/1000

	switch status {
	case ARIA2_TASK_ERROR_STATUS:
		note := fmt.Sprintf("Deal status for %s, code:%s, message:%s, status:%s", gid, code, message, status)
		swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_DOWNLOAD_FAILED_STATUS, note)
		logger.Error(note)
	case ARIA2_TASK_ACTIVE_STATUS:
		if deal.Status != DEAL_DOWNLOADING_STATUS {
			swanClient.UpdateOfflineDealDetails(deal.Id, DEAL_DOWNLOADING_STATUS, gid, filePath, strconv.Itoa(fileSize))
		}
		logger.Info("Deal downloading, id: %s, complete: %s%%, speed: %sKiB", deal.Id, completePercent, downloadSpeed)
	case ARIA2_TASK_COMPLETE_STATUS:
		swanClient.UpdateOfflineDealDetails(deal.Id, DEAL_DOWNLOADED_STATUS, gid, filePath, strconv.Itoa(fileSize))
	default:
		note := fmt.Sprintf("download failed, cause: %s", result.ErrorMessage)
		if note != deal.Note{
			swanClient.UpdateOfflineDealDetails(deal.Id, DEAL_DOWNLOAD_FAILED_STATUS, note, filePath, strconv.Itoa(fileSize))
		}
		logger.Error(note + " dealId:" + strconv.Itoa(deal.Id))
	}
}

func (self *Aria2Service) findDealsByStatus(status string, swanClient *utils.SwanClient) ([]models.OfflineDeal){
	deals := swanClient.GetOfflineDeals(self.MinerFid, status, strconv.Itoa(common.GET_OFFLINEDEAL_LIMIT_DEFAULT))
	return deals
}

func (self *Aria2Service) CheckDownloadStatus(aria2Client *utils.Aria2Client, swanClient *utils.SwanClient) {
	downloadingDeals := self.findDealsByStatus(DEAL_DOWNLOADING_STATUS, swanClient)

	for _, deal := range downloadingDeals {
		//fmt.Println(deal)
		gid := deal.Note
		if len(gid) <= 0 {
			note := "download gid not found in offline_deals.note"
			if note != deal.Note{
				swanClient.UpdateOfflineDealDetails(deal.Id, DEAL_DOWNLOAD_FAILED_STATUS, note, deal.FilePath, deal.FileSize)
			}
			continue
		}

		self.CheckDownloadStatus4Deal(aria2Client, swanClient, &deal, gid)
	}
}

func (self *Aria2Service) StartDownload4Deal(deal *models.OfflineDeal, aria2Client *utils.Aria2Client, swanClient *utils.SwanClient) {
	logs.GetLogger().Info("start downloading deal id ", deal.Id)
	url, err := url.Parse(deal.SourceFileUrl)
	if err != nil {
		log.Fatal(err)
	}

	filename := url.Path
	if strings.HasPrefix(url.RawQuery, "filename=") {
		filename = strings.TrimLeft(url.RawQuery, "filename=")
		filename = utils.GetDir(url.Path, filename)
	}
	today := time.Now()
	timeStr := fmt.Sprintf("%d%02d", today.Year(), today.Month())
	outDir := utils.GetDir(self.OutDir, strconv.Itoa(deal.UserId), timeStr)
	option := DownloadOption{
		Out: filename,
		Dir: outDir,
	}

	if utils.IsFileExists(outDir, filename) {
		utils.RemoveFile(outDir, filename)
	}

	response := aria2Client.DownloadFile(deal.SourceFileUrl, option)
	fmt.Println(response)

	gid := utils.GetFieldStrFromJson(response, "result")
	self.CheckDownloadStatus4Deal(aria2Client, swanClient, deal, gid)
}

func (self *Aria2Service) StartDownloading(aria2Client *utils.Aria2Client, swanClient *utils.SwanClient) {
	downloadingDeals := self.findDealsByStatus(DEAL_DOWNLOADING_STATUS, swanClient)
	countDownloadingDeals := len(downloadingDeals)
	if countDownloadingDeals >= MAX_DOWNLOADING_TASKS {
		return
	}

	for i := 1; i <= MAX_DOWNLOADING_TASKS - countDownloadingDeals; i++ {
		deal2Download := self.findNextDealReady2Download(swanClient)

		if deal2Download == nil {
			break
		}

		self.StartDownload4Deal(deal2Download, aria2Client, swanClient)
		time.Sleep(1 * time.Second)
	}
}

func Downloader(){
	aria2Client := utils.GetAria2Client()
	swanClient := utils.GetSwanClient()
	aria2Service := GetAria2Service()

	gocron.Every(1).Minute().Do(func (){
		//fmt.Println(1)
		aria2Service.CheckDownloadStatus(aria2Client, swanClient)
	})

	for {
		aria2Service.StartDownloading(aria2Client, swanClient)
		time.Sleep(60 * time.Second)
	}
}



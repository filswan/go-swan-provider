package offlineDealAdmin

import (
	"encoding/json"
	"fmt"
	"github.com/jasonlvhit/gocron"
	"net/url"
	"strconv"
	"strings"
	"swan-miner/common/utils"
	"swan-miner/config"
	"swan-miner/models"
	"time"
)

type DownloadOption struct {
	Out string   `json:"out"`
	Dir string   `json:"dir"`
}

type Aria2Service struct {
	MinerFid string
	OutDir   string
}

type Aria2GetStatusSuccess struct {
	Id 		string             `json:"id"`
	JsonRpc string             `json:"jsonrpc"`
	Result 	*Aria2StatusResult `json:"result"`
}

type Aria2GetStatusFail struct {
	Id 		string             `json:"id"`
	JsonRpc string             `json:"jsonrpc"`
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
	ErrorCode       int                     `json:"errorCode"`
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
	CompletedLength int64                      `json:"completedLength"`
	Index           string                     `json:"index"`
	Length          int64                      `json:"length"`
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

func  (self *Aria2Service) findNextDealReady2Download(swanClient *utils.SwanClient) (*models.OfflineDeal) {
	deals := swanClient.GetOfflineDeals(self.MinerFid, DEAL_STATUS_CREATED, "1")
	if len(deals) == 0 {
		deals = swanClient.GetOfflineDeals(self.MinerFid, DEAL_STATUS_WAITING, "1")
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
		msg := fmt.Sprintf("Get status for %s, code:%d, message:%s", gid, code, message)
		swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, msg)
		logger.Error(msg)
		return
	}

	if len(aria2GetStatusSuccess.Result.Files) != 1 {
		note := "Wrong file amount"
		swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note)
		logger.Error(note)
		return
	}

	result := aria2GetStatusSuccess.Result
	code := result.ErrorCode
	message := result.ErrorMessage
	status := result.Status
	file := aria2GetStatusSuccess.Result.Files[0]
	filePath := file.Path
	fileSize := file.Length
	completedLen := file.CompletedLength
	var completePercent int64 = 0
	if fileSize > 0 {
		completePercent = completedLen / fileSize * 100
	}
	downloadSpeed := result.DownloadSpeed/1000

	switch status {
	case ARIA2_TASK_STATUS_ERROR:
		note := fmt.Sprintf("Deal status for %s, code:%d, message:%s, status:%s", gid, code, message, status)
		swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note)
		logger.Error(note)
	case ARIA2_TASK_STATUS_ACTIVE:
		fileSizeDownloaded := utils.GetFileSize(filePath)
		if deal.Status != DEAL_STATUS_DOWNLOADING {
			swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOADING, gid, filePath, utils.GetStrFromInt64(fileSizeDownloaded))
		}
		msg := fmt.Sprintf("Deal downloading, id: %d, file size: %d, complete: %d%%, speed: %dKiB", deal.Id, fileSize, completePercent, downloadSpeed)
		logger.Info(msg)
	case ARIA2_TASK_STATUS_COMPLETE:
		fileSizeDownloaded := utils.GetFileSize(filePath)
		if fileSizeDownloaded >= 0 {
			swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOADED, gid, filePath, utils.GetStrFromInt64(fileSizeDownloaded))
		} else {
			note := fmt.Sprintf("File %s not found on", filePath)
			swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note, filePath, utils.GetStrFromInt64(fileSize))
			logger.Error(note)
		}
	default:
		note := fmt.Sprintf("Download failed, cause: %s", result.ErrorMessage)
		if note != deal.Note{
			swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note, filePath, utils.GetStrFromInt64(fileSize))
		}
		logger.Error(note, " dealId:", strconv.Itoa(deal.Id))
	}
}

func (self *Aria2Service) CheckDownloadStatus(aria2Client *utils.Aria2Client, swanClient *utils.SwanClient) {
	logger.Info("Start checking download status.")
	downloadingDeals := swanClient.GetOfflineDeals(self.MinerFid, DEAL_STATUS_DOWNLOADING)

	for _, deal := range downloadingDeals {
		//fmt.Println(deal)
		gid := deal.Note
		if len(gid) <= 0 {
			note := "Download gid not found in offline_deals.note"
			if note != deal.Note{
				swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note)
			}
			continue
		}

		self.CheckDownloadStatus4Deal(aria2Client, swanClient, &deal, gid)
	}
}

func (self *Aria2Service) StartDownload4Deal(deal *models.OfflineDeal, aria2Client *utils.Aria2Client, swanClient *utils.SwanClient) {
	logger.Info("start downloading deal id ", deal.Id)
	url, err := url.Parse(deal.SourceFileUrl)
	if err != nil {
		msg := fmt.Sprintf("parse source file url error:%s", err)
		swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, msg)
		msg = fmt.Sprintf("Deal id:%d, %s", deal.Id, msg)
		logger.Error(msg)
		return
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

func (self *Aria2Service) StartDownload(aria2Client *utils.Aria2Client, swanClient *utils.SwanClient) {
	logger.Info("Start download.")
	downloadingDeals := swanClient.GetOfflineDeals(self.MinerFid, DEAL_STATUS_DOWNLOADING)
	countDownloadingDeals := len(downloadingDeals)
	if countDownloadingDeals >= ARIA2_MAX_DOWNLOADING_TASKS {
		return
	}

	for i := 1; i <= ARIA2_MAX_DOWNLOADING_TASKS- countDownloadingDeals; i++ {
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
		aria2Service.StartDownload(aria2Client, swanClient)
		logger.Info("Sleeping...")
		time.Sleep(60 * time.Second)
	}
}



package offlineDealAdmin

import (
	"encoding/json"
	"fmt"
	"github.com/jasonlvhit/gocron"
	"log"
	"net/url"
	"strconv"
	"strings"
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

const ARIA2_TASK_ACTIVE_STATUS = "active"
const ARIA2_TASK_COMPLETE_STATUS = "complete"

type DownloadOption struct {
	Out string   `json:"out"`
	Dir string   `json:"dir"`
}

type Aria2Service struct {
	MinerFid string
	OutDir   string
}

type Aria2StatusSuccess struct {
	Id 		string            `json:"id"`
	JsonRpc string            `json:"jsonrpc"`
	Result 	Aria2StatusResult `json:"result"`
}

type Aria2StatusFail struct {
	Id 		string            `json:"id"`
	JsonRpc string            `json:"jsonrpc"`
	Error 	Aria2StatusError  `json:"error"`
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
	CompletedLength string                     `json:"completedLength"`
	Index           string                     `json:"index"`
	Length          string                     `json:"length"`
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

func isCompleted(taskState Aria2StatusResult) (bool){
	if taskState.ErrorCode != "0" || taskState.TotalLength == 0{
		return false
	}

	if taskState.Status ==ARIA2_TASK_COMPLETE_STATUS && taskState.CompletedLength == taskState.TotalLength{
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

func (self *Aria2Service) findDealsByStatus(status string, swanClient *utils.SwanClient) ([]models.OfflineDeal){
	deals := swanClient.GetOfflineDeals(self.MinerFid, status, "50")
	return deals
}

func (self *Aria2Service) StartDownloadForDeal(offlineDeal *models.OfflineDeal, aria2Client *utils.Aria2Client, swanClient *utils.SwanClient) {
	logs.GetLogger().Info("start downloading deal id ", offlineDeal.Id)
	url, err := url.Parse(offlineDeal.SourceFileUrl)
	if err != nil {
		log.Fatal(err)
	}
	filename := url.Path
	today := time.Now()
	timeStr := fmt.Sprintf("%d%02d", today.Year(), today.Month())
	option := DownloadOption{
		Out: filename,
		Dir: utils.GetDir(self.OutDir, strconv.Itoa(offlineDeal.UserId), timeStr),
	}
	response := aria2Client.DownloadFile(offlineDeal.SourceFileUrl, option)
	fmt.Println(response)

	gid := utils.GetFieldStrFromJson(response, "result")
	response = aria2Client.GetDownloadStatus(gid)
	if strings.Contains(response, "\"error\""){
		aria2StatusFail := Aria2StatusFail{}
		json.Unmarshal([]byte(response),&aria2StatusFail)
		code := aria2StatusFail.Error.Code
		message := aria2StatusFail.Error.Message
		msg := fmt.Sprintf("Get status for %s, code:%s, message:%s", gid,code,message)
		logs.GetLogger().Error(msg)
		return
	}

	aria2StatusSuccess := Aria2StatusSuccess{}
	json.Unmarshal([]byte(response),&aria2StatusSuccess)

	if len(aria2StatusSuccess.Result.Files) !=1 {
		logs.GetLogger().Error("wrong file amount")
		return
	}
	filePath := aria2StatusSuccess.Result.Files[0].Path
	fileSize := aria2StatusSuccess.Result.Files[0].Length
	swanClient.UpdateOfflineDealDetails(DEAL_DOWNLOADING_STATUS, gid, offlineDeal.Id, filePath, fileSize)
}

func (self *Aria2Service) CheckDownloadStatus(aria2Client *utils.Aria2Client, swanClient *utils.SwanClient) {
	downloadingDeals := self.findDealsByStatus(DEAL_DOWNLOADING_STATUS, swanClient)

	for i := 0; i < len(downloadingDeals); i++ {
		deal :=downloadingDeals[i]
		//fmt.Println(deal)
		gid := deal.Note
		if len(gid) <= 0 {
			note := "download gid not found in offline_deals.note"
			if note != deal.Note{
				swanClient.UpdateOfflineDealDetails(DEAL_DOWNLOAD_FAILED_STATUS, note, deal.Id, deal.FilePath, deal.FileSize)
			}
			continue
		}

		response := aria2Client.GetDownloadStatus(gid)
		if strings.Contains(response, "error"){
			aria2StatusFail := Aria2StatusFail{}
			json.Unmarshal([]byte(response),&aria2StatusFail)
			note := aria2StatusFail.Error.Message
			if note != deal.Note {
				swanClient.UpdateOfflineDealDetails(DEAL_DOWNLOAD_FAILED_STATUS, note, deal.Id, deal.FilePath, deal.FileSize)
			}
			continue
		}

		aria2StatusSuccess := Aria2StatusSuccess{}
		json.Unmarshal([]byte(response),&aria2StatusSuccess)

		taskState := aria2StatusSuccess.Result //. utils.GetFieldStrFromJson(response, "result")
		status := taskState.Status //status := utils.GetFieldSt

		if status == ARIA2_TASK_ACTIVE_STATUS {
			completedLen := taskState.CompletedLength // utils.GetFieldStrFromJson(taskState, "completedLength")
			totalLen := taskState.TotalLength
			completePercent := completedLen / totalLen * 100
			downloadSpeed := taskState.DownloadSpeed/1000

			logs.GetLogger().Info("continue downloading deal id %s complete %s%% speed %s KiB", deal.Id, completePercent, downloadSpeed)
			continue
		}

		if isCompleted(taskState) {
			fileSize := taskState.CompletedLength
			swanClient.UpdateOfflineDealDetails(DEAL_DOWNLOADED_STATUS, gid, deal.Id, deal.FilePath, string(fileSize))
			continue
		}

		note := fmt.Sprintf("download failed, cause: %s",taskState.ErrorMessage)
		if note!=deal.Note{
			swanClient.UpdateOfflineDealDetails(DEAL_DOWNLOAD_FAILED_STATUS, note, deal.Id, deal.FilePath, deal.FileSize)
		}
	}
}

func (self *Aria2Service) StartDownloading(aria2Client *utils.Aria2Client, swanClient *utils.SwanClient) {
	for{
		downloadingDeals := self.findDealsByStatus(DEAL_DOWNLOADING_STATUS, swanClient)
		countDownloadingDeals := len(downloadingDeals)
		if MAX_DOWNLOADING_TASKS > countDownloadingDeals {
			newTaskNum := MAX_DOWNLOADING_TASKS - countDownloadingDeals
			i := 1
			for i<=newTaskNum{
				deal2Download := self.findNextDealReady2Download(swanClient)

				if deal2Download==nil {
					break
				}

				self.StartDownloadForDeal(deal2Download, aria2Client, swanClient)
				time.Sleep(1 * time.Second)
			}

			time.Sleep(60 * time.Second)
		}
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

	aria2Service.StartDownloading(aria2Client, swanClient)
}



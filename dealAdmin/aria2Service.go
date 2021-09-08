package dealAdmin

import (
	"fmt"
	"log"
	"net/url"
	"reflect"
	"strconv"
	"swan-miner/common/utils"
	"swan-miner/logs"
	"time"
)

const DEAL_DOWNLOADING_STATUS = "Downloading"
const DEAL_DOWNLOADED_STATUS = "Downloaded"
const DEAL_DOWNLOAD_FAILED_STATUS = "DownloadFailed"
const DEAL_CREATED_STATUS = "Created"
const DEAL_WAITING_STATUS = "Waiting"

const ARIA2_TASK_ACTIVE_STATUS = "active"
const ARIA2_TASK_COMPLETE_STATUS = "complete"


func isCompleted(task string) (bool){
	errCode := utils.GetFieldFromJson(task, "errorCode")
	if errCode!="0"{
		return false
	}

	totalLength := utils.GetFieldFromJson(task, "totalLength")
	if totalLength=="0"{
		return false
	}

	status := utils.GetFieldFromJson(task, "status")
	comletedLength := utils.GetFieldFromJson(task, "completedLength")
	if status ==ARIA2_TASK_COMPLETE_STATUS && comletedLength == totalLength{
		return true
	}

	return false
}

func updateOfflineDealDetails(status, note, dealId string, swanClient *utils.SwanClient, filePath string, fileSize string) {
	swanClient.UpdateOfflineDealDetails(status,note,dealId, filePath, fileSize)
}

func findNextDealReady2Download(minerFid string, swanClient *utils.SwanClient) (interface{}) {
	deals := swanClient.GetOfflineDeals(minerFid, DEAL_CREATED_STATUS, "1")
	if len(deals) == 0 {
		deals = swanClient.GetOfflineDeals(minerFid, DEAL_WAITING_STATUS, "1")
	}

	return deals
}

func findDealsByStatus(status, minerFid string, swanClient *utils.SwanClient) ([]utils.OfflineDeal){
	deals := swanClient.GetOfflineDeals(minerFid, status, "50")
	return deals
}

func startDownloadForDeal(deal, outDir string, aria2Client Aria2c, swanClient *utils.SwanClient) {
	logs.GetLogger().Info("start downloading deal id ",utils.GetFieldFromJson(deal,"id"))
	fileResourceUrl := utils.GetFieldFromJson(deal, "file_source_url")
	url, err := url.Parse(fileResourceUrl.(string))
	if err != nil {
		log.Fatal(err)
	}
	filename := url.Path
	userId := utils.GetFieldFromJson(deal, "user_id").(string)
	today := time.Now()
	timeStr := fmt.Sprintf("%d%02d", today.Year(), today.Month())
	option := fmt.Sprintf("{\"out\": %s,\"dir\": %s + %s + '/' + %s})",filename, outDir, userId, timeStr)
	response := aria2Client.addUri(fileResourceUrl.(string), option)
	gid := utils.GetFieldFromJson(response, "result")
	response = aria2Client.post(STATUS, gid.(string))
}

func checkDownloadStatus(aria2c Aria2c, swanClient *utils.SwanClient, minerFid string) {
	downloadingDeals := findDealsByStatus(DEAL_DOWNLOADING_STATUS, minerFid, swanClient)

	switch reflect.TypeOf(downloadingDeals).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(downloadingDeals)

		for i := 0; i < s.Len(); i++ {
			fmt.Println(s.Index(i))

			deal :=s.Index(i).String()
			dealId := utils.GetFieldStrFromJson(deal, "id")
			currentStatus := utils.GetFieldFromJson(deal, "status")
			note:=utils.GetFieldStrFromJson(deal, "note")
			response := aria2c.post(note,"")

			var fileSize string
			var newStatus string

			if (len(note)>0) {
				taskState := utils.GetFieldStrFromJson(response, "result")
				if (len(response) > 0) {
					status := utils.GetFieldStrFromJson(taskState, "status")
					if status == ARIA2_TASK_ACTIVE_STATUS {
						completedLenStr := utils.GetFieldStrFromJson(taskState, "completedLength")
						completedLen, _ := strconv.ParseInt(completedLenStr, 10, 64)
						totalLenStr := utils.GetFieldStrFromJson(taskState, "totalLength")
						totalLen, _ := strconv.ParseInt(totalLenStr, 10, 64)
						completePercent := completedLen / totalLen * 100

						speedStr := utils.GetFieldFromJson(taskState, "downloadSpeed")
						speed := speedStr.(int) / 1000

						logs.GetLogger().Info("continue downloading deal id %s complete %s%% speed %s KiB", dealId, completePercent, speed)
					}

					if isCompleted(taskState) {
						fileSize = utils.GetFieldStrFromJson(taskState, "completedLength")
						newStatus = DEAL_DOWNLOADED_STATUS
					}
				}else{
					newStatus = DEAL_DOWNLOAD_FAILED_STATUS
					errMsg := utils.GetFieldStrFromJson(taskState, "errorMessage")
					note =fmt.Sprintf("download failed, cause: %s",errMsg)
				}
			}else{
				newStatus = DEAL_DOWNLOAD_FAILED_STATUS
				note = "download gid not found in offline_deals.note"
			}

			if newStatus != currentStatus{
				msg := fmt.Sprintf("deal id %s status %s -> %s", dealId, currentStatus, newStatus)
				logs.GetLogger().Info(msg)
				updateOfflineDealDetails(newStatus,note,dealId, swanClient,"", fileSize)

			}
		}
	}
}

func startDownloading(maxDownloadingTaskNum int, minerFid, outDir string, aria2Client Aria2c, swanClient *utils.SwanClient) {
	for{
		downloadingDeals := findDealsByStatus(DEAL_DOWNLOADING_STATUS, minerFid, swanClient)
		countDownloadingDeals := len(downloadingDeals)
		if maxDownloadingTaskNum > countDownloadingDeals {
			newTaskNum := maxDownloadingTaskNum - countDownloadingDeals
			i := 1
			for i<=newTaskNum{
				deal2Download := findNextDealReady2Download(minerFid,swanClient)

				if deal2Download==nil || len(deal2Download.(string))==0{
					break
				}

				startDownloadForDeal(deal2Download.(string), outDir, aria2Client, swanClient)
				time.Sleep(1 * time.Second)
			}

			time.Sleep(60 * time.Second)
		}
	}
}



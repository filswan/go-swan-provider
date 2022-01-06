package service

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"swan-provider/common/constants"
	"swan-provider/config"
	"time"

	"github.com/filswan/go-swan-lib/client"
	"github.com/filswan/go-swan-lib/client/swan"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
)

type Aria2Service struct {
	MinerFid    string
	DownloadDir string
}

func GetAria2Service() *Aria2Service {
	aria2Service := &Aria2Service{
		MinerFid:    config.GetConfig().Main.MinerFid,
		DownloadDir: config.GetConfig().Aria2.Aria2DownloadDir,
	}

	_, err := os.Stat(aria2Service.DownloadDir)
	if err != nil {
		logs.GetLogger().Error(constants.ERROR_LAUNCH_FAILED)
		logs.GetLogger().Error("Your download directory:", aria2Service.DownloadDir, " not exists.")
		logs.GetLogger().Fatal(constants.INFO_ON_HOW_TO_CONFIG)
	}

	return aria2Service
}

func (aria2Service *Aria2Service) findNextDealReady2Download(swanClient *swan.SwanClient) *libmodel.OfflineDeal {
	deals := swanClient.SwanGetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_CREATED, "1")
	if len(deals) == 0 {
		deals = swanClient.SwanGetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_WAITING, "1")
	}

	if len(deals) > 0 {
		offlineDeal := deals[0]
		return &offlineDeal
	}

	return nil
}

func (aria2Service *Aria2Service) CheckDownloadStatus4Deal(aria2Client *client.Aria2Client, swanClient *swan.SwanClient, deal libmodel.OfflineDeal, gid string) {
	aria2Status := aria2Client.GetDownloadStatus(gid)
	if aria2Status == nil {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "get download status failed for gid:"+gid, "no response from aria2")
		return
	}

	if aria2Status.Error != nil {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "get download status failed for gid:"+gid, aria2Status.Error.Message)
		return
	}

	if len(aria2Status.Result.Files) != 1 {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "get download status failed for gid:"+gid, "wrong file amount")
		return
	}

	result := aria2Status.Result
	file := result.Files[0]
	filePath := file.Path
	fileSize := utils.GetInt64FromStr(file.Length)

	msg := fmt.Sprintf("current status:,%s,%s", result.Status, result.ErrorMessage)
	logs.GetLogger().Info(GetLog(deal, msg))
	switch result.Status {
	case ARIA2_TASK_STATUS_ERROR:
		UpdateDealInfoAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, &filePath, result.Status, "download gid:"+gid, result.ErrorCode, result.ErrorMessage)
	case ARIA2_TASK_STATUS_ACTIVE, ARIA2_TASK_STATUS_WAITING:
		fileSizeDownloaded := utils.GetFileSize(filePath)
		completedLen := utils.GetInt64FromStr(file.CompletedLength)
		var completePercent float64 = 0
		if fileSize > 0 {
			completePercent = float64(completedLen) / float64(fileSize) * 100
		}
		downloadSpeed := utils.GetInt64FromStr(result.DownloadSpeed) / 1024
		fileSizeDownloaded = fileSizeDownloaded / 1024
		note := fmt.Sprintf("downloading, complete: %.2f%%, speed: %dKiB, downloaded:%dKiB, %s, download gid:%s", completePercent, downloadSpeed, fileSizeDownloaded, result.Status, gid)
		logs.GetLogger().Info(GetLog(deal, note))
		UpdateDealInfoAndLog(deal, DEAL_STATUS_DOWNLOADING, &filePath, gid)
		if result.Status == ARIA2_TASK_STATUS_WAITING {
			msg := fmt.Sprintf("waiting to download,%s,%s", result.Status, result.ErrorMessage)
			logs.GetLogger().Info(GetLog(deal, msg))
		}
	case ARIA2_TASK_STATUS_COMPLETE:
		fileSizeDownloaded := utils.GetFileSize(filePath)
		logs.GetLogger().Info(GetLog(deal, "downloaded"))
		if fileSizeDownloaded >= 0 {
			UpdateDealInfoAndLog(deal, DEAL_STATUS_DOWNLOADED, &filePath, "download gid:"+gid)
		} else {
			UpdateDealInfoAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, &filePath, "file not found on its download path", "download gid:"+gid)
		}
	default:
		UpdateDealInfoAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, &filePath, result.Status, "download gid:"+gid, result.ErrorCode, result.ErrorMessage)
	}
}

func (aria2Service *Aria2Service) CheckDownloadStatus(aria2Client *client.Aria2Client, swanClient *swan.SwanClient) {
	downloadingDeals := swanClient.SwanGetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_DOWNLOADING)

	for _, deal := range downloadingDeals {
		gid := strings.Trim(deal.Note, " ")
		if gid == "" {
			UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "download gid not found in offline_deals.note")
			continue
		}

		aria2Service.CheckDownloadStatus4Deal(aria2Client, swanClient, deal, gid)
	}
}

func (aria2Service *Aria2Service) CheckAndRestoreSuspendingStatus(aria2Client *client.Aria2Client, swanClient *swan.SwanClient) {
	suspendingDeals := swanClient.SwanGetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_SUSPENDING)

	for _, deal := range suspendingDeals {
		onChainStatus, _ := lotusService.LotusMarket.LotusGetDealOnChainStatus(deal.DealCid)

		if onChainStatus == ONCHAIN_DEAL_STATUS_WAITTING {
			swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_WAITING)
		} else if onChainStatus == ONCHAIN_DEAL_STATUS_ERROR {
			swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED)
		}
	}
}

func (aria2Service *Aria2Service) StartDownload4Deal(deal libmodel.OfflineDeal, aria2Client *client.Aria2Client, swanClient *swan.SwanClient) {
	logs.GetLogger().Info(GetLog(deal, "start downloading"))
	urlInfo, err := url.Parse(deal.FileSourceUrl)
	if err != nil {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "parse source file url error,", err.Error())
		return
	}

	outFilename := urlInfo.Path
	if strings.HasPrefix(urlInfo.RawQuery, "filename=") {
		outFilename = strings.TrimPrefix(urlInfo.RawQuery, "filename=")
		outFilename = filepath.Join(urlInfo.Path, outFilename)
	}
	outFilename = strings.TrimLeft(outFilename, "/")

	today := time.Now()
	timeStr := fmt.Sprintf("%d%02d", today.Year(), today.Month())
	outDir := filepath.Join(aria2Service.DownloadDir, strconv.Itoa(deal.UserId), strconv.Itoa(deal.Id), timeStr)
	aria2Download := aria2Client.DownloadFile(deal.FileSourceUrl, outDir, outFilename)

	if aria2Download == nil {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "no response when asking aria2 to download")
		return
	}

	if aria2Download.Error != nil {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, aria2Download.Error.Message)
		return
	}

	if aria2Download.Gid == "" {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "no gid returned when asking aria2 to download")
		return
	}

	aria2Service.CheckDownloadStatus4Deal(aria2Client, swanClient, deal, aria2Download.Gid)
}

func (aria2Service *Aria2Service) StartDownload(aria2Client *client.Aria2Client, swanClient *swan.SwanClient) {
	downloadingDeals := swanClient.SwanGetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_DOWNLOADING)

	countDownloadingDeals := len(downloadingDeals)
	if countDownloadingDeals >= ARIA2_MAX_DOWNLOADING_TASKS {
		return
	}

	for i := 1; i <= ARIA2_MAX_DOWNLOADING_TASKS-countDownloadingDeals; i++ {
		deal2Download := aria2Service.findNextDealReady2Download(swanClient)
		if deal2Download == nil {
			logs.GetLogger().Info("No offline deal to download")
			break
		}

		onChainStatus, _ := lotusService.LotusMarket.LotusGetDealOnChainStatus(deal2Download.DealCid)

		if onChainStatus == ONCHAIN_DEAL_STATUS_WAITTING {
			aria2Service.StartDownload4Deal(*deal2Download, aria2Client, swanClient)
		} else if onChainStatus == ONCHAIN_DEAL_STATUS_ERROR {
			UpdateStatusAndLog(*deal2Download, DEAL_STATUS_IMPORT_FAILED, "Lotus deal has error status")
		} else {
			UpdateStatusAndLog(*deal2Download, DEAL_STATUS_SUSPENDING)
		}

		time.Sleep(1 * time.Second)
	}
}

func (aria2Service *Aria2Service) PurgeDownloadFile(aria2Client *client.Aria2Client, swanClient *swan.SwanClient) {
	completedDeals := swanClient.SwanGetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_COMPLETED)
	for _, deal := range completedDeals {
		DeleteFile(&deal)
	}
	expiredDeals := swanClient.SwanGetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_EXPIRED)
	for _, deal := range expiredDeals {
		DeleteFile(&deal)
	}
	importFailedDeals := swanClient.SwanGetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_IMPORT_FAILED)
	for _, deal := range importFailedDeals {
		onChainStatus, _ := lotusService.LotusMarket.LotusGetDealOnChainStatus(deal.DealCid)
		GetLog(deal, "lotus deal status is "+onChainStatus)
		if onChainStatus == ONCHAIN_DEAL_STATUS_ERROR {
			DeleteFile(&deal)
		}
	}
}

func DeleteFile(deal *libmodel.OfflineDeal) {
	filePath := deal.FilePath
	if filePath == "" {
		GetLog(*deal, "file path is blank.")
		return
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		logs.GetLogger().Info("car file does not exist.")
	} else {
		if !fileInfo.IsDir() {
			err := os.Remove(filePath)
			if err != nil {
				logs.GetLogger().Error(err)
			} else {
				GetLog(*deal, "car file has successfully been deleted.")
			}
		} else {
			GetLog(*deal, "filepath is a directory and cannot be removed.")
		}
	}

}

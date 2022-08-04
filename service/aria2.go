package service

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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

func (aria2Service *Aria2Service) FindNextDealReady2Download(swanClient *swan.SwanClient) *libmodel.OfflineDeal {
	pageNum := 0
	pageSize := 1
	statuses := []string{}
	statuses = append(statuses, DEAL_STATUS_CREATED)
	statuses = append(statuses, DEAL_STATUS_WAITING)

	for _, status := range statuses {
		params := swan.GetOfflineDealsByStatusParams{
			DealStatus: status,
			ForMiner:   true,
			MinerFid:   &aria2Service.MinerFid,
			PageNum:    &pageNum,
			PageSize:   &pageSize,
		}
		deals, err := swanClient.GetOfflineDealsByStatus(params)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil
		}

		if len(deals) > 0 {
			offlineDeal := deals[0]
			return offlineDeal
		}
	}

	return nil
}

func (aria2Service *Aria2Service) CheckDownloadStatus4Deal(aria2Client *client.Aria2Client, swanClient *swan.SwanClient, deal *libmodel.OfflineDeal, gid string) {
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
	downloadingDeals := GetOfflineDeals(swanClient, DEAL_STATUS_DOWNLOADING, aria2Service.MinerFid, nil)

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
	suspendingDeals := GetOfflineDeals(swanClient, DEAL_STATUS_SUSPENDING, aria2Service.MinerFid, nil)

	for _, deal := range suspendingDeals {
		onChainStatus, onChainMessage, err := lotusService.LotusMarket.LotusGetDealOnChainStatus(deal.DealCid)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		if onChainStatus == nil {
			logs.GetLogger().Info("no on chain status for deal%", *deal.TaskName+":"+deal.DealCid)
			continue
		}

		if *onChainStatus == ONCHAIN_DEAL_STATUS_WAITTING {
			UpdateStatusAndLog(deal, DEAL_STATUS_WAITING, "deal waiting for downloading after suspending", *onChainStatus, *onChainMessage)
		} else if *onChainStatus == ONCHAIN_DEAL_STATUS_ERROR {
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal error after suspending", *onChainMessage)
		}
	}
}

func (aria2Service *Aria2Service) StartDownload4Deal(deal *libmodel.OfflineDeal, aria2Client *client.Aria2Client, swanClient *swan.SwanClient) {
	logs.GetLogger().Info(GetLog(deal, "start downloading"))
	urlInfo, err := url.Parse(deal.CarFileUrl)
	if err != nil {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "parse source file url error,", err.Error())
		return
	}

	outFilename := urlInfo.Path
	if strings.HasPrefix(urlInfo.RawQuery, "filename=") {
		outFilename = strings.TrimPrefix(urlInfo.RawQuery, "filename=")
		outFilename = filepath.Join(urlInfo.Path, outFilename)
	}
	_, outFilename = filepath.Split(outFilename)
	outDir := strings.TrimSuffix(aria2Service.DownloadDir, "/")
	filePath := outDir + "/" + outFilename
	if IsExist(filePath) {
		UpdateDealInfoAndLog(deal, DEAL_STATUS_DOWNLOADED, &filePath, outFilename+", the car file already exists, skip downloading it")
		return
	}

	aria2Download := aria2Client.DownloadFile(deal.CarFileUrl, outDir, outFilename)
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
	downloadingDeals := GetOfflineDeals(swanClient, DEAL_STATUS_DOWNLOADING, aria2Service.MinerFid, nil)

	countDownloadingDeals := len(downloadingDeals)
	if countDownloadingDeals >= ARIA2_MAX_DOWNLOADING_TASKS {
		return
	}

	for i := 1; i <= ARIA2_MAX_DOWNLOADING_TASKS-countDownloadingDeals; i++ {
		deal2Download := aria2Service.FindNextDealReady2Download(swanClient)
		if deal2Download == nil {
			logs.GetLogger().Info("No offline deal to download")
			break
		}

		//logs.GetLogger().Info("deal:", deal2Download.Id, " ", deal2Download.DealCid, deal2Download)
		onChainStatus, onChainMessage, err := lotusService.LotusMarket.LotusGetDealOnChainStatus(deal2Download.DealCid)
		if err != nil {
			logs.GetLogger().Error(err)
			break
		}

		if onChainStatus == nil {
			logs.GetLogger().Info("not found the deal on the chain", *deal2Download.TaskName+":"+deal2Download.DealCid)
			UpdateStatusAndLog(deal2Download, DEAL_STATUS_IMPORT_FAILED, "not found the deal on the chain")
			continue
		} else if *onChainStatus == ONCHAIN_DEAL_STATUS_WAITTING {
			aria2Service.StartDownload4Deal(deal2Download, aria2Client, swanClient)
		} else if *onChainStatus == ONCHAIN_DEAL_STATUS_ERROR {
			UpdateStatusAndLog(deal2Download, DEAL_STATUS_IMPORT_FAILED, "deal error before downloading", *onChainStatus, *onChainMessage)
		} else {
			UpdateStatusAndLog(deal2Download, DEAL_STATUS_SUSPENDING, "deal not ready for downloading", *onChainStatus, *onChainMessage)
		}

		time.Sleep(1 * time.Second)
	}
}

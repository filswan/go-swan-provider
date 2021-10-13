package offlineDealAdmin

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"swan-provider/common/client"
	"swan-provider/common/utils"
	"swan-provider/config"
	"swan-provider/logs"
	"swan-provider/models"
	"time"
)

type Aria2Service struct {
	MinerFid string
	OutDir   string
}

func GetAria2Service() *Aria2Service {
	aria2Service := &Aria2Service{
		MinerFid: config.GetConfig().Main.MinerFid,
		OutDir:   config.GetConfig().Aria2.Aria2DownloadDir,
	}

	_, err := os.Stat(aria2Service.OutDir)
	if err != nil {
		logs.GetLogger().Error("Swan provider launch failed.")
		logs.GetLogger().Error("Your download directory:", aria2Service.OutDir, " not exists.")
		logs.GetLogger().Fatal("For more information about how to config, please check https://docs.filswan.com/run-swan-provider/config-swan-provider")
	}

	return aria2Service
}

func (aria2Service *Aria2Service) findNextDealReady2Download(swanClient *client.SwanClient) *models.OfflineDeal {
	deals := swanClient.GetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_CREATED, "1")
	if len(deals) == 0 {
		deals = swanClient.GetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_WAITING, "1")
	}

	if len(deals) > 0 {
		offlineDeal := deals[0]
		return &offlineDeal
	}

	return nil
}

func (aria2Service *Aria2Service) CheckDownloadStatus4Deal(aria2Client *client.Aria2Client, swanClient *client.SwanClient, deal *models.OfflineDeal, gid string) {
	aria2Status := aria2Client.GetDownloadStatus(gid)
	if aria2Status == nil {
		note := fmt.Sprintf("Get status for %s failed, no response", gid)
		updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note)
		if !updated {
			logs.GetLogger().Error("Failed to update offline deal status")
		}
		logs.GetLogger().Error(note)
		return
	}

	if aria2Status.Error != nil {
		note := fmt.Sprintf("Get status for %s failed, code:%d, message:%s", gid, aria2Status.Error.Code, aria2Status.Error.Message)
		updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note)
		if !updated {
			logs.GetLogger().Error("Failed to update offline deal status")
		}
		logs.GetLogger().Error(note)
		return
	}

	if len(aria2Status.Result.Files) != 1 {
		note := "Wrong file amount"
		updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note)
		if !updated {
			logs.GetLogger().Error("Failed to update offline deal status")
		}
		logs.GetLogger().Error(note)
		return
	}

	result := aria2Status.Result
	code := result.ErrorCode
	message := result.ErrorMessage
	status := result.Status
	file := result.Files[0]
	filePath := file.Path
	fileSize := utils.GetInt64FromStr(file.Length)
	completedLen := utils.GetInt64FromStr(file.CompletedLength)
	var completePercent float64 = 0
	if fileSize > 0 {
		completePercent = float64(completedLen) / float64(fileSize) * 100
	}
	downloadSpeed := utils.GetInt64FromStr(result.DownloadSpeed) / 1000

	switch status {
	case ARIA2_TASK_STATUS_ERROR:
		note := fmt.Sprintf("Deal:%s status for %s, code:%s, message:%s, status:%s", deal.DealCid, gid, code, message, status)
		if !utils.IsFileExistsFullPath(aria2Service.OutDir) {
			note = fmt.Sprintf("%s.aria2 download directory: %s not exists", note, aria2Service.OutDir)
		}
		updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note)
		if !updated {
			logs.GetLogger().Error("Failed to update offline deal status")
		}
		logs.GetLogger().Error(note)
	case ARIA2_TASK_STATUS_ACTIVE:
		fileSizeDownloaded := utils.GetFileSize(filePath)
		if deal.Status != DEAL_STATUS_DOWNLOADING {
			updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOADING, gid, filePath, utils.GetStrFromInt64(fileSizeDownloaded))
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
		}
		msg := fmt.Sprintf("Deal downloading, CID: %s, complete: %.2f%%, speed: %dKiB", deal.DealCid, completePercent, downloadSpeed)
		logs.GetLogger().Info(msg)
	case ARIA2_TASK_STATUS_COMPLETE:
		fileSizeDownloaded := utils.GetFileSize(filePath)
		if fileSizeDownloaded >= 0 {
			note := fmt.Sprintf("Deal CID: %s downloaded", deal.DealCid)
			logs.GetLogger().Info(note)
			updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOADED, gid, filePath, utils.GetStrFromInt64(fileSizeDownloaded))
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
		} else {
			note := fmt.Sprintf("File %s not found on ", filePath)
			logs.GetLogger().Error(note)
			updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note, filePath, utils.GetStrFromInt64(fileSize))
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
		}
	default:
		note := fmt.Sprintf("Download failed, cause: %s", result.ErrorMessage)
		updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note, filePath, utils.GetStrFromInt64(fileSize))
		if !updated {
			logs.GetLogger().Error("Failed to update offline deal status")
		}
		logs.GetLogger().Error(note, " dealId:", strconv.Itoa(deal.Id))
	}
}

func (aria2Service *Aria2Service) CheckDownloadStatus(aria2Client *client.Aria2Client, swanClient *client.SwanClient) {
	downloadingDeals := swanClient.GetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_DOWNLOADING)

	for _, deal := range downloadingDeals {
		gid := deal.Note
		if len(gid) <= 0 {
			note := "Download gid not found in offline_deals.note"
			if note != deal.Note {
				updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note)
				if !updated {
					logs.GetLogger().Error("Failed to update offline deal status")
				}
			}
			continue
		}

		aria2Service.CheckDownloadStatus4Deal(aria2Client, swanClient, &deal, gid)
	}
}

func (aria2Service *Aria2Service) StartDownload4Deal(deal *models.OfflineDeal, aria2Client *client.Aria2Client, swanClient *client.SwanClient) {
	logs.GetLogger().Info("start downloading deal id ", deal.Id)
	urlInfo, err := url.Parse(deal.SourceFileUrl)
	if err != nil {
		msg := fmt.Sprintf("parse source file url error:%s", err)
		updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, msg)
		if !updated {
			logs.GetLogger().Error("Failed to update offline deal status")
		}
		msg = fmt.Sprintf("Deal:%s, %s", deal.DealCid, msg)
		logs.GetLogger().Error(msg)
		return
	}

	outFilename := urlInfo.Path
	if strings.HasPrefix(urlInfo.RawQuery, "filename=") {
		outFilename = strings.TrimLeft(urlInfo.RawQuery, "filename=")
		outFilename = utils.GetDir(urlInfo.Path, outFilename)
	}
	outFilename = strings.TrimLeft(outFilename, "/")

	today := time.Now()
	timeStr := fmt.Sprintf("%d%02d", today.Year(), today.Month())
	outDir := utils.GetDir(aria2Service.OutDir, strconv.Itoa(deal.UserId), timeStr)

	aria2Download := aria2Client.DownloadFile(deal.SourceFileUrl, outDir, outFilename)

	if aria2Download == nil {
		note := "No response when asking aria2 to download"
		logs.GetLogger().Error(note)
		updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note)
		if !updated {
			logs.GetLogger().Error("Failed to update offline deal status")
		}
		return
	}

	if aria2Download.Error != nil {
		note := fmt.Sprintf("Error: code(%d), %s", aria2Download.Error.Code, aria2Download.Error.Message)
		logs.GetLogger().Error(note)
		updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note)
		if !updated {
			logs.GetLogger().Error("Failed to update offline deal status")
		}
		return
	}

	if aria2Download.Gid == "" {
		note := "Error: no gid returned"
		logs.GetLogger().Error(note)
		updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_DOWNLOAD_FAILED, note)
		if !updated {
			logs.GetLogger().Error("Failed to update offline deal status")
		}
		return
	}

	aria2Service.CheckDownloadStatus4Deal(aria2Client, swanClient, deal, aria2Download.Gid)
}

func (aria2Service *Aria2Service) StartDownload(aria2Client *client.Aria2Client, swanClient *client.SwanClient) {
	downloadingDeals := swanClient.GetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_DOWNLOADING)
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

		aria2Service.StartDownload4Deal(deal2Download, aria2Client, swanClient)
		time.Sleep(1 * time.Second)
	}
}

package offlineDealAdmin

import (
	"swan-miner/common/utils"
	"swan-miner/logs"
	"time"
)

const ARIA2_TASK_STATUS_ERROR = "error"
const ARIA2_TASK_STATUS_ACTIVE = "active"
const ARIA2_TASK_STATUS_COMPLETE = "complete"

const DEAL_STATUS_CREATED = "Created"
const DEAL_STATUS_WAITING = "Waiting"
const DEAL_STATUS_DOWNLOADING = "Downloading"
const DEAL_STATUS_DOWNLOADED = "Downloaded"
const DEAL_STATUS_DOWNLOAD_FAILED = "DownloadFailed"

const DEAL_STATUS_IMPORT_READY = "ReadyForImport"
const DEAL_STATUS_IMPORT_FAILED = "ImportFailed"
const DEAL_STATUS_IMPORTING = "FileImporting"
const DEAL_STATUS_IMPORTED = "FileImported"
const DEAL_STATUS_ACTIVE = "DealActive"

const ONCHAIN_DEAL_STATUS_ERROR = "StorageDealError"
const ONCHAIN_DEAL_STATUS_ACTIVE = "StorageDealActive"
const ONCHAIN_DEAL_STATUS_NOTFOUND = "StorageDealNotFound"
const ONCHAIN_DEAL_STATUS_WAITTING = "StorageDealWaitingForData"
const ONCHAIN_DEAL_STATUS_ACCEPT = "StorageDealAcceptWait"
const ONCHAIN_DEAL_STATUS_AWAITING = "StorageDealAwaitingPreCommit"

const ARIA2_MAX_DOWNLOADING_TASKS = 10
const LOTUS_IMPORT_NUMNBER = "20" //Max number of deals to be imported at a time
const LOTUS_SCAN_NUMBER = "100"   //Max number of deals to be scanned at a time

func AdminOfflineDeal()  {
	aria2Client := utils.GetAria2Client()
	swanClient := utils.GetSwanClient()

	aria2Service := GetAria2Service()
	lotusService := GetLotusService()

	go func() {
		for {
			logs.GetLogger().Info("CheckDownloadStatus begin...")
			aria2Service.CheckDownloadStatus(aria2Client, swanClient)
			logs.GetLogger().Info("CheckDownloadStatus end... Sleeping...")
			time.Sleep(time.Minute)
		}
	}()

	go func() {
		for {
			logs.GetLogger().Info("StartDownload begin...")
			aria2Service.StartDownload(aria2Client, swanClient)
			logs.GetLogger().Info("StartDownload end... Sleeping...")
			time.Sleep(time.Minute)
		}
	}()

	go func() {
		for {
			logs.GetLogger().Info("StartImport begin...")
			lotusService.StartImport(swanClient)
			logs.GetLogger().Info("StartImport end... Sleeping...")
			time.Sleep(lotusService.ImportIntervalSecond)
		}
	}()

	go func() {
		for {
			logs.GetLogger().Info("StartScan begin...")
			lotusService.StartScan(swanClient)
			logs.GetLogger().Info("StartScan end... Sleeping...")
			time.Sleep(lotusService.ImportIntervalSecond)
		}
	}()
}

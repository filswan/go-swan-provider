package offlineDealAdmin

import (
	"swan-provider/common/utils"
	"swan-provider/logs"
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

var aria2Client = utils.GetAria2Client()
var swanClient = utils.GetSwanClient()

var aria2Service = GetAria2Service()
var lotusService = GetLotusService()

func AdminOfflineDeal()  {
	go swanSendHeartbeatRequest()
	go aria2CheckDownloadStatus()
	go aria2StartDownload()
	go lotusStartImport()
	go lotusStartScan()
}

func swanSendHeartbeatRequest() {
	for {
		logs.GetLogger().Info("Start...")
		SendHeartbeatRequest(swanClient)
		logs.GetLogger().Info("Sleeping...")
		time.Sleep(time.Minute)
	}
}

func aria2CheckDownloadStatus() {
	for {
		logs.GetLogger().Info("Start...")
		aria2Service.CheckDownloadStatus(aria2Client, swanClient)
		logs.GetLogger().Info("Sleeping...")
		time.Sleep(time.Minute)
	}
}

func aria2StartDownload() {
	for {
		logs.GetLogger().Info("Start...")
		aria2Service.StartDownload(aria2Client, swanClient)
		logs.GetLogger().Info("Sleeping...")
		time.Sleep(time.Minute)
	}
}

func lotusStartImport() {
	for {
		logs.GetLogger().Info("Start...")
		lotusService.StartImport(swanClient)
		logs.GetLogger().Info("Sleeping...")
		time.Sleep(lotusService.ImportIntervalSecond)
	}
}

func lotusStartScan() {
	for {
		logs.GetLogger().Info("Start...")
		lotusService.StartScan(swanClient)
		logs.GetLogger().Info("Sleeping...")
		time.Sleep(lotusService.ImportIntervalSecond)
	}
}

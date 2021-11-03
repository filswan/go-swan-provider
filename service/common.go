package service

import (
	"go-swan-provider/common/constants"
	"go-swan-provider/config"
	"strings"
	"time"

	"github.com/filswan/go-swan-lib/client"
	"github.com/filswan/go-swan-lib/logs"

	"github.com/filswan/go-swan-lib/client/swan"
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
const DEAL_STATUS_IMPORTING = "FileImporting"
const DEAL_STATUS_IMPORTED = "FileImported"
const DEAL_STATUS_IMPORT_FAILED = "ImportFailed"
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

var aria2Client *client.Aria2Client
var swanClient *swan.SwanClient

var swanService = GetSwanService()
var aria2Service = GetAria2Service()
var lotusService = GetLotusService()

func AdminOfflineDeal() {
	var err error

	aria2Host := config.GetConfig().Aria2.Aria2Host
	aria2Port := config.GetConfig().Aria2.Aria2Port
	aria2Secret := config.GetConfig().Aria2.Aria2Secret
	aria2Client = client.GetAria2Client(aria2Host, aria2Secret, aria2Port)

	apiUrl := config.GetConfig().Main.SwanApiUrl
	apiKey := config.GetConfig().Main.SwanApiKey
	accessToken := config.GetConfig().Main.SwanAccessToken
	swanClient, err = swan.SwanGetClient(apiUrl, apiKey, accessToken, "")
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Error("Swan provider launch failed.")
		logs.GetLogger().Info("For more information about how to config, please check https://docs.filswan.com/run-swan-provider/config-swan-provider")
		return
	}
	checkMinerExists()
	checkLotusConfig()
	swanService.UpdateBidConf(swanClient)
	go swanSendHeartbeatRequest()
	go aria2CheckDownloadStatus()
	go aria2StartDownload()
	go lotusStartImport()
	go lotusStartScan()
}

func checkMinerExists() {
	err := swanService.SendHeartbeatRequest(swanClient)
	if err != nil {
		logs.GetLogger().Info(err)
		if strings.Contains(err.Error(), "Miner Not found") {
			logs.GetLogger().Error(constants.ERROR_LAUNCH_FAILED)
			logs.GetLogger().Error("Cannot find your miner:", swanService.MinerFid)
			logs.GetLogger().Fatal(constants.INFO_ON_HOW_TO_CONFIG)
		}
	}
}

func checkLotusConfig() {
	logs.GetLogger().Info("Start testing lotus config.")

	if lotusService == nil {
		logs.GetLogger().Fatal("error in config")
	}

	lotusMarket := lotusService.LotusMarket
	lotusClient := lotusService.LotusClient
	if len(lotusMarket.ApiUrl) == 0 {
		logs.GetLogger().Fatal("please set config:lotus->market_api_url")
	}

	if len(lotusMarket.AccessToken) == 0 {
		logs.GetLogger().Fatal("please set config:lotus->market_access_token")
	}

	if len(lotusMarket.ClientApiUrl) == 0 {
		logs.GetLogger().Fatal("please set config:lotus->client_api_url")
	}

	err := lotusMarket.LotusImportData("bafyreib7azyg2yubucdhzn64gvyekdma7nbrbnfafcqvhsz2mcnvbnkktu", "test")

	if err != nil {
		logs.GetLogger().Fatal(err)
	}

	currentEpoch := lotusClient.LotusGetCurrentEpoch()
	if currentEpoch < 0 {
		logs.GetLogger().Fatal("please check config:lotus->api_url")
	}

	logs.GetLogger().Info("Pass testing lotus config.")
}

func swanSendHeartbeatRequest() {
	for {
		logs.GetLogger().Info("Start...")
		swanService.SendHeartbeatRequest(swanClient)
		logs.GetLogger().Info("Sleeping...")
		time.Sleep(swanService.ApiHeartbeatInterval)
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
		time.Sleep(lotusService.ScanIntervalSecond)
	}
}

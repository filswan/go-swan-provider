package service

import (
	"fmt"
	"strings"
	"swan-provider/common/constants"
	"swan-provider/config"
	"time"

	"github.com/filswan/go-swan-lib/client"
	libconstants "github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
)

const ARIA2_TASK_STATUS_ERROR = "error"
const ARIA2_TASK_STATUS_WAITING = "waiting"
const ARIA2_TASK_STATUS_ACTIVE = "active"
const ARIA2_TASK_STATUS_COMPLETE = "complete"

const DEAL_STATUS_CREATED = "Created"
const DEAL_STATUS_WAITING = "Waiting"
const DEAL_STATUS_SUSPENDING = "Suspending"

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
const ONCHAIN_DEAL_STATUS_SEALING = "StorageDealSealing"
const ONCHAIN_DEAL_STATUS_AWAITING = "StorageDealAwaitingPreCommit"

const ARIA2_MAX_DOWNLOADING_TASKS = 10
const LOTUS_IMPORT_NUMNBER = 20 //Max number of deals to be imported at a time
const LOTUS_SCAN_NUMBER = 100   //Max number of deals to be scanned at a time

var aria2Client *client.Aria2Client
var swanClient *swan.SwanClient

var swanService *SwanService
var aria2Service *Aria2Service
var lotusService *LotusService

func AdminOfflineDeal() {
	swanService = GetSwanService()
	aria2Service = GetAria2Service()
	lotusService = GetLotusService()

	aria2Client = SetAndCheckAria2Config()
	swanClient = SetAndCheckSwanConfig()
	checkMinerExists()
	checkLotusConfig()

	//logs.GetLogger().Info("swan token:", swanClient.SwanToken)
	swanService.UpdateBidConf(swanClient)
	go swanSendHeartbeatRequest()
	go aria2CheckDownloadStatus()
	go aria2StartDownload()
	go lotusStartImport()
	go lotusStartScan()
}

func SetAndCheckAria2Config() *client.Aria2Client {
	aria2DownloadDir := config.GetConfig().Aria2.Aria2DownloadDir
	aria2Host := config.GetConfig().Aria2.Aria2Host
	aria2Port := config.GetConfig().Aria2.Aria2Port
	aria2Secret := config.GetConfig().Aria2.Aria2Secret
	aria2MaxDownloadingTasks := config.GetConfig().Aria2.Aria2MaxDownloadingTasks

	if !utils.IsDirExists(aria2DownloadDir) {
		err := fmt.Errorf("aria2 down load dir:%s not exits, please set config:aria2->aria2_download_dir", aria2DownloadDir)
		logs.GetLogger().Fatal(err)
	}

	if len(aria2Host) == 0 {
		logs.GetLogger().Fatal("please set config:aria2->aria2_host")
	}

	aria2Client = client.GetAria2Client(aria2Host, aria2Secret, aria2Port)
	if aria2MaxDownloadingTasks <= 0 {
		logs.GetLogger().Warning("config [aria2].aria2_max_downloading_tasks is " + strconv.Itoa(aria2MaxDownloadingTasks) + ", no CAR file will be downloaded")
	}
	aria2ChangeMaxConcurrentDownloads := aria2Client.ChangeMaxConcurrentDownloads(strconv.Itoa(aria2MaxDownloadingTasks))
	if aria2ChangeMaxConcurrentDownloads == nil {
		err := fmt.Errorf("failed to set [aria2].aria2_max_downloading_tasks, please check the Aria2 service")
		logs.GetLogger().Fatal(err)
	}

	if aria2ChangeMaxConcurrentDownloads.Error != nil {
		err := fmt.Errorf(aria2ChangeMaxConcurrentDownloads.Error.Message)
		logs.GetLogger().Fatal(err)
	}
	return aria2Client
}

func SetAndCheckSwanConfig() *swan.SwanClient {
	var err error
	swanApiUrl := config.GetConfig().Main.SwanApiUrl
	swanApiKey := config.GetConfig().Main.SwanApiKey
	swanAccessToken := config.GetConfig().Main.SwanAccessToken

	if len(swanApiUrl) == 0 {
		logs.GetLogger().Fatal("please set config:main->api_url")
	}

	if len(swanApiKey) == 0 {
		logs.GetLogger().Fatal("please set config:main->api_key")
	}

	if len(swanAccessToken) == 0 {
		logs.GetLogger().Fatal("please set config:main->access_token")
	}

	swanClient, err := swan.GetClient(swanApiUrl, swanApiKey, swanAccessToken, "")
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Error(constants.ERROR_LAUNCH_FAILED)
		logs.GetLogger().Fatal(constants.INFO_ON_HOW_TO_CONFIG)
	}

	return swanClient
}

func checkMinerExists() {
	err := swanService.SendHeartbeatRequest(swanClient)
	if err != nil {
		logs.GetLogger().Info(err)
		if strings.Contains(err.Error(), "Miner Not found") {
			logs.GetLogger().Error("Cannot find your miner:", swanService.MinerFid)
		}
		logs.GetLogger().Error(constants.ERROR_LAUNCH_FAILED)
		logs.GetLogger().Fatal(constants.INFO_ON_HOW_TO_CONFIG)
	}
}

func checkLotusConfig() {
	logs.GetLogger().Info("Start testing lotus config.")

	if lotusService == nil {
		logs.GetLogger().Fatal("error in config")
	}

	lotusMarket := lotusService.LotusMarket
	lotusClient := lotusService.LotusClient
	if utils.IsStrEmpty(&lotusMarket.ApiUrl) {
		logs.GetLogger().Fatal("please set config:lotus->market_api_url")
	}

	if utils.IsStrEmpty(&lotusMarket.AccessToken) {
		logs.GetLogger().Fatal("please set config:lotus->market_access_token")
	}

	if utils.IsStrEmpty(&lotusMarket.ClientApiUrl) {
		logs.GetLogger().Fatal("please set config:lotus->client_api_url")
	}

	isWriteAuth, err := lotus.LotusCheckAuth(lotusMarket.ApiUrl, lotusMarket.AccessToken, libconstants.LOTUS_AUTH_WRITE)
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Fatal("please check config:lotus->market_api_url, lotus->market_access_token")
	}

	if !isWriteAuth {
		logs.GetLogger().Fatal("market access token should have write access right")
	}

	currentEpoch, err := lotusClient.LotusGetCurrentEpoch()
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Fatal("please check config:lotus->client_api_url")
	}

	logs.GetLogger().Info("current epoch:", *currentEpoch)
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
		aria2Service.CheckAndRestoreSuspendingStatus(aria2Client, swanClient)
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

func UpdateDealInfoAndLog(deal *libmodel.OfflineDeal, newSwanStatus string, filefullpath *string, messages ...string) {
	note := ""
	if newSwanStatus != DEAL_STATUS_DOWNLOADING {
		note = GetNote(messages...)
		note = utils.FirstLetter2Upper(note)
	} else {
		note = messages[0]
	}

	if newSwanStatus == DEAL_STATUS_IMPORT_FAILED || newSwanStatus == DEAL_STATUS_DOWNLOAD_FAILED {
		logs.GetLogger().Warn(GetLog(deal, note))
	} else {
		logs.GetLogger().Info(GetLog(deal, note))
	}

	filefullpathTemp := deal.FilePath
	if filefullpath != nil {
		filefullpathTemp = *filefullpath
	}

	if deal.Status == newSwanStatus && deal.Note == note && deal.FilePath == filefullpathTemp {
		return
	}

	err := UpdateOfflineDeal(swanClient, deal.Id, newSwanStatus, &note, &filefullpathTemp)
	if err != nil {
		logs.GetLogger().Error(GetLog(deal, constants.UPDATE_OFFLINE_DEAL_STATUS_FAIL))
	} else {
		msg := GetLog(deal, "set status to:"+newSwanStatus, "set note to:"+note, "set filepath to:"+filefullpathTemp)
		if newSwanStatus == DEAL_STATUS_IMPORT_FAILED || newSwanStatus == DEAL_STATUS_DOWNLOAD_FAILED {
			logs.GetLogger().Warn(msg)
		} else {
			logs.GetLogger().Info(msg)
		}
	}
}

func UpdateStatusAndLog(deal *libmodel.OfflineDeal, newSwanStatus string, messages ...string) {
	UpdateDealInfoAndLog(deal, newSwanStatus, nil, messages...)
}

func GetLog(deal *libmodel.OfflineDeal, messages ...string) string {
	text := GetNote(messages...)
	msg := fmt.Sprintf("taskName:%s, dealCid:%s, %s", *deal.TaskName, deal.DealCid, text)
	return msg
}

func GetNote(messages ...string) string {
	separator := ","
	result := ""
	if messages == nil {
		return result
	}
	for _, message := range messages {
		if message != "" {
			result = result + separator + message
		}
	}

	result = strings.TrimPrefix(result, separator)
	result = strings.TrimSuffix(result, separator)
	return result
}

func GetOfflineDeals(swanClient *swan.SwanClient, dealStatus string, minerFid string, limit *int) []*libmodel.OfflineDeal {
	pageNum := 1
	params := swan.GetOfflineDealsByStatusParams{
		DealStatus: dealStatus,
		ForMiner:   true,
		MinerFid:   &minerFid,
		PageNum:    &pageNum,
		PageSize:   limit,
	}

	offlineDeals, err := swanClient.GetOfflineDealsByStatus(params)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return offlineDeals
}

func UpdateOfflineDeal(swanClient *swan.SwanClient, dealId int, status string, note, filePath *string) error {
	params := &swan.UpdateOfflineDealParams{
		DealId:   dealId,
		Status:   status,
		Note:     note,
		FilePath: filePath,
	}

	err := swanClient.UpdateOfflineDeal(*params)
	if err != nil {
		logs.GetLogger().Error()
		return err
	}

	return nil
}

func UpdateOfflineDealStatus(swanClient *swan.SwanClient, dealId int, status string) error {
	params := &swan.UpdateOfflineDealParams{
		DealId: dealId,
		Status: status,
	}

	err := swanClient.UpdateOfflineDeal(*params)
	if err != nil {
		logs.GetLogger().Error()
		return err
	}

	return nil
}

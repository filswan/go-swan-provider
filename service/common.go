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
	"github.com/filswan/go-swan-lib/model"
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
const ONCHAIN_DEAL_STATUS_AWAITING = "StorageDealAwaitingPreCommit"

const ARIA2_MAX_DOWNLOADING_TASKS = 10
const LOTUS_IMPORT_NUMNBER = 20 //Max number of deals to be imported at a time
const LOTUS_SCAN_NUMBER = 100   //Max number of deals to be scanned at a time

var aria2Client *client.Aria2Client
var swanClient *swan.SwanClient

var swanService = GetSwanService()
var aria2Service = GetAria2Service()
var lotusService = GetLotusService()

func AdminOfflineDeal() {
	setAndCheckAria2Config()
	setAndCheckSwanConfig()
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

func setAndCheckAria2Config() {
	aria2DownloadDir := config.GetConfig().Aria2.Aria2DownloadDir
	aria2Host := config.GetConfig().Aria2.Aria2Host
	aria2Port := config.GetConfig().Aria2.Aria2Port
	aria2Secret := config.GetConfig().Aria2.Aria2Secret

	if !utils.IsDirExists(aria2DownloadDir) {
		err := fmt.Errorf("aria2 down load dir:%s not exits, please set config:aria2->aria2_download_dir", aria2DownloadDir)
		logs.GetLogger().Fatal(err)
	}

	if len(aria2Host) == 0 {
		logs.GetLogger().Fatal("please set config:aria2->aria2_host")
	}

	aria2Client = client.GetAria2Client(aria2Host, aria2Secret, aria2Port)
}

func setAndCheckSwanConfig() {
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

	swanClient, err = swan.GetClient("", swanApiUrl, swanApiKey, swanAccessToken, "")
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Error(constants.ERROR_LAUNCH_FAILED)
		logs.GetLogger().Fatal(constants.INFO_ON_HOW_TO_CONFIG)
	}
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
	if len(lotusMarket.ApiUrl) == 0 {
		logs.GetLogger().Fatal("please set config:lotus->market_api_url")
	}

	if len(lotusMarket.AccessToken) == 0 {
		logs.GetLogger().Fatal("please set config:lotus->market_access_token")
	}

	if len(lotusMarket.ClientApiUrl) == 0 {
		logs.GetLogger().Fatal("please set config:lotus->client_api_url")
	}

	isWriteAuth, err := lotus.LotusCheckAuth(lotusMarket.ApiUrl, lotusMarket.AccessToken, libconstants.LOTUS_AUTH_WRITE)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

	if !isWriteAuth {
		logs.GetLogger().Fatal("market access token should have write access right")
	}

	currentEpoch := lotusClient.LotusGetCurrentEpoch()
	if currentEpoch < 0 {
		logs.GetLogger().Fatal("please check config:lotus->client_api_url")
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

func getDealCost(dealCost lotus.ClientDealCostStatus) string {
	if dealCost.DealProposalAccepted != "" {
		return dealCost.DealProposalAccepted
	}

	if dealCost.ReserveClientFunds != "" {
		return dealCost.ReserveClientFunds
	}

	return dealCost.CostComputed
}

func UpdateDealInfoAndLog(deal *model.OfflineDeal, newSwanStatus string, filefullpath *string, messages ...string) {
	noteFunds := ""
	cost := deal.Cost
	if deal.DealCid != "" {
		dealCost, err := lotusService.LotusClient.LotusClientGetDealInfo(deal.DealCid)
		if err == nil {
			cost = getDealCost(*dealCost)
			noteFunds = GetNote("funds computed:"+dealCost.CostComputed, "funds reserved:"+dealCost.ReserveClientFunds, "funds released:"+dealCost.DealProposalAccepted)
		}
	}
	note := ""
	if newSwanStatus != DEAL_STATUS_DOWNLOADING {
		note = GetNote(messages...)
		note = GetNote(note, noteFunds)
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

	if deal.Status == newSwanStatus && deal.Note == note && deal.FilePath == filefullpathTemp && deal.Cost == cost {
		logs.GetLogger().Info(GetLog(deal, constants.NOT_UPDATE_OFFLINE_DEAL_STATUS))
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

func UpdateStatusAndLog(deal *model.OfflineDeal, newSwanStatus string, messages ...string) {
	UpdateDealInfoAndLog(deal, newSwanStatus, nil, messages...)
}

func GetLog(deal *model.OfflineDeal, messages ...string) string {
	text := GetNote(messages...)
	msg := fmt.Sprintf("deal(id=%d):%s,%s", deal.Id, deal.DealCid, text)
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
	offlineDeals, err := swanClient.GetOfflineDealsByStatus(dealStatus, &minerFid, nil, &pageNum, limit)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return offlineDeals
}

func GetOfflineDeal(swanClient *swan.SwanClient, dealStatus string, minerFid string) *libmodel.OfflineDeal {
	pageNum := 1
	pageSize := 1
	offlineDeals, err := swanClient.GetOfflineDealsByStatus(dealStatus, &minerFid, nil, &pageNum, &pageSize)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	if len(offlineDeals) > 0 {
		return offlineDeals[0]
	}

	return nil
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

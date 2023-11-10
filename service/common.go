package service

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/filswan/swan-boost-lib/provider"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"swan-provider/common/constants"
	"swan-provider/config"
	"syscall"
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

const LOTUS_IMPORT_NUMNBER = 20 //Max number of deals to be imported at a time
const LOTUS_SCAN_NUMBER = 100   //Max number of deals to be scanned at a time

var aria2Client *client.Aria2Client
var swanClient *swan.SwanClient

var swanService *SwanService
var aria2Service *Aria2Service
var lotusService *LotusService

var BoostPid int

func AdminOfflineDeal() {
	swanService = GetSwanService()
	aria2Service = GetAria2Service()
	lotusService = GetLotusService()

	if lotusService.MarketVersion == constants.MARKET_VERSION_1 {
		fmt.Println(color.YellowString("You are using the MARKET(version=1.1 built-in Lotus) import deals, but it is deprecated, will remove soon. Please set [main.market_version=“1.2”]"))
	}

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

	if lotusService.MarketVersion == constants.MARKET_VERSION_1 {
		marketApiUrl := config.GetConfig().Lotus.MarketApiUrl
		marketAccessToken := config.GetConfig().Lotus.MarketAccessToken

		if utils.IsStrEmpty(&lotusService.LotusClient.ApiUrl) {
			logs.GetLogger().Fatal("please set config:lotus->client_api_url")
		}
		if utils.IsStrEmpty(&marketApiUrl) {
			logs.GetLogger().Fatal("please set config:lotus->market_api_url")
		}
		if utils.IsStrEmpty(&marketAccessToken) {
			logs.GetLogger().Fatal("please set config:lotus->market_access_token")
		}

		lotusMarket, err := lotus.GetLotusMarket(marketApiUrl, marketAccessToken, lotusService.LotusClient.ApiUrl)
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}

		lotusService.LotusMarket = lotusMarket
		isWriteAuth, err := lotus.LotusCheckAuth(marketApiUrl, marketAccessToken, libconstants.LOTUS_AUTH_WRITE)
		if err != nil {
			logs.GetLogger().Error(err)
			logs.GetLogger().Fatal("please check config:lotus->market_api_url, lotus->market_access_token")
		}
		if !isWriteAuth {
			logs.GetLogger().Fatal("market access token should have write access right")
		}
	} else if lotusService.MarketVersion == constants.MARKET_VERSION_2 {
		market := config.GetConfig().Market
		if _, err := os.Stat(market.Repo); err != nil {
			if err := initBoost(market.Repo, market.MinerApi, market.FullNodeApi, market.PublishWallet, market.CollateralWallet); err != nil {
				os.Exit(0)
				return
			}
			logs.GetLogger().Info("init boostd successful")

			// enable Leveldb
			if err = boostEnableLeveldb(filepath.Join(market.Repo, "config.toml")); err != nil {
				logs.GetLogger().Warning("enable leveldb failed, please manually update [LocalIndexDirectory.Leveldb] Enabled=true")
				os.Exit(0)
			}
			logs.GetLogger().Info("boostd enable leveldb successful")
		}

		rpcApi, _, err := config.GetRpcInfoByFile(filepath.Join(market.Repo, "config.toml"))
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}

		// start boostd-data
		if _, err = startBoostData(market.Repo, market.BoostDataLog); err != nil {
			logs.GetLogger().Errorf("start boostd-data service failed, error: %+v", err)
			os.Exit(0)
		}

		// start boostd
		boostPid, err := startBoost(market.Repo, market.BoostLog, market.FullNodeApi)
		if err != nil {
			logs.GetLogger().Fatal(err)
			return
		}
		boostToken, err := GetBoostToken(market.Repo)
		boostClient, closer, err := provider.NewClient(boostToken, rpcApi)
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}
		defer closer()

		for {
			if _, err = boostClient.GetDealsConsiderOfflineStorageDeals(context.TODO()); err == nil {
				break
			} else {
				logs.GetLogger().Errorf("boost started failed, error: %v", err)
			}
			time.Sleep(1 * time.Second)
		}

		logs.GetLogger().Infof("start boostd rpc service successful, pid: %d", boostPid)
		BoostPid = boostPid
	}

	currentEpoch, err := lotusService.LotusClient.LotusGetCurrentEpoch()
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

	err := UpdateOfflineDeal(swanClient, deal.Id, newSwanStatus, &note, &filefullpathTemp, deal.ChainDealId)
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
	msg := fmt.Sprintf("taskName:%s, dealCid|dealUuid:%s, %s", *deal.TaskName, deal.DealCid, text)
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

func UpdateOfflineDeal(swanClient *swan.SwanClient, dealId int, status string, note, filePath *string, chainDealId int64) error {
	var params *swan.UpdateOfflineDealParams
	if chainDealId != 0 {
		params = &swan.UpdateOfflineDealParams{
			DealId:      dealId,
			Status:      status,
			Note:        note,
			FilePath:    filePath,
			ChainDealId: chainDealId,
		}
	} else {
		params = &swan.UpdateOfflineDealParams{
			DealId:   dealId,
			Status:   status,
			Note:     note,
			FilePath: filePath,
		}
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
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func initBoost(repo, minerApi, fullNodeApi, publishWallet, collatWallet string) error {
	ctx, cancelFunc := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancelFunc()
	args := make([]string, 0)
	args = append(args, "--vv")
	args = append(args, "--boost-repo="+repo)
	args = append(args, "init")
	args = append(args, "--api-sealer="+minerApi)
	args = append(args, "--api-sector-index="+minerApi)
	args = append(args, "--wallet-publish-storage-deals="+publishWallet)
	args = append(args, "--wallet-deal-collateral="+collatWallet)
	args = append(args, "--max-staging-deals-bytes=5000000000000000")

	cmd := exec.CommandContext(ctx, "boostd", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("MINER_API_INFO=%s", minerApi), fmt.Sprintf("FULLNODE_API_INFO=%s", fullNodeApi))

	if data, err := cmd.CombinedOutput(); err != nil {
		logs.GetLogger().Errorf("init boostd failed, output: %s,error: %+v", string(data), err)
		return err
	}
	return nil
}

func startBoost(repo, logFile, fullNodeApi string) (int, error) {
	args := make([]string, 0)
	args = append(args, "boostd")
	args = append(args, "--vv")
	args = append(args, "--boost-repo="+repo)
	args = append(args, "run")

	outFile, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, errors.Wrap(err, "open log file failed")
	}
	boostProcess, err := os.StartProcess("/usr/local/bin/boostd", args, &os.ProcAttr{
		Env: append(os.Environ(), fmt.Sprintf("FULLNODE_API_INFO=%s", fullNodeApi)),
		Sys: &syscall.SysProcAttr{
			Setsid: true,
		},
		Files: []*os.File{
			nil,
			outFile,
			outFile},
	})

	if err != nil {
		logs.GetLogger().Error(err)
		return 0, errors.Wrap(err, "start boostd process failed")
	}
	time.Sleep(10 * time.Second)
	return boostProcess.Pid, nil
}

func startBoostData(repo, logFile string) (int, error) {
	args := make([]string, 0)
	args = append(args, "boostd-data")
	args = append(args, "--repo="+repo)
	args = append(args, "--vv")
	args = append(args, "run")

	outFile, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, errors.Wrap(err, "open log file failed")
	}
	boostProcess, err := os.StartProcess("/usr/local/bin/boostd", args, &os.ProcAttr{
		Sys: &syscall.SysProcAttr{
			Setsid: true,
		},
		Files: []*os.File{
			nil,
			outFile,
			outFile},
	})

	if err != nil {
		logs.GetLogger().Error(err)
		return 0, errors.Wrap(err, "start boostd-data process failed")
	}
	logs.GetLogger().Warn("wait for the boostd-data startup to be finished...")
	time.Sleep(10 * time.Second)
	return boostProcess.Pid, nil
}

func boostEnableLeveldb(configFile string) error {
	args := []string{"-i", "/\\[LocalIndexDirectory.Leveldb\\]/,/Enabled/s/#Enabled = false/Enabled = true/", configFile}
	cmd := exec.Command("sed", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return errors.New("exec sed cmd failed")
	}
	return nil
}

func StopProcessById(processName string, pid int) {
	if pid == 0 {
		return
	}
	cmd := exec.Command("bash", "-c", fmt.Sprintf("sudo kill %d", pid))
	if _, err := cmd.CombinedOutput(); err != nil {
		logs.GetLogger().Errorf("stop %s failed, error: %s", processName, err.Error())
		return
	}
	logs.GetLogger().Infof("stop %s successfully", processName)
}

func GetBoostToken(repo string) (string, error) {
	tokenFile, err := ioutil.ReadFile(path.Join(repo, "token"))
	if err != nil {
		log.Println(err)
		return "", errors.Wrap(err, "open token file failed")
	}
	return string(tokenFile), nil
}

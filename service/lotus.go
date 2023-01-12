package service

import (
	"context"
	"fmt"
	"github.com/filswan/go-swan-lib/client/boost"
	"github.com/filswan/go-swan-lib/model"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"swan-provider/common/constants"
	"swan-provider/common/hql"
	"swan-provider/config"
	"sync"
	"time"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
)

type LotusService struct {
	MinerFid             string
	ImportIntervalSecond time.Duration
	ExpectedSealingTime  int
	ScanIntervalSecond   time.Duration
	LotusMarket          *lotus.LotusMarket
	LotusClient          *lotus.LotusClient
	importingDirs        sync.Map
	MarketVersion        string
}

func GetLotusService() *LotusService {
	confMain := config.GetConfig().Main

	lotusService := &LotusService{
		MinerFid:             confMain.MinerFid,
		ImportIntervalSecond: confMain.LotusImportInterval * time.Second,
		ExpectedSealingTime:  config.GetConfig().Bid.ExpectedSealingTime,
		ScanIntervalSecond:   confMain.LotusScanInterval * time.Second,
	}

	lotusService.MarketVersion = confMain.MarketVersion
	clientApiUrl := config.GetConfig().Lotus.ClientApiUrl
	lotusClient, err := lotus.LotusGetClient(clientApiUrl, "")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}
	lotusService.LotusClient = lotusClient
	return lotusService
}

func (lotusService *LotusService) StartImport(swanClient *swan.SwanClient) {
	maxImportNum := LOTUS_IMPORT_NUMNBER
	deals := GetOfflineDeals(swanClient, DEAL_STATUS_IMPORT_READY, aria2Service.MinerFid, &maxImportNum)
	if len(deals) == 0 {
		logs.GetLogger().Info("no pending offline deals found")
		return
	}

	aria2AutoDeleteCarFile := config.GetConfig().Aria2.Aria2AutoDeleteCarFile
	for _, deal := range deals {
		if _, ok := lotusService.importingDirs.Load(filepath.Dir(deal.FilePath)); ok {
			continue
		}

		var onChainStatus, onChainMessage *string
		var minerId string
		var err error
		var dealId uint64
		if lotusService.MarketVersion == constants.MARKET_VERSION_1 {
			minerId, dealId, onChainStatus, onChainMessage, err = lotusService.LotusMarket.LotusGetDealOnChainStatus(deal.DealCid)
			if err != nil {
				logs.GetLogger().Error(err)
				return
			}
			if onChainStatus == nil && onChainMessage == nil {
				UpdateStatusAndLog(deal, ONCHAIN_DEAL_STATUS_ERROR, "can not find from lotus-miner DagStore")
				continue
			}
		} else if lotusService.MarketVersion == constants.MARKET_VERSION_2 {
			_, graphqlApi, err := config.GetRpcInfoByFile(filepath.Join(config.GetConfig().Market.Repo, "config.toml"))
			if err != nil {
				logs.GetLogger().Error(err)
				return
			}
			hqlClient, err := hql.NewClient(graphqlApi)
			if err != nil {
				logs.GetLogger().Error(err)
				return
			}

			if _, err := uuid.Parse(deal.DealCid); err == nil {
				dealResp, err := hqlClient.GetDealByUuid(deal.DealCid)
				if err != nil {
					logs.GetLogger().Error(err)
					return
				}
				minerId = dealResp.Deal.GetProviderAddress()
				dealId, err = strconv.ParseUint(dealResp.Deal.GetChainDealID().Value, 10, 64)
				dealStatus := hql.DealStatus(dealResp.Deal.Checkpoint, dealResp.Deal.Err)
				onChainStatus = &dealStatus
				onChainMessage = &dealResp.Deal.Message
			} else {
				dealResp, err := hqlClient.GetProposalCid(deal.DealCid)
				if err != nil {
					logs.GetLogger().Error(err)
					continue
				}

				minerId = dealResp.LegacyDeal.GetProviderAddress()
				dealId, err = strconv.ParseUint(dealResp.LegacyDeal.GetChainDealID().Value, 10, 64)
				onChainStatus = &dealResp.LegacyDeal.Status
				onChainMessage = &dealResp.LegacyDeal.Message
			}

		}

		lotusService.importingDirs.Store(filepath.Dir(deal.FilePath), struct{}{})
		go func(minerId string, dealId uint64, onChainStatus *string, onChainMessage string, deal *model.OfflineDeal, aria2AutoDeleteCarFile bool) {
			UpdateSwanDealStatus(minerId, dealId, onChainStatus, onChainMessage, deal, aria2AutoDeleteCarFile)
			lotusService.importingDirs.Delete(filepath.Dir(deal.FilePath))
		}(minerId, dealId, onChainStatus, *onChainMessage, deal, aria2AutoDeleteCarFile)
	}
}

func (lotusService *LotusService) StartScan(swanClient *swan.SwanClient) {
	maxScanNum := LOTUS_SCAN_NUMBER
	importedDeals := GetOfflineDeals(swanClient, DEAL_STATUS_IMPORTED, aria2Service.MinerFid, &maxScanNum)
	importingDeals := GetOfflineDeals(swanClient, DEAL_STATUS_IMPORTING, aria2Service.MinerFid, &maxScanNum)

	deals := make([]*model.OfflineDeal, 0)
	deals = append(deals, importedDeals...)
	deals = append(deals, importingDeals...)
	if len(deals) == 0 {
		logs.GetLogger().Info("no ongoing offline deals found")
		return
	}

	if lotusService.MarketVersion == constants.MARKET_VERSION_1 {
		lotusDeals, err := lotusService.LotusMarket.LotusGetDeals()
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}
		if len(lotusDeals) == 0 {
			logs.GetLogger().Error("no deals returned from lotus")
			return
		}
		aria2AutoDeleteCarFile := config.GetConfig().Aria2.Aria2AutoDeleteCarFile
		for _, deal := range deals {
			minerId, dealId, onChainStatus, onChainMessage, err := lotusService.LotusMarket.LotusGetDealOnChainStatusFromDeals(lotusDeals, deal.DealCid)
			if err != nil {
				logs.GetLogger().Error(GetLog(deal, err.Error()))
				return
			}
			if onChainStatus == nil && onChainMessage == nil {
				UpdateStatusAndLog(deal, ONCHAIN_DEAL_STATUS_ERROR, "can not find from lotus-miner DagStore")
				continue
			}
			UpdateSwanDealStatus(minerId, dealId, onChainStatus, *onChainMessage, deal, aria2AutoDeleteCarFile)
		}
	} else {
		aria2AutoDeleteCarFile := config.GetConfig().Aria2.Aria2AutoDeleteCarFile
		for _, deal := range deals {
			_, graphqlApi, err := config.GetRpcInfoByFile(filepath.Join(config.GetConfig().Market.Repo, "config.toml"))
			if err != nil {
				logs.GetLogger().Error(err)
				return
			}
			hqlClient, err := hql.NewClient(graphqlApi)
			if err != nil {
				logs.GetLogger().Error(err)
				return
			}

			var onChainStatus *string
			var minerId, message string
			var dealId uint64
			if _, err := uuid.Parse(deal.DealCid); err == nil {
				dealResp, err := hqlClient.GetDealByUuid(deal.DealCid)
				if err != nil {
					logs.GetLogger().Errorf("taskName: %s, dealUuid: %s, get deal info failed, error: %+v", *deal.TaskName, deal.DealCid, err)
					UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "not found the deal in the db")
					continue
				}
				minerId = dealResp.Deal.GetProviderAddress()
				dealId, err = strconv.ParseUint(dealResp.Deal.GetChainDealID().Value, 10, 64)
				if err != nil {
					logs.GetLogger().Error(err)
					return
				}
				dealStatus := hql.DealStatus(dealResp.Deal.Checkpoint, dealResp.Deal.Err)
				onChainStatus = &dealStatus
				message = dealResp.Deal.Message
			} else {
				dealResp, err := hqlClient.GetProposalCid(deal.DealCid)
				if err != nil {
					logs.GetLogger().Errorf("taskName: %s, dealCid: %s, get deal info failed, error: %+v", *deal.TaskName, deal.DealCid, err)
					UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "not found the deal in the db")
					continue
				}

				minerId = dealResp.LegacyDeal.GetProviderAddress()
				dealId, err = strconv.ParseUint(dealResp.LegacyDeal.GetChainDealID().Value, 10, 64)
				onChainStatus = &dealResp.LegacyDeal.Status
				message = dealResp.LegacyDeal.Message
			}

			UpdateSwanDealStatus(minerId, dealId, onChainStatus, message, deal, aria2AutoDeleteCarFile)
		}
	}
}

func IsExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil || os.IsExist(err)
}

func DeleteDownloadedFiles(filePath string) {
	aria2AutoDeleteCarFile := config.GetConfig().Aria2.Aria2AutoDeleteCarFile
	if aria2AutoDeleteCarFile {
		if IsExist(filePath) {
			err := os.Remove(filePath)
			if err != nil {
				logs.GetLogger().Error("failed to delete file ", err, " file path ", filePath)
			} else {
				logs.GetLogger().Info("delete file successfully ", " file path ", filePath)
			}
		}
	}
}

func CorrectDealStatus(startEpoch int, minerId string, dealId uint64, onChainStatus string) (*string, error) {
	dealInfo, err := lotusService.LotusClient.LotusGetDealById(dealId)
	if err != nil {
		logs.GetLogger().Errorf("get market deal info by dealId failed,dealId: %d,error: %s ", dealId, err.Error())
		return nil, err
	}
	compatibleMinerId := strings.ReplaceAll(minerId, "f", "t")

	if dealInfo.State.SectorStartEpoch > -1 && dealInfo.State.SlashEpoch == -1 &&
		(dealInfo.Proposal.Provider == minerId || dealInfo.Proposal.Provider == compatibleMinerId) {
		onChainStatus = "StorageDealActive"
		return &onChainStatus, err
	}

	currentEpoch, err := lotusService.LotusClient.LotusGetCurrentEpoch()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	if startEpoch < int(*currentEpoch)+lotusService.ExpectedSealingTime {
		onChainStatus = "StorageDealError"
	} else {
		onChainStatus = ONCHAIN_DEAL_STATUS_SEALING
	}
	return &onChainStatus, nil
}

func UpdateSwanDealStatus(minerId string, dealId uint64, onChainStatus *string, onChainMessage string, deal *model.OfflineDeal, aria2AutoDeleteCarFile bool) {
	if dealId > 0 {
		status, err := CorrectDealStatus(deal.StartEpoch, minerId, dealId, *onChainStatus)
		if err != nil {
			logs.GetLogger().Error(GetLog(deal, err.Error()))
			return
		}
		onChainStatus = status
	}

	if utils.IsStrEmpty(onChainStatus) {
		logs.GetLogger().Info(GetLog(deal, "not found the deal on the chain"))
		UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "not found the deal on the chain")
		return
	}

	switch *onChainStatus {
	case ONCHAIN_DEAL_STATUS_ERROR:
		UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal error", *onChainStatus, onChainMessage)
		if aria2AutoDeleteCarFile {
			logs.GetLogger().Infof("dealId:%d, taskName:%s, dealCid|dealUuid:%s, has been %s, delete the car file, filePath:%s", dealId, *deal.TaskName, deal.DealCid, *onChainStatus, deal.FilePath)
			DeleteDownloadedFiles(deal.FilePath)
		}
	case ONCHAIN_DEAL_STATUS_ACTIVE:
		if lotusService.MarketVersion == constants.MARKET_VERSION_2 {
			UpdateStatusAndLog(deal, DEAL_STATUS_ACTIVE, "deal has been completed", *onChainStatus)
		} else {
			UpdateStatusAndLog(deal, DEAL_STATUS_ACTIVE, "deal has been completed", *onChainStatus, onChainMessage)
		}
		if aria2AutoDeleteCarFile {
			logs.GetLogger().Infof("dealId:%d, taskName:%s, dealCid|dealUuid:%s, has been %s, delete the car file, filePath:%s", dealId, *deal.TaskName, deal.DealCid, *onChainStatus, deal.FilePath)
			DeleteDownloadedFiles(deal.FilePath)
		}
	case ONCHAIN_DEAL_STATUS_ACCEPT:
		UpdateStatusAndLog(deal, deal.Status, "deal will be ready shortly", *onChainStatus, onChainMessage)
	case ONCHAIN_DEAL_STATUS_NOTFOUND:
		UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal not found", *onChainStatus, onChainMessage)
	case ONCHAIN_DEAL_STATUS_AWAITING, ONCHAIN_DEAL_STATUS_SEALING:
		UpdateStatusAndLog(deal, DEAL_STATUS_IMPORTED, "deal is sealing", *onChainStatus)
	case ONCHAIN_DEAL_STATUS_WAITTING:
		if deal.Status == DEAL_STATUS_IMPORTING {
			return
		}
		currentEpoch, err := lotusService.LotusClient.LotusGetCurrentEpoch()
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}

		if int64(deal.StartEpoch)-*currentEpoch < int64(lotusService.ExpectedSealingTime) {
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal expired before importing", *onChainStatus, onChainMessage)
			return
		}

		UpdateStatusAndLog(deal, DEAL_STATUS_IMPORTING, "importing deal")

		if lotusService.MarketVersion == constants.MARKET_VERSION_1 {
			err = lotusService.LotusMarket.LotusImportData(deal.DealCid, deal.FilePath)
			if err != nil { //There should be no output if everything goes well
				UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "import deal failed", err.Error())
				return
			}
		} else {
			market := config.GetConfig().Market
			boostToken, err := GetBoostToken(market.Repo)
			if err != nil {
				logs.GetLogger().Error(err)
				return
			}
			rpcApi, _, err := config.GetRpcInfoByFile(filepath.Join(config.GetConfig().Market.Repo, "config.toml"))
			if err != nil {
				logs.GetLogger().Error(err)
				return
			}

			boostClient, closer, err := boost.NewClient(boostToken, rpcApi)
			if err != nil {
				logs.GetLogger().Error(err)
				return
			}
			defer closer()

			if _, err := uuid.Parse(deal.DealCid); err == nil {
				rej, err := boostClient.OfflineDealWithData(context.TODO(), deal.DealCid, deal.FilePath)
				var msg string
				if err != nil {
					msg = fmt.Sprintf("import deal failed: %w", err.Error())
				}
				if rej != nil && rej.Reason != "" {
					msg = fmt.Sprintf("offline deal %s rejected: %s", deal.DealCid, rej.Reason)
				}
				if msg != "" {
					UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, msg)
				}
			} else {
				err = boostClient.OfflineDealWithDataByMarket(context.TODO(), deal.DealCid, deal.FilePath)
				if err != nil {
					UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "import deal failed", err.Error())
					return
				}
			}
		}

		UpdateStatusAndLog(deal, DEAL_STATUS_IMPORTED, "deal imported")
	default:
		if *onChainStatus == "StorageDealPublish" || *onChainStatus == "StorageDealPublishing" {
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORTED, "deal imported")
			return
		}
		UpdateStatusAndLog(deal, deal.Status, *onChainStatus, onChainMessage)
	}
}

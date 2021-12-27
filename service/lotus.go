package service

import (
	"swan-provider/config"
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
}

func GetLotusService() *LotusService {
	confMain := config.GetConfig().Main

	lotusService := &LotusService{
		MinerFid:             confMain.MinerFid,
		ImportIntervalSecond: confMain.LotusImportInterval * time.Second,
		ExpectedSealingTime:  config.GetConfig().Bid.ExpectedSealingTime,
		ScanIntervalSecond:   confMain.LotusScanInterval * time.Second,
	}

	marketApiUrl := config.GetConfig().Lotus.MarketApiUrl
	marketAccessToken := config.GetConfig().Lotus.MarketAccessToken
	clientApiUrl := config.GetConfig().Lotus.ClientApiUrl
	lotusMarket, err := lotus.GetLotusMarket(marketApiUrl, marketAccessToken, clientApiUrl)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	lotusService.LotusMarket = lotusMarket

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

	for _, deal := range deals {
		onChainStatus, onChainMessage, err := lotusService.LotusMarket.LotusGetDealOnChainStatus(deal.DealCid)
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}

		if utils.IsStrEmpty(onChainStatus) {
			logs.GetLogger().Error(GetLog(deal, "failed to get on chain status, please check if lotus miner is running properly"))
			continue
		}

		logs.GetLogger().Info(GetLog(deal, *onChainStatus, *onChainMessage))

		switch *onChainStatus {
		case ONCHAIN_DEAL_STATUS_ERROR:
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal is error before importing", *onChainStatus, *onChainMessage)
		case ONCHAIN_DEAL_STATUS_ACTIVE:
			UpdateStatusAndLog(deal, DEAL_STATUS_ACTIVE, "deal is active before importing", *onChainStatus, *onChainMessage)
		case ONCHAIN_DEAL_STATUS_ACCEPT:
			UpdateStatusAndLog(deal, deal.Status, "deal will be ready shortly", *onChainStatus, *onChainMessage)
		case ONCHAIN_DEAL_STATUS_NOTFOUND:
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal not found", *onChainStatus, *onChainMessage)
		case ONCHAIN_DEAL_STATUS_WAITTING:
			currentEpoch, err := lotusService.LotusClient.LotusGetCurrentEpoch()
			if err != nil {
				logs.GetLogger().Error(err)
				return
			}

			if int64(deal.StartEpoch)-*currentEpoch < int64(lotusService.ExpectedSealingTime) {
				UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal expired before importing", *onChainStatus, *onChainMessage)
				continue
			}

			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORTING, "importing deal")

			err = lotusService.LotusMarket.LotusImportData(deal.DealCid, deal.FilePath)
			if err != nil { //There should be no output if everything goes well
				UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "import deal failed", err.Error())
				continue
			}
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORTED, "deal imported")
		default:
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORTED, "deal already imported", *onChainStatus, *onChainMessage)
		}

		logs.GetLogger().Info("Sleeping...")
		time.Sleep(lotusService.ImportIntervalSecond)
	}
}

func (lotusService *LotusService) StartScan(swanClient *swan.SwanClient) {
	maxScanNum := LOTUS_SCAN_NUMBER
	deals := GetOfflineDeals(swanClient, DEAL_STATUS_IMPORTED, aria2Service.MinerFid, &maxScanNum)
	if len(deals) == 0 {
		logs.GetLogger().Info("no ongoing offline deals found")
		return
	}

	lotusDeals, err := lotusService.LotusMarket.LotusGetDeals()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	if len(lotusDeals) == 0 {
		logs.GetLogger().Error("no deals returned from lotus")
		return
	}

	for _, deal := range deals {
		//logs.GetLogger().Info(GetLog(deal, "current status in swan:"+deal.Status, "current note in swan:"+deal.Note))
		onChainStatus, onChainMessage, err := lotusService.LotusMarket.LotusGetDealOnChainStatusFromDeals(lotusDeals, deal.DealCid)
		if err != nil {
			logs.GetLogger().Error(GetLog(deal, err.Error()))
			return
		}

		if utils.IsStrEmpty(onChainStatus) {
			logs.GetLogger().Error(GetLog(deal, "on chain status is empty"))
			continue
		}

		switch *onChainStatus {
		case ONCHAIN_DEAL_STATUS_ERROR:
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal error when scan", *onChainStatus, *onChainMessage)
		case ONCHAIN_DEAL_STATUS_ACTIVE:
			UpdateStatusAndLog(deal, DEAL_STATUS_ACTIVE, "deal has been completed", *onChainStatus, *onChainMessage)
		case ONCHAIN_DEAL_STATUS_AWAITING:
			currentEpoch, err := lotusService.LotusClient.LotusGetCurrentEpoch()
			if err != nil {
				logs.GetLogger().Error(GetLog(deal, err.Error()))
				return
			}

			if *currentEpoch > int64(deal.StartEpoch) {
				UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "sector is proved and active, on chain status bug", *onChainStatus, *onChainMessage)
			} else {
				UpdateStatusAndLog(deal, deal.Status, *onChainStatus, *onChainMessage)
			}
		default:
			UpdateStatusAndLog(deal, deal.Status, *onChainStatus, *onChainMessage)
		}
	}
}

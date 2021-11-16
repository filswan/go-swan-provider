package service

import (
	"swan-provider/config"
	"time"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	"github.com/filswan/go-swan-lib/logs"
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
	deals := swanClient.SwanGetOfflineDeals(lotusService.MinerFid, DEAL_STATUS_IMPORT_READY, LOTUS_IMPORT_NUMNBER)
	if len(deals) == 0 {
		logs.GetLogger().Info("no pending offline deals found")
		return
	}

	for _, deal := range deals {
		logs.GetLogger().Info(GetLog(deal, "filepath:"+deal.FilePath))

		onChainStatus, onChainMessage := lotusService.LotusMarket.LotusGetDealOnChainStatus(deal.DealCid)
		if len(onChainStatus) == 0 {
			logs.GetLogger().Error(GetLog(deal, "failed to get on chain status, please check if lotus-miner is running properly"))
			continue
		}

		logs.GetLogger().Info(GetLog(deal, onChainStatus, onChainMessage))

		switch onChainStatus {
		case ONCHAIN_DEAL_STATUS_ERROR:
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal is error before importing", onChainStatus, onChainMessage)
		case ONCHAIN_DEAL_STATUS_ACTIVE:
			UpdateStatusAndLog(deal, DEAL_STATUS_ACTIVE, "deal is active before importing", onChainStatus, onChainMessage)
		case ONCHAIN_DEAL_STATUS_ACCEPT:
			UpdateStatusAndLog(deal, deal.Status, "deal will be ready shortly", onChainStatus, onChainMessage)
		case ONCHAIN_DEAL_STATUS_NOTFOUND:
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal not found", onChainStatus, onChainMessage)
		case ONCHAIN_DEAL_STATUS_WAITTING:
			currentEpoch := lotusService.LotusClient.LotusGetCurrentEpoch()
			if currentEpoch < 0 {
				UpdateStatusAndLog(deal, deal.Status, "failed to get current epoch", onChainStatus, onChainMessage)
				continue
			}

			if deal.StartEpoch-currentEpoch < lotusService.ExpectedSealingTime {
				UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal expired before importing", onChainStatus, onChainMessage)
				continue
			}

			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORTING, "importing deal")

			err := lotusService.LotusMarket.LotusImportData(deal.DealCid, deal.FilePath)

			if err != nil { //There should be no output if everything goes well
				UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "import deal failed", err.Error())
				continue
			}
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORTED, "deal imported")
		default:
			UpdateDealInfoAndLog(deal, DEAL_STATUS_IMPORTED, nil, "deal already imported", onChainStatus, onChainMessage)
		}

		logs.GetLogger().Info("Sleeping...")
		time.Sleep(lotusService.ImportIntervalSecond)
	}
}

func (lotusService *LotusService) StartScan(swanClient *swan.SwanClient) {
	deals := swanClient.SwanGetOfflineDeals(lotusService.MinerFid, DEAL_STATUS_IMPORTED, LOTUS_SCAN_NUMBER)
	if len(deals) == 0 {
		logs.GetLogger().Info("no ongoing offline deals found")
		return
	}

	lotusDeals := lotusService.LotusMarket.LotusGetDeals()
	if len(lotusDeals) == 0 {
		logs.GetLogger().Error("failed to get deals from lotus")
		return
	}

	for _, deal := range deals {
		logs.GetLogger().Info(GetLog(deal, "current status in swan:"+deal.Status, "current note in swan:"+deal.Note))

		onChainStatus, onChainMessage := lotusService.LotusMarket.LotusGetDealOnChainStatusFromDeals(lotusDeals, deal.DealCid)
		if len(onChainStatus) == 0 {
			logs.GetLogger().Error(GetLog(deal, "failed to get on chain status"))
			continue
		}

		switch onChainStatus {
		case ONCHAIN_DEAL_STATUS_ERROR:
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal error when scan", onChainStatus, onChainMessage)
		case ONCHAIN_DEAL_STATUS_ACTIVE:
			UpdateStatusAndLog(deal, DEAL_STATUS_ACTIVE, "deal has been completed", onChainStatus, onChainMessage)
		case ONCHAIN_DEAL_STATUS_AWAITING:
			currentEpoch := lotusService.LotusClient.LotusGetCurrentEpoch()
			if currentEpoch < 0 {
				UpdateStatusAndLog(deal, deal.Status, "failed to get current epoch", onChainStatus, onChainMessage)
				continue
			}

			if currentEpoch > deal.StartEpoch {
				UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "sector is proved and active, on chain status bug", onChainStatus, onChainMessage)
			} else {
				UpdateStatusAndLog(deal, deal.Status, onChainStatus, onChainMessage)
			}
		default:
			UpdateStatusAndLog(deal, deal.Status, onChainStatus, onChainMessage)
		}
	}
}

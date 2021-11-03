package service

import (
	"fmt"
	"go-swan-provider/config"
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

func GetNote(message, onChainStatus, onChainMessage string) string {
	result := fmt.Sprintf("%sOn chain status:%s", message, onChainStatus)
	if onChainMessage != "" {
		result = result + ", message:" + onChainMessage
	}
	return result
}

func (lotusService *LotusService) StartImport(swanClient *swan.SwanClient) {
	deals := swanClient.SwanGetOfflineDeals(lotusService.MinerFid, DEAL_STATUS_IMPORT_READY, LOTUS_IMPORT_NUMNBER)
	if len(deals) == 0 {
		logs.GetLogger().Info("No pending offline deals found.")
		return
	}

	for _, deal := range deals {
		msg := fmt.Sprintf("Deal ID: %d, Deal CID: %s. File Path: %s", deal.Id, deal.DealCid, deal.FilePath)
		logs.GetLogger().Info(msg)

		onChainStatus, message := lotusService.LotusMarket.LotusGetDealOnChainStatus(deal.DealCid)

		if len(onChainStatus) == 0 {
			logs.GetLogger().Error("Failed to get on chain status for :", deal.DealCid)
			continue
		}

		logs.GetLogger().Info("Deal on chain status: ", onChainStatus)

		switch onChainStatus {
		case ONCHAIN_DEAL_STATUS_ERROR:
			note := GetNote("Deal is error before importing.", onChainStatus, message)
			logs.GetLogger().Warn("Deal id:", deal.Id, " CID: ", deal.DealCid, " ", note)
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
		case ONCHAIN_DEAL_STATUS_ACTIVE:
			note := GetNote("Deal is active before importing.", onChainStatus, message)
			logs.GetLogger().Info(note)
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_ACTIVE, note)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
		case ONCHAIN_DEAL_STATUS_ACCEPT:
			logs.GetLogger().Info("Deal on chain status is ", ONCHAIN_DEAL_STATUS_ACCEPT, ". Deal will be ready shortly.")
		case ONCHAIN_DEAL_STATUS_NOTFOUND:
			note := GetNote("Deal not found.", onChainStatus, message)
			logs.GetLogger().Info(note)
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
		case ONCHAIN_DEAL_STATUS_WAITTING:
			currentEpoch := lotusService.LotusClient.LotusGetCurrentEpoch()
			if currentEpoch < 0 {
				return
			}

			if deal.StartEpoch-currentEpoch < lotusService.ExpectedSealingTime {
				note := GetNote("Deal expired before importing.", onChainStatus, message)
				updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
				if !updated {
					logs.GetLogger().Error("Failed to update offline deal status")
				}
				note = fmt.Sprintf("Deal id:%d, CID:%s, Deal will start too soon. Do not import this deal.", deal.Id, deal.DealCid)
				logs.GetLogger().Warn(note)
				continue
			}

			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORTING)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}

			err := lotusService.LotusMarket.LotusImportData(deal.DealCid, deal.FilePath)

			if err != nil { //There should be no output if everything goes well
				updated = swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, err.Error())
				if !updated {
					logs.GetLogger().Error("Failed to update offline deal status")
				}
				msg = fmt.Sprintf("Import deal failed. id: %s. Error message: %s", deal.DealCid, err.Error())
				logs.GetLogger().Warn(msg)
				continue
			}

			updated = swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORTED)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
			msg = fmt.Sprintf("Deal CID %s imported.", deal.DealCid)
			logs.GetLogger().Info(msg)
		default:
			note := GetNote("Deal is already imported.", onChainStatus, message)
			logs.GetLogger().Info("Deal CID:", deal.DealCid, " ", note)
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORTED, note)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
		}

		logs.GetLogger().Info("Sleeping...")
		time.Sleep(lotusService.ImportIntervalSecond)
	}
}

func (lotusService *LotusService) StartScan(swanClient *swan.SwanClient) {
	deals := swanClient.SwanGetOfflineDeals(lotusService.MinerFid, DEAL_STATUS_IMPORTED, LOTUS_SCAN_NUMBER)

	if len(deals) == 0 {
		logs.GetLogger().Info("No ongoing offline deals found.")
		return
	}

	lotusDeals := lotusService.LotusMarket.LotusGetDeals()
	if len(lotusDeals) == 0 {
		logs.GetLogger().Error("Failed to get deals from lotus.")
		return
	}

	for _, deal := range deals {
		msg := fmt.Sprintf("ID: %d. Deal CID: %s. Deal Status: %s.", deal.Id, deal.DealCid, deal.Status)
		logs.GetLogger().Info(msg)

		onChainStatus, message := lotusService.LotusMarket.LotusGetDealOnChainStatusFromDeals(lotusDeals, deal.DealCid)

		if len(onChainStatus) == 0 {
			logs.GetLogger().Error("Failed to get on chain status for :", deal.DealCid)
			continue
		}

		logs.GetLogger().Info("Deal on chain status: ", onChainStatus)

		switch onChainStatus {
		case ONCHAIN_DEAL_STATUS_ERROR:
			note := GetNote("Deal error when scan.", onChainStatus, message)
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
			msg = fmt.Sprintf("Setting deal %s status as %s", deal.DealCid, DEAL_STATUS_IMPORT_FAILED)
			logs.GetLogger().Info(msg)
		case ONCHAIN_DEAL_STATUS_ACTIVE:
			note := GetNote("Deal has been completed.", onChainStatus, message)
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_ACTIVE, note)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
			msg = fmt.Sprintf("Setting deal %s status as %s", deal.DealCid, DEAL_STATUS_ACTIVE)
			logs.GetLogger().Info(msg)
		case ONCHAIN_DEAL_STATUS_AWAITING:
			currentEpoch := lotusService.LotusClient.LotusGetCurrentEpoch()
			if currentEpoch < 0 {
				return
			}

			if currentEpoch > deal.StartEpoch {
				note := GetNote("Sector is proved and active.", onChainStatus, message)
				updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
				if !updated {
					logs.GetLogger().Error("Failed to update offline deal status")
				}
				msg = fmt.Sprintf("Setting deal %s status as ImportFailed due to on chain status bug.", deal.DealCid)
				logs.GetLogger().Info(msg)
			}
		}

		msg = fmt.Sprintf("On chain offline_deal message created. Message Body: on_chain_status:%s, on_chain_message:%s.", onChainStatus, message)
		logs.GetLogger().Info(msg)
	}
}

package service

import (
	"fmt"
	"swan-provider/common/client"
	"swan-provider/config"
	"swan-provider/logs"
	"time"
)

type LotusService struct {
	MinerFid             string
	ImportIntervalSecond time.Duration
	ExpectedSealingTime  int
	ScanIntervalSecond   time.Duration
}

func GetLotusService() *LotusService {
	confMain := config.GetConfig().Main

	lotusService := &LotusService{
		MinerFid:             confMain.MinerFid,
		ImportIntervalSecond: confMain.LotusImportInterval * time.Second,
		ExpectedSealingTime:  config.GetConfig().Bid.ExpectedSealingTime,
		ScanIntervalSecond:   confMain.LotusScanInterval * time.Second,
	}

	return lotusService
}

func (lotusService *LotusService) StartImport(swanClient *client.SwanClient) {
	deals := swanClient.GetOfflineDeals(lotusService.MinerFid, DEAL_STATUS_IMPORT_READY, LOTUS_IMPORT_NUMNBER)
	if len(deals) == 0 {
		logs.GetLogger().Info("No pending offline deals found.")
		return
	}

	for _, deal := range deals {
		msg := fmt.Sprintf("Deal CID: %s. File Path: %s", deal.DealCid, deal.FilePath)
		logs.GetLogger().Info(msg)

		onChainStatus, message := client.LotusGetDealOnChainStatus(deal.DealCid)

		if len(onChainStatus) == 0 {
			logs.GetLogger().Error("Failed to get on chain status for :", deal.DealCid)
			continue
		}

		logs.GetLogger().Info("Deal on chain status: ", onChainStatus)

		switch onChainStatus {
		case ONCHAIN_DEAL_STATUS_ERROR:
			note := "Deal on chain status is error before importing." + message
			logs.GetLogger().Warn("Deal id:", deal.Id, " ", note)
			updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
		case ONCHAIN_DEAL_STATUS_ACTIVE:
			note := "Deal on chain status is active before importing."
			logs.GetLogger().Info(note)
			updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_ACTIVE, note)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
		case ONCHAIN_DEAL_STATUS_ACCEPT:
			logs.GetLogger().Info("Deal on chain status is ", ONCHAIN_DEAL_STATUS_ACCEPT, ". Deal will be ready shortly.")
		case ONCHAIN_DEAL_STATUS_NOTFOUND:
			note := "Deal on chain status not found."
			logs.GetLogger().Info(note)
			updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
		case ONCHAIN_DEAL_STATUS_WAITTING:
			currentEpoch := client.LotusGetCurrentEpoch()
			if currentEpoch < 0 {
				return
			}

			if deal.StartEpoch-currentEpoch < lotusService.ExpectedSealingTime {
				updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, "Deal expired.")
				if !updated {
					logs.GetLogger().Error("Failed to update offline deal status")
				}
				note := fmt.Sprintf("Deal id:%d, Deal will start too soon. Do not import this deal.", deal.Id)
				logs.GetLogger().Warn(note)
				continue
			}

			updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORTING)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}

			result := client.LotusImportData(deal.DealCid, deal.FilePath)

			if len(result) > 0 { //There should be no output if everything goes well
				updated = swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, result)
				if !updated {
					logs.GetLogger().Error("Failed to update offline deal status")
				}
				msg = fmt.Sprintf("Import deal failed. id: %s. Error message: %s", deal.DealCid, result)
				logs.GetLogger().Warn(msg)
				continue
			}

			updated = swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORTED)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
			msg = fmt.Sprintf("Deal CID %s imported.", deal.DealCid)
			logs.GetLogger().Info(msg)
		default:
			logs.GetLogger().Info("Deal is already imported, please check.")
			updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORTED, onChainStatus)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
		}

		logs.GetLogger().Info("Sleeping...")
		time.Sleep(lotusService.ImportIntervalSecond)
	}
}

func (lotusService *LotusService) StartScan(swanClient *client.SwanClient) {
	deals := swanClient.GetOfflineDeals(lotusService.MinerFid, DEAL_STATUS_IMPORTED, LOTUS_SCAN_NUMBER)

	if len(deals) == 0 {
		logs.GetLogger().Info("No ongoing offline deals found.")
		return
	}

	lotusDeals := client.LotusGetDeals()
	if len(lotusDeals) == 0 {
		logs.GetLogger().Error("Failed to get deals from lotus.")
		return
	}

	for _, deal := range deals {
		msg := fmt.Sprintf("ID: %d. Deal CID: %s. Deal Status: %s.", deal.Id, deal.DealCid, deal.Status)
		logs.GetLogger().Info(msg)

		onChainStatus, onChainMessage := client.LotusGetDealOnChainStatusFromDeals(lotusDeals, deal.DealCid)

		if len(onChainStatus) == 0 {
			logs.GetLogger().Error("Failed to get on chain status for :", deal.DealCid)
			continue
		}

		logs.GetLogger().Info("Deal on chain status: ", onChainStatus)

		switch onChainStatus {
		case ONCHAIN_DEAL_STATUS_ERROR:
			updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, onChainMessage)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
			msg = fmt.Sprintf("Setting deal %s status as %s", deal.DealCid, DEAL_STATUS_IMPORT_FAILED)
			logs.GetLogger().Info(msg)
		case ONCHAIN_DEAL_STATUS_ACTIVE:
			note := "Deal has been completed"
			updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_ACTIVE, note)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			}
			msg = fmt.Sprintf("Setting deal %s status as %s", deal.DealCid, DEAL_STATUS_ACTIVE)
			logs.GetLogger().Info(msg)
		case ONCHAIN_DEAL_STATUS_AWAITING:
			currentEpoch := client.LotusGetCurrentEpoch()
			if currentEpoch < 0 {
				return
			}

			if currentEpoch > deal.StartEpoch {
				note := fmt.Sprintf("Sector is proved and active, while deal on chain status is %s. Set deal status as %s.", ONCHAIN_DEAL_STATUS_AWAITING, DEAL_STATUS_IMPORT_FAILED)
				updated := swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
				if !updated {
					logs.GetLogger().Error("Failed to update offline deal status")
				}
				msg = fmt.Sprintf("Setting deal %s status as ImportFailed due to on chain status bug.", deal.DealCid)
				logs.GetLogger().Info(msg)
			}
		}

		msg = fmt.Sprintf("On chain offline_deal message created. Message Body: on_chain_status:%s, on_chain_message:%s.", onChainStatus, onChainMessage)
		logs.GetLogger().Info(msg)
	}
}

package offlineDealAdmin

import (
	"fmt"
	"swan-miner/common/utils"
	"swan-miner/config"
	"time"
)

func Importer() {
	conf:=config.GetConfig()
	confMain:=conf.Main

	importIntervalSecond := confMain.ImportInterval * time.Second
	expectedSealingTime := confMain.ExpectedSealingTime
	minerFid := confMain.MinerFid

	swanClient := utils.GetSwanClient()

	for {
		deals := swanClient.GetOfflineDeals(minerFid, DEAL_STATUS_IMPORT_READY, LOTUS_IMPORT_NUMNBER)
		if deals == nil || len(deals) == 0 {
			logger.Info("No pending offline deals found.")
			logger.Info("Sleeping...")
			time.Sleep(importIntervalSecond)
			continue
		}

		for _, deal := range deals {
			msg := fmt.Sprintf("Deal CID: %s. File Path: %s", deal.DealCid, deal.FilePath)
			logger.Info(msg)

			onChainStatus, _ := utils.GetDealOnChainStatus(deal.DealCid)

			if len(onChainStatus) == 0 {
				logger.Info("Sleeping...")
				time.Sleep(importIntervalSecond)
				break
			}

			logger.Info("Deal on chain status: ", onChainStatus)

			if onChainStatus == ONCHAIN_DEAL_STATUS_ERROR {
				note := "Deal on chain status is error before importing."
				logger.Info(note)
				swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
				continue
			}

			if onChainStatus == ONCHAIN_DEAL_STATUS_ACTIVE {
				note := "Deal on chain status is active before importing."
				logger.Info(note)
				swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_ACTIVE, note)
				continue
			}

			if onChainStatus == ONCHAIN_DEAL_STATUS_ACCEPT {
				logger.Info("Deal on chain status is ", ONCHAIN_DEAL_STATUS_ACCEPT, ". Deal will be ready shortly.")
				continue
			}

			if onChainStatus == ONCHAIN_DEAL_STATUS_NOTFOUND {
				note := "Deal on chain status not found."
				logger.Info(note)
				swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
				continue
			}

			if onChainStatus != ONCHAIN_DEAL_STATUS_WAITTING {
				logger.Info("Deal is already imported, please check.")
				swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORTED, onChainStatus)
				continue
			}

			currentEpoch := utils.GetCurrentEpoch()

			if currentEpoch < 0 {
				logger.Error("Failed to get current epoch. Please check if miner is running properly.")
				time.Sleep(importIntervalSecond)
				break
			}

			if deal.StartEpoch - currentEpoch < expectedSealingTime {
				note := "Deal will start too soon, expired. Do not import this deal."
				logger.Info(note)
				swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
				continue
			}

			swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORTING)

			result := utils.LotusImportData(deal.DealCid, deal.FilePath)

			//There should be no output if everything goes well
			if len(result) > 0 {
				swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, result)
				msg = fmt.Sprintf("Import deal failed. CID: %s. Error message: %s", deal.Id, result)
				logger.Error(msg)
				continue
			}

			swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORTED)
			msg = fmt.Sprintf("Deal CID %s imported.", deal.DealCid)
			logger.Info(msg)
			logger.Info("Sleeping...")
			time.Sleep(importIntervalSecond)
		}
	}
}

func Scanner() {
	confMain := config.GetConfig().Main

	swanClient := utils.GetSwanClient()
	for {
		deals := swanClient.GetOfflineDeals(confMain.MinerFid, DEAL_STATUS_IMPORTED, LOTUS_SCAN_NUMBER)

		if len(deals) == 0 {
			logger.Info("No ongoing offline deals found.")
			logger.Info("Sleeping...")
			time.Sleep(confMain.ScanInterval * time.Second)
			continue
		}

		for _, deal := range deals {
			//fmt.Println(deal)
			msg := fmt.Sprintf("ID: %s. Deal CID: %s. Deal Status: %s.", deal.Id, deal.DealCid, deal.Status)
			logger.Info(msg)

			onChainStatus, onChainMessage := utils.GetDealOnChainStatus(deal.DealCid)

			if len(onChainStatus) == 0 {
				logger.Info("Sleeping...")
				time.Sleep(confMain.ScanInterval * time.Second)
				break
			}

			logger.Info("Deal on chain status: ", onChainStatus)

			if onChainStatus == ONCHAIN_DEAL_STATUS_ERROR {
				swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, onChainMessage)
				msg = fmt.Sprintf("Setting deal %s status as %s", deal.DealCid, DEAL_STATUS_IMPORT_FAILED)
				logger.Info(msg)
			}

			if onChainStatus ==ONCHAIN_DEAL_STATUS_ACTIVE{
				note := "Deal has been completed"
				swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_ACTIVE, note)
				msg = fmt.Sprintf("Setting deal %s status as %s", deal.DealCid, DEAL_STATUS_ACTIVE)
				logger.Info(msg)
			}

			if onChainStatus == ONCHAIN_DEAL_STATUS_AWAITING {
				currentEpoch := utils.GetCurrentEpoch()
				if currentEpoch != -1 && currentEpoch > deal.StartEpoch {
					note := fmt.Sprintf("Sector is proved and active, while deal on chain status is %s. Set deal status as %s.", ONCHAIN_DEAL_STATUS_AWAITING, DEAL_STATUS_IMPORT_FAILED)
					swanClient.UpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
					msg := fmt.Sprintf("Setting deal %s status as ImportFailed due to on chain status bug.", deal.DealCid)
					logger.Info(msg)
				}
			}

			msg = fmt.Sprintf("On chain offline_deal message created. Message Body: on_chain_status:%s, on_chain_message:%s.", onChainStatus, onChainMessage)
			logger.Info(msg)
		}

		logger.Info("Sleeping...")
		time.Sleep(confMain.ScanInterval * time.Second)
	}
}


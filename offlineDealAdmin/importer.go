package offlineDealAdmin

import (
	"fmt"
	"swan-miner/common/utils"
	"swan-miner/config"
	"swan-miner/logs"
	"time"
)

const DEAL_STATUS_READY = "ReadyForImport"
const ONCHAIN_DEAL_STATUS_NOTFOUND = "StorageDealNotFound"
const ONCHAIN_DEAL_STATUS_WAITTING = "StorageDealWaitingForData"
const ONCHAIN_DEAL_STATUS_ACCEPT = "StorageDealAcceptWait"

const IMPORT_NUMNBER = "20"  //Max number of deals to be imported at a time

func Importer() {
	conf:=config.GetConfig()
	confMain:=conf.Main

	importInterval := confMain.ImportInterval
	expectedSealingTime := confMain.ExpectedSealingTime
	minerFid := confMain.MinerFid

	swanClient := utils.GetSwanClient()

	logger := logs.GetLogger()

	for {
		deals := swanClient.GetOfflineDeals(minerFid,DEAL_STATUS_READY, IMPORT_NUMNBER)
		if deals == nil || len(deals) == 0 {
			logger.Info("No pending offline deals found.")
			logger.Info("Sleeping...")
			continue
		}

		for i := 0; i < len(deals); i++ {
			deal := deals[i]
			//fmt.Println(deal)

			msg := fmt.Sprintf("Deal CID: %s. File Path: %s", deal.DealCid, deal.FilePath)
			logger.Error(msg)

			onChainStatus := utils.GetDealOnChainStatus(deal.DealCid)

			if len(onChainStatus) == 0 {
				logger.Error("Failed to get deal on chain status, please check if lotus-miner is running properly.")
				logger.Info("Sleeping...")
				time.Sleep(importInterval * time.Second)
				break
			}

			msg = fmt.Sprintf("Deal on chain status: %s.", onChainStatus)
			logger.Info(msg)

			if onChainStatus == ONCHAIN_DEAL_STATUS_ERROR {
				note := "Deal on chain status is error before importing."
				logger.Info(note)
				swanClient.UpdateOfflineDealStatus(DEAL_STATUS_FAILED, note, deal.Id)
				continue
			}

			if onChainStatus == ONCHAIN_DEAL_STATUS_ACTIVE {
				note := "Deal on chain status is active before importing."
				logger.Info(note)
				swanClient.UpdateOfflineDealStatus(DEAL_STATUS_ACTIVE, note, deal.Id)
				continue
			}

			if onChainStatus == ONCHAIN_DEAL_STATUS_ACCEPT {
				logger.Info("Deal on chain status is " + ONCHAIN_DEAL_STATUS_ACCEPT + ". Deal will be ready shortly.")
				continue
			}

			if onChainStatus == ONCHAIN_DEAL_STATUS_NOTFOUND {
				note := "Deal on chain status not found."
				logger.Info(note)
				swanClient.UpdateOfflineDealStatus(DEAL_STATUS_FAILED, note, deal.Id)
				continue
			}

			if onChainStatus != ONCHAIN_DEAL_STATUS_WAITTING {
				logger.Info("Deal is already imported, please check.")
				swanClient.UpdateOfflineDealStatus(DEAL_STATUS_FILE_IMPORTED, onChainStatus, deal.Id)
				continue
			}

			currentEpoch := utils.GetCurrentEpoch()

			if currentEpoch < 0 { //when exception occurs for the above os command
				logger.Error("Failed to get current epoch. Please check if miner is running properly.")
				time.Sleep(importInterval * time.Second)
				break
			}

			msg = fmt.Sprintf("Current epoch: %s. Deal starting epoch: %d", currentEpoch, deal.StartEpoch)

			if deal.StartEpoch - currentEpoch < expectedSealingTime {
				note := "Deal will start too soon, expired. Do not import this deal."
				logger.Info(note)
				swanClient.UpdateOfflineDealStatus(DEAL_STATUS_FAILED, note, deal.Id)
				continue
			}

			swanClient.UpdateOfflineDealStatus(DEAL_STATUS_FILE_IMPORTING, "", deal.Id)

			result := utils.LotusImportData(deal.DealCid, deal.FilePath)

			//There should be no output if everything goes well
			if len(result) > 0 {
				swanClient.UpdateOfflineDealStatus(DEAL_STATUS_FAILED, result, deal.Id)
				msg = fmt.Sprintf("Import deal failed. CID: %s. Error message: %s", deal.Id, result)
				logger.Error(msg)
				continue
			}

			swanClient.UpdateOfflineDealStatus(DEAL_STATUS_FILE_IMPORTED, "", deal.Id)
			msg = fmt.Sprintf("Deal CID %s imported.", deal.DealCid)
			logger.Info(msg)
			logger.Info("Sleeping...")
			time.Sleep(importInterval * time.Second)
		}
	}
}
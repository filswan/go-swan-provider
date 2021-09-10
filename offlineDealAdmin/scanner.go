package offlineDealAdmin

import (
	"fmt"
	"swan-miner/common/utils"
	"swan-miner/config"
	"swan-miner/logs"
	"time"
)

const DEAL_STATUS_WAITING = "ReadyForImport"
const MESSAGE_TYPE_ON_CHAIN = "ON CHAIN"
const MESSAGE_TYPE_SWAN = "SWAN"
const ONCHAIN_DEAL_STATUS_AWAITING = "StorageDealAwaitingPreCommit"

const SCAN_NUMBER = "100" //Max number of deals to be scanned at a time

type OfflineDealMessage struct {
	MessageType string
	MessageBody string
	OfflineDealCid string
}

func NewOfflineDealMessage(messageType, messageBody, offlineDealCid string) (*OfflineDealMessage){
	p := new(OfflineDealMessage)
	p.MessageType=messageType
	p.MessageBody=messageBody
	p.OfflineDealCid=offlineDealCid
	return p
}

func Scanner() {
	confMain := config.GetConfig().Main
	logger := logs.GetLogger()

	swanClient := utils.GetSwanClient()
	for {
		deals := swanClient.GetOfflineDeals(confMain.MinerFid, DEAL_STATUS_FILE_IMPORTED, SCAN_NUMBER)

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

			msg = fmt.Sprintf("Deal on chain status: %s.", onChainStatus)
			logger.Info(msg)

			if onChainStatus == ONCHAIN_DEAL_STATUS_ERROR {
				swanClient.UpdateOfflineDealStatus(DEAL_STATUS_FAILED, onChainMessage, deal.Id)
				msg := fmt.Sprintf("Setting deal %s status as %s", deal.DealCid, DEAL_STATUS_FAILED)
				logger.Info(msg)
			}

			if onChainStatus ==ONCHAIN_DEAL_STATUS_ACTIVE{
				note := "Deal has been completed"
				swanClient.UpdateOfflineDealStatus(DEAL_STATUS_ACTIVE, note, deal.Id)
				msg := fmt.Sprintf("Setting deal %s status as %s", deal.DealCid, DEAL_STATUS_ACTIVE)
				logger.Info(msg)
			}

			if onChainStatus == ONCHAIN_DEAL_STATUS_AWAITING {
				currentEpoch := utils.GetCurrentEpoch()
				if currentEpoch != -1 && currentEpoch > deal.StartEpoch {
					note := fmt.Sprintf("Sector is proved and active, while deal on chain status is %s. Set deal status as %s.", ONCHAIN_DEAL_STATUS_AWAITING, DEAL_STATUS_FAILED)
					swanClient.UpdateOfflineDealStatus(DEAL_STATUS_FAILED, note, deal.Id)
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


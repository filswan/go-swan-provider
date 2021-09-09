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

func getCurrentEpoch() int {
	result,err := utils.ExecOsCmd("lotus-miner", "proving info")
	currentEpoch := 1 //something get from result

	fmt.Println(result)
	if len(err) != 0 {
		logs.GetLogger().Error("Failed to get current epoch. Please check if miner is running properly.")
		return -1
	}
	return currentEpoch
}

func Scanner() {
	confMain := config.GetConfig().Main
	logger := logs.GetLogger()

	swanClient := GetSwanClient()
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
			cmd :="lotus-miner storage-deals list -v | grep " + deal.DealCid
			result, err := utils.ExecOsCmd(cmd, "")
			if len(err)>0{
				logger.Error(err)
				continue
			}

			if len(result) == 0 {
				note := "Failed to find deal on chain."
				swanClient.UpdateOfflineDealStatus(DEAL_STATUS_FAILED, note, deal.Id)
				logger.Info(note + " Deal ID: " + deal.Id)
				continue
			}

			onChainMessage := ""
			//dealStatusIndex := utils.GetFieldStrFromJson(result, "StorageDeal")
			onChainStatus := result //some value get from result
			if onChainStatus == ONCHAIN_DEAL_STATUS_ERROR {
				onChainMessage = result // some value get from result
				swanClient.UpdateOfflineDealStatus(DEAL_STATUS_FAILED, onChainMessage, deal.Id)
				msg := fmt.Sprintf("Setting deal %s status as ImportFailed", deal.DealCid)
				logger.Info(msg)
			}

			if onChainStatus ==ONCHAIN_DEAL_STATUS_ACTIVE{
				note := "Deal has been completed"
				swanClient.UpdateOfflineDealStatus(ONCHAIN_DEAL_STATUS_ACTIVE, note, deal.Id)
				msg := fmt.Sprintf("Setting deal %s status as Active", deal.DealCid)
				logger.Info(msg)
			}

			if onChainStatus == ONCHAIN_DEAL_STATUS_AWAITING {
				currentEpoch := getCurrentEpoch()
				if currentEpoch != -1 && currentEpoch > deal.StartEpoch {
					note := "Sector is proved and active, while deal on chain status is StorageDealAwaitingPreCommit. Set deal status as ImportFailed."
					swanClient.UpdateOfflineDealStatus(DEAL_STATUS_FAILED, note, deal.Id)
					msg := fmt.Sprintf("Setting deal %s status as ImportFailed due to on chain status bug.", deal.DealCid)
					logger.Info(msg)
				}
			}

			message:= fmt.Sprintf("{\"on_chain_status\": %s,\"on_chain_message\": %s}", onChainStatus, onChainMessage)
			msg = fmt.Sprintf("On chain offline_deal message created. Message Body: %s.", message)
			logger.Info(msg)
		}

		logger.Info("Sleeping...")
		time.Sleep(confMain.ScanInterval * time.Second)
	}
}


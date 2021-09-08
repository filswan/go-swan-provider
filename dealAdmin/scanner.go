package dealAdmin

import (
	"fmt"
	"reflect"
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
	result,ok := utils.ExecOsCmd("lotus-miner", "proving info")
	currentEpoch := 1 //something get from result

	fmt.Println(result)
	if !ok {
		logs.GetLogger().Error("Failed to get current epoch. Please check if miner is running properly.")
		return -1
	}
	return currentEpoch
}

func Scanner() {
	confMain := config.GetConfig().Main
	scanInterval := confMain.ScanInterval
	minerFid := confMain.MinerFid

	swanClient := GetSwanClient()
	for {
		deals := swanClient.GetOfflineDeals(minerFid, DEAL_STATUS_FILE_IMPORTED, SCAN_NUMBER)

		if len(deals)==0{
			logs.GetLogger().Info("No ongoing offline deals found.")
			logs.GetLogger().Info("Sleeping...")
			time.Sleep(time.Duration(scanInterval) * time.Second)
			continue
		}

		switch reflect.TypeOf(deals).Kind() {
		case reflect.Slice:
			dealArr := reflect.ValueOf(deals)

			for i := 0; i < dealArr.Len(); i++ {
				deal := dealArr.Index(i).String()
				fmt.Println(deal)
				dealId := utils.GetFieldStrFromJson(deal, "id")
				dealCid := utils.GetFieldStrFromJson(deal, "deal_cid")
				dealStatus := utils.GetFieldStrFromJson(deal, "status")
				msg := fmt.Sprintf("ID: %s. Deal CID: %s. Deal Status: %s.", dealId, dealCid, dealStatus)
				logs.GetLogger().Info(msg)
				cmd :="lotus-miner storage-deals list -v | grep " + dealCid
				result,_ := utils.ExecOsCmd(cmd, "")
				if result == ""{
					note := "Failed to find deal on chain."
					swanClient.UpdateOfflineDealDetails(DEAL_STATUS_FAILED, note, dealId, "","")
					msg := fmt.Sprintf("Deal details: %s", result)
					logs.GetLogger().Info(msg)
					logs.GetLogger().Info()
				}

				onChainMessage := ""
				//dealStatusIndex := utils.GetFieldStrFromJson(result, "StorageDeal")
				onChainStatus := result //some value get from result
				if onChainStatus == ONCHAIN_DEAL_STATUS_ERROR {
					onChainMessage= result // some value get from result
					swanClient.UpdateOfflineDealDetails(DEAL_STATUS_FAILED, onChainMessage, dealId, "","")
					msg := fmt.Sprintf("Setting deal %s status as ImportFailed", dealCid)
					logs.GetLogger().Info(msg)
				}

				if onChainStatus ==ONCHAIN_DEAL_STATUS_ACTIVE{
					note:="Deal has been completed"
					swanClient.UpdateOfflineDealDetails(ONCHAIN_DEAL_STATUS_ACTIVE, note, dealId, "","")
					msg := fmt.Sprintf("Setting deal %s status as Active", dealCid)
					logs.GetLogger().Info(msg)
				}

				if onChainStatus == ONCHAIN_DEAL_STATUS_AWAITING {
					currentEpoch := getCurrentEpoch()
					startEpoch := utils.GetFieldFromJson(deal, "start_epoch").(int)
					if currentEpoch != -1 && currentEpoch > startEpoch{
						note := "Sector is proved and active, while deal on chain status is StorageDealAwaitingPreCommit. Set deal status as ImportFailed."
						swanClient.UpdateOfflineDealDetails(DEAL_STATUS_FAILED, note, dealId, "","")
						msg := fmt.Sprintf("Setting deal %s status as ImportFailed due to on chain status bug.", dealCid)
						logs.GetLogger().Info(msg)
						message:= fmt.Sprintf("{\"on_chain_status\": %s,\"on_chain_message\": %s}", onChainStatus, onChainMessage)
						msg = fmt.Sprintf("On chain offline_deal message created. Message Body: %s.", message)
						logs.GetLogger().Info(msg)
						continue
					}
				}

				logs.GetLogger().Info("Sleeping...")
				time.Sleep(time.Duration(scanInterval) * time.Second)
			}
		}
	}
}


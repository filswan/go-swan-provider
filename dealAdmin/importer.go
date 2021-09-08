package dealAdmin

import (
	"fmt"
	"reflect"
	"strings"
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


func getDealOnChainStatus(dealCid string) (string) {
	cmd := "lotus-miner storage-deals list -v | grep " + dealCid
	result,_ := utils.ExecOsCmd(cmd,"")

	return result
}

func Importer() {
	conf:=config.GetConfig()
	confMain:=conf.Main

	importInterval := confMain.ImportInterval
	expectedSealingTime := confMain.ExpectedSealingTime
	minerFid := confMain.MinerFid

	swanClient := GetJwtToken()

	for {
		deals := swanClient.GetOfflineDeals(minerFid,DEAL_STATUS_READY, IMPORT_NUMNBER)
		if deals==nil{
			logs.GetLogger().Info("No pending offline deals found.")
			logs.GetLogger().Info("Sleeping...")
			continue
		}

		switch reflect.TypeOf(deals).Kind() {
		case reflect.Slice:
			dealArr := reflect.ValueOf(deals)

			for i := 0; i < dealArr.Len(); i++ {
				deal := dealArr.Index(i).String()
				fmt.Println(deal)
				dealCid := utils.GetFieldStrFromJson(deal,"deal_cid")
				filePath := utils.GetFieldStrFromJson(deal, "file_path")

				msg := fmt.Sprintf("Deal CID: %s. File Path: %s", dealCid, filePath)
				logs.GetLogger().Error(msg)

				onChainStatus := getDealOnChainStatus(dealCid)
				if !strings.HasPrefix(onChainStatus,"StorageDeal") {
					logs.GetLogger().Error(onChainStatus)
					logs.GetLogger().Error("Failed to get deal on chain status, please check if lotus-miner is running properly.")
					logs.GetLogger().Info("Sleeping...")

					time.Sleep(time.Duration(importInterval) * time.Second)
					break
				}

				if !strings.HasPrefix(onChainStatus, "StorageDeal"){
					logs.GetLogger().Error(onChainStatus)
					logs.GetLogger().Error("Failed to get deal on chain status, please check if lotus-miner is running properly.")
					logs.GetLogger().Info("Sleeping...")

					time.Sleep(time.Duration(importInterval) * time.Second)
					break
				}

				msg = fmt.Sprintf("Deal on chain status: %s.", onChainStatus)
				logs.GetLogger().Info(msg)

				if onChainStatus == ONCHAIN_DEAL_STATUS_ERROR{
					logs.GetLogger().Info("Deal on chain status is error before importing.")
					note := "Deal error before importing."
					dealId := utils.GetFieldStrFromJson(deal, "id")
					swanClient.UpdateOfflineDealDetails(DEAL_STATUS_FAILED, note, dealId, "","")
					continue
				}

				if onChainStatus == ONCHAIN_DEAL_STATUS_ACTIVE {
					logs.GetLogger().Info("Deal on chain status is active before importing.")
					note := "Deal active before importing."
					dealId := utils.GetFieldStrFromJson(deal, "id")
					swanClient.UpdateOfflineDealDetails(DEAL_STATUS_FAILED, note, dealId, "","")
					continue
				}

				if onChainStatus == ONCHAIN_DEAL_STATUS_ACCEPT {
					logs.GetLogger().Info("Deal on chain status is StorageDealAcceptWait. Deal will be ready shortly.")
					continue
				}

				if onChainStatus == ONCHAIN_DEAL_STATUS_NOTFOUND {
					logs.GetLogger().Info("Deal on chain status not found.")
					note := "Deal not found."
					dealId := utils.GetFieldStrFromJson(deal, "id")
					swanClient.UpdateOfflineDealDetails(DEAL_STATUS_FAILED, note, dealId, "","")
					continue
				}

				if onChainStatus != ONCHAIN_DEAL_STATUS_WAITTING {
					logs.GetLogger().Info("Deal is already imported, please check.")
					note := onChainStatus
					dealId := utils.GetFieldStrFromJson(deal, "id")
					swanClient.UpdateOfflineDealDetails(DEAL_STATUS_FAILED, note, dealId, "","")
					continue
				}

				result,_ := utils.ExecOsCmd("lotus-miner", " proving info")
				currentEpoch := 1 //something get from result

				if currentEpoch<0{
					logs.GetLogger().Error("Failed to get current epoch. Please check if miner is running properly.")

					time.Sleep(time.Duration(importInterval) * time.Second)
					break
				}

				startEpoch := utils.GetFieldFromJson(deal, "start_epoch").(int)
				msg = fmt.Sprintf("Current epoch: %s. Deal starting epoch: %s", currentEpoch)

				dealId := utils.GetFieldStrFromJson(deal, "id")
				if startEpoch-currentEpoch<expectedSealingTime{
					logs.GetLogger().Info("Deal will start too soon. Do not import this deal.")
					note := "Deal expired."
					swanClient.UpdateOfflineDealDetails(DEAL_STATUS_FAILED, note, dealId, "","")
					continue
				}

				command := "lotus-miner storage-deals import-data " + dealId + " " + filePath
				logs.GetLogger().Info("Command: "+command)
				swanClient.UpdateOfflineDealDetails(DEAL_STATUS_FAILED, "note", dealId, "","")

				result,_ = utils.ExecOsCmd(command,"")

				if result==""{
					swanClient.UpdateOfflineDealDetails(DEAL_STATUS_FAILED, "note", dealId, "","")
					msg = fmt.Sprintf("Import deal failed. CID: %s. Error message: %s", dealId,result)
					logs.GetLogger().Error()
					continue
				}

				swanClient.UpdateOfflineDealDetails(DEAL_STATUS_FAILED, "note", dealId, "","")
				msg = fmt.Sprintf("Deal CID %s imported.", dealId)
				logs.GetLogger().Info("Sleeping...")
				time.Sleep(time.Duration(importInterval) * time.Second)
			}
		}
	}
}
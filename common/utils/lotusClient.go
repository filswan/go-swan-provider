package utils

import (
	"strings"
	"swan-miner/logs"
)

func GetDealOnChainStatus(dealCid string) (string, string){
	cmd := "lotus-miner storage-deals list -v | grep " + dealCid
	result, err := ExecOsCmd(cmd)

	if err != nil {
		logs.GetLogger().Error(err)
		return "", ""
	}

	if len(result) == 0 {
		logs.GetLogger().Error("Failed to get deal on chain status, please check if lotus-miner is running properly.")
		logs.GetLogger().Error("Deal does not found on chain. DealCid:" + dealCid)
		return "", ""
	}

	words := strings.Fields(result)
	status := ""
	for _, word := range words {
		if strings.HasPrefix(word,"StorageDeal") {
			status = word
			break
		}
	}

	if len(status) == 0 {
		return "", ""
	}

	message := ""

	for i :=11; i < len(words); i++ {
		message = message + words[i] + " "
	}

	return status, message
}

func GetCurrentEpoch() (int) {
	cmd := "lotus-miner proving info | grep 'Current Epoch'"
	result, err := ExecOsCmd(cmd)

	if err != nil {
		logs.GetLogger().Error(err)
		return -1
	}

	if len(result) == 0 {
		logs.GetLogger().Error("Failed to get current epoch. Please check if miner is running properly.")
		return -1
	}

	words := strings.Fields(result)
	currentEpoch := GetInt64FromStr(words[1])

	return int(currentEpoch)
}

func LotusImportData(dealCid string, filepath string) (string) {
	cmd := "lotus-miner storage-deals import-data " + dealCid + " " + filepath
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd)

	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}

	return result
}

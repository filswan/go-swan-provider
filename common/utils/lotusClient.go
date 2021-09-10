package utils

import (
	"strconv"
	"strings"
	"swan-miner/logs"
)

func GetDealOnChainStatus(dealCid string) (string, string){
	logger := logs.GetLogger()
	cmd := "lotus-miner storage-deals list -v | grep " + dealCid
	result, err := ExecOsCmd(cmd)

	if len(err) > 0 {
		logger.Error(err)
		return "", ""
	}

	if len(result) == 0 {
		logger.Error("Failed to get deal on chain status, please check if lotus-miner is running properly.")
		logger.Error("Deal does not exist on chain. DealCid:"+dealCid)
		return "", ""
	}

	words := strings.Fields(result)
	var status string
	for _, word := range words {
		status = strings.Trim(word, " ")
		if strings.HasPrefix(word,"StorageDeal") {
			break
		}
	}

	if len(status) == 0 {
		return "", ""
	}

	message := ""

	for i :=11; i < len(words); i++ {
		message = message + words[i]
	}

	return status, message
}

func GetCurrentEpoch() (int) {
	cmd := "lotus-miner proving info | grep 'Current Epoch'"
	result, err := ExecOsCmd(cmd)

	if len(err) > 0 {
		logs.GetLogger().Error(err)
		return -1
	}

	if len(result) == 0 {
		logs.GetLogger().Error("Failed to get current epoch. Please check if miner is running properly.")
		return -1
	}

	words := strings.Split(result, ":")
	currentEpoch, err1 := strconv.ParseInt(words[1], 10, 64)
	if err1 != nil {
		logs.GetLogger().Error(err1.Error())
		return -1
	}

	return int(currentEpoch)
}

func LotusImportData(dealCid string, filepath string) (string) {
	cmd := "lotus-miner storage-deals import-data " + dealCid + " " + filepath
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd)

	if len(err) > 0 {
		logs.GetLogger().Error(err)
		return ""
	}

	return result
}

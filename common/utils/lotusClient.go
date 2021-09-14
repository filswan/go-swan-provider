package utils

import (
	"regexp"
	"strings"
	"swan-provider/logs"
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
		logs.GetLogger().Error("Deal does not found on chain. DealCid:", dealCid)
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

	message = strings.TrimRight(message, " ")
	return status, message
}

func GetCurrentEpoch() (int) {
	cmd := "lotus-miner proving info | grep 'Current Epoch'"
	logs.GetLogger().Info(cmd)
	result, err := ExecOsCmd(cmd)

	if err != nil {
		logs.GetLogger().Error(err)
		return -1
	}

	if len(result) == 0 {
		logs.GetLogger().Error("Failed to get current epoch. Please check if miner is running properly.")
		return -1
	}

	logs.GetLogger().Info(result)

	re := regexp.MustCompile("[0-9]+")
	words := re.FindAllString(result, -1)
	logs.GetLogger().Info("words:",words)
	var currentEpoch int64 = -1
	if words != nil && len(words) > 0 {
		currentEpoch = GetInt64FromStr(words[0])
	}

	logs.GetLogger().Info("currentEpoch: ", currentEpoch)
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

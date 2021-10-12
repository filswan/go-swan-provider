package client

import (
	"regexp"
	"strings"
	"swan-provider/common/utils"
	"swan-provider/logs"
)

const (
	JSON_RPC_VERSION = "2.0"

	//StorageDealStaged  29
)

type JsonRpcParams struct {
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type GetDealInfoParam struct {
	DealCid string `json:"/"`
}

type LotusClient struct {
	ApiUrl string
	Token  string
}

func GenJsonRpcParams(dealCid string) JsonRpcParams {
	var params []interface{}
	getDealInfoParam := GetDealInfoParam{
		DealCid: dealCid,
	}
	params = append(params, getDealInfoParam)

	jsonRpcParams := JsonRpcParams{
		JsonRpc: JSON_RPC_VERSION,
		Method:  "Filecoin.ClientGetDealInfo",
		Params:  params,
		Id:      7878,
	}

	return jsonRpcParams
}

func LotusGetDealOnChainStatus1(dealCid string, token string) (string, string) {
	url := "http://192.168.88.41:1234/rpc/v0"

	params := GenJsonRpcParams(dealCid)
	response := utils.HttpPostNoToken(url, params)

	logs.GetLogger().Info(response)

	result := utils.GetFieldMapFromJson(response, "result")
	if result == nil {
		logs.GetLogger().Error("Failed to get result from:", url)
		return "", ""
	}
	state := result["State"]
	if state == nil {
		logs.GetLogger().Error("Failed to get state from:", url)
		return "", ""
	}
	message := result["Message"]
	if message == nil {
		logs.GetLogger().Error("Failed to get message from:", url)
		return "", ""
	}

	logs.GetLogger().Info(state)
	logs.GetLogger().Info(message)

	return state.(string), message.(string)
}

func GetDealOnChainStatus(dealCid string) (string, string) {
	cmd := "lotus-miner storage-deals list -v | grep -a " + dealCid
	result, err := ExecOsCmd(cmd)

	if err != nil {
		logs.GetLogger().Error("Failed to get deal on chain status, please check if lotus-miner is running properly.")
		logs.GetLogger().Error(err)
		return "", ""
	}

	if len(result) == 0 {
		logs.GetLogger().Error("Deal does not found on chain. DealCid:", dealCid)
		return "", ""
	}

	words := strings.Fields(result)
	status := ""
	for _, word := range words {
		if strings.HasPrefix(word, "StorageDeal") {
			status = word
			break
		}
	}

	if len(status) == 0 {
		return "", ""
	}

	message := ""

	for i := 11; i < len(words); i++ {
		message = message + words[i] + " "
	}

	message = strings.TrimRight(message, " ")
	return status, message
}

func GetCurrentEpoch() int {
	cmd := "lotus-miner proving info | grep 'Current Epoch'"
	logs.GetLogger().Info(cmd)
	result, err := utils.ExecOsCmd(cmd)

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
	logs.GetLogger().Info("words:", words)
	var currentEpoch int64 = -1
	if len(words) > 0 {
		currentEpoch = utils.GetInt64FromStr(words[0])
	}

	logs.GetLogger().Info("currentEpoch: ", currentEpoch)
	return int(currentEpoch)
}

func LotusImportData(dealCid string, filepath string) string {
	cmd := "lotus-miner storage-deals import-data " + dealCid + " " + filepath
	logs.GetLogger().Info(cmd)

	result, err := utils.ExecOsCmd(cmd)

	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}

	return result
}

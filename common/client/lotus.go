package client

import (
	"regexp"
	"swan-provider/common/utils"
	"swan-provider/logs"
)

const (
	JSON_RPC_VERSION       = "2.0"
	CLIENT_GET_DEAL_INFO   = "Filecoin.ClientGetDealInfo"
	CLIENT_GET_DEAL_STATUS = "Filecoin.ClientGetDealStatus"

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

func GenJsonRpcParams4ClientGetDealInfo(dealCid string) JsonRpcParams {
	var params []interface{}
	getDealInfoParam := GetDealInfoParam{
		DealCid: dealCid,
	}
	params = append(params, getDealInfoParam)

	jsonRpcParams := JsonRpcParams{
		JsonRpc: JSON_RPC_VERSION,
		Method:  CLIENT_GET_DEAL_INFO,
		Params:  params,
		Id:      7878,
	}

	return jsonRpcParams
}

func GenJsonRpcParams4ClientGetDealStatus(state int) JsonRpcParams {
	var params []interface{}
	params = append(params, state)

	jsonRpcParams := JsonRpcParams{
		JsonRpc: JSON_RPC_VERSION,
		Method:  CLIENT_GET_DEAL_STATUS,
		Params:  params,
		Id:      7878,
	}

	return jsonRpcParams
}

//"lotus-miner storage-deals list -v | grep -a " + dealCid
func (lotusClient *LotusClient) LotusClientGetDealStatus(state int) string {
	params := GenJsonRpcParams4ClientGetDealStatus(state)
	response := utils.HttpPostNoToken(lotusClient.ApiUrl, params)

	logs.GetLogger().Info(response)

	result := utils.GetFieldStrFromJson(response, "result")
	if result == "" {
		logs.GetLogger().Error("Failed to get result from:", lotusClient.ApiUrl)
		return ""
	}

	return result
}

//"lotus-miner storage-deals list -v | grep -a " + dealCid
func (lotusClient *LotusClient) LotusGetDealOnChainStatus(dealCid string, token string) (string, string) {
	params := GenJsonRpcParams4ClientGetDealInfo(dealCid)
	response := utils.HttpPostNoToken(lotusClient.ApiUrl, params)

	logs.GetLogger().Info(response)

	result := utils.GetFieldMapFromJson(response, "result")
	if result == nil {
		logs.GetLogger().Error("Failed to get result from:", lotusClient.ApiUrl)
		return "", ""
	}
	state := result["State"]
	if state == nil {
		logs.GetLogger().Error("Failed to get state from:", lotusClient.ApiUrl)
		return "", ""
	}
	message := result["Message"]
	if message == nil {
		logs.GetLogger().Error("Failed to get message from:", lotusClient.ApiUrl)
		return "", ""
	}

	status := lotusClient.LotusClientGetDealStatus(state.(int))

	logs.GetLogger().Info(status)
	logs.GetLogger().Info(message)

	return status, message.(string)
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

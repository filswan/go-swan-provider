package client

import (
	"swan-provider/common/utils"
	"swan-provider/config"
	"swan-provider/logs"
)

const (
	LOTUS_JSON_RPC_ID            = 7878
	LOTUS_JSON_RPC_VERSION       = "2.0"
	LOTUS_CLIENT_GET_DEAL_INFO   = "Filecoin.ClientGetDealInfo"
	LOTUS_CLIENT_GET_DEAL_STATUS = "Filecoin.ClientGetDealStatus"
	LOTUS_CHAIN_HEAD             = "Filecoin.ChainHead"
)

type JsonRpcParams struct {
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type LotusGetDealInfoParam struct {
	DealCid string `json:"/"`
}

type LotusClient struct {
	ApiUrl      string
	AccessToken string
}

func LotusGetClient() *LotusClient {
	lotusClient := &LotusClient{
		ApiUrl:      config.GetConfig().Lotus.ApiUrl,
		AccessToken: config.GetConfig().Lotus.AccessToken,
	}

	return lotusClient
}

//"lotus-miner storage-deals list -v | grep -a " + dealCid
func (lotusClient *LotusClient) LotusClientGetDealStatus(state int) string {
	var params []interface{}
	params = append(params, state)

	jsonRpcParams := JsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_GET_DEAL_STATUS,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := utils.HttpPost(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)

	//logs.GetLogger().Info(response)

	result := utils.GetFieldStrFromJson(response, "result")
	if result == "" {
		logs.GetLogger().Error("Failed to get result from:", lotusClient.ApiUrl)
		return ""
	}

	return result
}

//"lotus-miner storage-deals list -v | grep -a " + dealCid
func (lotusClient *LotusClient) LotusGetDealOnChainStatus(dealCid string) (string, string) {
	var params []interface{}
	getDealInfoParam := LotusGetDealInfoParam{DealCid: dealCid}
	params = append(params, getDealInfoParam)

	jsonRpcParams := JsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_GET_DEAL_INFO,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := utils.HttpPost(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)

	//logs.GetLogger().Info(response)

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

	stateInt := int(state.(float64))

	status := lotusClient.LotusClientGetDealStatus(stateInt)

	logs.GetLogger().Info(status)
	logs.GetLogger().Info(message)

	return status, message.(string)
}

func (lotusClient *LotusClient) GetCurrentEpoch() int {
	var params []interface{}

	jsonRpcParams := JsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CHAIN_HEAD,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := utils.HttpPost(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)

	//logs.GetLogger().Info(response)

	result := utils.GetFieldMapFromJson(response, "result")
	if result == nil {
		logs.GetLogger().Error("Failed to get result from:", lotusClient.ApiUrl)
		return -1
	}

	height := result["Height"]
	if height == nil {
		logs.GetLogger().Error("Failed to get height from:", lotusClient.ApiUrl)
		return -1
	}

	heightFloat := height.(float64)
	return int(heightFloat)
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

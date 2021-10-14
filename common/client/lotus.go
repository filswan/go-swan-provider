package client

import (
	"encoding/json"
	"fmt"
	"swan-provider/common/utils"
	"swan-provider/config"
	"swan-provider/logs"
)

const (
	LOTUS_JSON_RPC_ID                  = 7878
	LOTUS_JSON_RPC_VERSION             = "2.0"
	LOTUS_CLIENT_GET_DEAL_INFO         = "Filecoin.ClientGetDealInfo"
	LOTUS_CLIENT_GET_DEAL_STATUS       = "Filecoin.ClientGetDealStatus"
	LOTUS_CHAIN_HEAD                   = "Filecoin.ChainHead"
	LOTUS_MARKET_IMPORT_DATA           = "Filecoin.MarketImportDealData"
	LOTUS_MARKET_LIST_INCOMPLETE_DEALS = "Filecoin.MarketListIncompleteDeals"
)

type DealCid struct {
	DealCid string `json:"/"`
}

type LotusClient struct {
	ApiUrl           string
	MinerApiUrl      string
	MinerAccessToken string
}

type MarketListIncompleteDeals struct {
	Id      int           `json:"id"`
	JsonRpc string        `json:"jsonrpc"`
	Result  []Deal        `json:"result"`
	Error   *JsonRpcError `json:"error"`
}

type JsonRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Deal struct {
	State       int     `json:"State"`
	Message     string  `json:"Message"`
	ProposalCid DealCid `json:"ProposalCid"`
}

func LotusGetClient() *LotusClient {
	lotusClient := &LotusClient{
		ApiUrl:           config.GetConfig().Lotus.ApiUrl,
		MinerApiUrl:      config.GetConfig().Lotus.MinerApiUrl,
		MinerAccessToken: config.GetConfig().Lotus.MinerAccessToken,
	}

	return lotusClient
}

//"lotus-miner storage-deals list -v | grep -a " + dealCid
func LotusGetDealStatus(state int) string {
	lotusClient := LotusGetClient()

	var params []interface{}
	params = append(params, state)

	jsonRpcParams := JsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_GET_DEAL_STATUS,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := HttpPostNoToken(lotusClient.ApiUrl, jsonRpcParams)

	//logs.GetLogger().Info(response)

	result := utils.GetFieldStrFromJson(response, "result")
	if result == "" {
		logs.GetLogger().Error("Failed to get result from:", lotusClient.ApiUrl)
		return ""
	}

	return result
}

func LotusGetDeals() []Deal {
	lotusClient := LotusGetClient()

	var params []interface{}
	jsonRpcParams := JsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_MARKET_LIST_INCOMPLETE_DEALS,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	logs.GetLogger().Info("Get deal list from ", lotusClient.MinerApiUrl)
	response := HttpGet(lotusClient.MinerApiUrl, lotusClient.MinerAccessToken, jsonRpcParams)
	logs.GetLogger().Info("Get deal list got from ", lotusClient.MinerApiUrl)
	deals := &MarketListIncompleteDeals{}
	err := json.Unmarshal([]byte(response), deals)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return deals.Result
}

func LotusGetDealOnChainStatusFromDeals(deals []Deal, dealCid string) (string, string) {
	if len(deals) == 0 {
		logs.GetLogger().Error("Deal list is empty.")
		return "", ""
	}

	for _, deal := range deals {
		if deal.ProposalCid.DealCid != dealCid {
			continue
		}
		status := LotusGetDealStatus(deal.State)
		logs.GetLogger().Info("deal: ", dealCid, " status:", status, " message:", deal.Message)
		return status, deal.Message
	}

	logs.GetLogger().Error("Did not find your deal:", dealCid, " in the returned list.")

	return "", ""
}

//"lotus-miner storage-deals list -v | grep -a " + dealCid
func LotusGetDealOnChainStatus(dealCid string) (string, string) {
	deals := LotusGetDeals()
	status, message := LotusGetDealOnChainStatusFromDeals(deals, dealCid)
	return status, message
}

func LotusGetCurrentEpoch() int {
	lotusClient := LotusGetClient()

	var params []interface{}

	jsonRpcParams := JsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CHAIN_HEAD,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := HttpPostNoToken(lotusClient.ApiUrl, jsonRpcParams)

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
	lotusClient := LotusGetClient()

	var params []interface{}
	getDealInfoParam := DealCid{DealCid: dealCid}
	params = append(params, getDealInfoParam)
	params = append(params, filepath)

	jsonRpcParams := JsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_MARKET_IMPORT_DATA,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := HttpPost(lotusClient.MinerApiUrl, lotusClient.MinerAccessToken, jsonRpcParams)
	if response == "" {
		msg := "no return"
		logs.GetLogger().Error(msg)
		return msg
	}
	logs.GetLogger().Info(response)

	errorInfo := utils.GetFieldMapFromJson(response, "error")

	if errorInfo == nil {
		return ""
	}

	logs.GetLogger().Error(errorInfo)
	errCode := int(errorInfo["code"].(float64))
	errMsg := errorInfo["message"].(string)
	msg := fmt.Sprintf("Error code:%d message:%s", errCode, errMsg)
	logs.GetLogger().Error(msg)
	return msg
}

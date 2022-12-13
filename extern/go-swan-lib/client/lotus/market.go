package lotus

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
)

const (
	LOTUS_MARKET_GET_ASK               = "Filecoin.MarketGetAsk"
	LOTUS_MARKET_IMPORT_DATA           = "Filecoin.MarketImportDealData"
	LOTUS_MARKET_LIST_INCOMPLETE_DEALS = "Filecoin.MarketListIncompleteDeals"
)

type LotusMarket struct {
	ApiUrl       string
	AccessToken  string
	ClientApiUrl string
}

type MarketGetAsk struct {
	LotusJsonRpcResult
	Result *struct {
		Ask MarketGetAskResultAsk
	} `json:"result"`
}

type MarketGetAskResultAsk struct {
	Price         string
	VerifiedPrice string
	MinPieceSize  int
	MaxPieceSize  int
	Miner         string
	Timestamp     int
	Expiry        int
	SeqNo         int
}

func GetLotusMarket(apiUrl, accessToken, clientApiUrl string) (*LotusMarket, error) {
	if len(apiUrl) == 0 {
		err := fmt.Errorf("lotus market api url is required")
		logs.GetLogger().Error(err)
		return nil, err
	}

	lotusMarket := &LotusMarket{
		ApiUrl:       apiUrl,
		AccessToken:  accessToken,
		ClientApiUrl: clientApiUrl,
	}

	return lotusMarket, nil
}

//"lotus client query-ask " + minerFid
func (lotusMarket *LotusMarket) LotusMarketGetAsk() (*MarketGetAskResultAsk, error) {
	var params []interface{}

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_MARKET_GET_ASK,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	//here the api url should be miner's api url, need to change later on
	response, err := web.HttpGetNoToken(lotusMarket.ApiUrl, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	marketGetAsk := &MarketGetAsk{}
	err = json.Unmarshal(response, marketGetAsk)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if marketGetAsk.Error != nil {
		err := fmt.Errorf("%d,%s", marketGetAsk.Error.Code, marketGetAsk.Error.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return &marketGetAsk.Result.Ask, nil
}

type DealCid struct {
	DealCid string `json:"/"`
}

type MarketListIncompleteDeals struct {
	Id      int           `json:"id"`
	JsonRpc string        `json:"jsonrpc"`
	Result  []Deal        `json:"result"`
	Error   *JsonRpcError `json:"error"`
}

type Deal struct {
	State       int     `json:"State"`
	Message     string  `json:"Message"`
	DealID      uint64  `json:"DealID"`
	ProposalCid DealCid `json:"ProposalCid"`
	Proposal    struct {
		Client   string `json:"Client"`
		Provider string `json:"Provider"`
	} `json:"Proposal"`
}

func (lotusMarket *LotusMarket) LotusGetDeals() ([]Deal, error) {
	var params []interface{}
	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_MARKET_LIST_INCOMPLETE_DEALS,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response, err := web.HttpGet(lotusMarket.ApiUrl, lotusMarket.AccessToken, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	deals := &MarketListIncompleteDeals{}
	err = json.Unmarshal(response, deals)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return deals.Result, nil
}

func (lotusMarket *LotusMarket) LotusGetDealOnChainStatusFromDeals(deals []Deal, dealCid string) (string, uint64, *string, *string, error) {
	if len(deals) == 0 {
		err := fmt.Errorf("Deal list is empty, maybe you initialized a new datastore. you have lost the deal, dealCid: %s", dealCid)
		logs.GetLogger().Error(err)
		return "", 0, nil, nil, err
	}

	lotusClient, err := LotusGetClient(lotusMarket.ClientApiUrl, "")
	if err != nil {
		logs.GetLogger().Error(err)
		return "", 0, nil, nil, err
	}
	for _, deal := range deals {
		if deal.ProposalCid.DealCid != dealCid {
			continue
		}

		status, err := lotusClient.LotusGetDealStatus(deal.State)
		if err != nil {
			logs.GetLogger().Error(err)
			return "", 0, nil, nil, err
		}

		return deal.Proposal.Provider, deal.DealID, status, &deal.Message, nil
	}

	return "", 0, nil, nil, nil
}

//"lotus-miner storage-deals list -v | grep -a " + dealCid
func (lotusMarket *LotusMarket) LotusGetDealOnChainStatus(dealCid string) (string, uint64, *string, *string, error) {
	deals, err := lotusMarket.LotusGetDeals()
	if err != nil {
		logs.GetLogger().Error(err)
		return "", 0, nil, nil, err
	}

	minerId, dealId, status, message, err := lotusMarket.LotusGetDealOnChainStatusFromDeals(deals, dealCid)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", 0, nil, nil, err
	}
	return minerId, dealId, status, message, nil
}

func (lotusMarket *LotusMarket) LotusImportData(dealCid string, filepath string) error {
	var params []interface{}
	getDealInfoParam := DealCid{DealCid: dealCid}
	params = append(params, getDealInfoParam)
	params = append(params, filepath)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_MARKET_IMPORT_DATA,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response, err := web.HttpPost(lotusMarket.ApiUrl, lotusMarket.AccessToken, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	errorInfo := utils.GetFieldMapFromJson(response, "error")

	if errorInfo == nil {
		return nil
	}

	errCode := int(errorInfo["code"].(float64))
	errMsg := errorInfo["message"].(string)
	err = fmt.Errorf("error code:%d message:%s", errCode, errMsg)
	if strings.Contains(string(response), "(need 'write')") {
		logs.GetLogger().Error("please check your access token, it should have write access")
		logs.GetLogger().Error(err)
	}
	return err
}

package lotus

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"

	"github.com/shopspring/decimal"
)

const (
	LOTUS_CLIENT_MINER_QUERY     = "Filecoin.ClientMinerQueryOffer"
	LOTUS_CLIENT_QUERY_ASK       = "Filecoin.ClientQueryAsk"
	LOTUS_CLIENT_GET_DEAL_INFO   = "Filecoin.ClientGetDealInfo"
	LOTUS_CLIENT_GET_DEAL_STATUS = "Filecoin.ClientGetDealStatus"
	LOTUS_CHAIN_HEAD             = "Filecoin.ChainHead"
	LOTUS_CLIENT_CALC_COMM_P     = "Filecoin.ClientCalcCommP"
	LOTUS_CLIENT_IMPORT          = "Filecoin.ClientImport"
	LOTUS_CLIENT_GEN_CAR         = "Filecoin.ClientGenCar"
	LOTUS_CLIENT_START_DEAL      = "Filecoin.ClientStartDeal"
	LOTUS_STATE_STORAGE_DEAL     = "Filecoin.StateMarketStorageDeal"

	STAGE_RESERVE_FUNDS     = "StorageDealReserveClientFunds"
	STAGE_PROPOSAL_ACCEPTED = "StorageDealProposalAccepted"

	FUNDS_RESERVED = "funds reserved"
	FUNDS_RELEASED = "funds released"
)

type LotusClient struct {
	ApiUrl      string
	AccessToken string
}

type ClientCalcCommP struct {
	LotusJsonRpcResult
	Result *struct {
		Root Cid
		Size int
	} `json:"result"`
}

type ClientImport struct {
	LotusJsonRpcResult
	Result *struct {
		Root     Cid
		ImportID int64
	} `json:"result"`
}

func LotusGetClient(apiUrl, accessToken string) (*LotusClient, error) {
	if len(apiUrl) == 0 {
		err := fmt.Errorf("config lotus api_url is required")
		logs.GetLogger().Error(err)
		return nil, err
	}

	lotusClient := &LotusClient{
		ApiUrl:      apiUrl,
		AccessToken: accessToken,
	}

	return lotusClient, nil
}

type ClientMinerQuery struct {
	LotusJsonRpcResult
	Result struct {
		MinerPeer struct {
			Address string
			ID      string
		}
	} `json:"result"`
}

type ClientDealInfo struct {
	LotusJsonRpcResult
	Result struct {
		State      int
		Message    string
		DealStages struct {
			Stages []struct {
				Name             string
				Description      string
				ExpectedDuration string
				CreatedTime      string
				UpdatedTime      string
				Logs             []struct {
					Log         string
					UpdatedTime string
				}
			}
		}
		PricePerEpoch string
		Duration      int
		DealID        int64
		Verified      bool
	} `json:"result"`
}

type ClientDealCostStatus struct {
	CostComputed         string
	ReserveClientFunds   string
	DealProposalAccepted string
	Status               string
	DealId               int64
	Verified             bool
}

func (lotusClient *LotusClient) LotusClientGetDealInfo(dealCid string) (*ClientDealCostStatus, error) {
	var params []interface{}
	cid := Cid{Cid: dealCid}
	params = append(params, cid)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_GET_DEAL_INFO,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response, err := web.HttpGetNoToken(lotusClient.ApiUrl, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	clientDealInfo := &ClientDealInfo{}
	err = json.Unmarshal(response, clientDealInfo)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if clientDealInfo.Error != nil {
		err := fmt.Errorf("deal:%s,code:%d,message:%s", dealCid, clientDealInfo.Error.Code, clientDealInfo.Error.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	pricePerEpoch, err := decimal.NewFromString(clientDealInfo.Result.PricePerEpoch)
	if err != nil {
		err := fmt.Errorf("deal:%s,%s", dealCid, err.Error())
		logs.GetLogger().Error(err)
		return nil, err
	}
	duration := decimal.NewFromInt(int64(clientDealInfo.Result.Duration))

	clientDealCostStatus := ClientDealCostStatus{}
	clientDealCostStatus.CostComputed = pricePerEpoch.Mul(duration).String()

	dealStages := clientDealInfo.Result.DealStages.Stages
	for _, stage := range dealStages {
		if strings.EqualFold(stage.Name, STAGE_RESERVE_FUNDS) {
			for _, log := range stage.Logs {
				if strings.Contains(log.Log, FUNDS_RESERVED) {
					clientDealCostStatus.ReserveClientFunds = utils.GetNumStrFromStr(log.Log)
					clientDealCostStatus.ReserveClientFunds = strings.TrimSuffix(clientDealCostStatus.ReserveClientFunds, ">")
				}
			}
		}
		if strings.EqualFold(stage.Name, STAGE_PROPOSAL_ACCEPTED) {
			for _, log := range stage.Logs {
				if strings.Contains(log.Log, FUNDS_RELEASED) {
					clientDealCostStatus.DealProposalAccepted = utils.GetNumStrFromStr(log.Log)
					clientDealCostStatus.DealProposalAccepted = strings.TrimSuffix(clientDealCostStatus.DealProposalAccepted, ">")
				}
			}
		}
	}

	dealStatus, err := lotusClient.LotusGetDealStatus(clientDealInfo.Result.State)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	clientDealCostStatus.Status = *dealStatus
	clientDealCostStatus.DealId = clientDealInfo.Result.DealID
	clientDealCostStatus.Verified = clientDealInfo.Result.Verified

	return &clientDealCostStatus, nil
}

func GetDealCost(dealCost ClientDealCostStatus) string {
	if dealCost.DealProposalAccepted != "" {
		return dealCost.DealProposalAccepted
	}

	if dealCost.ReserveClientFunds != "" {
		return dealCost.ReserveClientFunds
	}

	return dealCost.CostComputed
}

func (lotusClient *LotusClient) LotusClientMinerQuery(minerFid string) (*string, error) {
	var params []interface{}
	params = append(params, minerFid)
	params = append(params, nil)
	params = append(params, nil)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_MINER_QUERY,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response, err := web.HttpGetNoToken(lotusClient.ApiUrl, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	clientMinerQuery := &ClientMinerQuery{}
	err = json.Unmarshal(response, clientMinerQuery)
	if err != nil {
		err := fmt.Errorf("miner:%s,%s", minerFid, err.Error())
		logs.GetLogger().Error(err)
		return nil, err
	}

	if clientMinerQuery.Error != nil {
		err := fmt.Errorf("miner:%s,code:%d,message:%s", minerFid, clientMinerQuery.Error.Code, clientMinerQuery.Error.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	minerPeerId := clientMinerQuery.Result.MinerPeer.ID
	return &minerPeerId, nil
}

type ClientQueryAsk struct {
	LotusJsonRpcResult
	Result struct {
		Price         string
		VerifiedPrice string
		MinPieceSize  int64
		MaxPieceSize  int64
	} `json:"result"`
}

type MinerConfig struct {
	Price         decimal.Decimal
	VerifiedPrice decimal.Decimal
	MinPieceSize  int64
	MaxPieceSize  int64
}

func (lotusClient *LotusClient) LotusClientQueryAsk(minerFid string) (*MinerConfig, error) {
	minerPeerId, err := lotusClient.LotusClientMinerQuery(minerFid)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	var params []interface{}
	params = append(params, minerPeerId)
	params = append(params, minerFid)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_QUERY_ASK,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response, err := web.HttpGetNoToken(lotusClient.ApiUrl, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	clientQueryAsk := &ClientQueryAsk{}
	err = json.Unmarshal(response, clientQueryAsk)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if clientQueryAsk.Error != nil {
		err := fmt.Errorf("miner:%s,code:%d,message:%s", minerFid, clientQueryAsk.Error.Code, clientQueryAsk.Error.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	price, err := decimal.NewFromString(clientQueryAsk.Result.Price)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	verifiedPrice, err := decimal.NewFromString(clientQueryAsk.Result.VerifiedPrice)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	minerConfig := &MinerConfig{
		Price:         price,
		VerifiedPrice: verifiedPrice,
		MinPieceSize:  clientQueryAsk.Result.MinPieceSize,
		MaxPieceSize:  clientQueryAsk.Result.MaxPieceSize,
	}

	return minerConfig, nil
}

func (lotusClient *LotusClient) LotusGetCurrentEpoch() (*int64, error) {
	var params []interface{}

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CHAIN_HEAD,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response, err := web.HttpPostNoToken(lotusClient.ApiUrl, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	result := utils.GetFieldMapFromJson(response, "result")
	if result == nil {
		err := fmt.Errorf("failed to get result from:%s", lotusClient.ApiUrl)
		logs.GetLogger().Error(err)
		return nil, err
	}

	height := result["Height"]
	if height == nil {
		err := fmt.Errorf("failed to get height from:%s", lotusClient.ApiUrl)
		logs.GetLogger().Error(err)
		return nil, err
	}

	heightFloat := height.(float64)
	heightInt64 := int64(heightFloat)
	return &heightInt64, nil
}

//"lotus-miner storage-deals list -v | grep -a " + dealCid
func (lotusClient *LotusClient) LotusGetDealStatus(state int) (*string, error) {
	var params []interface{}
	params = append(params, state)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_GET_DEAL_STATUS,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response, err := web.HttpPostNoToken(lotusClient.ApiUrl, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	result := utils.GetFieldStrFromJson(response, "result")
	if result == "" {
		logs.GetLogger().Error("no response from:", lotusClient.ApiUrl)
		return nil, err
	}

	return &result, nil
}

//"lotus client commP " + carFilePath
func (lotusClient *LotusClient) LotusClientCalcCommP(filepath string) (*string, error) {
	var params []interface{}
	params = append(params, filepath)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_CALC_COMM_P,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response, err := web.HttpPost(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	clientCalcCommP := &ClientCalcCommP{}
	err = json.Unmarshal(response, clientCalcCommP)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if clientCalcCommP.Error != nil {
		err := fmt.Errorf("get piece CID failed for:%s, error code:%d, message:%s", filepath, clientCalcCommP.Error.Code, clientCalcCommP.Error.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	if clientCalcCommP.Result == nil {
		err := fmt.Errorf("no result from:%s", lotusClient.ApiUrl)
		logs.GetLogger().Error()
		return nil, err
	}

	pieceCid := clientCalcCommP.Result.Root.Cid
	return &pieceCid, nil
}

type ClientFileParam struct {
	Path  string
	IsCAR bool
}

//"lotus client import --car " + carFilePath
func (lotusClient *LotusClient) LotusClientImport(filepath string, isCar bool) (*string, error) {
	var params []interface{}
	clientFileParam := ClientFileParam{
		Path:  filepath,
		IsCAR: isCar,
	}
	params = append(params, clientFileParam)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_IMPORT,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response, err := web.HttpGet(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	clientImport := &ClientImport{}
	err = json.Unmarshal(response, clientImport)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if clientImport.Error != nil {
		err := fmt.Errorf("lotus import file %s failed, error code:%d, message:%s", filepath, clientImport.Error.Code, clientImport.Error.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	if clientImport.Result == nil {
		err := fmt.Errorf("lotus import file %s failed, result is null from %s", filepath, lotusClient.ApiUrl)
		logs.GetLogger().Error(err)
		return nil, err
	}

	dataCid := clientImport.Result.Root.Cid

	return &dataCid, nil
}

//"lotus client generate-car " + srcFilePath + " " + destCarFilePath
func (lotusClient *LotusClient) LotusClientGenCar(srcFilePath, destCarFilePath string, srcFilePathIsCar bool) error {
	var params []interface{}
	clientFileParam := ClientFileParam{
		Path:  srcFilePath,
		IsCAR: srcFilePathIsCar,
	}
	params = append(params, clientFileParam)
	params = append(params, destCarFilePath)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_GEN_CAR,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response, err := web.HttpGet(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	//logs.GetLogger().Info(response)
	lotusJsonRpcResult := &LotusJsonRpcResult{}
	err = json.Unmarshal(response, lotusJsonRpcResult)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if lotusJsonRpcResult.Error != nil {
		err := fmt.Errorf("error, code:%d, message:%s", lotusJsonRpcResult.Error.Code, lotusJsonRpcResult.Error.Message)
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

type ClientStartDealParamData struct {
	TransferType string
	Root         Cid
	PieceCid     *Cid
	PieceSize    int
}

type ClientStartDealParam struct {
	Data              ClientStartDealParamData
	Wallet            string
	Miner             string
	EpochPrice        string
	MinBlocksDuration int
	DealStartEpoch    int64
	FastRetrieval     bool
	VerifiedDeal      bool
}

type ClientStartDeal struct {
	LotusJsonRpcResult
	Result Cid `json:"result"`
}

func (lotusClient *LotusClient) CheckDuration(duration int, startEpoch int64) error {
	if duration < constants.DURATION_MIN || duration > constants.DURATION_MAX {
		err := fmt.Errorf("deal duration out of bounds (min, max, provided): %d, %d, %d", constants.DURATION_MIN, constants.DURATION_MAX, duration)
		logs.GetLogger().Error(err)
		return err
	}

	currentEpoch, err := lotusClient.LotusGetCurrentEpoch()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	endEpoch := startEpoch + (int64)(duration)

	epoch2EndfromNow := endEpoch - *currentEpoch
	if epoch2EndfromNow >= constants.DURATION_MAX {
		err := fmt.Errorf("invalid deal end epoch %d: cannot be more than %d past current epoch %d", endEpoch, constants.DURATION_MAX, *currentEpoch)
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func (lotusClient *LotusClient) CheckDealConfig(dealConfig *model.DealConfig) (*decimal.Decimal, error) {
	if dealConfig == nil {
		err := fmt.Errorf("parameter dealConfig is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	if dealConfig.SenderWallet == "" {
		err := fmt.Errorf("wallet should be set")
		logs.GetLogger().Error(err)
		return nil, err
	}

	minerConfig, err := lotusClient.LotusClientQueryAsk(dealConfig.MinerFid)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	var e18 decimal.Decimal = decimal.NewFromFloat(constants.LOTUS_PRICE_MULTIPLE_1E18)

	var minerPrice decimal.Decimal
	if dealConfig.VerifiedDeal {
		minerPrice = minerConfig.VerifiedPrice.Div(e18)
	} else {
		minerPrice = minerConfig.Price.Div(e18)
	}
	logs.GetLogger().Info("miner:", dealConfig.MinerFid, ",price:", minerPrice)

	priceCmp := dealConfig.MaxPrice.Cmp(minerPrice)
	if priceCmp < 0 {
		err := fmt.Errorf("miner price:%s > deal max price:%s", minerPrice.String(), dealConfig.MaxPrice.String())
		logs.GetLogger().Error(err)
		return nil, err
	}

	if dealConfig.Duration == 0 {
		dealConfig.Duration = constants.DURATION_DEFAULT
	}

	err = lotusClient.CheckDuration(dealConfig.Duration, dealConfig.StartEpoch)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return &minerPrice, nil
}

func (lotusClient *LotusClient) LotusClientStartDeal(dealConfig *model.DealConfig) (*string, error) {
	minerPrice, err := lotusClient.CheckDealConfig(dealConfig)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	pieceSize, sectorSize := utils.CalculatePieceSize(dealConfig.FileSize)
	cost := utils.CalculateRealCost(sectorSize, *minerPrice)

	epochPrice := cost.Mul(decimal.NewFromFloat(constants.LOTUS_PRICE_MULTIPLE_1E18))

	if !dealConfig.SkipConfirmation {
		logs.GetLogger().Info("Do you confirm to submit the deal?")
		logs.GetLogger().Info("Press Y/y to continue, other key to quit")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		response = strings.TrimRight(response, "\n")

		if !strings.EqualFold(response, "Y") {
			logs.GetLogger().Info("Your input is ", response, ". Now give up submit the deal.")
			return nil, nil
		}
	}

	clientStartDealParamData := ClientStartDealParamData{
		TransferType: dealConfig.TransferType, //constants.LOTUS_TRANSFER_TYPE_MANUAL,
		Root: Cid{
			Cid: dealConfig.PayloadCid,
		},
		PieceCid:  nil,
		PieceSize: int(pieceSize),
	}

	dealConfig.PieceCid = strings.Trim(dealConfig.PieceCid, " ")
	if dealConfig.PieceCid != "" {
		clientStartDealParamData.PieceCid = &Cid{
			Cid: dealConfig.PieceCid,
		}
	}

	clientStartDealParam := ClientStartDealParam{
		Data:              clientStartDealParamData,
		Wallet:            dealConfig.SenderWallet,
		Miner:             dealConfig.MinerFid,
		EpochPrice:        epochPrice.BigInt().String(),
		MinBlocksDuration: dealConfig.Duration,
		DealStartEpoch:    dealConfig.StartEpoch,
		FastRetrieval:     dealConfig.FastRetrieval,
		VerifiedDeal:      dealConfig.VerifiedDeal,
	}

	var params []interface{}
	params = append(params, clientStartDealParam)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_START_DEAL,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response, err := web.HttpGet(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	clientStartDeal := &ClientStartDeal{}
	err = json.Unmarshal([]byte(response), clientStartDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if clientStartDeal.Error != nil {
		err := fmt.Errorf("error, code:%d, message:%s", clientStartDeal.Error.Code, clientStartDeal.Error.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return &clientStartDeal.Result.Cid, nil
}

func (lotusClient *LotusClient) LotusGetDealById(dealId uint64) (*DealInfo, error) {
	var params []interface{}
	params = append(params, dealId)
	params = append(params, []interface{}{})
	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_STATE_STORAGE_DEAL,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response, err := web.HttpGet(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	deal := &MarketStorageDeal{}
	err = json.Unmarshal(response, deal)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return &deal.Result, nil
}

type MarketStorageDeal struct {
	Id      int           `json:"id"`
	JsonRpc string        `json:"jsonrpc"`
	Result  DealInfo      `json:"result"`
	Error   *JsonRpcError `json:"error"`
}

type DealInfo struct {
	Proposal struct {
		PieceCID struct {
			PieceCid string `json:"/"`
		} `json:"PieceCID"`
		VerifiedDeal bool   `json:"VerifiedDeal"`
		Client       string `json:"Client"`
		Provider     string `json:"Provider"`
		StartEpoch   int    `json:"StartEpoch"`
		EndEpoch     int    `json:"EndEpoch"`
	} `json:"Proposal"`
	State struct {
		SectorStartEpoch int `json:"SectorStartEpoch"`
		LastUpdatedEpoch int `json:"LastUpdatedEpoch"`
		SlashEpoch       int `json:"SlashEpoch"`
	} `json:"State"`
}

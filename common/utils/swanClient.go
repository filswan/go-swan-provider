package utils

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"swan-provider/config"
	"swan-provider/logs"
	"swan-provider/models"
)

const GET_OFFLINEDEAL_LIMIT_DEFAULT = 50
const RESPONSE_STATUS_SUCCESS = "SUCCESS"

type TokenAccessInfo struct {
	ApiKey      string   `json:"apikey"`
	AccessToken string   `json:"access_token"`
}

type SwanClient struct {
	ApiUrl string
	ApiKey string
	Token  string
}

type MinerResponse struct {
	Status      string        `json:"status"`
	Message     string        `json:"message"`
	Data        models.Miner  `json:"data"`
}

type GetOfflineDealResponse struct {
	Data   GetOfflineDealData `json:"data"`
	Status string             `json:"status"`
}

type GetOfflineDealData struct {
	Deal    []models.OfflineDeal `json:"deal""`
}

type UpdateOfflineDealResponse struct {
	Data   UpdateOfflineDealData `json:"data"`
	Status string                `json:"status"`
}

type UpdateOfflineDealData struct {
	Deal    models.OfflineDeal   `json:"deal""`
	Message string               `json:"message"`
}

func GetSwanClient() *SwanClient {
	mainConf := config.GetConfig().Main
	uri := mainConf.SwanApiUrl + "/user/api_keys/jwt"
	data := TokenAccessInfo{ApiKey: mainConf.SwanApiKey, AccessToken: mainConf.SwanAccessToken}
	response := HttpPostNoToken(uri, data)

	if strings.Index(response, "fail") >= 0 {
		message := GetFieldStrFromJson(response, "message")
		status := GetFieldStrFromJson(response, "status")
		logs.GetLogger().Fatal(status, ": ", message)
	}

	jwtToken := GetFieldMapFromJson(response,"data")
	if jwtToken == nil {
		logs.GetLogger().Fatal("Error: fail to connect swan api")
	}

	jwt:= jwtToken["jwt"].(string)

	swanClient := &SwanClient {
		ApiUrl: mainConf.SwanApiUrl,
		ApiKey: mainConf.SwanApiKey,
		Token: jwt,
	}

	return swanClient
}

func (self *SwanClient) GetMiner(minerFid string) *MinerResponse {
	apiUrl := self.ApiUrl + "/miner/info/" + minerFid

	response := HttpGetNoToken(apiUrl, "")
	minerResponse := &MinerResponse{}
	err := json.Unmarshal([]byte(response), minerResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return minerResponse
}

func (self *SwanClient) UpdateMinerBidConf(minerFid string) {
	minerResponse := self.GetMiner(minerFid)
	if minerResponse == nil || strings.ToUpper(minerResponse.Status) != RESPONSE_STATUS_SUCCESS {
		logs.GetLogger().Error("Error: Get miner information failed")
		return
	}

	miner := minerResponse.Data

	confBid := config.GetConfig().Bid
	if miner.BidMode == confBid.BidMode &&
		miner.StartEpoch == confBid.StartEpoch &&
		miner.Price == confBid.Price &&
		miner.VerifiedPrice == confBid.VerifiedPrice &&
		miner.MinPieceSize == confBid.MinPieceSize &&
		miner.MaxPieceSize == confBid.MaxPieceSize &&
		miner.AutoBidTaskPerDay == confBid.AutoBidTaskPerDay {
		logs.GetLogger().Info("No changes in bid configuration")
		return
	}

	logs.GetLogger().Info("Begin updating bid configuration")
	apiUrl := self.ApiUrl + "/miner/info"

	params := url.Values{}
	params.Add("miner_fid", minerFid)
	params.Add("bid_mode", strconv.Itoa(confBid.BidMode))
	params.Add("start_epoch", strconv.Itoa(confBid.StartEpoch))
	params.Add("price", confBid.Price)
	params.Add("verified_price", confBid.VerifiedPrice)
	params.Add("min_piece_size", confBid.MinPieceSize)
	params.Add("max_piece_size", confBid.MaxPieceSize)
	params.Add("auto_bid_task_per_day", strconv.Itoa(confBid.AutoBidTaskPerDay))

	response := HttpPost(apiUrl, self.Token, strings.NewReader(params.Encode()))

	minerResponse = &MinerResponse{}
	err := json.Unmarshal([]byte(response), minerResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	if strings.ToUpper(minerResponse.Status) != RESPONSE_STATUS_SUCCESS {
		logs.GetLogger().Error("Error: failed to update bid configuration.", minerResponse.Message)
		return
	}

	logs.GetLogger().Info("Bid configuration updated.")
}

func (self *SwanClient) GetOfflineDeals(minerFid, status string, limit ...string) []models.OfflineDeal {
	rowLimit := strconv.Itoa(GET_OFFLINEDEAL_LIMIT_DEFAULT)
	if limit != nil && len(limit) >0 {
		rowLimit = limit[0]
	}

	urlStr := self.ApiUrl + "/offline_deals/" + minerFid + "?deal_status=" + status + "&limit=" + rowLimit + "&offset=0"
	response := HttpGet(urlStr, self.Token, "")
	getOfflineDealResponse := GetOfflineDealResponse{}
	err := json.Unmarshal([]byte(response), &getOfflineDealResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	if strings.ToUpper(getOfflineDealResponse.Status) != RESPONSE_STATUS_SUCCESS {
		logs.GetLogger().Error("Get offline deal with status ", status, " failed")
		return nil
	}

	return getOfflineDealResponse.Data.Deal
}

func (self *SwanClient) UpdateOfflineDealStatus(dealId int, status string, statusInfo ...string) bool {
	if len(status) == 0 {
		logs.GetLogger().Error("Please provide status")
		return false
	}

	apiUrl := self.ApiUrl + "/my_miner/deals/" + strconv.Itoa(dealId)

	params := url.Values{}
	params.Add("status", status)

	if len(statusInfo) > 0 {
		params.Add("note", statusInfo[0])
	}

	if len(statusInfo) > 1 {
		params.Add("file_path", statusInfo[1])
	}

	if len(statusInfo) > 2 {
		params.Add("file_size", statusInfo[2])
	}

	response := HttpPut(apiUrl, self.Token, strings.NewReader(params.Encode()))

	updateOfflineDealResponse := &UpdateOfflineDealResponse{}
	err := json.Unmarshal([]byte(response), updateOfflineDealResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	if strings.ToUpper(updateOfflineDealResponse.Status) != RESPONSE_STATUS_SUCCESS {
		logs.GetLogger().Error("Update offline deal with status ", status, " failed.", updateOfflineDealResponse.Data.Message)
		return false
	}

	return true
}

func (self *SwanClient) SendHeartbeatRequest(minerFid string) string {
	apiUrl := self.ApiUrl + "/heartbeat"
	params := url.Values{}
	params.Add("miner_id", minerFid)

	response := HttpPost(apiUrl, self.Token , strings.NewReader(params.Encode()))
	return response
}


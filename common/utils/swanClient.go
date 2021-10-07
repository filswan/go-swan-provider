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
	ApiKey      string `json:"apikey"`
	AccessToken string `json:"access_token"`
}

type SwanClient struct {
	ApiUrl string
	ApiKey string
	Token  string
}

type MinerResponse struct {
	Status  string       `json:"status"`
	Message string       `json:"message"`
	Data    models.Miner `json:"data"`
}

type GetOfflineDealResponse struct {
	Data   GetOfflineDealData `json:"data"`
	Status string             `json:"status"`
}

type GetOfflineDealData struct {
	Deal []models.OfflineDeal `json:"deal"`
}

type UpdateOfflineDealResponse struct {
	Data   UpdateOfflineDealData `json:"data"`
	Status string                `json:"status"`
}

type UpdateOfflineDealData struct {
	Deal    models.OfflineDeal `json:"deal"`
	Message string             `json:"message"`
}

func (swanClient *SwanClient) GetJwtToken() bool {
	for i := 0; i < 3; i++ {
		uri := swanClient.ApiUrl + "/user/api_keys/jwt"
		data := TokenAccessInfo{ApiKey: swanClient.ApiKey, AccessToken: config.GetConfig().Main.SwanAccessToken}
		response := HttpPostNoToken(uri, data)

		if strings.Contains(response, "fail") {
			message := GetFieldStrFromJson(response, "message")
			status := GetFieldStrFromJson(response, "status")
			logs.GetLogger().Error(status, ": ", message)
			if message == "api_key Not found" {
				logs.GetLogger().Fatal(message, " please check api_key,access_token in ~/.swan/provider/config.toml")
			}
			if i < 3 {
				continue
			} else {
				logs.GetLogger().Error("Failed to get token after trying 3 times.")
				return false
			}
		}

		jwtToken := GetFieldMapFromJson(response, "data")
		if jwtToken == nil {
			logs.GetLogger().Error("Error: fail to connect swan api")
			if i < 3 {
				continue
			} else {
				logs.GetLogger().Error("Failed to get token after trying 3 times.")
				return false
			}
		}

		swanClient.Token = jwtToken["jwt"].(string)

		return true
	}

	return false
}

func GetSwanClient() *SwanClient {
	swanClient := &SwanClient{
		ApiUrl: config.GetConfig().Main.SwanApiUrl,
		ApiKey: config.GetConfig().Main.SwanApiKey,
	}

	if !swanClient.GetJwtToken() {
		logs.GetLogger().Fatal("Failed to get jwt token from swan")
	}

	return swanClient
}

func (swanClient *SwanClient) GetMiner(minerFid string) *MinerResponse {
	apiUrl := swanClient.ApiUrl + "/miner/info/" + minerFid

	response := HttpGetNoToken(apiUrl, "")
	msg := GetFieldStrFromJson(response, "message")
	if msg == "Miner Not found" {
		logs.GetLogger().Fatal("Cannot find your miner:", minerFid)
	}

	minerResponse := &MinerResponse{}
	err := json.Unmarshal([]byte(response), minerResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return minerResponse
}

func (swanClient *SwanClient) UpdateMinerBidConf(minerFid string) {
	swanClient.GetJwtToken()

	minerResponse := swanClient.GetMiner(minerFid)
	if minerResponse == nil || strings.ToUpper(minerResponse.Status) != RESPONSE_STATUS_SUCCESS {
		logs.GetLogger().Error("Error: Get miner information failed")
		return
	}

	miner := minerResponse.Data

	confBid := config.GetConfig().Bid
	if miner.BidMode == confBid.BidMode &&
		miner.ExpectedSealingTime == confBid.ExpectedSealingTime &&
		miner.StartEpoch == confBid.StartEpoch &&
		miner.AutoBidTaskPerDay == confBid.AutoBidTaskPerDay {
		logs.GetLogger().Info("No changes in bid configuration")
		return
	}

	logs.GetLogger().Info("Begin updating bid configuration")
	apiUrl := swanClient.ApiUrl + "/miner/info"

	params := url.Values{}
	params.Add("miner_fid", minerFid)
	params.Add("bid_mode", strconv.Itoa(confBid.BidMode))
	params.Add("expected_sealing_time", strconv.Itoa(confBid.ExpectedSealingTime))
	params.Add("start_epoch", strconv.Itoa(confBid.StartEpoch))
	params.Add("auto_bid_task_per_day", strconv.Itoa(confBid.AutoBidTaskPerDay))

	response := HttpPost(apiUrl, swanClient.Token, strings.NewReader(params.Encode()))

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

func (swanClient *SwanClient) GetOfflineDeals(minerFid, status string, limit ...string) []models.OfflineDeal {
	if !swanClient.GetJwtToken() {
		return nil
	}

	rowLimit := strconv.Itoa(GET_OFFLINEDEAL_LIMIT_DEFAULT)
	if len(limit) > 0 {
		rowLimit = limit[0]
	}

	urlStr := swanClient.ApiUrl + "/offline_deals/" + minerFid + "?deal_status=" + status + "&limit=" + rowLimit + "&offset=0"
	response := HttpGet(urlStr, swanClient.Token, "")
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

func (swanClient *SwanClient) UpdateOfflineDealStatus(dealId int, status string, statusInfo ...string) bool {
	if !swanClient.GetJwtToken() {
		return false
	}

	if len(status) == 0 {
		logs.GetLogger().Error("Please provide status")
		return false
	}

	apiUrl := swanClient.ApiUrl + "/my_miner/deals/" + strconv.Itoa(dealId)

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

	response := HttpPut(apiUrl, swanClient.Token, strings.NewReader(params.Encode()))

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

func (swanClient *SwanClient) SendHeartbeatRequest(minerFid string) string {
	apiUrl := swanClient.ApiUrl + "/heartbeat"
	params := url.Values{}
	params.Add("miner_id", minerFid)

	response := HttpPost(apiUrl, swanClient.Token, strings.NewReader(params.Encode()))

	if strings.Contains(response, "fail") {
		logs.GetLogger().Error("Failed to send heartbeat.")
	}
	return response
}

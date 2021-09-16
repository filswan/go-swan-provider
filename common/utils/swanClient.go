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

type TokenAccessInfo struct {
	ApiKey      string   `json:"apikey"`
	AccessToken string   `json:"access_token"`
}

type SwanClient struct {
	ApiUrl string
	ApiKey string
	Token  string
}

type OfflineDealResponse struct {
	Data   OfflineDealData `json:"data"`
	Status string          `json:"status"`
}

type OfflineDealData struct {
	Deal  []models.OfflineDeal `json:"deal""`
}

func GetSwanClient() *SwanClient {
	mainConf := config.GetConfig().Main
	uri := mainConf.SwanApiUrl +"/user/api_keys/jwt"
	data := TokenAccessInfo{ApiKey: mainConf.SwanApiKey, AccessToken: mainConf.SwanAccessToken}
	response := HttpPostNoToken(uri, data)

	jwtToken := GetFieldMapFromJson(response,"data")
	if jwtToken == nil {
		logs.GetLogger().Fatal("Error: fail to connect swan api")
		return nil
	}

	jwt:= jwtToken["jwt"].(string)

	swanClient := &SwanClient{
		ApiUrl: mainConf.SwanApiUrl,
		ApiKey: mainConf.SwanApiKey,
		Token: jwt,
	}

	return swanClient
}

func (self *SwanClient) UpdateBidConf(minerFid string) string {
	logs.GetLogger().Info("Begin updating bid configuration")
	apiUrl := self.ApiUrl + "/miner/info/"

	confBid := config.GetConfig().Bid
	params := url.Values{}
	params.Add("miner_fid", minerFid)
	params.Add("bid_mode", strconv.Itoa(confBid.BidMode))
	params.Add("start_epoch", strconv.Itoa(confBid.StartEpoch))
	params.Add("price", strconv.FormatFloat(confBid.Price, 'E', -1, 64))
	params.Add("verified_price", strconv.FormatFloat(confBid.Price, 'E', -1, 64))
	params.Add("min_piece_size", confBid.MinPieceSize)
	params.Add("max_piece_size", confBid.MaxPieceSize)

	response := HttpPost(apiUrl, self.Token, strings.NewReader(params.Encode()))

	logs.GetLogger().Info("Finished")
	return response
}

func (self *SwanClient) GetOfflineDeals(minerFid, status string, limit ...string) []models.OfflineDeal {
	rowLimit := strconv.Itoa(GET_OFFLINEDEAL_LIMIT_DEFAULT)
	if limit != nil && len(limit) >0 {
		rowLimit = limit[0]
	}

	urlStr := self.ApiUrl + "/offline_deals/" + minerFid + "?deal_status=" + status + "&limit=" + rowLimit + "&offset=0"
	response := HttpGet(urlStr, self.Token, "")
	offlineDealResponse := OfflineDealResponse{}
	err := json.Unmarshal([]byte(response),&offlineDealResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return offlineDealResponse.Data.Deal
}

func (self *SwanClient) UpdateOfflineDealStatus(dealId int, status string, statusInfo ...string) string {
	if len(status) == 0 {
		logs.GetLogger().Error("Please provide status")
		return ""
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

	return response
}

func (self *SwanClient) SendHeartbeatRequest(minerFid string) string {
	apiUrl := self.ApiUrl + "/heartbeat"
	params := url.Values{}
	params.Add("miner_id", minerFid)

	response := HttpPost(apiUrl, self.Token , strings.NewReader(params.Encode()))
	return response
}


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

func GetSwanClient() (*SwanClient) {
	mainConf := config.GetConfig().Main
	uri := mainConf.ApiUrl+"/user/api_keys/jwt"
	data := TokenAccessInfo{ApiKey: mainConf.ApiKey, AccessToken: mainConf.AccessToken}
	response := HttpPostJsonParamNoToken(uri, data)

	jwtToken := GetFieldMapFromJson(response,"data")
	jwt:= jwtToken["jwt"].(string)

	swanClient := &SwanClient{
		ApiUrl: mainConf.ApiUrl,
		ApiKey: mainConf.ApiKey,
		Token: jwt,
	}

	return swanClient
}

func (self *SwanClient) GetOfflineDeals(minerFid, status string, limit ...string) ([]models.OfflineDeal){
	rowLimit := strconv.Itoa(GET_OFFLINEDEAL_LIMIT_DEFAULT)
	if limit != nil && len(limit) >0 {
		rowLimit = limit[0]
	}

	urlStr := config.GetConfig().Main.ApiUrl+ "/offline_deals/" + minerFid + "?deal_status=" + status + "&limit=" + rowLimit + "&offset=0"
	response := HttpGetJsonParam(urlStr, self.Token, "")
	offlineDealResponse := OfflineDealResponse{}
	err := json.Unmarshal([]byte(response),&offlineDealResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return offlineDealResponse.Data.Deal
}

func (self *SwanClient) UpdateOfflineDealStatus(dealId int, status string, statusInfo ...string) (string) {
	apiUrl := config.GetConfig().Main.ApiUrl + "/my_miner/deals/" + strconv.Itoa(dealId)

	params := url.Values{}
	if len(status) > 0 {
		params.Add("status", status)
	}

	if len(statusInfo) > 0 {
		params.Add("note", statusInfo[0])
	}

	if len(statusInfo) > 1 {
		params.Add("file_path", statusInfo[1])
	}

	if len(statusInfo) > 2 {
		params.Add("file_size", statusInfo[2])
	}

	response := HttpPutFormParam(apiUrl, self.Token, strings.NewReader(params.Encode()))

	return response
}

func (self *SwanClient) SendHeartbeatRequest(minerFid string) string {
	apiUrl := config.GetConfig().Main.ApiUrl + "/heartbeat"
	params := url.Values{}
	params.Add("miner_id", minerFid)

	response := HttpPostFormParam(apiUrl, self.Token , strings.NewReader(params.Encode()))
	return response
}


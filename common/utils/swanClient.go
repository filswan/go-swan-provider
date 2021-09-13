package utils

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"swan-miner/config"
	"swan-miner/logs"
	"swan-miner/models"
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

func GetSwanClient() (*SwanClient){
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

	url := config.GetConfig().Main.ApiUrl+ "/offline_deals/" + minerFid + "?deal_status=" + status + "&limit=" + rowLimit + "&offset=0"
	response := HttpGetJsonParam(url, self.Token, "")
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

	form := url.Values{}
	if len(status) > 0 {
		form.Add("status", status)
	}

	if len(statusInfo) > 0 {
		form.Add("note", statusInfo[0])
	}

	if len(statusInfo) > 1 {
		form.Add("file_path", statusInfo[1])
	}

	if len(statusInfo) > 2 {
		form.Add("file_size", statusInfo[2])
	}

	response := HttpPutFormParam(apiUrl, self.Token, strings.NewReader(form.Encode()))

	return response
}



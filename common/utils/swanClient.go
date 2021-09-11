package utils

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"swan-miner/config"
	"swan-miner/models"
)

type TokenAccessInfo struct {
	ApiKey      string   `json:"apikey"`
	AccessToken string   `json:"access_token"`
}

type SwanClient struct {
	ApiUrl string
	ApiKey string
	Token  string
}

type DealDetail struct {
	Status   string   `json:"status"`
	Note     string   `json:"note"`
	FilePath string   `json:"file_path"`
	FileSize string   `json:"file_size"`
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

/*func (self *SwanClient) UpdateTaskByUuid(taskUuid, minerFid string, csvFile interface{}){
	logs.GetLogger().Info("Updating Swan task.")
	uri := config.GetConfig().Main.ApiUrl + "/uuid_tasks/" + taskUuid
	tokenString :=""
	payloadData := "{\"miner_fid\": "+minerFid+"}"

	Put(uri,tokenString,payloadData)
	logs.GetLogger().Info("Swan task updated.")
}
*/

func (self *SwanClient) GetOfflineDeals(minerFid, status string, limit ...string) ([]models.OfflineDeal){
	rowLimit := strconv.Itoa(GET_OFFLINEDEAL_LIMIT_DEFAULT)
	if limit != nil && len(limit) >0 {
		rowLimit = limit[0]
	}

	url := config.GetConfig().Main.ApiUrl+ "/offline_deals/" + minerFid + "?deal_status=" + status + "&limit=" + rowLimit + "&offset=0"
	//fmt.Println(url)
	response := HttpGetJsonParam(url, self.Token, "")
	//fmt.Println(response)
	offlineDealResponse := OfflineDealResponse{}
	json.Unmarshal([]byte(response),&offlineDealResponse)
	deals:=offlineDealResponse.Data.Deal

	return deals
}

func (self *SwanClient) UpdateOfflineDealStatus(dealId int, status string, statusInfo ...interface{}) (string) {
	apiUrl := config.GetConfig().Main.ApiUrl + "/my_miner/deals/" + strconv.Itoa(dealId)

	form := url.Values{}
	if len(status) > 0 {
		form.Add("status", status)
	}

	if len(statusInfo) > 0 {
		form.Add("note", statusInfo[0].(string))
	}

	if len(statusInfo) > 1 {
		form.Add("file_path", statusInfo[1].(string))
	}

	if len(statusInfo) > 2 {
		form.Add("file_size", statusInfo[2].(string))
	}

	response := HttpPutFormParam(apiUrl, self.Token, strings.NewReader(form.Encode()))
/*	fmt.Println(apiUrl)
	fmt.Println(response)*/

	return response
}



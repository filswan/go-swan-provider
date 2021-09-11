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
	//fmt.Println("Refreshing token")
	mainConf := config.GetConfig().Main
	uri := mainConf.ApiUrl+"/user/api_keys/jwt"
	data := TokenAccessInfo{ApiKey: mainConf.ApiKey, AccessToken: mainConf.AccessToken} //
	//dataJson := fmt.Sprintf(`{\"apikey\":\"%s\",\"access_token\":\"%s\"}`, mainConf.ApiKey, mainConf.AccessToken)//ToJson(data)
	response := HttpPostJsonParamNoToken(uri,data)
	//fmt.Println(response)

	jwtToken := GetFieldMapFromJson(response,"data")
	jwt:= jwtToken["jwt"].(string)
	//fmt.Println(jwt)

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

/*	i:=0
	for i < len(deals){
		deal := deals[i]
		fmt.Println(deal)
		i++
	}*/

	return deals
}

func (self *SwanClient) UpdateOfflineDealStatus(dealId int, status, note string) (string) {
	apiUrl := config.GetConfig().Main.ApiUrl + "/my_miner/deals/" + strconv.Itoa(dealId)
	form := url.Values{}
	if len(status)>0 {
		form.Add("status", status)
	}

	if len(note)>0{
		form.Add("note", note)
	}

	response := HttpPutFormParam(apiUrl,self.Token, strings.NewReader(form.Encode()))
/*	fmt.Println(apiUrl)
	fmt.Println(response)*/
	return response
}

func (self *SwanClient) UpdateOfflineDealDetails(dealId int, status, note, filePath, fileSize string) (string) {
	apiUrl := config.GetConfig().Main.ApiUrl + "/my_miner/deals/" + strconv.Itoa(dealId)

	form := url.Values{}
	if len(status)>0 {
		form.Add("status", status)
	}

	if len(note)>0{
		form.Add("note", note)
	}

	if len(filePath)>0 {
		form.Add("file_path", filePath)
	}

	if len(note)>0{
		form.Add("file_size", fileSize)
	}

	response := HttpPutFormParam(apiUrl, self.Token,strings.NewReader(form.Encode()))
/*	fmt.Println(apiUrl)
	fmt.Println(response)*/

	return response
}



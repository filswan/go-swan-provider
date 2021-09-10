package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	response := HttpPostNoToken(uri,data)
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

func (self *SwanClient) GetOfflineDeals(minerFid, status, limit string) ([]models.OfflineDeal){
	url := config.GetConfig().Main.ApiUrl+ "/offline_deals/" + minerFid + "?deal_status=" + status + "&limit=" + limit + "&offset=0"
	//fmt.Println(url)
	response := HttpGet(url, self.Token, "")
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

func (self *SwanClient) UpdateOfflineDealStatus(status, note string, dealId int) (string) {
	url := config.GetConfig().Main.ApiUrl + "/my_miner/deals/" + strconv.Itoa(dealId)
	dealDetail := DealDetail{
		Status: status,
		Note: note,
	}
	fmt.Println(ToJson(dealDetail))
	response := HttpPut(url,self.Token,dealDetail)
	fmt.Println(url)
	fmt.Println(response)
	return response
}

func (self *SwanClient) UpdateOfflineDealDetails(status, note string, dealId int, filePath string, fileSize string)  {
	url := config.GetConfig().Main.ApiUrl + "/my_miner/deals/" + strconv.Itoa(dealId)
	dealDetail := DealDetail{
		Status: status,
		Note: note,
		FilePath: filePath,
		FileSize: fileSize,
	}
	fmt.Println(ToJson(dealDetail))
	response := HttpPut(url,self.Token,dealDetail)
	fmt.Println(url)
	fmt.Println(response)
}



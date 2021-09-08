package utils

import (
	"fmt"
	"swan-miner/config"
	"swan-miner/logs"
)

type TokenAccessInfo struct {
	ApiKey      string   `json:"apikey"`
	AccessToken string   `json:"access_token"`
}

type SwanClient struct {
	ApiUrl string
	ApiKey string
	Token string
}

func GetJwtToken() (*SwanClient){
	fmt.Println("Refreshing token")
	mainConf := config.GetConfig().Main
	uri := mainConf.ApiUrl+"/user/api_keys/jwt"
	data := TokenAccessInfo{ApiKey: mainConf.ApiKey, AccessToken: mainConf.AccessToken}//
	//dataJson := fmt.Sprintf(`{\"apikey\":\"%s\",\"access_token\":\"%s\"}`, mainConf.ApiKey, mainConf.AccessToken)//ToJson(data)
	response := Post(uri,data)
	fmt.Println(response)

	jwtToken := GetFieldMapFromJson(response,"data")
	jwt:= jwtToken["jwt"].(string)
	fmt.Println(jwt)

	swanClient := &SwanClient{
		ApiUrl: mainConf.ApiUrl,
		ApiKey: mainConf.ApiKey,
		Token: jwt,
	}

	return swanClient
}

func (self *SwanClient) UpdateTaskByUuid(taskUuid, minerFid string, csvFile interface{}){
	logs.GetLogger().Info("Updating Swan task.")
	uri := config.GetConfig().Main.ApiUrl + "/uuid_tasks/" + taskUuid
	tokenString :=""
	payloadData := "{\"miner_fid\": "+minerFid+"}"

	Put(uri,tokenString,payloadData)
	logs.GetLogger().Info("Swan task updated.")
}


func (self *SwanClient) GetOfflineDeals(minerFid, status, limit string) (interface{}){
	uri := config.GetConfig().Main.ApiUrl+ "/offline_deals/" + minerFid + "?deal_status=" + status + "&limit=" + limit + "&offset=0"
	response := Get(uri)
	deal := GetFieldFromJson(response, "deal")
	return deal
}

func (self *SwanClient) UpdateOfflineDealDetails(status,note string, dealId string, filePath string, fileSize string)  {
	url := config.GetConfig().Main.ApiUrl + "/my_miner/deals/" + string(dealId)
	body := fmt.Sprintf("{\"status\": %s, \"note\": %s, \"file_path\": %s, \"file_size\": %s}", status, note,filePath, fileSize)
	Put(url,"",body)
}



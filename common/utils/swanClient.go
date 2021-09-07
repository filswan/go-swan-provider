package utils

import (
	"fmt"
	"swan-miner/config"
	"swan-miner/logs"
)

type TokenAccessInfo struct {
	apikey       string
	access_token string
}

type SwanClient struct {
	ApiUrl string
	ApiKey string
	Token string
}

func (self *SwanClient) GetJwtToken(){
	fmt.Println("Refreshing token")
	mainConf := config.GetConfig().Main
	uri := mainConf.ApiUrl+"/user/api_keys/jwt"
	data := TokenAccessInfo{mainConf.ApiKey, mainConf.AccessToken}
	response := Post(uri,data)
	fmt.Println(response)

	jwtToken := GetFieldFromJson(response,"jwt")
	fmt.Println(jwtToken)

	tokenString,ok := jwtToken.(string)
	fmt.Println(tokenString,ok)
	jwtTokenExpiration := GetTokenExpiration(tokenString)
	fmt.Println(jwtTokenExpiration)
/*
	payload = jwt.decode(jwt=self.jwt_token, verify=False, algorithm='HS256')
	self.jwt_token_expiration = payload['exp']*/
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



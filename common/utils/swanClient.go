package utils

import (
	"fmt"
	"swan-miner/config"
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

type OfflineDeal struct {
	uuid             string
	source_file_name string
	source_file_path string
	source_file_md5  string
	source_file_url  string
	source_file_size string
	car_file_name    string
	car_file_path    string
	car_file_md5     string
	car_file_url     string
	car_file_size    string
	deal_cid         string
	data_cid         string
	piece_cid        string
	miner_id         string
	start_epoch      string
}


func GetJwtToken() (*SwanClient){
	//fmt.Println("Refreshing token")
	mainConf := config.GetConfig().Main
	uri := mainConf.ApiUrl+"/user/api_keys/jwt"
	data := TokenAccessInfo{ApiKey: mainConf.ApiKey, AccessToken: mainConf.AccessToken}//
	//dataJson := fmt.Sprintf(`{\"apikey\":\"%s\",\"access_token\":\"%s\"}`, mainConf.ApiKey, mainConf.AccessToken)//ToJson(data)
	response := Post(uri,data)
	//fmt.Println(response)

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

/*func (self *SwanClient) UpdateTaskByUuid(taskUuid, minerFid string, csvFile interface{}){
	logs.GetLogger().Info("Updating Swan task.")
	uri := config.GetConfig().Main.ApiUrl + "/uuid_tasks/" + taskUuid
	tokenString :=""
	payloadData := "{\"miner_fid\": "+minerFid+"}"

	Put(uri,tokenString,payloadData)
	logs.GetLogger().Info("Swan task updated.")
}
*/

func (self *SwanClient) GetOfflineDeals(minerFid, status, limit string) ([]OfflineDeal){
	url := config.GetConfig().Main.ApiUrl+ "/offline_deals/" + minerFid + "?deal_status=" + status + "&limit=" + limit + "&offset=0"
	fmt.Println(url)
	response := Get(url, self.Token, "")
	fmt.Println(response)
	data := GetFieldMapFromJson(response, "data")
	fmt.Println(data)
	deals := data["deal"].([]OfflineDeal)
	fmt.Println(deals)
	return deals
}

func (self *SwanClient) UpdateOfflineDealDetails(status,note string, dealId string, filePath string, fileSize string)  {
	url := config.GetConfig().Main.ApiUrl + "/my_miner/deals/" + string(dealId)
	body := fmt.Sprintf("{\"status\": %s, \"note\": %s, \"file_path\": %s, \"file_size\": %s}", status, note,filePath, fileSize)
	Put(url,"",body)
}



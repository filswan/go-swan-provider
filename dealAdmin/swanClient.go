package dealAdmin

import (
	"fmt"
	"swan-miner/common/utils"
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

type DealDetail struct {
	Status   string   `json:"status"`
	Note     string   `json:"note"`
	FilePath string   `json:"file_path"`
	FileSize string   `json:"file_size"`
}

func GetSwanClient() (*SwanClient){
	//fmt.Println("Refreshing token")
	mainConf := config.GetConfig().Main
	uri := mainConf.ApiUrl+"/user/api_keys/jwt"
	data := TokenAccessInfo{ApiKey: mainConf.ApiKey, AccessToken: mainConf.AccessToken} //
	//dataJson := fmt.Sprintf(`{\"apikey\":\"%s\",\"access_token\":\"%s\"}`, mainConf.ApiKey, mainConf.AccessToken)//ToJson(data)
	response := utils.Post(uri,data)
	//fmt.Println(response)

	jwtToken := utils.GetFieldMapFromJson(response,"data")
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

func (self *SwanClient) GetOfflineDeals(minerFid, status, limit string) ([]interface{}){
	url := config.GetConfig().Main.ApiUrl+ "/offline_deals/" + minerFid + "?deal_status=" + status + "&limit=" + limit + "&offset=0"
	fmt.Println(url)
	response := utils.Get(url, self.Token, "")
	fmt.Println(response)
	data := utils.GetFieldMapFromJson(response, "data")
	fmt.Println(data)
	deals := data["deal"].([]interface{})
	fmt.Println(deals)
	return deals
}

func (self *SwanClient) UpdateOfflineDealDetails(status,note,dealId string, filePath string, fileSize string)  {
	url := config.GetConfig().Main.ApiUrl + "/my_miner/deals/" + dealId
	dealDetail := DealDetail{
		Status: status,
		Note: note,
		FilePath: filePath,
		FileSize: fileSize,
	}
	response := utils.Put(url,self.Token,dealDetail)
	fmt.Println(url)
	fmt.Println(response)
}



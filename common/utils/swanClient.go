package utils

import (
	"fmt"
	"swan-miner/config"
)

type TokenAccessInfo struct {
	apikey       string
	access_token string
}

func getJwtToken(){
	fmt.Println("Refreshing token")
	mainConf := config.GetConfig().Main
	uri := mainConf.ApiUrl+"/user/api_keys/jwt"
	data := TokenAccessInfo{mainConf.ApiKey, mainConf.AccessToken}
	response := Post(uri,data)
	fmt.Println(response)
}

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

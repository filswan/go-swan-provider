package dealAdmin

import (
	"fmt"
	"strings"
	"swan-miner/common/utils"
	"swan-miner/config"
	"swan-miner/logs"
)

const IDPREFIX = "nbfs"
const ADD_URI = "aria2.addUri"
const GET_VER = "aria2.getVersion"
const STOPPED = "aria2.tellStopped"
const ACTIVE = "aria2.tellActive"
const STATUS = "aria2.tellStatus"

type Aria2Client struct {
	Host      string
	port      int
	token     string
	serverUrl string
}

type Payload struct {
	JsonRpc string        `json:"jsonrpc"`
	Id      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

func GetAria2Client() (*Aria2Client){
	confAria2c := config.GetConfig().Aria2
	aria2cClient := &Aria2Client{
		Host:  confAria2c.Aria2Host,
		port:  confAria2c.Aria2Port,
		token: confAria2c.Aria2Secret,
	}

	aria2cClient.serverUrl = fmt.Sprintf("http://%s:%d/jsonrpc", aria2cClient.Host, aria2cClient.port)

	return aria2cClient
}

func (self *Aria2Client) GenPayload(method string, uri string , options interface{}) (interface{}){
	var params []interface{}
	params = append(params, "token:"+self.token)
	var urls [] string
	urls = append(urls, uri)
	params = append(params, urls)
	params = append(params, options)

	payload := Payload{
		JsonRpc: "2.0",
		Id: uri,
		Method: method,
		Params: params,
	}

	return payload
}

func (self *Aria2Client) DownloadFile(uri string, options interface{}) (string) {
	payloads := self.GenPayload(ADD_URI, uri, options)
	result := utils.HttpPostNoToken(self.serverUrl,payloads)
	//fmt.Println(result)
	if strings.Contains(result,"error"){
		errorInfo := utils.GetFieldMapFromJson(result, "error")
		errorCode := errorInfo["code"]
		errorMsg := errorInfo["message"]
		msg := fmt.Sprintf("ERROR: %s, %s",errorCode, errorMsg)
		logs.GetLogger().Error(msg)
		return ""
	}else{
		return result
	}
}

func (self *Aria2Client) GenPayloadForStatus(gid string) (interface{}){
	var params []interface{}
	params = append(params, "token:"+self.token)
/*	var urls [] string
	urls = append(urls, uri)*/
	params = append(params, gid)

	payload := Payload{
		JsonRpc: "2.0",
		//Id: uri,
		Method: STATUS,
		Params: params,
	}

	return payload
}


func (self *Aria2Client) GetDownloadStatus(gid string) (string) {
	payload := self.GenPayloadForStatus(gid)
	result := utils.HttpPostNoToken(self.serverUrl,payload)
	fmt.Println(result)
	return result
}



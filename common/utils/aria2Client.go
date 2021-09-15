package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"swan-provider/config"
	"swan-provider/logs"
)

const ADD_URI = "aria2.addUri"
const STATUS = "aria2.tellStatus"

type Aria2Client struct {
	Host      string
	port      int
	token     string
	serverUrl string
}

type Payload struct {
	JsonRpc   string        `json:"jsonrpc"`
	Id        string        `json:"id"`
	Method    string        `json:"method"`
	Params    []interface{} `json:"params"`
}

type Aria2GetStatusFail struct {
	Id 		string             `json:"id"`
	JsonRpc string             `json:"jsonrpc"`
	Error 	*Aria2StatusError  `json:"error"`
}

type Aria2StatusError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
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
	response := HttpPostNoToken(self.serverUrl,payloads)
	if strings.Contains(response,"error"){
		aria2GetStatusFail := Aria2GetStatusFail{}
		err := json.Unmarshal([]byte(response), &aria2GetStatusFail)
		if err != nil {
			logs.GetLogger().Error(err)
		}

		msg := fmt.Sprintf("Error: code(%d), %s",aria2GetStatusFail.Error.Code, aria2GetStatusFail.Error.Message)
		logs.GetLogger().Error(msg)
		return ""
	}else{
		return response
	}
}

func (self *Aria2Client) GenPayloadForStatus(gid string) (interface{}){
	var params []interface{}
	params = append(params, "token:"+self.token)
	params = append(params, gid)

	payload := Payload{
		JsonRpc: "2.0",
		Method: STATUS,
		Params: params,
	}

	return payload
}


func (self *Aria2Client) GetDownloadStatus(gid string) (string) {
	payload := self.GenPayloadForStatus(gid)
	result := HttpPostNoToken(self.serverUrl, payload)
	return result
}



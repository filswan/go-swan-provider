package dealAdmin

import (
	"container/list"
	"fmt"
	"swan-miner/common/utils"
	"swan-miner/config"
)

const IDPREFIX = "nbfs"
const ADD_URI = "aria2.addUri"
const GET_VER = "aria2.getVersion"
const STOPPED = "aria2.tellStopped"
const ACTIVE = "aria2.tellActive"
const STATUS = "aria2.tellStatus"

type Aria2Client struct {
	host string
	port int
	token string
	serverUrl string "http://{host}:{port}/jsonrpc"
}

func GetAria2Client() (*Aria2Client){
	confAria2c := config.GetConfig().Aria2
	aria2cClient := &Aria2Client{
		host: confAria2c.Aria2Host,
		port: confAria2c.Aria2Port,
		token: confAria2c.Aria2Secret,
	}

	aria2cClient.serverUrl = fmt.Sprintf("http://%s:%d/jsonrpc", aria2cClient.host, aria2cClient.port)

	return aria2cClient
}

func (self *Aria2Client) GenPayload(method string, uris string , options string, cid string, IDPREFIX string) (string){
	if cid!=""{
		cid = IDPREFIX+cid
	}else {
		cid =IDPREFIX+IDPREFIX
	}


	l := list.New()
	if len(self.token)>0{
		l.PushBack("token:"+self.token)
	}

	if (len(uris)>0){
		l.PushBack(uris)
	}

	if (len(options)>0){
		l.PushBack(options)
	}

	var p map[string]interface{}
	p["jsonrpc"]="2.0"
	p["id"]=cid
	p["method"]=method
	p["params"]=l

	return utils.ToJson(p)
}

func (self *Aria2Client) post(action, params string) (string) {
	payloads := self.GenPayload(action, "", "", "", "")
	result := utils.Post(self.serverUrl,payloads)
	return result
}

func (self *Aria2Client) addUri(uri, options string) (string) {
	result := self.post(ADD_URI,options)
	return result
}



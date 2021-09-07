package dealAdmin

import (
	"container/list"
	"swan-miner/common/utils"
)

const IDPREFIX = "nbfs"
const ADD_URI = "aria2.addUri"
const GET_VER = "aria2.getVersion"
const STOPPED = "aria2.tellStopped"
const ACTIVE = "aria2.tellActive"
const STATUS = "aria2.tellStatus"

type Aria2c struct {
	host string
	port string
	token string
	serverUrl string "http://{host}:{port}/jsonrpc"
}

func (self *Aria2c) GenPayload(method string, uris string , options string, cid string, IDPREFIX string) (string){
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

func (self *Aria2c) post(action, params string) (string) {
	payloads := self.GenPayload(action, "", "", "", "")
	result := utils.Post(self.serverUrl,payloads)
	return result
}

func (self *Aria2c) addUri(uri, options string) (string) {
	result := self.post(ADD_URI,options)
	return result
}



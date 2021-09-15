package utils

import (
	"encoding/json"
	"fmt"
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

type DownloadOption struct {
	Out string   `json:"out"`
	Dir string   `json:"dir"`
}

type Aria2GetStatusResponse struct {
	Id 		string               `json:"id"`
	JsonRpc string               `json:"jsonrpc"`
	Error 	*Aria2Error          `json:"error"`
	Result 	*Aria2StatusResult   `json:"result"`
}

type Aria2DownloadResponse struct {
	Id 		string               `json:"id"`
	JsonRpc string               `json:"jsonrpc"`
	Error 	*Aria2Error          `json:"error"`
	Gid 	string               `json:"result"`
}


type Aria2Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

/*type Aria2GetStatusSuccess struct {
	Id 		string             `json:"id"`
	JsonRpc string             `json:"jsonrpc"`
	Result 	*Aria2StatusResult `json:"result"`
}*/

type Aria2StatusResult struct {
	Bitfield        string                  `json:"bitfield"`
	CompletedLength string                  `json:"completedLength"`
	Connections     string                  `json:"connections"`
	Dir             string                  `json:"dir"`
	DownloadSpeed   string                  `json:"downloadSpeed"`
	ErrorCode       string                  `json:"errorCode"`
	ErrorMessage    string                  `json:"errorMessage"`
	Gid             string                  `json:"gid"`
	NumPieces       string                  `json:"numPieces"`
	PieceLength     string                  `json:"pieceLength"`
	Status          string                  `json:"status"`
	TotalLength     string                  `json:"totalLength"`
	UploadLength    string                  `json:"uploadLength"`
	UploadSpeed     string                  `json:"uploadSpeed"`
	Files           []Aria2StatusResultFile `json:"files"`
}

type Aria2StatusResultFile struct {
	CompletedLength string                     `json:"completedLength"`
	Index           string                     `json:"index"`
	Length          string                     `json:"length"`
	Path            string                     `json:"path"`
	Selected        string                     `json:"selected"`
	Uris            []Aria2StatusResultFileUri `json:"uris"`
}

type Aria2StatusResultFileUri struct {
	Status string `json:"status"`
	Uri    string `json:"uri"`
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

func (self *Aria2Client) GenPayload4Download(method string, uri string, outDir, outFilename string) Payload {
	options := DownloadOption{
		Out: outFilename,
		Dir: outDir,
	}

	var params []interface{}
	params = append(params, "token:"+self.token)
	var urls [] string
	urls = append(urls, uri)
	params = append(params, urls)
	params = append(params, options)

	payload := Payload {
		JsonRpc: "2.0",
		Id: uri,
		Method: method,
		Params: params,
	}

	return payload
}

func (self *Aria2Client) DownloadFile(uri string, outDir, outFilename string) *Aria2DownloadResponse {
	payload := self.GenPayload4Download(ADD_URI, uri, outDir, outFilename)

	if IsFileExists(outDir, outFilename) {
		RemoveFile(outDir, outFilename)
	}

	response := HttpPostNoToken(self.serverUrl, payload)
	aria2DownloadResponse := &Aria2DownloadResponse{}
	err := json.Unmarshal([]byte(response), aria2DownloadResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return aria2DownloadResponse
}

func (self *Aria2Client) GenPayload4Status(gid string) Payload {
	var params []interface{}
	params = append(params, "token:"+self.token)
	params = append(params, gid)

	payload := Payload {
		JsonRpc: "2.0",
		Method: STATUS,
		Params: params,
	}

	return payload
}


func (self *Aria2Client) GetDownloadStatus(gid string) *Aria2GetStatusResponse {
	payload := self.GenPayload4Status(gid)
	response := HttpPostNoToken(self.serverUrl, payload)

	aria2GetStatusResponse := &Aria2GetStatusResponse{}
	err := json.Unmarshal([]byte(response), aria2GetStatusResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return aria2GetStatusResponse
}



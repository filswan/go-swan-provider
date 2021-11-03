package client

import (
	"encoding/json"
	"fmt"
	"swan-provider/common/utils"
	"swan-provider/config"
	"swan-provider/logs"

	libclient "github.com/filswan/go-swan-lib/client"
)

const ADD_URI = "aria2.addUri"
const STATUS = "aria2.tellStatus"

type JsonRpcParams struct {
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type Aria2Payload struct {
	JsonRpc string        `json:"jsonrpc"`
	Id      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type Aria2Client struct {
	Host      string
	port      int
	token     string
	serverUrl string
}

type Aria2DownloadOption struct {
	Out string `json:"out"`
	Dir string `json:"dir"`
}

type Aria2Status struct {
	Id      string             `json:"id"`
	JsonRpc string             `json:"jsonrpc"`
	Error   *Aria2Error        `json:"error"`
	Result  *Aria2StatusResult `json:"result"`
}

type Aria2Download struct {
	Id      string      `json:"id"`
	JsonRpc string      `json:"jsonrpc"`
	Error   *Aria2Error `json:"error"`
	Gid     string      `json:"result"`
}

type Aria2Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

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

func GetAria2Client() *Aria2Client {
	confAria2c := config.GetConfig().Aria2
	aria2cClient := &Aria2Client{
		Host:  confAria2c.Aria2Host,
		port:  confAria2c.Aria2Port,
		token: confAria2c.Aria2Secret,
	}

	aria2cClient.serverUrl = fmt.Sprintf("http://%s:%d/jsonrpc", aria2cClient.Host, aria2cClient.port)

	return aria2cClient
}

func (aria2Client *Aria2Client) GenPayload4Download(method string, uri string, outDir, outFilename string) Aria2Payload {
	options := Aria2DownloadOption{
		Out: outFilename,
		Dir: outDir,
	}

	var params []interface{}
	params = append(params, "token:"+aria2Client.token)
	var urls []string
	urls = append(urls, uri)
	params = append(params, urls)
	params = append(params, options)

	payload := Aria2Payload{
		JsonRpc: "2.0",
		Id:      uri,
		Method:  method,
		Params:  params,
	}

	return payload
}

func (aria2Client *Aria2Client) DownloadFile(uri string, outDir, outFilename string) *Aria2Download {
	payload := aria2Client.GenPayload4Download(ADD_URI, uri, outDir, outFilename)

	if utils.IsFileExists(outDir, outFilename) {
		utils.RemoveFile(outDir, outFilename)
	}

	response := libclient.HttpPostNoToken(aria2Client.serverUrl, payload)
	aria2Download := &Aria2Download{}
	err := json.Unmarshal([]byte(response), aria2Download)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return aria2Download
}

func (aria2Client *Aria2Client) GenPayload4Status(gid string) Aria2Payload {
	var params []interface{}
	params = append(params, "token:"+aria2Client.token)
	params = append(params, gid)

	payload := Aria2Payload{
		JsonRpc: "2.0",
		Method:  STATUS,
		Params:  params,
	}

	return payload
}

func (aria2Client *Aria2Client) GetDownloadStatus(gid string) *Aria2Status {
	payload := aria2Client.GenPayload4Status(gid)
	response := libclient.HttpPostNoToken(aria2Client.serverUrl, payload)
	//logs.GetLogger().Info(gid, " download status:", response)

	aria2Status := &Aria2Status{}
	err := json.Unmarshal([]byte(response), aria2Status)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return aria2Status
}

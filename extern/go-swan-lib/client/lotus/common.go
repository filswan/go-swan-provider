package lotus

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/logs"
)

const (
	LOTUS_JSON_RPC_ID      = 7878
	LOTUS_JSON_RPC_VERSION = "2.0"
)

type LotusJsonRpcParams struct {
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type LotusJsonRpcResult struct {
	Id      int           `json:"id"`
	JsonRpc string        `json:"jsonrpc"`
	Error   *JsonRpcError `json:"error"`
}

type JsonRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Cid struct {
	Cid string `json:"/"`
}

const (
	LOTUS_VERSION = "Filecoin.Version"
)

type LotusVersionResult struct {
	Version    string
	APIVersion int
	BlockDelay int
}

type LotusVersionResponse struct {
	LotusJsonRpcResult
	Result LotusVersionResult `json:"result"`
}

//when using lotus node api url it returns version of lotus node
//when using lotus miner api url it returns version of lotus miner
func LotusVersion(apiUrl string) (*string, error) {
	var params []interface{}

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_VERSION,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	//here the api url should be miner's api url, need to change later on
	response, err := web.HttpGetNoToken(apiUrl, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	lotusVersionResponse := &LotusVersionResponse{}
	err = json.Unmarshal(response, lotusVersionResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if lotusVersionResponse.Error != nil {
		msg := fmt.Sprintf("error, code:%d, message:%s", lotusVersionResponse.Error.Code, lotusVersionResponse.Error.Message)
		err := errors.New(msg)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return &lotusVersionResponse.Result.Version, nil
}

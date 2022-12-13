package lotus

import (
	"encoding/json"
	"fmt"

	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/logs"
)

const (
	FILECOIN_AUTH_VERIFY = "Filecoin.AuthVerify"
)

type AuthVerify struct {
	LotusJsonRpcResult
	Result []string `json:"result"`
}

func LotusCheckAuth(apiUrl, token, expectedAuth string) (bool, error) {
	auths, err := LotusAuthVerify(apiUrl, token)
	if err != nil {
		logs.GetLogger().Error(err)
		return false, err
	}
	for _, auth := range auths {
		if auth == expectedAuth {
			return true, nil
		}
	}
	return false, nil
}

func LotusAuthVerify(apiUrl, token string) ([]string, error) {
	if len(apiUrl) == 0 {
		err := fmt.Errorf("api url is required")
		logs.GetLogger().Error(err)
		return nil, err
	}

	if len(token) == 0 {
		err := fmt.Errorf("token is required")
		logs.GetLogger().Error(err)
		return nil, err
	}

	var params []interface{}
	params = append(params, token)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  FILECOIN_AUTH_VERIFY,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	//here the api url should be miner's api url, need to change later on
	response, err := web.HttpGetNoToken(apiUrl, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	authVerify := &AuthVerify{}
	err = json.Unmarshal(response, authVerify)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if authVerify.Error != nil {
		err := fmt.Errorf("error, code:%d, message:%s", authVerify.Error.Code, authVerify.Error.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return authVerify.Result, nil
}

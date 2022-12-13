package swan

import (
	"fmt"
	"strings"

	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
)

type LoginByEmailParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginByApikeyParams struct {
	Apikey      string `json:"apikey"`
	AccessToken string `json:"access_token"`
}

func (swanClient *SwanClient) GetJwtTokenByApiKey() error {
	data := LoginByApikeyParams{
		Apikey:      swanClient.ApiKey,
		AccessToken: swanClient.AccessToken,
	}

	if len(swanClient.ApiUrl) == 0 {
		err := fmt.Errorf("api url is required")
		logs.GetLogger().Error(err)
		return err
	}

	if len(data.Apikey) == 0 {
		err := fmt.Errorf("apikey is required")
		logs.GetLogger().Error(err)
		return err
	}

	if len(data.AccessToken) == 0 {
		err := fmt.Errorf("acess token is required")
		logs.GetLogger().Error(err)
		return err
	}

	apiUrl := utils.UrlJoin(swanClient.ApiUrl, "user/login_by_apikey")

	response, err := web.HttpPostNoToken(apiUrl, data)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if strings.Contains(string(response), "fail") {
		message := utils.GetFieldStrFromJson(response, "message")
		status := utils.GetFieldStrFromJson(response, "status")
		err := fmt.Errorf("status:%s, message:%s", status, message)
		logs.GetLogger().Error(err)

		if message == "apikey not found" {
			logs.GetLogger().Error("please check your api key")
		}

		if message == "access token wrong" {
			logs.GetLogger().Error("Please check your access token")
		}

		logs.GetLogger().Info("for more information about how to config, please check https://docs.filswan.com/run-swan-provider/config-swan-provider")

		return err
	}

	jwtData := utils.GetFieldMapFromJson(response, "data")
	if jwtData == nil {
		err := fmt.Errorf("error: fail to connect to swan api")
		logs.GetLogger().Error(err)
		return err
	}

	swanClient.SwanToken = jwtData["jwt_token"].(string)

	return nil
}

func (swanClient *SwanClient) GetJwtTokenUp3Times() error {
	if len(swanClient.ApiUrl) == 0 {
		err := fmt.Errorf("api url is required")
		logs.GetLogger().Error(err)
		return err
	}

	if len(swanClient.ApiKey) == 0 {
		err := fmt.Errorf("api key is required")
		logs.GetLogger().Error(err)
		return err
	}

	if len(swanClient.AccessToken) == 0 {
		err := fmt.Errorf("access token is required")
		logs.GetLogger().Error(err)
		return err
	}

	var err error
	for i := 0; i < 3; i++ {
		err = swanClient.GetJwtTokenByApiKey()
		if err == nil {
			break
		}
		logs.GetLogger().Error(err)
	}

	if err != nil {
		err = fmt.Errorf("failed to connect to swan platform after trying 3 times")
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

package swan

import (
	"fmt"

	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
)

type SwanClient struct {
	ApiUrl      string
	SwanToken   string
	ApiKey      string
	AccessToken string
}
type SwanServerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func GetClient(apiUrl, apiKey, accessToken, swanToken string) (*SwanClient, error) {
	if len(apiUrl) == 0 {
		err := fmt.Errorf("api url is required")
		logs.GetLogger().Error(err)
		return nil, err
	}

	swanClient := &SwanClient{
		ApiUrl:      apiUrl,
		ApiKey:      apiKey,
		AccessToken: accessToken,
		SwanToken:   swanToken,
	}

	if swanToken == constants.EMPTY_STRING {
		err := swanClient.GetJwtTokenUp3Times()
		return swanClient, err
	}

	return swanClient, nil
}

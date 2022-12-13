package swan

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
)

func (swanClient *SwanClient) CheckDatacap(wallet string) (bool, error) {
	apiUrl := swanClient.ApiUrl + "/tools/check_datacap?address=" + wallet
	params := url.Values{}

	response, err := web.HttpGetNoToken(apiUrl, strings.NewReader(params.Encode()))

	if err != nil {
		logs.GetLogger().Error(err)
		return false, err
	}

	status := utils.GetFieldStrFromJson(response, "status")

	if !strings.EqualFold(status, constants.SWAN_API_STATUS_SUCCESS) {
		message := utils.GetFieldStrFromJson(response, "message")
		err := fmt.Errorf("error:%s,%s", status, message)
		logs.GetLogger().Error(err)
		return false, err
	}

	data := utils.GetFieldMapFromJson(response, "data")
	isVerified := data["is_verified"].(bool)

	return isVerified, nil
}

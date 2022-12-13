package swan

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/model"
)

type GetCarFileByUuidUrlResult struct {
	Data   GetCarFileByUuidUrlResultData `json:"data"`
	Status string                        `json:"status"`
}
type GetCarFileByUuidUrlResultData struct {
	CarFile          model.CarFile        `json:"car_file"`
	OfflineDeals     []*model.OfflineDeal `json:"offline_deals"`
	TotalItems       int                  `json:"total_items"`
	TotalTaskCount   int                  `json:"total_task_count"`
	BidCount         int                  `json:"bid_count"`
	DealCompleteRate string               `json:"deal_complete_rate"`
}

func (swanClient *SwanClient) GetCarFileByUuidUrl(taskUuid, carFileUrl string) (*GetCarFileByUuidUrlResultData, error) {
	if len(taskUuid) == 0 {
		err := fmt.Errorf("please provide task uuid")
		logs.GetLogger().Error(err)
		return nil, err
	}

	if len(carFileUrl) == 0 {
		err := fmt.Errorf("please provide car file url")
		logs.GetLogger().Error(err)
		return nil, err
	}

	apiUrl := fmt.Sprintf("%s/car_files/car_file?task_uuid=%s&car_file_url=%s", swanClient.ApiUrl, taskUuid, carFileUrl)

	response, err := web.HttpGet(apiUrl, swanClient.SwanToken, "")

	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	getCarFileByUuidUrlResult := &GetCarFileByUuidUrlResult{}
	err = json.Unmarshal(response, getCarFileByUuidUrlResult)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if !strings.EqualFold(getCarFileByUuidUrlResult.Status, constants.SWAN_API_STATUS_SUCCESS) {
		err := fmt.Errorf("error:%s", getCarFileByUuidUrlResult.Status)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return &getCarFileByUuidUrlResult.Data, nil
}

type GetAutoBidCarFilesByStatusResult struct {
	Data   GetAutoBidCarFilesByStatusResultData `json:"data"`
	Status string                               `json:"status"`
}
type GetAutoBidCarFilesByStatusResultData struct {
	CarFile          model.CarFile        `json:"car_file"`
	OfflineDeals     []*model.OfflineDeal `json:"offline_deals"`
	TotalItems       int                  `json:"total_items"`
	TotalTaskCount   int                  `json:"total_task_count"`
	BidCount         int                  `json:"bid_count"`
	DealCompleteRate string               `json:"deal_complete_rate"`
}

func (swanClient *SwanClient) GetAutoBidCarFilesByStatus(carFileStatus string) (*GetAutoBidCarFilesByStatusResultData, error) {
	carFileStatus = strings.Trim(carFileStatus, " ")
	if len(carFileStatus) == 0 {
		err := fmt.Errorf("please provide car file status")
		logs.GetLogger().Error(err)
		return nil, err
	}

	apiUrl := fmt.Sprintf("%s/car_files/auto_bid/get_by_status?car_file_status=%s", swanClient.ApiUrl, carFileStatus)

	response, err := web.HttpGet(apiUrl, swanClient.SwanToken, "")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	getAutoBidCarFilesByStatusResult := &GetAutoBidCarFilesByStatusResult{}
	err = json.Unmarshal(response, getAutoBidCarFilesByStatusResult)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if !strings.EqualFold(getAutoBidCarFilesByStatusResult.Status, constants.SWAN_API_STATUS_SUCCESS) {
		err := fmt.Errorf("error:%s", getAutoBidCarFilesByStatusResult.Status)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return &getAutoBidCarFilesByStatusResult.Data, nil
}

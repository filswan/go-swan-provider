package swan

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
)

func (swanClient *SwanClient) CreateTask(task model.Task, fileDescs []*model.FileDesc) (*SwanServerResponse, error) {
	apiUrl := utils.UrlJoin(swanClient.ApiUrl, "tasks/create_task")
	params := map[string]interface{}{
		"task":       task,
		"file_descs": fileDescs,
	}

	response, err := web.HttpPost(apiUrl, swanClient.SwanToken, params)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	swanServerResponse := &SwanServerResponse{}
	err = json.Unmarshal(response, swanServerResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if !strings.EqualFold(swanServerResponse.Status, constants.SWAN_API_STATUS_SUCCESS) {
		err := fmt.Errorf("error:%s,%s", swanServerResponse.Status, swanServerResponse.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return swanServerResponse, nil
}

type GetTaskResult struct {
	Data   GetTaskResultData `json:"data"`
	Status string            `json:"status"`
}

type GetTaskResultData struct {
	Task           []model.Task `json:"task"`
	TotalItems     int          `json:"total_items"`
	TotalTaskCount int          `json:"total_task_count"`
}

func (swanClient *SwanClient) GetTasks(limit *int, status *string) (*GetTaskResult, error) {
	apiUrl := utils.UrlJoin(swanClient.ApiUrl, "tasks")
	filters := ""
	if limit != nil {
		filters = filters + "?limit=" + strconv.Itoa(*limit)
	}

	if status != nil {
		if filters == "" {
			filters = filters + "?"
		} else {
			filters = filters + "&"
		}
		filters = filters + "status=" + *status
	}

	apiUrl = apiUrl + filters

	response, err := web.HttpGet(apiUrl, swanClient.SwanToken, "")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	getTaskResult := &GetTaskResult{}
	err = json.Unmarshal(response, getTaskResult)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if !strings.EqualFold(getTaskResult.Status, constants.SWAN_API_STATUS_SUCCESS) {
		err := fmt.Errorf("error:%s", getTaskResult.Status)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return getTaskResult, nil
}

func (swanClient *SwanClient) GetAllTasks(status string) ([]model.Task, error) {
	limit := -1
	getTaskResult, err := swanClient.GetTasks(&limit, &status)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return getTaskResult.Data.Task, err
}

type GetTaskByUuidResult struct {
	Data   GetTaskByUuidResultData `json:"data"`
	Status string                  `json:"status"`
}
type GetTaskByUuidResultData struct {
	//AverageBid       string              `json:"average_bid"`
	Task             model.Task           `json:"task"`
	CarFiles         []model.CarFile      `json:"car_file"`
	Miner            model.Miner          `json:"miner"`
	Deal             []*model.OfflineDeal `json:"deal"`
	TotalItems       int                  `json:"total_items"`
	TotalTaskCount   int                  `json:"total_task_count"`
	BidCount         int                  `json:"bid_count"`
	DealCompleteRate string               `json:"deal_complete_rate"`
}

func (swanClient *SwanClient) GetTaskByUuid(taskUuid string) (*GetTaskByUuidResult, error) {
	if len(taskUuid) == 0 {
		err := fmt.Errorf("please provide task uuid")
		logs.GetLogger().Error(err)
		return nil, err
	}
	apiUrl := utils.UrlJoin(swanClient.ApiUrl, "tasks", taskUuid)

	response, err := web.HttpGet(apiUrl, swanClient.SwanToken, "")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	getTaskByUuidResult := &GetTaskByUuidResult{}
	err = json.Unmarshal(response, getTaskByUuidResult)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if !strings.EqualFold(getTaskByUuidResult.Status, constants.SWAN_API_STATUS_SUCCESS) {
		err := fmt.Errorf("error:%s", getTaskByUuidResult.Status)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return getTaskByUuidResult, nil
}

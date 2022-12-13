package swan

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
)

type MinerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Miner model.Miner `json:"miner"`
	} `json:"data"`
}

func (swanClient *SwanClient) GetMiner(minerFid string) (*MinerResponse, error) {
	apiUrl := utils.UrlJoin(swanClient.ApiUrl, "miners", minerFid)

	response, err := web.HttpGetNoToken(apiUrl, "")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	minerResponse := &MinerResponse{}
	err = json.Unmarshal(response, minerResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if !strings.EqualFold(minerResponse.Status, constants.SWAN_API_STATUS_SUCCESS) {
		err := fmt.Errorf("status:%s, message:%s", minerResponse.Status, minerResponse.Message)
		logs.GetLogger().Error(err)
		return nil, err

	}

	return minerResponse, nil
}

type UpdateMinerConfigParams struct {
	MinerFid            string `json:"miner_fid"`
	BidMode             int    `json:"bid_mode"`
	ExpectedSealingTime int    `json:"expected_sealing_time"`
	StartEpoch          int    `json:"start_epoch"`
	AutoBidDealPerDay   int    `json:"auto_bid_deal_per_day"`
}

func (swanClient *SwanClient) UpdateMinerBidConf(minerFid string, confMiner model.Miner) error {
	err := swanClient.GetJwtTokenUp3Times()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	minerResponse, err := swanClient.GetMiner(minerFid)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if minerResponse == nil || minerResponse.Status != constants.SWAN_API_STATUS_SUCCESS {
		logs.GetLogger().Error("Error: Get miner information failed")
		return err
	}

	miner := minerResponse.Data.Miner

	if miner.BidMode == confMiner.BidMode &&
		miner.ExpectedSealingTime == confMiner.ExpectedSealingTime &&
		miner.StartEpoch == confMiner.StartEpoch &&
		miner.AutoBidDealPerDay == confMiner.AutoBidDealPerDay {
		logs.GetLogger().Info("No changes in bid configuration")
		return err
	}

	logs.GetLogger().Info("Begin updating bid configuration")
	apiUrl := utils.UrlJoin(swanClient.ApiUrl, "miners/update_miner_config")

	params := UpdateMinerConfigParams{
		MinerFid:            minerFid,
		BidMode:             confMiner.BidMode,
		ExpectedSealingTime: confMiner.ExpectedSealingTime,
		StartEpoch:          confMiner.StartEpoch,
		AutoBidDealPerDay:   confMiner.AutoBidDealPerDay,
	}

	response, err := web.HttpPost(apiUrl, swanClient.SwanToken, params)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	swanServerResponse := &SwanServerResponse{}
	err = json.Unmarshal(response, swanServerResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if !strings.EqualFold(minerResponse.Status, constants.SWAN_API_STATUS_SUCCESS) {
		err := fmt.Errorf("%s,%s", minerResponse.Status, minerResponse.Message)
		logs.GetLogger().Error(err)
		return err
	}

	logs.GetLogger().Info("Bid configuration updated.")
	return nil
}

type SetHeartbeatOnlineParams struct {
	MinerFid string `json:"miner_fid"`
}

func (swanClient *SwanClient) SendHeartbeatRequest(minerFid string) error {
	err := swanClient.GetJwtTokenUp3Times()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	apiUrl := utils.UrlJoin(swanClient.ApiUrl, "miners/set_heartbeat_online")
	params := &SetHeartbeatOnlineParams{
		MinerFid: minerFid,
	}

	response, err := web.HttpPost(apiUrl, swanClient.SwanToken, params)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	swanServerResponse := &SwanServerResponse{}
	err = json.Unmarshal([]byte(response), swanServerResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if !strings.EqualFold(swanServerResponse.Status, constants.SWAN_API_STATUS_SUCCESS) {
		err := fmt.Errorf("%s,%s", swanServerResponse.Status, swanServerResponse.Message)
		logs.GetLogger().Error(err)
		return err
	}

	msg := fmt.Sprintf("status:%s, message:%s", swanServerResponse.Status, swanServerResponse.Message)
	logs.GetLogger().Info(msg)
	return nil
}

package service

import (
	"swan-provider/config"
	"swan-provider/logs"
	"time"

	"github.com/filswan/go-swan-lib/client/swan"
	libmodel "github.com/filswan/go-swan-lib/model"
)

type SwanService struct {
	MinerFid             string
	ApiHeartbeatInterval time.Duration
}

func GetSwanService() *SwanService {
	mainConf := config.GetConfig().Main
	swanService := &SwanService{
		MinerFid:             mainConf.MinerFid,
		ApiHeartbeatInterval: mainConf.SwanApiHeartbeatInterval * time.Second,
	}

	return swanService
}

func (swanService *SwanService) SendHeartbeatRequest(swanClient *swan.SwanClient) {
	response := swanClient.SendHeartbeatRequest(swanService.MinerFid)
	logs.GetLogger().Info(response)
}

func (swanService *SwanService) UpdateBidConf(swanClient *swan.SwanClient) {
	confMiner := &libmodel.Miner{
		BidMode:             config.GetConfig().Bid.BidMode,
		ExpectedSealingTime: config.GetConfig().Bid.ExpectedSealingTime,
		StartEpoch:          config.GetConfig().Bid.StartEpoch,
		AutoBidTaskPerDay:   config.GetConfig().Bid.AutoBidTaskPerDay,
	}

	swanClient.UpdateMinerBidConf(swanService.MinerFid, *confMiner)
}

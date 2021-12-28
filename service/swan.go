package service

import (
	"swan-provider/config"
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

func (swanService *SwanService) SendHeartbeatRequest(swanClient *swan.SwanClient) error {
	err := swanClient.SendHeartbeatRequest(swanService.MinerFid)
	return err
}

func (swanService *SwanService) UpdateBidConf(swanClient *swan.SwanClient) {
	confMiner := &libmodel.Miner{
		BidMode:             config.GetConfig().Bid.BidMode,
		ExpectedSealingTime: config.GetConfig().Bid.ExpectedSealingTime,
		StartEpoch:          config.GetConfig().Bid.StartEpoch,
		AutoBidDealPerDay:   config.GetConfig().Bid.AutoBidDealPerDay,
	}

	swanClient.UpdateMinerBidConf(swanService.MinerFid, *confMiner)
}

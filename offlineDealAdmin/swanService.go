package offlineDealAdmin

import (
	"swan-provider/common/utils"
	"swan-provider/config"
	"swan-provider/logs"
	"time"
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

func (swanService *SwanService) SendHeartbeatRequest(swanClient *utils.SwanClient) {
	response := swanClient.SendHeartbeatRequest(swanService.MinerFid)
	logs.GetLogger().Info(response)
}

func (swanService *SwanService) UpdateBidConf(swanClient *utils.SwanClient) {
	swanClient.UpdateMinerBidConf(swanService.MinerFid)
}

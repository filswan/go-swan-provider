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
	swanService := &SwanService {
		MinerFid: mainConf.MinerFid,
		ApiHeartbeatInterval: mainConf.SwanApiHeartbeatInterval,
	}

	return swanService
}

func (self *SwanService) SendHeartbeatRequest(swanClient *utils.SwanClient) {
	response := swanClient.SendHeartbeatRequest(self.MinerFid)
	logs.GetLogger().Info(response)
}

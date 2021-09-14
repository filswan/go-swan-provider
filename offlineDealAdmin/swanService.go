package offlineDealAdmin

import (
	"swan-provider/common/utils"
	"swan-provider/config"
	"swan-provider/logs"
)

func SendHeartbeatRequest(swanClient *utils.SwanClient) {
	minerFid := config.GetConfig().Main.MinerFid
	response := swanClient.SendHeartbeatRequest(minerFid)
	logs.GetLogger().Info(response)
}

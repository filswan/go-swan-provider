package commonRouters

import (
	"runtime"
	"swan-miner/common"
	"swan-miner/models"
)

func getSwanMinerHostInfo() *models.HostInfo {
	info := new(models.HostInfo)
	info.SwanMinerVersion = common.GetVersion()
	info.OperatingSystem = runtime.GOOS
	info.Architecture = runtime.GOARCH
	info.CPUnNumber = runtime.NumCPU()
	return info
}

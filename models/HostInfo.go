package models

type HostInfo struct {
	SwanMinerVersion string `json:"swan_miner_version"`
	OperatingSystem  string `json:"operating_system"`
	Architecture     string `json:"architecture"`
	CPUnNumber       int    `json:"cpu_number"`
}

package config

import (
	"log"
	"swan-provider/logs"
	"time"

	"github.com/BurntSushi/toml"
)

type Configuration struct {
	Port  int   `toml:"port"`
	Lotus lotus `toml:"lotus"`
	Aria2 aria2 `toml:"aria2"`
	Main  main  `toml:"main"`
	Bid   bid   `toml:"bid"`
}

type lotus struct {
	ApiUrl           string `toml:"api_url"`
	MinerApiUrl      string `toml:"miner_api_url"`
	MinerAccessToken string `toml:"miner_access_token"`
}

type aria2 struct {
	Aria2DownloadDir string `toml:"aria2_download_dir"`
	Aria2Host        string `toml:"aria2_host"`
	Aria2Port        int    `toml:"aria2_port"`
	Aria2Secret      string `toml:"aria2_secret"`
}

type main struct {
	SwanApiUrl               string        `toml:"api_url"`
	SwanApiKey               string        `toml:"api_key"`
	SwanAccessToken          string        `toml:"access_token"`
	SwanApiHeartbeatInterval time.Duration `toml:"api_heartbeat_interval"`
	MinerFid                 string        `toml:"miner_fid"`
	LotusImportInterval      time.Duration `toml:"import_interval"`
	LotusScanInterval        time.Duration `toml:"scan_interval"`
}

type bid struct {
	BidMode             int `toml:"bid_mode"`
	ExpectedSealingTime int `toml:"expected_sealing_time"`
	StartEpoch          int `toml:"start_epoch"`
	AutoBidTaskPerDay   int `toml:"auto_bid_task_per_day"`
}

var config *Configuration

func InitConfig() {
	configFile := generateConfigFile()
	logs.GetLogger().Info("Your config file is:", configFile)

	if metaData, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatal("error:", err)
	} else {
		if !requiredFieldsAreGiven(metaData) {
			log.Fatal("required fields not given")
		}
	}
}

func GetConfig() Configuration {
	if config == nil {
		InitConfig()
	}
	return *config
}

func requiredFieldsAreGiven(metaData toml.MetaData) bool {
	requiredFields := [][]string{
		{"port"},

		{"lotus"},
		{"aria2"},
		{"main"},
		{"bid"},

		{"lotus", "api_url"},
		{"lotus", "miner_api_url"},
		{"lotus", "miner_access_token"},

		{"aria2", "aria2_download_dir"},
		{"aria2", "aria2_host"},
		{"aria2", "aria2_port"},
		{"aria2", "aria2_secret"},

		{"main", "api_url"},
		{"main", "miner_fid"},
		{"main", "import_interval"},
		{"main", "scan_interval"},
		{"main", "api_key"},
		{"main", "access_token"},
		{"main", "api_heartbeat_interval"},

		{"bid", "bid_mode"},
		{"bid", "expected_sealing_time"},
		{"bid", "start_epoch"},
		{"bid", "auto_bid_task_per_day"},
	}

	for _, v := range requiredFields {
		if !metaData.IsDefined(v...) {
			log.Fatal("required conf fields ", v)
		}
	}

	return true
}

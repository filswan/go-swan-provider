package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/filswan/go-swan-lib/logs"
)

type Configuration struct {
	Port    int    `toml:"port"`
	Release bool   `toml:"release"`
	Lotus   lotus  `toml:"lotus"`
	Aria2   aria2  `toml:"aria2"`
	Main    main   `toml:"main"`
	Bid     bid    `toml:"bid"`
	Market  market `toml:"market"`
}

type lotus struct {
	ClientApiUrl      string `toml:"client_api_url"`
	MarketApiUrl      string `toml:"market_api_url"`
	MarketAccessToken string `toml:"market_access_token"`
}

type aria2 struct {
	Aria2DownloadDir         string `toml:"aria2_download_dir"`
	Aria2Host                string `toml:"aria2_host"`
	Aria2Port                int    `toml:"aria2_port"`
	Aria2Secret              string `toml:"aria2_secret"`
	Aria2AutoDeleteCarFile   bool   `toml:"aria2_auto_delete_car_file"`
	Aria2MaxDownloadingTasks int    `toml:"aria2_max_downloading_tasks"`
}

type main struct {
	SwanApiUrl               string        `toml:"api_url"`
	SwanApiKey               string        `toml:"api_key"`
	SwanAccessToken          string        `toml:"access_token"`
	SwanApiHeartbeatInterval time.Duration `toml:"api_heartbeat_interval"`
	MinerFid                 string        `toml:"miner_fid"`
	LotusImportInterval      time.Duration `toml:"import_interval"`
	LotusScanInterval        time.Duration `toml:"scan_interval"`
	MarketType               string        `toml:"market_type"`
}

type bid struct {
	BidMode             int `toml:"bid_mode"`
	ExpectedSealingTime int `toml:"expected_sealing_time"`
	StartEpoch          int `toml:"start_epoch"`
	AutoBidDealPerDay   int `toml:"auto_bid_deal_per_day"`
}
type market struct {
	FullNodeApi      string `toml:"full_node_api"`
	MinerApiInfo     string `toml:"miner_api_info"`
	CollateralWallet string `toml:"collateral_wallet"`
	PublishWallet    string `toml:"publish_wallet"`
	RpcUrl           string `toml:"rpc_url"`
	GraphqlUrl       string `toml:"graphql_url"`
}

var config *Configuration

func InitConfig() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Fatal("Cannot get home directory.")
	}

	configFile := filepath.Join(homedir, ".swan/provider/config.toml")

	logs.GetLogger().Info("Your config file is:", configFile)

	if metaData, err := toml.DecodeFile(configFile, &config); err != nil {
		logs.GetLogger().Fatal("error:", err)
	} else {
		if !requiredFieldsAreGiven(metaData) {
			logs.GetLogger().Fatal("required fields not given")
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
		{"release"},

		{"lotus"},
		{"aria2"},
		{"main"},
		{"bid"},
		{"market"},

		{"lotus", "client_api_url"},
		{"lotus", "market_api_url"},
		{"lotus", "market_access_token"},

		{"aria2", "aria2_download_dir"},
		{"aria2", "aria2_host"},
		{"aria2", "aria2_port"},
		{"aria2", "aria2_secret"},
		{"aria2", "aria2_max_downloading_tasks"},
		{"aria2", "aria2_auto_delete_car_file"},

		{"main", "api_url"},
		{"main", "miner_fid"},
		{"main", "import_interval"},
		{"main", "scan_interval"},
		{"main", "api_key"},
		{"main", "access_token"},
		{"main", "api_heartbeat_interval"},
		{"main", "market_type"},

		{"bid", "bid_mode"},
		{"bid", "expected_sealing_time"},
		{"bid", "start_epoch"},
		{"bid", "auto_bid_deal_per_day"},

		{"market", "full_node_api"},
		{"market", "miner_api_info"},
		{"market", "collateral_wallet"},
		{"market", "publish_wallet"},
	}

	for _, v := range requiredFields {
		if !metaData.IsDefined(v...) {
			logs.GetLogger().Fatal("required conf fields ", v)
		}
	}

	return true
}

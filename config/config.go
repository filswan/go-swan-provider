package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"swan-provider/common/constants"
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

type ConfigurationBak struct {
	Port    int      `toml:"port"`
	Release bool     `toml:"release"`
	Lotus   lotus    `toml:"lotus"`
	Aria2   aria2Bak `toml:"aria2"`
	Main    main     `toml:"main"`
	Bid     bid      `toml:"bid"`
	Market  market   `toml:"market"`
}

type lotus struct {
	ClientApiUrl      string `toml:"client_api_url"`
	ClientApiToken    string `toml:"client_api_token"`
	MarketApiUrl      string `toml:"market_api_url"`
	MarketAccessToken string `toml:"market_access_token"`
}

type aria2 struct {
	Aria2DownloadDir         string   `toml:"aria2_download_dir"`
	Aria2Host                string   `toml:"aria2_host"`
	Aria2Port                int      `toml:"aria2_port"`
	Aria2Secret              string   `toml:"aria2_secret"`
	Aria2AutoDeleteCarFile   bool     `toml:"aria2_auto_delete_car_file"`
	Aria2MaxDownloadingTasks int      `toml:"aria2_max_downloading_tasks"`
	Aria2CandidateDirs       []string `toml:"aria2_candidate_dirs"`
}

type aria2Bak struct {
	Aria2DownloadDir         string `toml:"aria2_download_dir"`
	Aria2Host                string `toml:"aria2_host"`
	Aria2Port                int    `toml:"aria2_port"`
	Aria2Secret              string `toml:"aria2_secret"`
	Aria2AutoDeleteCarFile   bool   `toml:"aria2_auto_delete_car_file"`
	Aria2MaxDownloadingTasks int    `toml:"aria2_max_downloading_tasks"`
	Aria2CandidateDirs       string `toml:"aria2_candidate_dirs"`
}

type main struct {
	SwanApiUrl                string        `toml:"api_url"`
	SwanApiKey                string        `toml:"api_key"`
	SwanAccessToken           string        `toml:"access_token"`
	SwanApiHeartbeatInterval  time.Duration `toml:"api_heartbeat_interval"`
	MinerFid                  string        `toml:"miner_fid"`
	LotusImportInterval       time.Duration `toml:"import_interval"`
	LotusScanInterval         time.Duration `toml:"scan_interval"`
	MarketVersion             string        `toml:"market_version"`
	LotusConcurrentImportings uint32        `toml:"concurrent_importings"`
}

type bid struct {
	BidMode             int `toml:"bid_mode"`
	ExpectedSealingTime int `toml:"expected_sealing_time"`
	StartEpoch          int `toml:"start_epoch"`
	AutoBidDealPerDay   int `toml:"auto_bid_deal_per_day"`
}
type market struct {
	FullNodeApi      string
	MinerApi         string
	CollateralWallet string `toml:"collateral_wallet"`
	PublishWallet    string `toml:"publish_wallet"`
	RpcUrl           string
	GraphqlUrl       string
	Repo             string
	BoostLog         string
}

var config *Configuration

func InitConfig() {
	swanPath, exist := os.LookupEnv("SWAN_PATH")
	var basePath, configFile string
	if exist {
		configFile = filepath.Join(swanPath, "provider/config.toml")
		basePath = filepath.Join(swanPath, "provider")
	} else {
		homedir, err := os.UserHomeDir()
		if err != nil {
			logs.GetLogger().Fatal("Cannot get home directory.")
		}
		configFile = filepath.Join(homedir, ".swan/provider/config.toml")
		basePath = filepath.Join(homedir, ".swan/provider")
	}

	logs.GetLogger().Info("Your config file is:", configFile)

	metaData, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		var configBak *ConfigurationBak
		metaData, err = toml.DecodeFile(configFile, &configBak)
		if err == nil {
			assignConfig(config, configBak)
		} else {
			logs.GetLogger().Fatal("error:", err)
		}
	}

	dirs := config.Aria2.Aria2CandidateDirs
	newDirs := make([]string, 0)
	for _, strPath := range dirs {
		newDirs = append(newDirs, strings.TrimSpace(strPath))
	}
	config.Aria2.Aria2CandidateDirs = newDirs

	if !requiredFieldsAreGiven(metaData) {
		logs.GetLogger().Fatal("required fields not given")
	}

	config.Market.Repo = filepath.Join(basePath, "boost")
	config.Market.BoostLog = filepath.Join(basePath, "boost.log")

	fullNodeApi, err := ChangeToFull(config.Lotus.ClientApiUrl, config.Lotus.ClientApiToken)
	if err != nil {
		logs.GetLogger().Fatal(err)
		return
	}
	minerApi, err := ChangeToFull(config.Lotus.MarketApiUrl, config.Lotus.MarketAccessToken)
	if err != nil {
		logs.GetLogger().Fatal(err)
		return
	}

	config.Market.MinerApi = minerApi
	config.Market.FullNodeApi = fullNodeApi
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
		{"lotus", "client_api_token"},

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
		{"main", "market_version"},

		{"bid", "bid_mode"},
		{"bid", "expected_sealing_time"},
		{"bid", "start_epoch"},
		{"bid", "auto_bid_deal_per_day"},

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

func GetRpcInfoByFile(configPath string) (string, string, error) {
	var config struct {
		API struct {
			ListenAddress string
		}
		Graphql struct {
			Port uint64
		}
	}

	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return "", "", err
	}

	var rpcUrl string
	splits := strings.Split(config.API.ListenAddress, "/")
	if len(splits) == 0 {
		rpcUrl = fmt.Sprintf("127.0.0.1:%d", constants.DEFAULT_API_PORT)
	} else {
		rpcUrl = fmt.Sprintf("127.0.0.1:%s", splits[4])
	}

	if config.Graphql.Port == 0 {
		config.Graphql.Port = constants.DEFAULT_GRAPHQL_PORT
	}
	graphqlUrl := fmt.Sprintf("http://127.0.0.1:%d/graphql/query", config.Graphql.Port)
	return rpcUrl, graphqlUrl, nil
}

func ChangeToFull(apiUrl, token string) (string, error) {
	u, err := url.Parse(apiUrl)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:/ip4/%s/tcp/%s/http", token, u.Hostname(), u.Port()), nil
}

func assignConfig(config *Configuration, configBak *ConfigurationBak) {
	config.Port = configBak.Port
	config.Release = configBak.Release

	config.Lotus.ClientApiUrl = configBak.Lotus.ClientApiUrl
	config.Lotus.ClientApiToken = configBak.Lotus.ClientApiToken
	config.Lotus.MarketApiUrl = configBak.Lotus.MarketApiUrl
	config.Lotus.MarketAccessToken = configBak.Lotus.MarketAccessToken
	config.Aria2.Aria2DownloadDir = configBak.Aria2.Aria2DownloadDir
	config.Aria2.Aria2Host = configBak.Aria2.Aria2Host
	config.Aria2.Aria2Port = configBak.Aria2.Aria2Port
	config.Aria2.Aria2Secret = configBak.Aria2.Aria2Secret
	config.Aria2.Aria2AutoDeleteCarFile = configBak.Aria2.Aria2AutoDeleteCarFile
	config.Aria2.Aria2MaxDownloadingTasks = configBak.Aria2.Aria2MaxDownloadingTasks

	splits := strings.Split(configBak.Aria2.Aria2CandidateDirs, ",")
	dirs := make([]string, 0)
	for _, strPath := range splits {
		dirs = append(dirs, strings.TrimSpace(strPath))
	}
	config.Aria2.Aria2CandidateDirs = dirs

	config.Main.SwanApiUrl = configBak.Main.SwanApiUrl
	config.Main.SwanApiKey = configBak.Main.SwanApiKey
	config.Main.SwanAccessToken = configBak.Main.SwanAccessToken
	config.Main.SwanApiHeartbeatInterval = configBak.Main.SwanApiHeartbeatInterval
	config.Main.MinerFid = configBak.Main.MinerFid
	config.Main.LotusImportInterval = configBak.Main.LotusImportInterval
	config.Main.LotusScanInterval = configBak.Main.LotusScanInterval
	config.Main.MarketVersion = configBak.Main.MarketVersion
	config.Main.LotusConcurrentImportings = configBak.Main.LotusConcurrentImportings

	config.Bid.BidMode = configBak.Bid.BidMode
	config.Bid.ExpectedSealingTime = configBak.Bid.ExpectedSealingTime
	config.Bid.StartEpoch = configBak.Bid.StartEpoch
	config.Bid.AutoBidDealPerDay = configBak.Bid.AutoBidDealPerDay

	config.Market.CollateralWallet = configBak.Market.CollateralWallet
	config.Market.PublishWallet = configBak.Market.PublishWallet
}

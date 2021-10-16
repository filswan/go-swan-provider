package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"swan-provider/common/utils"
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

func generateConfigFile() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Fatal("Cannot get home directory.")
	}

	targetDir := filepath.Join(homedir, ".swan/provider")
	err = os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Fatal("Cannot create .swan/provider directory under your home directory.")
	}

	configTargetFile := filepath.Join(targetDir, "config.toml")
	_, err = os.Stat(configTargetFile)
	if err != nil {
		pwdDir, err1 := os.Getwd()
		if err1 != nil {
			logs.GetLogger().Error(err)
			logs.GetLogger().Fatal("Cannot get your current directory.")
		}

		configSrcFile := filepath.Join(pwdDir, "config/config.toml")
		_, err2 := os.Stat(configSrcFile)
		if err2 == nil {
			logs.GetLogger().Info("Copying source config file:", configSrcFile, " to ", configTargetFile)
			_, err = utils.CopyFile(configSrcFile, configTargetFile)
			if err != nil {
				logs.GetLogger().Error(err)
				logs.GetLogger().Fatal("Cannot copy ", configSrcFile, " to ", configTargetFile)
			}
			logs.GetLogger().Info("Copy source config file:", configSrcFile, " to ", configTargetFile, " succeed.")

			return configTargetFile
		}
		downloadDir := filepath.Join(targetDir, "download")
		os.MkdirAll(downloadDir, os.ModePerm)
		logs.GetLogger().Info("Generating config file:", configTargetFile)
		configs := []string{
			"port = 8888",
			"",
			"[lotus]",
			"api_url=\"http://<ip>:<port>/rpc/v0\"   # Url of lotus web api, generally the <port> is 1234",
			"miner_api_url=\"http://<ip>:<port>/rpc/v0\"   # Url of lotus miner web api, generally the <port> is 2345",
			"miner_access_token=\"\"  # Access token of lotus miner web api",
			"",

			"[aria2]",
			fmt.Sprintf("aria2_download_dir = \"%s\"   # Directory where offline deal files will be downloaded for importing", downloadDir),
			"aria2_host = \"127.0.0.1\"  # Aria2 server address",
			"aria2_port = 6800         # Aria2 server port",
			"aria2_secret = \"my_aria2_secret\"  # Must be the same value as rpc-secure in aria2.conf",
			"",
			"[main]",
			"api_url = \"https://api.filswan.com\"  # Swan API address. For Swan production, it is \"https://api.filswan.com\"",
			"miner_fid = \"f0xxxx\"          # Your filecoin Miner ID",
			"import_interval = 600         # 600 seconds or 10 minutes. Importing interval between each deal.",
			"scan_interval = 600           # 600 seconds or 10 minutes. Time interval to scan all the ongoing deals and update status on Swan platform.",
			"api_key = \"\"                  # Your api key. Acquire from Filswan -> \"My Profile\"->\"Developer Settings\". You can also check the Guide.",
			"access_token = \"\"             # Your access token. Acquire from Filswan -> \"My Profile\"->\"Developer Settings\". You can also check the Guide.",
			"api_heartbeat_interval = 300  # 300 seconds or 5 minutes. Time interval to send heartbeat.",
			"",
			"[bid]",
			"bid_mode = 1                  # 0: manual, 1: auto",
			"expected_sealing_time = 1920  # 1920 epoch or 16 hours. The time expected for sealing deals. Deals starting too soon will be rejected.",
			"start_epoch = 2880            # 2880 epoch or 24 hours. Relative value to current epoch",
			"auto_bid_task_per_day = 20    # auto-bid task limit per day for your miner defined above",
		}

		utils.CreateFileWithContents(configTargetFile, configs)
		logs.GetLogger().Info("Generate config file:", configTargetFile, " succeed.")
	}

	return configTargetFile
}

package config

import (
	"github.com/BurntSushi/toml"
	"log"
	"time"
)

type Configuration struct {
	Port  string
	Dev   bool
	Aria2 aria2
	Main  main
}

type aria2 struct {
	DiskCache                    int         `toml:"disk-cache"`
	FileAllocation               string      `toml:"file-allocation"`
	IsContinue                   bool        `toml:"continue"`
	MaxTries                     int         `toml:"max-tries"`
	RpcListenPort                int         `toml:"rpc-listen-port"`
	MaxConcurrentDownloads       int         `toml:"max-concurrent-downloads"`
	MaxConnectionPerServer       int         `toml:"max-connection-per-server"`
	MinSplitSize                 string      `toml:"min-split-size"`
	Split                        int         `toml:"split"`
	DisableIpv6                  bool        `toml:"disable-ipv6"`
	AlwaysResume                 bool        `toml:"always-resume"`
	KeepUnfinishedDownloadResult bool        `toml:"keep-unfinished-download-result"`
	InputFile                    string      `toml:"input-file"`
	SaveSession                  string      `toml:"save-session"`
	SaveSessionInterval          int         `toml:"save-session-interval"`
	EnableRpc                    bool        `toml:"enable-rpc"`
	Pause                        bool        `toml:"pause"`
	RpcAllowOriginAll            bool        `toml:"rpc-allow-origin-all"`
	RpcListenAll                 bool        `toml:"rpc-listen-all"`
	RpcSaveUploadMetadata        bool        `toml:"rpc-save-upload-metadata"`
	RpcSecure                    bool        `toml:"rpc-secure"`
	RpcSecret                    string      `toml:"rpc-secret"`
	Aria2DownloadDir             string      `toml:"aria2_download_dir"`
	Aria2Conf                    string      `toml:"aria2_conf"`
	Aria2Host                    string      `toml:"aria2_host"`
	Aria2Port                    int         `toml:"aria2_port"`
	Aria2Secret                  string      `toml:"aria2_secret"`
}

type main struct {
	ApiUrl              string               `toml:"api_url"`
	MinerFid            string               `toml:"miner_fid"`
	ExpectedSealingTime int                  `toml:"expected_sealing_time"`
	ImportInterval      time.Duration        `toml:"import_interval"`
	ScanInterval        time.Duration        `toml:"scan_interval"`
	ApiKey              string               `toml:"api_key"`
	AccessToken         string               `toml:"access_token"`
}

var config *Configuration

func InitConfig() {
	//if strings.Trim(configFile, " ") == "" {
	configFile := "./config/config.toml"
	//}
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
	}

	for _, v := range requiredFields {
		if !metaData.IsDefined(v...) {
			log.Fatal("required fields ", v)
		}
	}

	return true
}

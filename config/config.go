package config

import (
	"github.com/BurntSushi/toml"
	"log"
	"time"
)

type Configuration struct {
	Port  string        `toml:"port"`
	Dev   bool          `toml:"dev"`
	Aria2 aria2         `toml:"aria2"`
	Main  main          `toml:"main"`
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
	SwanApiUrl               string        `toml:"api_url"`
	SwanApiKey               string        `toml:"api_key"`
	SwanAccessToken          string        `toml:"access_token"`
	SwanApiHeartbeatInterval time.Duration `toml:"api_heartbeat_interval"`
	MinerFid                 string        `toml:"miner_fid"`
	ExpectedSealingTime      int           `toml:"expected_sealing_time"`
	LotusImportInterval      time.Duration `toml:"import_interval"`
	LotusScanInterval        time.Duration `toml:"scan_interval"`
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
	requiredFields := [][]string {
		{"port"},
		{"dev"},

		{"aria2"},
		{"main"},

		{"aria2", "disk-cache"},
		{"aria2", "file-allocation"},
		{"aria2", "continue"},
		{"aria2", "max-tries"},
		{"aria2", "rpc-listen-port"},
		{"aria2", "max-concurrent-downloads"},
		{"aria2", "max-connection-per-server"},
		{"aria2", "min-split-size"},
		{"aria2", "split"},
		{"aria2", "disable-ipv6"},
		{"aria2", "always-resume"},
		{"aria2", "keep-unfinished-download-result"},
		{"aria2", "input-file"},
		{"aria2", "save-session"},
		{"aria2", "save-session-interval"},
		{"aria2", "enable-rpc"},
		{"aria2", "pause"},
		{"aria2", "rpc-allow-origin-all"},
		{"aria2", "rpc-listen-all"},
		{"aria2", "rpc-save-upload-metadata"},
		{"aria2", "rpc-secure"},
		{"aria2", "rpc-secret"},
		{"aria2", "aria2_download_dir"},
		{"aria2", "aria2_conf"},
		{"aria2", "aria2_host"},
		{"aria2", "aria2_port"},
		{"aria2", "aria2_secret"},

		{"main", "api_url"},
		{"main", "miner_fid"},
		{"main", "expected_sealing_time"},
		{"main", "import_interval"},
		{"main", "scan_interval"},
		{"main", "api_key"},
		{"main", "access_token"},
		{"main", "api_heartbeat_interval"},
	}

	for _, v := range requiredFields {
		if !metaData.IsDefined(v...) {
			log.Fatal("required conf fields ", v)
		}
	}

	return true
}

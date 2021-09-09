package config

import (
	"github.com/BurntSushi/toml"
	"log"
	"strings"
	"time"
)

type Configuration struct {
	Port  string
	Dev   bool
	Aria2 aria2
	Main  main
}

type aria2 struct {
	DiskCache                    int
	FileAllocation               string
	IsContinue                   bool
	MaxTries                     int
	RpcListenPort                int
	MaxConcurrentDownloads       int
	MaxConnectionPerServer       int
	MinSplitSize                 string
	Split                        int
	DisableIpv6                  bool
	AlwaysResume                 bool
	KeepUnfinishedDownloadResult bool
	InputFile                    string
	SaveSession                  string
	SaveSessionInterval          int
	EnableRpc                    bool
	Pause                        bool
	RpcAllowOriginAll            bool
	RpcListenAll                 bool
	RpcSaveUploadMetadata        bool
	RpcSecure                    bool
	RpcSecret                    string
	Aria2DownloadDir             string
	Aria2Conf                    string
	Aria2Host                    string
	Aria2Port                    int
	Aria2Secret                  string
}

type main struct {
	ApiUrl              string
	MinerFid            string
	ExpectedSealingTime int
	ImportInterval      time.Duration
	ScanInterval        time.Duration
	ApiKey              string
	AccessToken         string
}

var config *Configuration

func InitConfig(configFile string) {
	if strings.Trim(configFile, " ") == "" {
		configFile = "./config/config.toml"
	}
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
		InitConfig("")
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

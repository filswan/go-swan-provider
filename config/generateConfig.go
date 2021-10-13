package config

import (
	"fmt"
	"os"
	"path/filepath"
	"swan-provider/common/utils"
	"swan-provider/logs"
)

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
			"api_url=\"http://192.168.88.41:1234/rpc/v0\"   # Url of lotus web api",
			"miner_api_url=\"http://192.168.88.41:2345/rpc/v0\"   # Url of lotus miner web api",
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

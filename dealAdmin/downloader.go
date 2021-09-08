package dealAdmin

import (
	"fmt"
	"swan-miner/common/utils"
	"swan-miner/config"
	"github.com/jasonlvhit/gocron"
)


const MAX_DOWNLOADING_TASKS = 10

func Downloader(){
	conf := config.GetConfig()
	confMain := conf.Main
	confAria := conf.Aria2

	minerFild := confMain.MinerFid
	aria2Host := confAria.Aria2Host
	aria2Port := confAria.Aria2Port
/*	aria2Secret := confAria.Aria2Secret
	ariaConf := confAria.Aria2Conf*/
	outDir := confAria.Aria2DownloadDir
	apiUrl := confMain.ApiUrl
	apiKey := confMain.ApiKey
	accessToken := confMain.AccessToken

	aria2Client := Aria2c{
		host: aria2Host,
		port: string(aria2Port),
		token: accessToken,
	}

	swanClient := &utils.SwanClient{
		apiUrl,
		apiKey,
		accessToken,
	}

	gocron.Every(1).Minute().Do(func (){
		fmt.Println(1)
		startDownloading(MAX_DOWNLOADING_TASKS, minerFild, outDir, aria2Client, swanClient)
	})
}

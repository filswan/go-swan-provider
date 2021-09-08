package dealAdmin

import (
	"fmt"
	"github.com/jasonlvhit/gocron"
	"swan-miner/config"
)


const MAX_DOWNLOADING_TASKS = 10

func Downloader(){
	conf := config.GetConfig()
	confMain := conf.Main
	confAria := conf.Aria2

	minerFild := confMain.MinerFid
/*	aria2Secret := confAria.Aria2Secret
	ariaConf := confAria.Aria2Conf*/
	outDir := confAria.Aria2DownloadDir
	apiUrl := confMain.ApiUrl
	apiKey := confMain.ApiKey
	accessToken := confMain.AccessToken

	aria2Client := GetAria2Client()

	swanClient := &SwanClient{
		apiUrl,
		apiKey,
		accessToken,
	}

	gocron.Every(1).Minute().Do(func (){
		fmt.Println(1)
		startDownloading(MAX_DOWNLOADING_TASKS, minerFild, outDir, aria2Client, swanClient)
	})
}

package dealAdmin

import (
	"github.com/jasonlvhit/gocron"
	"swan-miner/common/utils"
)


const MAX_DOWNLOADING_TASKS = 10

func Downloader(){
	aria2Client := utils.GetAria2Client()
	swanClient := GetSwanClient()
	aria2Service := GetAria2Service()

	gocron.Every(1).Minute().Do(func (){
		//fmt.Println(1)
		aria2Service.CheckDownloadStatus(aria2Client,swanClient)
	})

	aria2Service.startDownloading(MAX_DOWNLOADING_TASKS, aria2Client, swanClient)
}

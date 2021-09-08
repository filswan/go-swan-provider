package dealAdmin

import (
	"fmt"
	"github.com/jasonlvhit/gocron"
)


const MAX_DOWNLOADING_TASKS = 10

func Downloader(){
	aria2Client := GetAria2Client()
	swanClient := GetSwanClient()

	gocron.Every(1).Minute().Do(func (){
		fmt.Println(1)
		startDownloading(MAX_DOWNLOADING_TASKS, aria2Client, swanClient)
	})
}

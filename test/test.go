package test

import (
	"time"

	"github.com/filswan/go-swan-lib/logs"
)

func GetCurrentEpoch() int {
	currentNanoSec := time.Now().UnixNano()
	currentEpoch := (currentNanoSec/1e9 - 1598306471) / 30
	logs.GetLogger().Info(currentEpoch)
	return int(currentEpoch)
}

//func TestAriaClient() {
//	swanClient := client.GetSwanClient()
//
//	aria2Client := client.GetAria2Client()
//	offlineDeal := &models.OfflineDeal{
//		Id:            163,
//		UserId:        163,
//		SourceFileUrl: "https://file-examples-com.github.io/uploads/2020/03/file_example_WEBP_500kB.webp",
//	}
//
//	aria2Service := service.GetAria2Service()
//	aria2Service.StartDownload4Deal(offlineDeal, aria2Client, swanClient)
//	aria2Client.GetDownloadStatus("f80d913a4dff40651")
//}

//func TestDownloader() {
//	aria2Client := client.GetAria2Client()
//	swanClient := client.GetSwanClient()
//	aria2Service := service.GetAria2Service()
//	aria2Service.StartDownload(aria2Client, swanClient)
//	aria2Service.CheckDownloadStatus(aria2Client, swanClient)
//}

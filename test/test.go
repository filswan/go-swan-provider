package test

import (
	"strings"
	"swan-provider/common/client"
	"swan-provider/common/utils"
	"swan-provider/config"
	"swan-provider/logs"
	"swan-provider/models"
	"swan-provider/service"
	"time"
)

type Todo struct {
	UserID    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func GetCurrentEpoch() int {
	currentNanoSec := time.Now().UnixNano()
	currentEpoch := (currentNanoSec/1e9 - 1598306471) / 30
	logs.GetLogger().Info(currentEpoch)
	return int(currentEpoch)
}

func TestRestApiClient() {
	response := client.HttpGet("https://jsonplaceholder.typicode.com/todos/1", "", "")
	logs.GetLogger().Info(response)

	todo := Todo{1, 2, "lorem ipsum dolor sit amet", true}
	response = client.HttpPostNoToken("https://jsonplaceholder.typicode.com/todos", todo)
	logs.GetLogger().Info(response)

	response = client.HttpPut("https://jsonplaceholder.typicode.com/todos/1", "", todo)
	logs.GetLogger().Info(response)

	title := utils.GetFieldFromJson(response, "title")
	logs.GetLogger().Info(title)

	response = client.HttpDelete("https://jsonplaceholder.typicode.com/todos/1", "", todo)
	logs.GetLogger().Info(response)
}

func TestSwanClient() {
	swanClient := client.GetSwanClient()
	mainConf := config.GetConfig().Main
	deals := swanClient.GetOfflineDeals(mainConf.MinerFid, "Downloading", "10")
	logs.GetLogger().Info(deals)

	response := swanClient.UpdateOfflineDealStatus(2455, "Downloaded", "test note")
	logs.GetLogger().Info(response)

	response = swanClient.UpdateOfflineDealStatus(2455, "Completed", "test note", "/test/test", "0003222")
	logs.GetLogger().Info(response)
}

func TestAriaClient() {
	swanClient := client.GetSwanClient()

	aria2Client := client.GetAria2Client()
	offlineDeal := &models.OfflineDeal{
		Id:            163,
		UserId:        163,
		SourceFileUrl: "https://file-examples-com.github.io/uploads/2020/03/file_example_WEBP_500kB.webp",
	}

	aria2Service := service.GetAria2Service()
	aria2Service.StartDownload4Deal(offlineDeal, aria2Client, swanClient)
	aria2Client.GetDownloadStatus("f80d913a4dff40651")
}

func TestDownloader() {
	aria2Client := client.GetAria2Client()
	swanClient := client.GetSwanClient()
	aria2Service := service.GetAria2Service()
	aria2Service.StartDownload(aria2Client, swanClient)
	aria2Service.CheckDownloadStatus(aria2Client, swanClient)
}

func TestOsCmdClient() {
	result, err := client.ExecOsCmd("ls -l")
	logs.GetLogger().Info(result, err)

	result, err = client.ExecOsCmd("pwd")
	logs.GetLogger().Info(result, err)

	result, err = client.ExecOsCmd("ls -l | grep common")
	logs.GetLogger().Info(result, err)

	words := strings.Fields(result)
	for _, word := range words {
		logs.GetLogger().Info(word)
	}
}

func TestOsCmdClient1() {
	/*result, err := */ client.ExecOsCmd2Screen("ls -l")
	//logs.GetLogger().Info(result, err)

	/*result, err = */
	client.ExecOsCmd2Screen("pwd")
	//logs.GetLogger().Info(result, err)

	/*result, err = */
	client.ExecOsCmd2Screen("ls -l | grep x")
	//logs.GetLogger().Info(result, err)
}

func TestSendHeartbeatRequest() {
	minerFid := config.GetConfig().Main.MinerFid

	swanClient := client.GetSwanClient()

	response := swanClient.SendHeartbeatRequest(minerFid)
	logs.GetLogger().Info(response)
}

func TestLotusClient() {
	currentEpoch := client.LotusGetCurrentEpoch()
	logs.GetLogger().Info("currentEpoch: ", currentEpoch)
	status, message := client.LotusGetDealOnChainStatus("bafyreigbcdmozbfyr5sfipu7xm4fj23r3g2idgk7jibaku4y4r2z4x55bq")
	logs.GetLogger().Info("status: ", status)
	logs.GetLogger().Info("message: ", message)
	message = client.LotusImportData("bafyreiaj7av2qgziwfyvo663a2kjg3n35rvfr2i5r2dyrexxukdbybz7ky", "/tmp/swan-downloads/185/202107/go1.15.5.linux-amd64.tar.gz.car")
	logs.GetLogger().Info("message: ", message)
	message = client.LotusImportData("bafyreia5qflut2hqbwfhhhiybes5uhnx6aehgg3ltvam2aqbkekkyuoboy", "/tmp/swan-downloads/185/202107/go1.15.5.linux-amd64.tar.gz.car")
	logs.GetLogger().Info("message: ", message)
}

func Test() {
	TestLotusClient()
}

package test

import (
	"strings"
	"swan-provider/common/client"
	"swan-provider/common/utils"
	"swan-provider/config"
	"swan-provider/logs"
	"swan-provider/models"
	"swan-provider/offlineDealAdmin"
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
	response := utils.HttpGet("https://jsonplaceholder.typicode.com/todos/1", "", "")
	logs.GetLogger().Info(response)

	todo := Todo{1, 2, "lorem ipsum dolor sit amet", true}
	response = utils.HttpPostNoToken("https://jsonplaceholder.typicode.com/todos", todo)
	logs.GetLogger().Info(response)

	response = utils.HttpPut("https://jsonplaceholder.typicode.com/todos/1", "", todo)
	logs.GetLogger().Info(response)

	title := utils.GetFieldFromJson(response, "title")
	logs.GetLogger().Info(title)

	response = utils.HttpDelete("https://jsonplaceholder.typicode.com/todos/1", "", todo)
	logs.GetLogger().Info(response)
}

func TestSwanClient() {
	swanClient := utils.GetSwanClient()
	mainConf := config.GetConfig().Main
	deals := swanClient.GetOfflineDeals(mainConf.MinerFid, "Downloading", "10")
	logs.GetLogger().Info(deals)

	response := swanClient.UpdateOfflineDealStatus(2455, "Downloaded", "test note")
	response = swanClient.UpdateOfflineDealStatus(2455, "Completed", "test note", "/test/test", "0003222")
	logs.GetLogger().Info(response)
}

func TestAriaClient() {
	swanClient := utils.GetSwanClient()

	aria2Client := utils.GetAria2Client()
	offlineDeal := &models.OfflineDeal{
		Id:            163,
		UserId:        163,
		SourceFileUrl: "https://file-examples-com.github.io/uploads/2020/03/file_example_WEBP_500kB.webp",
	}

	aria2Service := offlineDealAdmin.GetAria2Service()
	aria2Service.StartDownload4Deal(offlineDeal, aria2Client, swanClient)
	aria2Client.GetDownloadStatus("f80d913a4dff40651")
}

func TestDownloader() {
	aria2Client := utils.GetAria2Client()
	swanClient := utils.GetSwanClient()
	aria2Service := offlineDealAdmin.GetAria2Service()
	aria2Service.StartDownload(aria2Client, swanClient)
	aria2Service.CheckDownloadStatus(aria2Client, swanClient)
}

func TestOsCmdClient() {
	result, err := utils.ExecOsCmd("ls -l")
	logs.GetLogger().Info(result, err)

	result, err = utils.ExecOsCmd("pwd")
	logs.GetLogger().Info(result, err)

	result, err = utils.ExecOsCmd("ls -l | grep common")
	logs.GetLogger().Info(result, err)

	words := strings.Fields(result)
	for _, word := range words {
		logs.GetLogger().Info(word)
	}
}

func TestOsCmdClient1() {
	/*result, err := */ utils.ExecOsCmd2Screen("ls -l")
	//logs.GetLogger().Info(result, err)

	/*result, err = */
	utils.ExecOsCmd2Screen("pwd")
	//logs.GetLogger().Info(result, err)

	/*result, err = */
	utils.ExecOsCmd2Screen("ls -l | grep x")
	//logs.GetLogger().Info(result, err)
}

func TestSendHeartbeatRequest() {
	minerFid := config.GetConfig().Main.MinerFid

	swanClient := utils.GetSwanClient()

	response := swanClient.SendHeartbeatRequest(minerFid)
	logs.GetLogger().Info(response)
}

func TestLotusClient() {
	lotusClinet := client.LotusGetClient()
	currentEpoch := lotusClinet.GetCurrentEpoch()
	logs.GetLogger().Info("currentEpoch: ", currentEpoch)
	status, message := lotusClinet.LotusGetDealOnChainStatus("")
	logs.GetLogger().Info("status: ", status)
	logs.GetLogger().Info("message: ", message)
}

func Test() {
	TestLotusClient()
}

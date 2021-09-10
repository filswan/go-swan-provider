package test

import (
	"fmt"
	"strings"
	"swan-miner/common/utils"
	"swan-miner/config"
	"swan-miner/models"
	"swan-miner/offlineDealAdmin"
)

func TestRestApiAccessor() {
	response := utils.HttpGet("https://jsonplaceholder.typicode.com/todos/1", "", "")
	fmt.Println(response)
	todo := models.Todo{1, 2, "lorem ipsum dolor sit amet", true}
	response = utils.HttpPostNoToken("https://jsonplaceholder.typicode.com/todos", todo)
	fmt.Println(response)

	response = utils.HttpPut("https://jsonplaceholder.typicode.com/todos/1", "",todo)
	fmt.Println(response)

	title := utils.GetFieldFromJson(response,"title")
	fmt.Println("title",title)

	response = utils.HttpDelete("https://jsonplaceholder.typicode.com/todos/1", "",todo)
	fmt.Println(response)
}

func TestSwanClient() {
	swanClient := utils.GetSwanClient()

	//fmt.Println(swanClient)

	mainConf := config.GetConfig().Main
	deals := swanClient.GetOfflineDeals(mainConf.MinerFid,"ReadyForImport", "10")
	fmt.Println(deals)
	response := swanClient.UpdateOfflineDealStatus("Completed","test note",2455)
	fmt.Println(response)
}

func TestAriaClient() {
	swanClient := utils.GetSwanClient()

	aria2Client := utils.GetAria2Client()
	offlineDeal := &models.OfflineDeal{
		Id: 163,
		UserId: 163,
		SourceFileUrl: "https://file-examples-com.github.io/uploads/2020/03/file_example_WEBP_500kB.webp",
	}

	aria2Service := offlineDealAdmin.GetAria2Service()
	aria2Service.StartDownloadForDeal(offlineDeal, aria2Client, swanClient)
	aria2Client.GetDownloadStatus("f80d913a4dff40651")
}

func TestDownloader() {
	aria2Client := utils.GetAria2Client()
	swanClient := utils.GetSwanClient()
	aria2Service := offlineDealAdmin.GetAria2Service()
	aria2Service.StartDownloading(aria2Client, swanClient)
}

func TestOsCmdClient()  {
	result, err := utils.ExecOsCmd("ls -l")
	fmt.Println(result, err)

	result, err = utils.ExecOsCmd("pwd")
	fmt.Println(result, err)

	result, err = utils.ExecOsCmd("ls -l | grep common")
	fmt.Println(result, err)
	words := strings.Fields(result)
	for _, word := range words {
		fmt.Println(word)
	}
}



func TestOsCmdClient1()  {
	/*result, err := */utils.ExecOsCmd2Screen("ls -l")
	//fmt.Println(result, err)

	/*result, err = */utils.ExecOsCmd2Screen("pwd")
	//fmt.Println(result, err)

	/*result, err = */utils.ExecOsCmd2Screen("ls -l | grep x")
	//fmt.Println(result, err)
}


package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/joho/godotenv"
	"os"
	"swan-miner/common/constants"
	"swan-miner/common/utils"
	"swan-miner/config"
	"swan-miner/logs"
	"swan-miner/models"
	"swan-miner/offlineDealAdmin"
	"swan-miner/routers/commonRouters"
	"time"
)

func main() {
	LoadEnv()

	config.InitConfig("")

	//offlineDealAdmin.AdminOfflineDeal()
	//testRestApiAccessor()

/*	swanClient := offlineDealAdmin.GetSwanClient()

	//fmt.Println(swanClient)

	mainConf := config.GetConfig().Main
	swanClient.GetOfflineDeals(mainConf.MinerFid,"Waiting", "10")
	swanClient.UpdateOfflineDealStatus("Completed","test note","2455")

	aria2Client := utils.GetAria2Client()
	offlineDeal := &offlineDealAdmin.OfflineDeal{
		Id: "163",
		UserId: string(163),
		SourceFileUrl: "https://file-examples-com.github.io/uploads/2020/03/file_example_WEBP_500kB.webp",
	}

	aria2Service := offlineDealAdmin.GetAria2Service()
	aria2Service.StartDownloadForDeal(*offlineDeal, aria2Client, swanClient)
	aria2Client.GetDownloadStatus("f80d913a4dff40651")*/

	offlineDealAdmin.Downloader()
	//
	//createServer()
}

func testRestApiAccessor(){
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

func createServer() {
	r := gin.Default()
	r.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))

	v1 := r.Group("/api/v1")
	commonRouters.HostManager(v1.Group(constants.URL_HOST_GET_COMMON))

	err := r.Run(":" + config.GetConfig().Port)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}
}

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		logs.GetLogger().Error(err)
	}
	fmt.Println("name: ", os.Getenv("privateKey"))
}

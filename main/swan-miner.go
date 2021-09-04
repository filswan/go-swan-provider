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
	"swan-miner/routers/commonRouters"
	"time"
)

func main() {
	LoadEnv()

	config.InitConfig("")

	testRestApiAccessor()
	//
	//createServer()
}

func testRestApiAccessor(){
	response := utils.Get("https://jsonplaceholder.typicode.com/todos/1")
	fmt.Println(response)
	todo := utils.Todo{1, 2, "lorem ipsum dolor sit amet", true}
	response = utils.Post("https://jsonplaceholder.typicode.com/todos", todo)
	fmt.Println(response)

	response = utils.Put("https://jsonplaceholder.typicode.com/todos/1", todo)
	fmt.Println(response)

	title := utils.GetFieldFromJson(response,"title")
	fmt.Println("title",title)

	response = utils.Delete("https://jsonplaceholder.typicode.com/todos/1", todo)
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

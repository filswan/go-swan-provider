package main

import (
	"os"
	"strconv"
	"swan-provider/common/constants"
	"swan-provider/config"
	"swan-provider/logs"
	"swan-provider/offlineDealAdmin"
	"swan-provider/routers/commonRouters"
	"time"

	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/joho/godotenv"
)

func main() {
	//LoadEnv()
	//test.Test()
	logs.InitLogger()
	//config.InitConfig()
	offlineDealAdmin.AdminOfflineDeal()
	createHttpServer()
}

func createHttpServer() {
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

	err := r.Run(":" + strconv.Itoa(config.GetConfig().Port))
	if err != nil {
		logs.GetLogger().Fatal(err)
	}
}

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		logs.GetLogger().Error(err)
	}

	logs.GetLogger().Info("name: ", os.Getenv("privateKey"))
}

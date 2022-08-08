package main

import (
	"fmt"
	"os"
	"strconv"
	"swan-provider/common"
	"swan-provider/common/constants"
	"swan-provider/config"
	"swan-provider/routers"
	"swan-provider/service"
	"time"

	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/joho/godotenv"

	"github.com/filswan/go-swan-lib/logs"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	subCmd := os.Args[1]
	switch subCmd {
	case "version":
		printVersion()
	case "daemon":
		service.AdminOfflineDeal()
		createHttpServer()
	default:
		printUsage()
	}
}

func printVersion() {
	fmt.Println(getVersion())
}

func getVersion() string {
	return common.VERSION
}

func printUsage() {
	fmt.Println("NAME:")
	fmt.Println("    swan-provider")
	fmt.Println("VERSION:")
	fmt.Println("    " + getVersion())
	fmt.Println("USAGE:")
	fmt.Println("    swan-provider version")
	fmt.Println("    swan-provider daemon")
}

func createHttpServer() {
	//logs.GetLogger().Info("release mode:", config.GetConfig().Release)
	if config.GetConfig().Release {
		gin.SetMode(gin.ReleaseMode)
	}

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
	routers.HostManager(v1.Group(constants.URL_HOST_GET_COMMON))

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

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/filswan/go-swan-lib/client/boost"
	"os"
	"os/signal"
	"strconv"
	"swan-provider/common"
	"swan-provider/common/constants"
	"swan-provider/config"
	"swan-provider/routers"
	"swan-provider/service"
	"syscall"
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
	case "set-ask":
		setAsk()
	case "version":
		printVersion()
	case "daemon":
		sigCh := make(chan os.Signal, 2)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		service.AdminOfflineDeal()
		go createHttpServer()
		select {
		case sig := <-sigCh:
			logs.GetLogger().Warn("received shutdown signal: ", sig)
			service.StopBoost(service.BoostPid)
		}
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
	fmt.Println("    swan-provider set-ask --price=xx --verified-price=xx --min-piece-size=xx --max-piece-size=xx")
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

func setAsk() {
	price := flag.String("price", "0", "Set the price of the ask for unverified deals (specified as FIL / GiB / Epoch) to `PRICE`.")
	verifiedPrice := flag.String("verified-price", "0", "Set the price of the ask for verified deals (specified as FIL / GiB / Epoch) to `PRICE`")
	minSize := flag.String("min-piece-size", "256B", "Set minimum piece size (w/bit-padding, in bytes) in ask to `SIZE`")
	maxSize := flag.String("max-piece-size", "0", "Set maximum piece size (w/bit-padding, in bytes) in ask to `SIZE`")

	market := config.GetConfig().Market
	boostToken, err := service.GetBoostToken(market.Repo)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	boostClient, closer, err := boost.NewClient(boostToken, market.RpcUrl)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	defer closer()
	if err = boostClient.MarketSetAsk(context.TODO(), *price, *verifiedPrice, *minSize, *maxSize); err != nil {
		logs.GetLogger().Error(err)
	}
}

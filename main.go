package main

import (
	"context"
	"fmt"
	"github.com/filswan/swan-boost-lib/provider"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
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
			service.StopProcessById("boostd", service.BoostPid)
		}
	default:
		printUsage()
	}
}

func printVersion() {
	fmt.Println("swan-provider version: ", getVersion())
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
	params := os.Args[2:]
	var price, verifiedPrice, minSize, maxSize string
	for _, param := range params {
		if strings.Contains(param, "verified-price") {
			index := strings.Index(param, "=")
			verifiedPrice = param[index+1:]
		} else if strings.Contains(param, "min-piece-size") {
			index := strings.Index(param, "=")
			minSize = param[index+1:]
		} else if strings.Contains(param, "max-piece-size") {
			index := strings.Index(param, "=")
			maxSize = param[index+1:]
		} else if strings.Contains(param, "price") {
			index := strings.Index(param, "=")
			price = param[index+1:]
		}
	}

	if price == "" {
		logs.GetLogger().Errorf("price is required")
		return
	}
	if verifiedPrice == "" {
		logs.GetLogger().Errorf("verified-price is required")
		return
	}
	if minSize == "" {
		logs.GetLogger().Errorf("min-piece-size is required")
		return
	}
	if maxSize == "" {
		logs.GetLogger().Errorf("max-piece-size is required")
		return
	}

	market := config.GetConfig().Market
	boostToken, err := service.GetBoostToken(market.Repo)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	rpcApi, _, err := config.GetRpcInfoByFile(filepath.Join(market.Repo, "config.toml"))
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	boostClient, closer, err := provider.NewClient(boostToken, rpcApi)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	defer closer()

	if err = boostClient.MarketSetAsk(context.TODO(), price, verifiedPrice, minSize, maxSize); err != nil {
		logs.GetLogger().Error(err)
	}
	fmt.Println("set-ask successfully! You can check it using “lotus client query-ask <minerID>”")
}

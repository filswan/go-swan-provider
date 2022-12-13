package main

import (
	"os"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/logs"
)

func main() {
	wallet := os.Args[1]
	isVerified, err := lotus.IsWalletVerified(wallet)
	logs.GetLogger().Info(isVerified, err)
}

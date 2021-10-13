package utils

import (
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"swan-provider/logs"
	"time"
)

// GetEpochInMillis get current timestamp
func GetEpochInMillis() (millis int64) {
	nanos := time.Now().UnixNano()
	millis = nanos / 1000000
	return
}

func ReadContractAbiJsonFile(aptpath string) (string, error) {
	jsonFile, err := os.Open(aptpath)

	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	return string(byteValue), nil
}

func GetRewardPerBlock() *big.Int {
	rewardBig, _ := new(big.Int).SetString("35000000000000000000", 10) // the unit is wei
	return rewardBig
}
func GetStrFromInt64(num int64) string {
	return strconv.FormatInt(num, 10)
}

func GetInt64FromStr(numStr string) int64 {
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		logs.GetLogger().Error(err)
		return -1
	}

	return num
}

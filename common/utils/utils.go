package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"
	"swan-miner/logs"
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

func GetFieldFromJson(jsonStr string, fieldName string) (interface{}){
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &result)
	fieldVal := result[fieldName].(interface{})

	//fmt.Println(fieldName,fieldVal)

	return fieldVal
}

func GetFieldStrFromJson(jsonStr string, fieldName string) (string){
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &result)
	fieldVal := result[fieldName].(interface{})

	//fmt.Println(fieldName,fieldVal)

	return fieldVal.(string)
}

func GetFieldMapFromJson(jsonStr string, fieldName string) (map[string]interface{}){
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &result)
	fieldVal := result[fieldName].(interface{})

	//fmt.Println(fieldName,fieldVal)

	return fieldVal.(map[string]interface{})
}

func ToJson(obj interface{}) (string){
	jsonBytes, _ := json.Marshal(obj)
	jsonString := string(jsonBytes)
	return jsonString
}

func GetDir(root string, dirs ...string) (string) {
	path := root

	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		if strings.HasSuffix(path,"/") {
			path = path + dir
		}else{
			path = path + "/" + dir
		}
	}

	return path
}

func IsFileExists(filePath, fileName string) (bool) {
	fileFullPath := GetDir(filePath, fileName)
	_, err := os.Stat(fileFullPath)
	if err == nil {
		return true
	}

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return false
}

func RemoveFile(filePath, fileName string) {
	fileFullPath := GetDir(filePath, fileName)
	err := os.Remove(fileFullPath)
	if err != nil {
		 logs.GetLogger().Error(err.Error())
	}
}

func GetFileSize(fileFullPath string) (int64) {
	fi, err := os.Stat(fileFullPath)
	if err != nil {
		return -1
	}
	size := fi.Size()

	return size
}

func GetStrFromInt64(num int64) (string) {
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
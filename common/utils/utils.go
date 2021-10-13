package utils

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"
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

func GetFieldFromJson(jsonStr string, fieldName string) interface{} {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	if result == nil {
		logs.GetLogger().Error("Failed to parse ", jsonStr, " as map[string]interface{}.")
		return nil
	}

	fieldVal := result[fieldName]
	return fieldVal
}

func GetFieldStrFromJson(jsonStr string, fieldName string) string {
	fieldVal := GetFieldFromJson(jsonStr, fieldName)
	if fieldVal == nil {
		return ""
	}

	switch fieldValType := fieldVal.(type) {
	case string:
		return fieldValType
	default:
		return ""
	}
}

func GetFieldMapFromJson(jsonStr string, fieldName string) map[string]interface{} {
	fieldVal := GetFieldFromJson(jsonStr, fieldName)
	if fieldVal == nil {
		logs.GetLogger().Error("Failed to get ", fieldName, " from ", jsonStr)
		return nil
	}

	switch fieldValType := fieldVal.(type) {
	case map[string]interface{}:
		return fieldValType
	default:
		return nil
	}
}

func ToJson(obj interface{}) string {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}

	jsonString := string(jsonBytes)
	return jsonString
}

func GetDir(root string, dirs ...string) string {
	path := root

	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		if strings.HasSuffix(path, "/") {
			if strings.HasPrefix(dir, "/") {
				dir = strings.TrimLeft(dir, "/")
			}
			path = path + dir
		} else {
			path = path + "/" + dir
		}
	}

	return path
}

func IsFileExists(filePath, fileName string) bool {
	fileFullPath := GetDir(filePath, fileName)
	_, err := os.Stat(fileFullPath)

	if err != nil {
		logs.GetLogger().Info(err)
		return false
	}

	return true
}

func IsFileExistsFullPath(fileFullPath string) bool {
	_, err := os.Stat(fileFullPath)

	if err != nil {
		logs.GetLogger().Info(err)
		return false
	}

	return true
}

func RemoveFile(filePath, fileName string) {
	fileFullPath := GetDir(filePath, fileName)
	err := os.Remove(fileFullPath)
	if err != nil {
		logs.GetLogger().Error(err.Error())
	}
}

func GetFileSize(fileFullPath string) int64 {
	fi, err := os.Stat(fileFullPath)
	if err != nil {
		logs.GetLogger().Info(err)
		return -1
	}

	return fi.Size()
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

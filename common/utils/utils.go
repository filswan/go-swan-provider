package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
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

	fmt.Println(fieldName,fieldVal)

	return fieldVal
}

func GetFieldStrFromJson(jsonStr string, fieldName string) (string){
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &result)
	fieldVal := result[fieldName].(interface{})

	fmt.Println(fieldName,fieldVal)

	return fieldVal.(string)
}

func GetFieldMapFromJson(jsonStr string, fieldName string) (map[string]interface{}){
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &result)
	fieldVal := result[fieldName].(interface{})

	fmt.Println(fieldName,fieldVal)

	return fieldVal.(map[string]interface{})
}

func ToJson(obj interface{}) (string){
	jsonBytes, _ := json.Marshal(obj)
	jsonString := string(jsonBytes)
	return jsonString
}

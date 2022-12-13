package utils

import (
	"encoding/json"
	"strings"

	"github.com/filswan/go-swan-lib/logs"
)

func GetFieldFromJson(jsonBytes []byte, fieldName string) interface{} {
	var result map[string]interface{}
	err := json.Unmarshal(jsonBytes, &result)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	if result == nil {
		logs.GetLogger().Error("Failed to parse ", jsonBytes, " as map[string]interface{}.")
		return nil
	}

	fieldVal := result[fieldName]
	return fieldVal
}

func GetFieldStrFromJson(jsonBytes []byte, fieldName string) string {
	fieldVal := GetFieldFromJson(jsonBytes, fieldName)
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

func GetFieldMapFromJson(jsonBytes []byte, fieldName string) map[string]interface{} {
	fieldVal := GetFieldFromJson(jsonBytes, fieldName)
	if fieldVal == nil {
		if strings.EqualFold(fieldName, "error") {
			return nil
		}
		logs.GetLogger().Info("No", fieldName, " in ", string(jsonBytes))
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

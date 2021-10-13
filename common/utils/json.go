package utils

import (
	"encoding/json"
	"swan-provider/logs"
)

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
		logs.GetLogger().Info("No ", fieldName, " in ", jsonStr)
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

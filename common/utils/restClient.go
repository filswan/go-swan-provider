package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"swan-provider/logs"
)

const HTTP_CONTENT_TYPE_FORM = "application/x-www-form-urlencoded"
const HTTP_CONTENT_TYPE_JSON = "application/json; charset=utf-8"

func HttpPostNoToken(uri string, jsonRequest interface{}) string {
	response := httpRequest(http.MethodPost, uri, "" , jsonRequest)
	return response
}

func HttpPost(uri, tokenString  string, params interface{}) string {
	response := httpRequest(http.MethodPost, uri, tokenString, params)
	return response
}

func HttpGet(uri, tokenString  string, jsonRequest interface{}) string {
	response := httpRequest(http.MethodGet, uri, tokenString , jsonRequest)
	return response
}

func HttpPut(uri, tokenString  string, params interface{}) string {
	response := httpRequest(http.MethodPut, uri, tokenString , params)
	return response
}

func HttpDelete(uri, tokenString  string, jsonRequest interface{}) string {
	response := httpRequest(http.MethodDelete, uri, tokenString , jsonRequest)
	return response
}

func httpRequest(httpMethod, uri, tokenString string, params interface{}) (string) {
	var request *http.Request
	var err error

	switch params.(type) {
	case io.Reader:
		request, err = http.NewRequest(httpMethod, uri, params.(io.Reader))
		if err != nil {
			logs.GetLogger().Error(err)
			return ""
		}
		request.Header.Set("Content-Type", HTTP_CONTENT_TYPE_FORM)
	default:
		jsonReq, errJson := json.Marshal(params)
		if errJson != nil {
			logs.GetLogger().Error(errJson)
			return ""
		}

		request, err = http.NewRequest(httpMethod, uri, bytes.NewBuffer(jsonReq))
		if err != nil {
			logs.GetLogger().Error(err)
			return ""
		}
		request.Header.Set("Content-Type", HTTP_CONTENT_TYPE_JSON)
	}

	if len(tokenString) > 0 {
		request.Header.Set("Authorization","Bearer "+tokenString)
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if response == nil {
		logs.GetLogger().Error(uri, " no response")
		return ""
	}

	if response.Body == nil {
		logs.GetLogger().Error(uri, " no response body")
		return ""
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}

	return string(responseBody)
}

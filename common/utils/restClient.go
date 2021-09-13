package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"swan-miner/logs"
)

const HTTP_CONTENT_TYPE_FORM = "application/x-www-form-urlencoded"
const HTTP_CONTENT_TYPE_JSON = "application/json; charset=utf-8"

func HttpPostJsonParamNoToken(uri string, jsonRequest interface{}) string {
	response := httpRequestJsonParam(http.MethodPost, uri, "" , jsonRequest)

	return response
}

func HttpPostJsonParam(uri, tokenString  string, jsonRequest interface{}) string {
	response := httpRequestJsonParam(http.MethodPost, uri, tokenString , jsonRequest)

	return response
}

func HttpGetJsonParam(uri, tokenString  string, jsonRequest interface{}) string {
	response := httpRequestJsonParam(http.MethodGet, uri, tokenString , jsonRequest)

	return response
}

func HttpPutJsonParam(uri, tokenString  string, jsonRequest interface{}) string {
	response := httpRequestJsonParam(http.MethodPut, uri, tokenString , jsonRequest)

	return response
}

func HttpPutFormParam(uri, tokenString  string, params io.Reader) string {
	response := httpRequestFormParam(http.MethodPut, uri, tokenString , params)

	return response
}

func HttpDeleteJsonParam(uri, tokenString  string, jsonRequest interface{}) string {
	response := httpRequestJsonParam(http.MethodDelete, uri, tokenString , jsonRequest)

	return response
}

func httpRequestJsonParam(httpMethod, uri, tokenString string, params interface{}) (string) {
	jsonReq, err := json.Marshal(params)
	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}

	request, err := http.NewRequest(httpMethod, uri, bytes.NewBuffer(jsonReq))
	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}
	request.Header.Set("Content-Type", HTTP_CONTENT_TYPE_JSON)

	if len(tokenString)>0{
		request.Header.Set("Authorization","Bearer "+tokenString)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	defer response.Body.Close()

	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}
	responseBody, _ := ioutil.ReadAll(response.Body)

	responseString := string(responseBody)

	return responseString
}

func httpRequestFormParam(httpMethod, uri, tokenString string, params io.Reader) (string) {
	request, err := http.NewRequest(httpMethod, uri, params)
	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}
	request.Header.Set("Content-Type", HTTP_CONTENT_TYPE_FORM)

	if len(tokenString)>0{
		request.Header.Set("Authorization","Bearer "+tokenString)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	defer response.Body.Close()

	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}
	responseBody, _ := ioutil.ReadAll(response.Body)

	responseString := string(responseBody)

	return responseString
}

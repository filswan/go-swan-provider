package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"swan-miner/config"
)

const contentType = "application/json; charset=utf-8"

func GetOfflineDeals(self, miner_fid, status, limit string) string {
	url := config.GetConfig().Main.ApiUrl + "/offline_deals/" + miner_fid + "?deal_status=" + status + "&limit=" + limit + "&offset=0"

	response := Get(url)

	return response
}

func Get(uri string) string {
	fmt.Println("Performing Http Get..." + uri)
	response, err := http.Get(uri)

	if err != nil {
		fmt.Print(err.Error())
		log.Fatalln(err)
	}

	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	result := string(responseBody)
	//fmt.Println(result)

	return result
}

func Post(uri string, jsonRequest interface{}) string {
	fmt.Println("Performing Http Post...", uri, contentType, jsonRequest)
	jsonReq, err := json.Marshal(jsonRequest)
	response, err := http.Post(uri, contentType, bytes.NewBuffer(jsonReq))
	if err != nil {
		log.Fatalln(err)
	}

	defer response.Body.Close()
	responseBody, _ := ioutil.ReadAll(response.Body)

	// Convert response body to string
	responseStr := string(responseBody)
	//fmt.Println(responseStr)

	return responseStr
}

func Put(uri string, jsonRequest interface{}) string {
	response := httpRequest(http.MethodPut, uri, jsonRequest)

	return response
}

func Delete(uri string, jsonRequest interface{}) string {
	response := httpRequest(http.MethodDelete, uri, jsonRequest)

	return response
}

func httpRequest(httpMethod, uri string, jsonRequest interface{}) string {
	fmt.Println("Performing Http "+httpMethod+"...", uri, jsonRequest)
	jsonReq, err := json.Marshal(jsonRequest)
	request, err := http.NewRequest(httpMethod, uri, bytes.NewBuffer(jsonReq))
	request.Header.Set("Content-Type", contentType)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	defer response.Body.Close()
	responseBody, _ := ioutil.ReadAll(response.Body)

	// Convert response body to string
	responseString := string(responseBody)
	//fmt.Println(responseString)

	return responseString
}

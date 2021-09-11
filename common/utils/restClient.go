package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"swan-miner/logs"
)

/*func Get(uri string) string {
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
}*/

/*func Post(uri string, jsonRequest interface{}) string {
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
}*/

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
	//fmt.Println("Performing Http "+httpMethod+"...", uri, jsonRequest)
	jsonReq, err := json.Marshal(params)
	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}
	//fmt.Println(string(jsonReq))
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

	// Convert response body to string
	responseString := string(responseBody)
	//fmt.Println(responseString)

	return responseString
}

func httpRequestFormParam(httpMethod, uri, tokenString string, params io.Reader) (string) {
	//fmt.Println("Performing Http "+httpMethod+"...", uri, jsonRequest)
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

	// Convert response body to string
	responseString := string(responseBody)
	//fmt.Println(responseString)

	return responseString
}

func UploadFileByStream(uri, filepath, filename string) {
	var fileReader io.Reader
	var err error
	fileFullPath := filepath+"/"+filename
	fileReader, err = os.Open(fileFullPath)

	boundary := "MyMultiPartBoundary12345"
	token := "DEPLOY_GATE_TOKEN"
	message := "Uploaded by Nebula"
	releaseNote := "Built by Nebula"
	fieldFormat := "--%s\r\nContent-Disposition: form-data; name=\"%s\"\r\n\r\n%s\r\n"
	tokenPart := fmt.Sprintf(fieldFormat, boundary, "token", token)
	messagePart := fmt.Sprintf(fieldFormat, boundary, "message", message)
	releaseNotePart := fmt.Sprintf(fieldFormat, boundary, "release_note", releaseNote)
	fileName := filename
	fileHeader := "Content-type: application/octet-stream"
	fileFormat := "--%s\r\nContent-Disposition: form-data; name=\"file\"; filename=\"%s\"\r\n%s\r\n\r\n"
	filePart := fmt.Sprintf(fileFormat, boundary, fileName, fileHeader)
	bodyTop := fmt.Sprintf("%s%s%s%s", tokenPart, messagePart, releaseNotePart, filePart)
	bodyBottom := fmt.Sprintf("\r\n--%s--\r\n", boundary)
	body := io.MultiReader(strings.NewReader(bodyTop), fileReader, strings.NewReader(bodyBottom))

	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)

	response, err := http.Post(uri, contentType, body)
	if err != nil {
		fmt.Println(err)
	}
	if response!=nil{
		content, err := ioutil.ReadAll(response.Body)
		responseContent:=string(content)
		fmt.Println(responseContent)
		if err != nil {
			fmt.Println(err)
		}

		response.Body.Close()
	}
}


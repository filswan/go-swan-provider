package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
)

const HTTP_CONTENT_TYPE_FORM = "application/x-www-form-urlencoded"
const HTTP_CONTENT_TYPE_JSON = "application/json; charset=UTF-8"

func HttpPostNoToken(uri string, params interface{}) ([]byte, error) {
	response, err := HttpRequest(http.MethodPost, uri, "", params)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return response, nil
}

func HttpPost(uri, tokenString string, params interface{}) ([]byte, error) {
	response, err := HttpRequest(http.MethodPost, uri, tokenString, params)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return response, nil
}

func HttpGetNoToken(uri string, params interface{}) ([]byte, error) {
	response, err := HttpRequest(http.MethodGet, uri, "", params)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return response, nil
}

func HttpGet(uri, tokenString string, params interface{}) ([]byte, error) {
	response, err := HttpRequest(http.MethodGet, uri, tokenString, params)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return response, nil
}

func HttpPut(uri, tokenString string, params interface{}) ([]byte, error) {
	response, err := HttpRequest(http.MethodPut, uri, tokenString, params)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return response, nil
}

func HttpDelete(uri, tokenString string, params interface{}) ([]byte, error) {
	response, err := HttpRequest(http.MethodDelete, uri, tokenString, params)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return response, nil
}

func HttpRequest(httpMethod, uri, tokenString string, params interface{}) ([]byte, error) {
	var request *http.Request
	var err error

	switch params := params.(type) {
	case io.Reader:
		request, err = http.NewRequest(httpMethod, uri, params)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		request.Header.Set("Content-Type", HTTP_CONTENT_TYPE_FORM)
	default:
		jsonReq, errJson := json.Marshal(params)
		if errJson != nil {
			logs.GetLogger().Error(errJson)
			return nil, errJson
		}

		request, err = http.NewRequest(httpMethod, uri, bytes.NewBuffer(jsonReq))
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		request.Header.Set("Content-Type", HTTP_CONTENT_TYPE_JSON)
	}

	if len(strings.Trim(tokenString, " ")) > 0 {
		request.Header.Set("Authorization", "Bearer "+tokenString)
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf("http status: %s, code:%d, url:%s", response.Status, response.StatusCode, uri)
		logs.GetLogger().Error(err)
		switch response.StatusCode {
		case http.StatusNotFound:
			logs.GetLogger().Error("please check your url:", uri)
		case http.StatusUnauthorized:
			logs.GetLogger().Error("Please check your token:", tokenString)
		}
		return nil, err
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return responseBody, nil
}

func HttpPutFile(url string, tokenString string, paramTexts map[string]string, paramFilename, paramFilepath string) (string, error) {
	response, err := HttpRequestFile(http.MethodPut, url, tokenString, paramTexts, paramFilename, paramFilepath)
	return response, err
}

func HttpPostFile(url string, tokenString string, paramTexts map[string]string, paramFilename, paramFilepath string) (string, error) {
	response, err := HttpRequestFile(http.MethodPost, url, tokenString, paramTexts, paramFilename, paramFilepath)
	return response, err
}

func HttpRequestFile(httpMethod, url string, tokenString string, paramTexts map[string]string, paramFilename, paramFilepath string) (string, error) {
	filename, fileContent, err := utils.ReadFile(paramFilepath)
	if err != nil {
		logs.GetLogger().Info(err)
		return "", err
	}

	bodyBuf := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile(paramFilename, filename)
	if err != nil {
		bodyWriter.Close()
		logs.GetLogger().Info(err)
		return "", err
	}

	fileWriter.Write(fileContent)

	for key, val := range paramTexts {
		err = bodyWriter.WriteField(key, val)
		if err != nil {
			bodyWriter.Close()
			logs.GetLogger().Info(err)
			return "", err
		}
	}

	bodyWriter.Close()

	request, err := http.NewRequest(httpMethod, url, bodyBuf)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", nil
	}

	request.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	if len(strings.Trim(tokenString, " ")) > 0 {
		request.Header.Set("Authorization", "Bearer "+tokenString)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", nil
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf("http status:%s, code:%d, url:%s", response.Status, response.StatusCode, url)
		logs.GetLogger().Error(err)
		switch response.StatusCode {
		case http.StatusNotFound:
			logs.GetLogger().Error("please check your url:", url)
		case http.StatusUnauthorized:
			logs.GetLogger().Error("Please check your token:", tokenString)
		}
		return "", err
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}

	responseStr := string(responseBody)

	return responseStr, nil
}

func HttpUploadFileByStream(uri, filefullpath string) ([]byte, error) {
	fileReader, err := os.Open(filefullpath)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	filename := filepath.Base(filefullpath)

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
		logs.GetLogger().Error(err)
		return nil, nil
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf("http status:%s, code:%d, url:%s", response.Status, response.StatusCode, uri)
		logs.GetLogger().Error(err)
		switch response.StatusCode {
		case http.StatusNotFound:
			logs.GetLogger().Error("please check your url:", uri)
		}
		return nil, err
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	responseStr := string(responseBody)
	//logs.GetLogger().Info(responseStr)
	filesInfo := strings.Split(responseStr, "\n")
	if len(filesInfo) < 4 {
		err := fmt.Errorf("not enough files info returned")
		logs.GetLogger().Error(err)
		return nil, err
	}
	responseStr = filesInfo[3]
	return []byte(responseStr), nil
}

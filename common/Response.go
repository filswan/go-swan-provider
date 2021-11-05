package common

import (
	libconstants "github.com/filswan/go-swan-lib/constants"
)

type BasicResponse struct {
	Status   string      `json:"status"`
	Code     string      `json:"code"`
	Data     interface{} `json:"data,omitempty"`
	Message  string      `json:"message,omitempty"`
	PageInfo *PageInfo   `json:"page_info,omitempty"`
}

type PageInfo struct {
	PageNumber       string `json:"page_number"`
	PageSize         string `json:"page_size"`
	TotalRecordCount string `json:"total_record_count"`
}

type MixedResponse struct {
	BasicResponse
	MixData struct {
		Success interface{} `json:"success"`
		Fail    interface{} `json:"fail"`
	} `json:"mix_data"`
}

func CreateSuccessResponse(_data interface{}) BasicResponse {
	return BasicResponse{
		Status: libconstants.SWAN_API_STATUS_SUCCESS,
		Data:   _data,
	}
}

func CreateErrorResponse(_errCode, _errMsg string) BasicResponse {
	return BasicResponse{
		Status:  libconstants.SWAN_API_STATUS_FAIL,
		Code:    _errCode,
		Message: _errMsg,
	}
}

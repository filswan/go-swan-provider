package model

import "github.com/shopspring/decimal"

type OfflineDeal struct {
	Id       int    `json:"id"`
	DealCid  string `json:"deal_cid"`
	FilePath string `json:"file_path"` //store file downloaded from server locally
	//FileName             string           `json:"file_name"`
	//FileSourceUrl        string           `json:"file_source_url"`
	//Md5Origin            string           `json:"md5_origin"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	Status     string `json:"status"`
	MinerId    int    `json:"miner_id"`
	Md5Local   string `json:"md5_local"`
	StartEpoch int    `json:"start_epoch"`
	//FileDownloadedStatus string           `json:"file_downloaded_status"`
	UserId int    `json:"user_id"`
	Note   string `json:"note"`
	TaskId int    `json:"task_id"`
	//IsPublic             int              `json:"is_public"`
	//FileSize             string           `json:"file_size"`
	PayloadCid string `json:"payload_cid"`
	PieceCid   string `json:"piece_cid"`
	//DownloadedAt         string           `json:"downloaded_at"`
	//Cost                 string           `json:"cost"`
	CarFileId     int              `json:"car_file_id"`
	CarFileUrl    string           `json:"car_file_url"`
	CarFileSize   int64            `json:"car_file_size"`
	MinerFid      string           `json:"miner_fid"`
	TaskName      *string          `json:"task_name"`
	TaskUuid      *string          `json:"task_uuid"`
	TaskType      *string          `json:"task_type"`
	FastRetrieval *int             `json:"fast_retrieval"`
	MaxPrice      *decimal.Decimal `json:"max_price"`
	Duration      *int             `json:"duration"`
	SourceId      *int             `json:"source_id"`
}

package models

type OfflineDeal struct {
	Id               int    `json:"id"`
	UserId           int    `json:"user_id"`
	SourceFileUrl    string `json:"file_source_url"`
	Status           string `json:"status"`
	Note             string `json:"note"`
	MinerId          int    `json:"miner_id"`
	StartEpoch       int    `json:"start_epoch"`
	FilePath         string `json:"file_path"`
	FileSize         string `json:"file_size"`
	DealCid          string `json:"deal_cid"`
	TaskId           int    `json:"task_id"`
}
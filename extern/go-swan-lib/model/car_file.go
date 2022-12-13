package model

type CarFile struct {
	Id         int     `json:"id"`
	TaskId     int     `json:"task_id"`
	OriginName string  `json:"origin_name"`
	StartEpoch int64   `json:"start_epoch"`
	FileUrl    string  `json:"file_url"`
	FileMd5    *string `json:"file_md5"`
	FileSize   int64   `json:"file_size"`
	PayloadCid string  `json:"payload_cid"`
	PieceCid   string  `json:"piece_cid"`
	PinStatus  *string `json:"pin_status"`
	CreatedAt  int64   `json:"created_at"`
	UpdatedAt  int64   `json:"updated_at"`
}

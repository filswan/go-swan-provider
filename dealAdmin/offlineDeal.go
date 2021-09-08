package dealAdmin

type OfflineDeal struct {
	Id               string `json:"id"`
	UserId           string `json:"user_id"`
	Uuid             string `json:"uuid"`
	SourceFileName   string `json:"source_file_name"`
	SourceFilePath   string `json:"source_file_path"`
	SourceFileMd5    string `json:"source_file_md5"`
	SourceFileUrl    string `json:"source_file_url"`
	SourceFileSize   string `json:"source_file_size"`
	CarFileName      string `json:"car_file_name"`
	CarFilePath      string `json:"car_file_path"`
	CarFileMd5       string `json:"car_file_md5"`
	CarFileUrl       string `json:"car_file_url"`
	CarFileSize      string `json:"car_file_size"`
	DealCid          string `json:"deal_cid"`
	DataCid          string `json:"data_cid"`
	PieceCid         string `json:"piece_cid"`
	MinerId          string `json:"miner_id"`
	StartEpoch       string `json:"start_epoch"`
}
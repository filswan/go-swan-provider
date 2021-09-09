package offlineDealAdmin

type OfflineDeal struct {
	Id               string `json:"id"`
	UserId           string `json:"user_id"`
	SourceFileUrl    string `json:"file_source_url"`
	Status           string `json:"status"`
	Note             string `json:"note"`
	MinerId          string `json:"miner_id"`
	StartEpoch       int    `json:"start_epoch"`
	FilePath         string `json:"file_path"`
	FileSize         string `json:"file_size"`
	DealCid          string `json:"deal_cid"`
}
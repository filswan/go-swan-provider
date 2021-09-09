package dealAdmin

type OfflineDeal struct {
	Id               string `json:"id"`
	UserId           string `json:"user_id"`
	SourceFileUrl    string `json:"file_source_url"`
	Status           string `json:"status"`
	Note             string `json:"note"`
	MinerId          string `json:"miner_id"`
	StartEpoch       string `json:"start_epoch"`
	FileSize         string `json:"file_size"`
	DealCid          string `json:"deal_cid"`
}
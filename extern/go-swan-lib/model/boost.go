package model

type ProviderDealState struct {
	DealUuid     string
	IsOffline    bool
	DealDataRoot string
	ChainDealID  uint64
	PublishCID   string
	DealStatus   string
	Err          string
}

type ProviderDealRejectionInfo struct {
	Accepted bool
	Reason   string
}

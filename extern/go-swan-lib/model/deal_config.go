package model

import (
	"github.com/shopspring/decimal"
)

type DealConfig struct {
	SkipConfirmation bool
	VerifiedDeal     bool
	FastRetrieval    bool
	StartEpoch       int64
	MinerFid         string
	MaxPrice         decimal.Decimal
	SenderWallet     string
	Duration         int
	TransferType     string
	PayloadCid       string
	PieceCid         string
	FileSize         int64
}

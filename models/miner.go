package models

type Miner struct {
    Id                int     `json:"id"`
    MinerFid          string  `json:"miner_fid"`
    BidMode           int     `json:"bid_mode"`
    StartEpoch        int     `json:"start_epoch"`
    Price             string  `json:"price"`
    VerifiedPrice     string  `json:"verified_price"`
    MinPieceSize      string  `json:"min_piece_size"`
    MaxPieceSize      string  `json:"max_piece_size"`
    AutoBidTaskPerDay int     `json:"auto_bid_task_per_day"`
}

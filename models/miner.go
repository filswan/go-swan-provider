package models

type Miner struct {
    Id                  int     `json:"id"`
    MinerFid            string  `json:"miner_fid"`
    BidMode             int     `json:"bid_mode"`
    ExpectedSealingTime int     `json:"expected_sealing_time"`
    StartEpoch          int     `json:"start_epoch"`
    AutoBidTaskPerDay   int     `json:"auto_bid_task_per_day"`
}

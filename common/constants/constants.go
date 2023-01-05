package constants

const (
	URL_HOST_GET_COMMON    = "/common"
	URL_HOST_GET_HOST_INFO = "/miner/host/info"

	ERROR_LAUNCH_FAILED   = "Swan provider launch failed."
	INFO_ON_HOW_TO_CONFIG = "For more information about how to config, please check https://docs.filswan.com/run-swan-provider/config-swan-provider"

	UPDATE_OFFLINE_DEAL_STATUS_FAIL = "failed to update offline deal status"
	NOT_UPDATE_OFFLINE_DEAL_STATUS  = "no need to update deal status in swan"

	CHECKPOINT_ACCEPTED    = "Accepted"
	CHECKPOINT_TRANSFERRED = "Transferred"
	CHECKPOINT_PULISHED    = "Published"
	CHECKPOINT_CONFIRMED   = "PublishConfirmed"
	CHECKPOINT_ADDPIECE    = "AddedPiece"
	CHECKPOINT_INDEX       = "IndexedAndAnnounced"
	CHECKPOINT_COMPLETE    = "Complete"

	MARKET_VERSION_1 = "1.0"
	MARKET_VERSION_2 = "2.0"
)

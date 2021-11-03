package constants

const (
	URL_HOST_GET_COMMON    = "/common"
	URL_HOST_GET_HOST_INFO = "/miner/host/info"

	HTTP_STATUS_SUCCESS = "success"
	HTTP_STATUS_FAIL    = "fail"
	HTTP_STATUS_ERROR   = "error"

	HTTP_CODE_200_OK                    = "200" //http.StatusOk
	HTTP_CODE_400_BAD_REQUEST           = "400" //http.StatusBadRequest
	HTTP_CODE_401_UNAUTHORIZED          = "401" //http.StatusUnauthorized
	HTTP_CODE_500_INTERNAL_SERVER_ERROR = "500" //http.StatusInternalServerError

	ERROR_LAUNCH_FAILED   = "Swan provider launch failed."
	INFO_ON_HOW_TO_CONFIG = "For more information about how to config, please check https://docs.filswan.com/run-swan-provider/config-swan-provider"
)

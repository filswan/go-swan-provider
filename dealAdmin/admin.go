package dealAdmin


const DEAL_STATUS_FAILED = "ImportFailed"

const DEAL_STATUS_FILE_IMPORTING = "FileImporting"
const DEAL_STATUS_FILE_IMPORTED = "FileImported"
const DEAL_STATUS_ACTIVE = "DealActive"

const ONCHAIN_DEAL_STATUS_ERROR = "StorageDealError"
const ONCHAIN_DEAL_STATUS_ACTIVE = "StorageDealActive"

func dealAdmin()  {
	go Downloader()
	go Importer()
	go Scanner()
}
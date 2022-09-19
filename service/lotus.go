package service

import (
	"github.com/filswan/go-swan-lib/model"
	"os"
	"swan-provider/config"
	"time"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
)

type LotusService struct {
	MinerFid             string
	ImportIntervalSecond time.Duration
	ExpectedSealingTime  int
	ScanIntervalSecond   time.Duration
	LotusMarket          *lotus.LotusMarket
	LotusClient          *lotus.LotusClient
}

func GetLotusService() *LotusService {
	confMain := config.GetConfig().Main

	lotusService := &LotusService{
		MinerFid:             confMain.MinerFid,
		ImportIntervalSecond: confMain.LotusImportInterval * time.Second,
		ExpectedSealingTime:  config.GetConfig().Bid.ExpectedSealingTime,
		ScanIntervalSecond:   confMain.LotusScanInterval * time.Second,
	}

	marketApiUrl := config.GetConfig().Lotus.MarketApiUrl
	marketAccessToken := config.GetConfig().Lotus.MarketAccessToken
	clientApiUrl := config.GetConfig().Lotus.ClientApiUrl
	lotusMarket, err := lotus.GetLotusMarket(marketApiUrl, marketAccessToken, clientApiUrl)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	lotusService.LotusMarket = lotusMarket

	lotusClient, err := lotus.LotusGetClient(clientApiUrl, "")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}
	lotusService.LotusClient = lotusClient

	return lotusService
}

func (lotusService *LotusService) StartImport(swanClient *swan.SwanClient) {
	maxImportNum := LOTUS_IMPORT_NUMNBER
	deals := GetOfflineDeals(swanClient, DEAL_STATUS_IMPORT_READY, aria2Service.MinerFid, &maxImportNum)
	if len(deals) == 0 {
		logs.GetLogger().Info("no pending offline deals found")
		return
	}

	aria2AutoDeleteCarFile := config.GetConfig().Aria2.Aria2AutoDeleteCarFile
	for _, deal := range deals {
		minerId, dealId, onChainStatus, onChainMessage, err := lotusService.LotusMarket.LotusGetDealOnChainStatus(deal.DealCid)
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}
		UpdateSwanDealStatus(minerId, dealId, onChainStatus, *onChainMessage, deal, aria2AutoDeleteCarFile)

		logs.GetLogger().Info("Sleeping...")
		time.Sleep(lotusService.ImportIntervalSecond)
	}
}

func (lotusService *LotusService) StartScan(swanClient *swan.SwanClient) {
	maxScanNum := LOTUS_SCAN_NUMBER
	importedDeals := GetOfflineDeals(swanClient, DEAL_STATUS_IMPORTED, aria2Service.MinerFid, &maxScanNum)
	importingDeals := GetOfflineDeals(swanClient, DEAL_STATUS_IMPORTING, aria2Service.MinerFid, &maxScanNum)

	deals := make([]*model.OfflineDeal, 0)
	deals = append(deals, importedDeals...)
	deals = append(deals, importingDeals...)
	if len(deals) == 0 {
		logs.GetLogger().Info("no ongoing offline deals found")
		return
	}

	lotusDeals, err := lotusService.LotusMarket.LotusGetDeals()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	if len(lotusDeals) == 0 {
		logs.GetLogger().Error("no deals returned from lotus")
		return
	}
	aria2AutoDeleteCarFile := config.GetConfig().Aria2.Aria2AutoDeleteCarFile
	for _, deal := range deals {
		minerId, dealId, onChainStatus, onChainMessage, err := lotusService.LotusMarket.LotusGetDealOnChainStatusFromDeals(lotusDeals, deal.DealCid)
		if err != nil {
			logs.GetLogger().Error(GetLog(deal, err.Error()))
			return
		}

		UpdateSwanDealStatus(minerId, dealId, onChainStatus, *onChainMessage, deal, aria2AutoDeleteCarFile)
	}
}

func IsExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil || os.IsExist(err)
}

func DeleteDownloadedFiles(filePath string) {
	aria2AutoDeleteCarFile := config.GetConfig().Aria2.Aria2AutoDeleteCarFile
	if aria2AutoDeleteCarFile {
		if IsExist(filePath) {
			err := os.Remove(filePath)
			if err != nil {
				logs.GetLogger().Error("failed to delete file ", err, " file path ", filePath)
			} else {
				logs.GetLogger().Info("delete file successfully ", " file path ", filePath)
			}
		}
	}
}

func CorrectDealStatus(startEpoch int, minerId string, dealId uint64, onChainStatus string) (*string, error) {
	dealInfo, err := lotusService.LotusClient.LotusGetDealById(dealId)
	if err != nil {
		logs.GetLogger().Errorf("get market deal info by dealId failed,dealId: %d,error: %s ", dealId, err.Error())
		return nil, err
	}
	if dealInfo.State.SectorStartEpoch > -1 && dealInfo.State.SlashEpoch == -1 && dealInfo.Proposal.Provider == minerId {
		onChainStatus = "StorageDealActive"
	}

	currentEpoch, err := lotusService.LotusClient.LotusGetCurrentEpoch()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	if startEpoch < int(*currentEpoch)+lotusService.ExpectedSealingTime {
		onChainStatus = "StorageDealError"
	} else {
		onChainStatus = ONCHAIN_DEAL_STATUS_SEALING
	}
	return &onChainStatus, nil
}

func UpdateSwanDealStatus(minerId string, dealId uint64, onChainStatus *string, onChainMessage string, deal *model.OfflineDeal, aria2AutoDeleteCarFile bool) {
	if dealId > 0 {
		status, err := CorrectDealStatus(deal.StartEpoch, minerId, dealId, *onChainStatus)
		if err != nil {
			logs.GetLogger().Error(GetLog(deal, err.Error()))
			return
		}
		onChainStatus = status
	}

	if utils.IsStrEmpty(onChainStatus) {
		logs.GetLogger().Info(GetLog(deal, "not found the deal on the chain"))
		UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "not found the deal on the chain")
		return
	}

	switch *onChainStatus {
	case ONCHAIN_DEAL_STATUS_ERROR:
		UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal error", *onChainStatus, onChainMessage)
		if aria2AutoDeleteCarFile {
			logs.GetLogger().Infof("dealId:%d, taskName:%s, dealCid:%s, has been %s, delete the car file, filePath:%s", dealId, *deal.TaskName, deal.DealCid, *onChainStatus, deal.FilePath)
			DeleteDownloadedFiles(deal.FilePath)
		}
	case ONCHAIN_DEAL_STATUS_ACTIVE:
		UpdateStatusAndLog(deal, DEAL_STATUS_ACTIVE, "deal has been completed", *onChainStatus, onChainMessage)
		if aria2AutoDeleteCarFile {
			logs.GetLogger().Infof("dealId:%d, taskName:%s, dealCid:%s, has been %s, delete the car file, filePath:%s", dealId, *deal.TaskName, deal.DealCid, *onChainStatus, deal.FilePath)
			DeleteDownloadedFiles(deal.FilePath)
		}
	case ONCHAIN_DEAL_STATUS_ACCEPT:
		UpdateStatusAndLog(deal, deal.Status, "deal will be ready shortly", *onChainStatus, onChainMessage)
	case ONCHAIN_DEAL_STATUS_NOTFOUND:
		UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal not found", *onChainStatus, onChainMessage)
	case ONCHAIN_DEAL_STATUS_AWAITING, ONCHAIN_DEAL_STATUS_SEALING:
		UpdateStatusAndLog(deal, DEAL_STATUS_IMPORTED, "deal is sealing", *onChainStatus, onChainMessage)
	case ONCHAIN_DEAL_STATUS_WAITTING:
		if deal.Status == DEAL_STATUS_IMPORTING {
			return
		}
		currentEpoch, err := lotusService.LotusClient.LotusGetCurrentEpoch()
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}

		if int64(deal.StartEpoch)-*currentEpoch < int64(lotusService.ExpectedSealingTime) {
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "deal expired before importing", *onChainStatus, onChainMessage)
			return
		}

		UpdateStatusAndLog(deal, DEAL_STATUS_IMPORTING, "importing deal")
		err = lotusService.LotusMarket.LotusImportData(deal.DealCid, deal.FilePath)
		if err != nil { //There should be no output if everything goes well
			UpdateStatusAndLog(deal, DEAL_STATUS_IMPORT_FAILED, "import deal failed", err.Error())
			return
		}
		UpdateStatusAndLog(deal, DEAL_STATUS_IMPORTED, "deal imported")
	default:
		UpdateStatusAndLog(deal, deal.Status, *onChainStatus, onChainMessage)
	}
}

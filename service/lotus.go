package service

import (
	"fmt"
	"swan-provider/config"
	"time"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/model"
)

type LotusService struct {
	MinerFid             string
	ImportIntervalSecond time.Duration
	ExpectedSealingTime  int
	ScanIntervalSecond   time.Duration
	LotusMarket          *lotus.LotusMarket
	LotusClient          *lotus.LotusClient
}

func GetLogFromStatus(deal model.OfflineDeal, status, reason string) string {
	msg := fmt.Sprintf("deal id%d, CID:%s, set deal status to:%s", deal.Id, deal.DealCid, status)
	if reason != "" {
		msg = msg + " due to " + reason
	}
	return msg
}

func GetLog(deal model.OfflineDeal, text string) string {
	msg := fmt.Sprintf("deal id%d, CID:%s, %s", deal.Id, deal.DealCid, text)
	return msg
}

func GetNote(message, onChainStatus, onChainMessage string) string {
	result := fmt.Sprintf("%sOn chain status:%s", message, onChainStatus)
	if onChainMessage != "" {
		result = result + ", message:" + onChainMessage
	}
	return result
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
	deals := swanClient.SwanGetOfflineDeals(lotusService.MinerFid, DEAL_STATUS_IMPORT_READY, LOTUS_IMPORT_NUMNBER)
	if len(deals) == 0 {
		logs.GetLogger().Info("No pending offline deals found.")
		return
	}

	for _, deal := range deals {
		logs.GetLogger().Info(GetLog(deal, "filepath:"+deal.FilePath))

		onChainStatus, onChainMessage := lotusService.LotusMarket.LotusGetDealOnChainStatus(deal.DealCid)

		if len(onChainStatus) == 0 {
			logs.GetLogger().Error(GetLog(deal, "failed to get on chain status."))
			continue
		}

		msg := GetNote("", onChainStatus, onChainMessage)
		logs.GetLogger().Info(GetLog(deal, msg))

		switch onChainStatus {
		case ONCHAIN_DEAL_STATUS_ERROR:
			note := GetNote("Deal is error before importing.", onChainStatus, onChainMessage)
			logs.GetLogger().Warn(GetLog(deal, note))
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
			if !updated {
				logs.GetLogger().Error(GetLog(deal, "failed to update offline deal status"))
			} else {
				logs.GetLogger().Warn(GetLogFromStatus(deal, DEAL_STATUS_IMPORT_FAILED, ""))
			}
		case ONCHAIN_DEAL_STATUS_ACTIVE:
			note := GetNote("Deal is active before importing.", onChainStatus, onChainMessage)
			logs.GetLogger().Info(GetLog(deal, note))
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_ACTIVE, note)
			if !updated {
				logs.GetLogger().Error(GetLog(deal, "failed to update offline deal status"))
			} else {
				logs.GetLogger().Info(GetLogFromStatus(deal, DEAL_STATUS_ACTIVE, ""))
			}
		case ONCHAIN_DEAL_STATUS_ACCEPT:
			note := GetNote("Deal will be ready shortly.", onChainStatus, onChainMessage)
			logs.GetLogger().Info(GetLog(deal, note))
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, deal.Status, note)
			if !updated {
				logs.GetLogger().Error(GetLog(deal, "failed to update offline deal status"))
			}
		case ONCHAIN_DEAL_STATUS_NOTFOUND:
			note := GetNote("Deal not found.", onChainStatus, onChainMessage)
			logs.GetLogger().Info(GetLog(deal, note))
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
			if !updated {
				logs.GetLogger().Error(GetLog(deal, "failed to update offline deal status"))
			} else {
				logs.GetLogger().Info(GetLogFromStatus(deal, DEAL_STATUS_IMPORT_FAILED, ""))
			}
		case ONCHAIN_DEAL_STATUS_WAITTING:
			currentEpoch := lotusService.LotusClient.LotusGetCurrentEpoch()
			if currentEpoch < 0 {
				return
			}

			if deal.StartEpoch-currentEpoch < lotusService.ExpectedSealingTime {
				note := GetNote("Deal expired before importing.", onChainStatus, onChainMessage)
				logs.GetLogger().Warn(GetLog(deal, note))
				updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
				if !updated {
					logs.GetLogger().Error(GetLog(deal, "failed to update offline deal status"))
				} else {
					logs.GetLogger().Warn(GetLogFromStatus(deal, DEAL_STATUS_IMPORT_FAILED, ""))
				}
				continue
			}

			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORTING)
			if !updated {
				logs.GetLogger().Error(GetLog(deal, "failed to update offline deal status"))
			} else {
				logs.GetLogger().Info(GetLogFromStatus(deal, DEAL_STATUS_IMPORTING, ""))
			}

			err := lotusService.LotusMarket.LotusImportData(deal.DealCid, deal.FilePath)

			if err != nil { //There should be no output if everything goes well
				updated = swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, err.Error())
				if !updated {
					logs.GetLogger().Error(GetLog(deal, "failed to update offline deal status"))
				} else {
					logs.GetLogger().Info(GetLogFromStatus(deal, DEAL_STATUS_IMPORT_FAILED, ""))
				}
				logs.GetLogger().Warn(GetLog(deal, "import deal failed. error:"+err.Error()))
				continue
			}

			updated = swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORTED)
			if !updated {
				logs.GetLogger().Error(GetLog(deal, "failed to update offline deal status"))
			} else {
				logs.GetLogger().Info(GetLogFromStatus(deal, DEAL_STATUS_IMPORTED, ""))
			}
			logs.GetLogger().Info(GetLog(deal, "deal imported"))
		default:
			note := GetNote("Deal is already imported.", onChainStatus, onChainMessage)
			logs.GetLogger().Info(GetLog(deal, note))
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORTED, note)
			if !updated {
				logs.GetLogger().Error(GetLog(deal, "failed to update offline deal status"))
			} else {
				logs.GetLogger().Info(GetLogFromStatus(deal, DEAL_STATUS_IMPORTED, ""))
			}
		}

		logs.GetLogger().Info("Sleeping...")
		time.Sleep(lotusService.ImportIntervalSecond)
	}
}

func (lotusService *LotusService) StartScan(swanClient *swan.SwanClient) {
	deals := swanClient.SwanGetOfflineDeals(lotusService.MinerFid, DEAL_STATUS_IMPORTED, LOTUS_SCAN_NUMBER)

	if len(deals) == 0 {
		logs.GetLogger().Info("No ongoing offline deals found.")
		return
	}

	lotusDeals := lotusService.LotusMarket.LotusGetDeals()
	if len(lotusDeals) == 0 {
		logs.GetLogger().Error("Failed to get deals from lotus.")
		return
	}

	for _, deal := range deals {
		logs.GetLogger().Info(GetLog(deal, deal.Status))

		onChainStatus, onChainMessage := lotusService.LotusMarket.LotusGetDealOnChainStatusFromDeals(lotusDeals, deal.DealCid)

		if len(onChainStatus) == 0 {
			logs.GetLogger().Error(GetLog(deal, "failed to get on chain status"))
			continue
		}

		msg := GetNote("", onChainStatus, onChainMessage)
		logs.GetLogger().Info(GetLog(deal, msg))

		switch onChainStatus {
		case ONCHAIN_DEAL_STATUS_ERROR:
			note := GetNote("Deal error when scan.", onChainStatus, onChainMessage)
			logs.GetLogger().Warn(GetLog(deal, note))
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			} else {
				logs.GetLogger().Warn(GetLogFromStatus(deal, DEAL_STATUS_IMPORT_FAILED, ""))
			}
		case ONCHAIN_DEAL_STATUS_ACTIVE:
			note := GetNote("Deal has been completed.", onChainStatus, onChainMessage)
			logs.GetLogger().Info(GetLog(deal, note))
			updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_ACTIVE, note)
			if !updated {
				logs.GetLogger().Error("Failed to update offline deal status")
			} else {
				logs.GetLogger().Info(GetLogFromStatus(deal, DEAL_STATUS_ACTIVE, ""))
			}
		case ONCHAIN_DEAL_STATUS_AWAITING:
			currentEpoch := lotusService.LotusClient.LotusGetCurrentEpoch()
			if currentEpoch < 0 {
				return
			}

			if currentEpoch > deal.StartEpoch {
				note := GetNote("Sector is proved and active.", onChainStatus, onChainMessage)
				logs.GetLogger().Warn(GetLog(deal, note))
				updated := swanClient.SwanUpdateOfflineDealStatus(deal.Id, DEAL_STATUS_IMPORT_FAILED, note)
				if !updated {
					logs.GetLogger().Error("Failed to update offline deal status")
				} else {
					logs.GetLogger().Warn(GetLogFromStatus(deal, DEAL_STATUS_IMPORT_FAILED, "on chain status bug"))
				}
			}
		}
	}
}

package boost

import (
	"context"
	"fmt"
	boostapi "github.com/filecoin-project/boost/api"
	"github.com/filecoin-project/boost/storagemarket/types"
	"github.com/filecoin-project/boost/storagemarket/types/dealcheckpoints"
	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	"github.com/filswan/go-swan-lib/model"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"net/http"
)

type Client struct {
	stub boostapi.BoostStruct
}

func NewClient(authToken, apiUrl string) (*Client, jsonrpc.ClientCloser, error) {
	var headers http.Header
	if authToken != "" {
		headers = http.Header{"Authorization": []string{"Bearer " + authToken}}
	} else {
		headers = nil
	}

	var apiSub boostapi.BoostStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+apiUrl+"/rpc/v0", "Filecoin",
		[]interface{}{&apiSub.Internal}, headers)
	if err != nil {
		return nil, nil, errors.Wrap(err, "connecting with boost failed")
	}

	return &Client{
		stub: apiSub,
	}, closer, nil
}

func (pc *Client) GetDealInfoByDealUuid(ctx context.Context, dealUuid string) (*model.ProviderDealState, error) {
	dealUid, err := uuid.Parse(dealUuid)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("dealUuid=[%s] parse failed", dealUid))
	}
	boostDeal, err := pc.stub.BoostDeal(ctx, dealUid)
	if err != nil {
		return nil, err
	}
	var pds model.ProviderDealState
	pds.DealUuid = boostDeal.DealUuid.String()
	pds.IsOffline = boostDeal.IsOffline
	pds.DealDataRoot = boostDeal.DealDataRoot.String()
	pds.ChainDealID = uint64(boostDeal.ChainDealID)
	pds.PublishCID = boostDeal.PublishCID.String()
	pds.DealStatus = statusMessage(&types.DealStatusResponse{
		DealUUID:  boostDeal.DealUuid,
		Error:     boostDeal.Err,
		IsOffline: boostDeal.IsOffline,
		DealStatus: &types.DealStatus{
			Error:  boostDeal.Err,
			Status: boostDeal.Checkpoint.String(),
		},
	})
	pds.Err = boostDeal.Err
	return &pds, nil
}

func (pc *Client) OfflineDealWithData(ctx context.Context, dealUuid, filePath string) (*model.ProviderDealRejectionInfo, error) {
	dealUid, err := uuid.Parse(dealUuid)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("dealUuid=[%s] parse failed", dealUid))
	}
	offlineDealWithData, err := pc.stub.BoostOfflineDealWithData(ctx, dealUid, filePath)
	if err != nil {
		return nil, err
	}
	return &model.ProviderDealRejectionInfo{
		Accepted: offlineDealWithData.Accepted,
		Reason:   offlineDealWithData.Reason,
	}, nil
}

func statusMessage(resp *types.DealStatusResponse) string {
	switch resp.DealStatus.Status {
	case dealcheckpoints.Accepted.String():
		if resp.IsOffline {
			return "Awaiting Offline Data Import"
		}
	case dealcheckpoints.Transferred.String():
		return "Ready to Publish"
	case dealcheckpoints.Published.String():
		return "Awaiting Publish Confirmation"
	case dealcheckpoints.PublishConfirmed.String():
		return "Adding to Sector"
	case dealcheckpoints.AddedPiece.String():
		return "Announcing"
	case dealcheckpoints.IndexedAndAnnounced.String():
		return "Sealing"
	case dealcheckpoints.Complete.String():
		if resp.DealStatus.Error != "" {
			return "Error: " + resp.DealStatus.Error
		}
		return "Expired"
	}
	return resp.DealStatus.Status
}

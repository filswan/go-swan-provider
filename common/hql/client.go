package hql

import (
	"context"
	"errors"
	"github.com/Khan/genqlient/graphql"
	"net/http"
	"swan-provider/common/hql/gen"
)

type Client struct {
	hqlClient graphql.Client
}

func NewClient(endpoint string) (*Client, error) {
	if endpoint == "" || len(endpoint) == 0 {
		return nil, errors.New("graphql url is required")
	}
	client := graphql.NewClient(endpoint, http.DefaultClient)
	return &Client{client}, nil
}

func (c Client) GetDealByUuid(dealUuid string) (*gen.DealResponse, error) {
	return gen.Deal(context.TODO(), c.hqlClient, dealUuid)
}

func (c Client) GetProposalCid(proposalCid string) (*gen.LegacyDealResponse, error) {
	return gen.LegacyDeal(context.TODO(), c.hqlClient, proposalCid)
}

func (c Client) GetSectorStates() (*gen.GetSectorStatesResponse, error) {
	return gen.GetSectorStates(context.TODO(), c.hqlClient)
}

func (c Client) GetDealListByStatus(checkPoint gen.Checkpoint) (*gen.GetDealListByStatusResponse, error) {
	return gen.GetDealListByStatus(context.TODO(), c.hqlClient, checkPoint)
}

func (c Client) GetLegacyDeals() (*gen.GetLegacyDealsResponse, error) {
	return gen.GetLegacyDeals(context.TODO(), c.hqlClient)
}

var Checkpoint = map[gen.Checkpoint]string{
	"Accepted":            "Accepted",
	"Transferred":         "Transferred",
	"Published":           "Published",
	"PublishConfirmed":    "PublishConfirmed",
	"AddedPiece":          "AddedPiece",
	"IndexedAndAnnounced": "IndexedAndAnnounced",
	"Complete":            "Complete",
}

func DealStatus(checkpoint gen.Checkpoint, err string) string {
	switch checkpoint {
	case "Accepted":
		return "StorageDealWaitingForData"
	case "Transferred":
		fallthrough
	case "Published":
		fallthrough
	case "PublishConfirmed":
		return "StorageDealAwaitingPreCommit"
	case "AddedPiece":
		fallthrough
	case "IndexedAndAnnounced":
		return "StorageDealSealing"
	case "Complete":
		switch err {
		case "":
			return "StorageDealActive"
		case "Cancelled":
			return "StorageDealNotFound"
		}
		return "StorageDealError"
	}
	return "StorageDealNotFound"
}

func Message(checkpoint gen.Checkpoint, err string) string {
	switch checkpoint {
	case "Accepted":
		return "Awaiting Offline Data Import"
	case "Transferred":
		return "Ready to Publish"
	case "Published":
		return "Awaiting Publish Confirmation"
	case "PublishConfirmed":
		return "Adding to Sector"
	case "AddedPiece":
		fallthrough
	case "IndexedAndAnnounced":
		return "sealing"
	case "Complete":
		switch err {
		case "":
			return "Complete"
		case "Cancelled":
			return "Cancelled"
		}
		return "Error: " + err
	}
	return "unknow"
}

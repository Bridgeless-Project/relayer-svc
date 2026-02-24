package containers

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/connector"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/logan/v3"
)

type updateSignersCatchupContainer struct {
	id uint32

	dbQ              db.SignaturesQ
	epoch            *db.Epoch
	chainClient      chain.ChildClient
	coreConnector    *connector.Connector
	tendermintClient *http.HTTP

	logger *logan.Entry
}

func NewUpdateSignersCatchUpContainer(
		chainClient chain.ChildClient,
		epoch db.Epoch, dbQ db.SignaturesQ,
		connector *connector.Connector,
		tendermintClient *http.HTTP,
		logger *logan.Entry,
	) UpdateSignersContainers {
	return &updateSignersCatchupContainer{
		id:               epoch.Id,
		chainClient:      chainClient,
		epoch:          &epoch,
		dbQ:              dbQ,
		tendermintClient: tendermintClient,
		coreConnector:    connector,
		logger:           logger.WithField("catchup_container", epoch.Id),
	}
}

func (c *updateSignersCatchupContainer) ID() uint32 {
	return c.id
}

func (c *updateSignersCatchupContainer) Run(ctx context.Context) (*db.Epoch, error) {
	id := db.SignatureIdentifier{Id: c.epoch.Id, ChainId: c.epoch.ChainId, Nonce: c.epoch.Nonce}

	err := executeUpdateSigners(ctx, c.chainClient, c.epoch, c.tendermintClient, c.logger)
	if err != nil {
		updateErr := c.dbQ.UpdateStatus(id, internalTypes.EpochStatus_EPOCH_STATUS_FAILED)
		if updateErr != nil {
			c.logger.WithError(updateErr).Error("failed to update epoch status to failed")
		}
		return nil, errors.Wrap(err, "failed to catch up epoch")
	}

	err = c.dbQ.UpdateStatus(id, internalTypes.EpochStatus_EPOCH_STATUS_PROCESSED)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update epoch status to processed")
	}

	return c.epoch, nil
}

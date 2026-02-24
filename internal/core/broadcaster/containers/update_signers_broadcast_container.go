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

type updateSignersBroadcastContainer struct {
	id               uint32
	dbQ              db.SignaturesQ
	epoch            *db.Epoch
	tendermintClient *http.HTTP
	chainClient      chain.ChildClient
	coreConnector    *connector.Connector

	logger *logan.Entry
}

func NewUpdateSignersBroadcastContainer(chainClient chain.ChildClient, epoch *db.Epoch, dbQ db.SignaturesQ,
	coreConnector *connector.Connector, tendermintClient *http.HTTP, logger *logan.Entry) UpdateSignersContainers {
	return &updateSignersBroadcastContainer{
		id:               epoch.Id,
		chainClient:      chainClient,
		epoch:            epoch,
		tendermintClient: tendermintClient,
		dbQ:              dbQ,
		coreConnector:    coreConnector,
		logger:           logger.WithField("broadcast_container", epoch.Id),
	}
}

func (b *updateSignersBroadcastContainer) ID() uint32 {
	return b.id
}

func (b *updateSignersBroadcastContainer) Run(ctx context.Context) (*db.Epoch, error) {
	id := db.SignatureIdentifier{Id: b.epoch.Id, ChainId: b.epoch.ChainId, Nonce: b.epoch.Nonce}

	err := executeUpdateSigners(ctx, b.chainClient, b.epoch, b.tendermintClient, b.logger)
	if err != nil {
		updateErr := b.dbQ.UpdateStatus(id, internalTypes.EpochStatus_EPOCH_STATUS_FAILED)
		if updateErr != nil {
			b.logger.WithError(updateErr).Error("failed to update epoch status to failed")
		}
		return nil, errors.Wrap(err, "failed to process update signers")
	}

	err = b.dbQ.UpdateStatus(id, internalTypes.EpochStatus_EPOCH_STATUS_PROCESSED)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update epoch status to processed")
	}

	return b.epoch, nil
}
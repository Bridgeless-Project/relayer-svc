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

type broadcastContainer struct {
	id               string
	dbQ              db.DepositsQ
	deposit          *db.Deposit
	tendermintClient *http.HTTP
	chainClient      chain.Client
	coreConnector    *connector.Connector

	logger *logan.Entry
}

func NewBroadcastContainer(chainClient chain.Client, deposit db.Deposit, dbQ db.DepositsQ,
	coreConnector *connector.Connector, tendermintClient *http.HTTP, logger *logan.Entry) WithdrawalContainer {
	return &broadcastContainer{
		id:               deposit.String(),
		chainClient:      chainClient,
		deposit:          &deposit,
		tendermintClient: tendermintClient,
		dbQ:              dbQ,
		coreConnector:    coreConnector,
		logger:           logger.WithField("broadcast_container", deposit.String()),
	}
}

func (b *broadcastContainer) ID() string {
	return b.id
}

func (b *broadcastContainer) Run(ctx context.Context) (*db.Deposit, error) {
	processed, err := b.isAlreadyProcessed(ctx)
	if err != nil {
		return b.deposit, errors.Wrap(err, "failed to validate deposit")
	}

	if processed {
		err = b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_ALREADY_EXISTS)
		if err != nil {
			return b.deposit, errors.Wrap(err, "failed to update deposit status to already exists")
		}

		return nil, internalTypes.ErrAlreadyExists
	}

	if err = b.dbQ.UpdateStatus(b.deposit.DepositIdentifier,
		internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSING); err != nil {
		return nil, errors.Wrap(err, "failed to update deposit status processing")
	}

	b.deposit, err = executeWithdrawal(ctx, b.chainClient, *b.deposit, b.tendermintClient, b.logger)
	if err != nil {
		updateErr := b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED)
		if updateErr != nil {
			b.logger.WithError(updateErr).Error("failed to update deposit status to FAILED")
		}

		if b.deposit.WithdrawalTxHash == nil {
			b.deposit.WithdrawalTxHash = ptr(defaultWithdrawalHash)
		}

		updateErr = b.dbQ.UpdateWithdrawalDetails(*b.deposit)
		if updateErr != nil {
			b.logger.WithError(updateErr).Error("failed to update withdrawal details")
		}

		return b.deposit, errors.Wrap(err, "failed to process withdrawal")
	}

	if err = b.dbQ.UpdateWithdrawalDetails(*b.deposit); err != nil {

		updateErr := b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED)
		if updateErr != nil {
			b.logger.WithError(updateErr).Error("failed to update deposit status to FAILED")
		}

		return nil, errors.Wrap(err, "failed to update withdrawal details")
	}

	b.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE
	err = b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update deposit withdrawal status to submit")
	}

	return b.deposit, nil
}

func (b *broadcastContainer) isAlreadyProcessed(ctx context.Context) (bool, error) {
	processed, err := b.chainClient.IsProcessed(ctx, *b.deposit)
	if err != nil {
		return false, errors.Wrap(err, "error validating withdrawal existence on chain")
	}

	return processed, nil
}

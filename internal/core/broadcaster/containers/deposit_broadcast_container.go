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

type depositBroadcastContainer struct {
	id               string
	dbQ              db.DepositsQ
	deposit          *db.Deposit
	tendermintClient *http.HTTP
	chainClient      chain.ChildClient
	coreConnector    *connector.Connector

	logger *logan.Entry
}

func NewDepositBroadcastContainer(chainClient chain.ChildClient, deposit db.Deposit, dbQ db.DepositsQ,
	coreConnector *connector.Connector, tendermintClient *http.HTTP, logger *logan.Entry) WithdrawalContainer {
	return &depositBroadcastContainer{
		id:               deposit.String(),
		chainClient:      chainClient,
		deposit:          &deposit,
		tendermintClient: tendermintClient,
		dbQ:              dbQ,
		coreConnector:    coreConnector,
		logger:           logger.WithField("broadcast_container", deposit.String()),
	}
}

func (b *depositBroadcastContainer) ID() string {
	return b.id
}

func (b *depositBroadcastContainer) Run(ctx context.Context) (*db.Deposit, error) {
	processed, err := b.chainClient.IsProcessed(ctx, *b.deposit)
	if err != nil {

		b.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED
		updateErr := b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, b.deposit.WithdrawalStatus)
		if updateErr != nil {
			b.logger.WithError(updateErr).Error("error updating deposit status to failed")
		}

		return b.deposit, errors.Wrap(err, "failed to validate deposit")
	}

	if processed {
		b.logger.Warnf("deposit %s already withdrawn", b.deposit.String())

		b.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_ALREADY_WITHDRAWN
		err = b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, b.deposit.WithdrawalStatus)
		if err != nil {
			b.logger.WithError(err).Error("failed to update deposit status to already-withdrawn")
		}

		b.deposit.WithdrawalTxHash = ptr(defaultWithdrawalHash)
		err = b.dbQ.UpdateWithdrawalDetails(*b.deposit)
		if err != nil {
			b.logger.WithError(err).Error("failed to update withdrawal details")
		}

		// let the service submit the null hash to core
		return b.deposit, nil
	}

	b.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSING
	if err = b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, b.deposit.WithdrawalStatus); err != nil {
		return nil, errors.Wrap(err, "failed to update deposit status processing")
	}

	err = executeWithdrawal(ctx, b.chainClient, b.deposit, b.tendermintClient, b.logger)
	if err != nil {

		b.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED
		updateErr := b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, b.deposit.WithdrawalStatus)
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

		b.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED
		updateErr := b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, b.deposit.WithdrawalStatus)
		if updateErr != nil {
			b.logger.WithError(updateErr).Error("failed to update deposit status to FAILED")
		}

		return b.deposit, errors.Wrap(err, "failed to update withdrawal details")
	}

	b.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE
	err = b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, b.deposit.WithdrawalStatus)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update deposit withdrawal status to submit")
	}

	return b.deposit, nil
}

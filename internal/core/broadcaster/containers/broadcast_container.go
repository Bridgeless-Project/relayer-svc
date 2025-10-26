package containers

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

type broadcastContainer struct {
	dbQ         db.DepositsQ
	deposit     db.Deposit
	chainClient chain.Client

	logger *logan.Entry
}

func NewBroadcastContainer(chainClient chain.Client, deposit db.Deposit, dbQ db.DepositsQ, logger *logan.Entry) WithdrawalContainer {
	return &broadcastContainer{
		chainClient: chainClient,
		deposit:     deposit,
		dbQ:         dbQ,
		logger:      logger.WithField("broadcast_container", deposit.String()),
	}
}

func (b *broadcastContainer) ID() string {
	return b.ID()
}

func (b *broadcastContainer) Run(ctx context.Context) (*db.Deposit, error) {
	processed, err := b.isAlreadyProcessed(ctx)
	if err != nil {
		return &b.deposit, errors.Wrap(err, "failed to validate deposit")
	}

	if processed {
		err = b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_ALREADY_EXISTS)
		if err != nil {
			return &b.deposit, errors.Wrap(err, "failed to update deposit status to already exists")
		}

		return &b.deposit, internalTypes.ErrAlreadyExists
	}

	if err = b.dbQ.UpdateStatus(b.deposit.DepositIdentifier,
		internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSING); err != nil {
		return &b.deposit, errors.Wrap(err, "failed to update deposit status processing")
	}

	if err = executeWithdrawal(ctx, b.chainClient, b.deposit, b.logger); err != nil {
		b.logger.WithError(err).Error("failed to process deposit")

		updateErr := b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED)
		if updateErr != nil {
			b.logger.WithError(updateErr).Error("failed to update deposit status to FAILED")
		}

		return &b.deposit, errors.Wrap(err, "failed to process deposit")
	}

	if err = b.dbQ.UpdateWithdrawalTx(b.deposit.DepositIdentifier, *b.deposit.WithdrawalTxHash); err != nil {

		updateErr := b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED)
		if updateErr != nil {
			b.logger.WithError(updateErr).Error("failed to update deposit status to FAILED")
		}

		return &b.deposit, errors.Wrap(err, "failed to update deposit withdrawal tx")
	}

	b.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE
	err = b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE)
	if err != nil {
		return &b.deposit, errors.Wrap(err, "failed to update deposit withdrawal status to submit")
	}

	return &b.deposit, nil
}

func (b *broadcastContainer) isAlreadyProcessed(ctx context.Context) (bool, error) {
	processed, err := b.chainClient.IsProcessed(ctx, b.deposit)
	if err != nil {
		return false, errors.Wrap(err, "error validating withdrawal existence on chain")
	}

	return processed, nil
}

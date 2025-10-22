package broadcaster

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

type broadcastContainer struct {
	id          uuid.UUID
	dbQ         db.DepositsQ
	deposit     db.Deposit
	chainClient chain.Client

	logger *logan.Entry
}

func NewContainer(id string, chainClient chain.Client, deposit db.Deposit, dbQ db.DepositsQ, logger *logan.Entry) *broadcastContainer {
	return &broadcastContainer{
		id:          uuid.New(),
		chainClient: chainClient,
		deposit:     deposit,
		dbQ:         dbQ,
		logger:      logger.WithField("container", id),
	}
}

func (b *broadcastContainer) Run(ctx context.Context) (*db.Deposit, error) {
	valid, err := b.validate(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate deposit")
	}

	if !valid {
		err = b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_ALREADY_EXISTS)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update deposit status")
		}
	}

	if err = b.process(ctx); err != nil {
		b.logger.WithError(err).Error("failed to process deposit")

		err = b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update deposit status")
		}

		return nil, nil
	}

	if err = b.dbQ.UpdateWithdrawalTx(b.deposit.DepositIdentifier, *b.deposit.WithdrawalTxHash); err != nil {
		return nil, errors.Wrap(err, "failed to update deposit withdrawal tx")
	}

	if err = b.dbQ.UpdateStatus(b.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED); err != nil {
		return nil, errors.Wrap(err, "failed to update deposit withdrawal status")
	}

	return &b.deposit, nil
}

func (b *broadcastContainer) validate(ctx context.Context) (bool, error) {
	processed, err := b.chainClient.IsProcessed(ctx, b.deposit)
	if err != nil {
		return false, errors.Wrap(err, "error validating withdrawal existence on chain")
	}

	return processed, nil
}

func (b *broadcastContainer) process(ctx context.Context) error {
	var (
		txHash string
		err    error
	)

	switch b.deposit.WithdrawalToken {
	case core.DefaultNativeTokenAddress:
		txHash, err = b.chainClient.WithdrawNative(ctx, b.deposit)
	default:
		txHash, err = b.chainClient.WithdrawToken(ctx, b.deposit)
	}
	if err != nil {
		return errors.Wrap(err, "error processing withdrawal")
	}

	b.logger.Infof("Processed deposit %s withdrawal hash %s", b.deposit.String(), txHash)
	b.deposit.WithdrawalTxHash = &txHash

	return nil
}

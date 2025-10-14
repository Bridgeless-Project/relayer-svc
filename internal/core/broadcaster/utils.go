package broadcaster

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
)

func (b *Broadcaster) validateExistence(ctx context.Context, deposit db.Deposit) error {
	_, exists := b.cache.Load(deposit.String())
	if exists {
		return errWithdrawalInProcess
	}

	client, err := b.clientsRepo.Client(deposit.WithdrawalChainId)
	if err != nil {
		b.logger.WithError(err).Error("error validating existence of withdrawal")
		return errWithdraw
	}

	exists, err = client.IsProcessed(ctx, deposit)
	if err != nil {
		b.logger.WithError(err).Error("error validating existence of withdrawal")
		return errWithdraw
	}
	if exists {
		return errWithdrawalInProcess
	}

	return nil
}

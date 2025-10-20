package broadcaster

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
)

func (b *Broadcaster) validateExistence(ctx context.Context, deposit db.Deposit) error {
	_, exists := b.cache.Load(deposit.String())
	if exists {
		return errAlreadyExists
	}

	client, err := b.clientsRepo.Client(deposit.WithdrawalChainId)
	if err != nil {
		return errors.Wrap(err, "failed to get withdrawal chain client")
	}

	exists, err = client.IsProcessed(ctx, deposit)
	if err != nil {
		return errors.Wrap(err, "failed to check if deposit processed on chain")
	}

	println(exists)
	if exists {
		return errAlreadyExists
	}

	return nil
}

package broadcaster

import (
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
)

func (b *Broadcaster) validateExistence(deposit db.Deposit) error {
	_, exists := b.cache.Load(deposit.DepositIdentifier.String())
	if exists {
		return errWithdrawalInProcess
	}

	depositData, err := b.dbConn.Get(deposit.DepositIdentifier)
	if err != nil {
		return errors.Wrap(err, "failed to get deposit data")
	}

	if depositData != nil {
		return errWithdrawalAlreadyExists
	}

	return nil
}

func isInternalError(err error) bool {
	return errors.Is(err, errWithdraw) || errors.Is(err, chain.ErrChainNotSupported)
}

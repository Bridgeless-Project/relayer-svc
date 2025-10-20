package broadcaster

import (
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
)

// checkExistence checks whether deposit persists in cache or database
func (b *Broadcaster) checkExistence(deposit db.Deposit) error {
	_, exists := b.cache.Load(deposit.String())
	if exists {
		return errAlreadyExists
	}

	depositData, err := b.dbConn.Get(deposit.DepositIdentifier)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve deposit data")
	}

	if depositData != nil {
		return errAlreadyExists
	}

	return nil
}

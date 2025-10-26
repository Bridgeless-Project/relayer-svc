package containers

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
)

type WithdrawalContainer interface {
	Run(ctx context.Context) (*db.Deposit, error)
	ID() string
}

package containers

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
)

const (
	defaultWithdrawalHash = "0x0000000000000000000000000000000000000000"
)

type WithdrawalContainer interface {
	Run(ctx context.Context) (*db.Deposit, error)
	ID() string
}

type UpdateSignersContainers interface {
	Run(ctx context.Context) (*db.Epoch, error)
	ID() uint32
}

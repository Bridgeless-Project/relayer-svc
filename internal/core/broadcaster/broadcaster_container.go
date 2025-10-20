package broadcaster

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/google/uuid"
	"gitlab.com/distributed_lab/logan/v3"
)

type broadcastContainer struct {
	id          uuid.UUID
	dbQ         db.DepositsQ
	deposit     db.Deposit
	chainClient chain.Client

	logger *logan.Entry
}

func NewContainer(id uuid.UUID, chainClient chain.Client, deposit db.Deposit, dbQ db.DepositsQ, logger *logan.Entry) *broadcastContainer {
	deposit.Id = id.String()
	return &broadcastContainer{
		id:          uuid.New(),
		chainClient: chainClient,
		deposit:     deposit,
		dbQ:         dbQ.New(),
		logger:      logger.WithField("container", id.String()),
	}
}

func Run(ctx context.Context) (*db.Deposit, error) {
	return nil, nil
}

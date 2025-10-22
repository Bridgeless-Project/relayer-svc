package broadcaster

import (
	"context"
	"fmt"
	"sync"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/connector"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

type Broadcaster struct {
	coreConnector *connector.Connector
	clientsRepo   chain.Repository
	handlerChan   chan *broadcastContainer

	dbConn db.DepositsQ
	logger *logan.Entry
	cache  sync.Map
}

func New(coreConnector *connector.Connector, dbConn db.DepositsQ, clientsRepo chain.Repository, logger *logan.Entry) *Broadcaster {
	return &Broadcaster{
		coreConnector: coreConnector,
		clientsRepo:   clientsRepo,
		handlerChan:   make(chan *broadcastContainer),
		dbConn:        dbConn,
		logger:        logger,
	}
}

func (b *Broadcaster) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			b.logger.Debug("context canceled. Stopping broadcaster")
			return
		case container, ok := <-b.handlerChan:
			if !ok {
				ctx.Done()
				b.logger.Debug("deposit channel is closed. Stopping broadcaster")
				return
			}

			_, err := container.Run(ctx)
			if err != nil {
				b.logger.WithError(err).Error(fmt.Sprintf("error processing withdrawal %s", container.id))
			}
		}
	}
}

func (b *Broadcaster) Broadcast(deposit db.Deposit) error {
	_, ok := b.cache.Load(deposit.String())
	if ok {
		return errAlreadyExists
	}

	deposit.WithdrawalStatus = types.WithdrawalStatus_WITHDRAWAL_STATUS_PENDING
	err := b.dbConn.Insert(deposit)
	if err != nil {
		if errors.Is(err, db.ErrAlreadySubmitted) {
			// Store duplicate deposit identifier to cache to avoid spamming db with get queries
			b.cache.Store(deposit.String(), nil)
			return errAlreadyExists
		}

		b.logger.WithError(err).Error("error inserting deposit")
		return types.ErrFailedToBroadcast
	}

	b.cache.Store(deposit.String(), nil)

	chainClient, err := b.clientsRepo.Client(deposit.WithdrawalChainId)
	if err != nil {
		b.logger.WithError(err).Error("failed to get the withdrawal chain client")
		return types.ErrFailedToBroadcast
	}

	go func() {
		b.handlerChan <- NewContainer(uuid.New().String(), chainClient, deposit, b.dbConn, b.logger)
	}()

	return nil
}

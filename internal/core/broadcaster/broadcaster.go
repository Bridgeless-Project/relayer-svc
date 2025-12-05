package broadcaster

import (
	"context"
	"sync"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster/containers"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/connector"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/logan/v3"
)

type Broadcaster struct {
	ctx context.Context

	coreConnector *connector.Connector
	clientsRepo   chain.Repository
	workersMap    map[string]chan containers.WithdrawalContainer
	submitChan    chan *db.Deposit

	tendermintClient *http.HTTP

	dbConn db.DepositsQ
	logger *logan.Entry
	cache  sync.Map

	wg *sync.WaitGroup

	chainTxPoolSize int64
	submitBatchSize int64
}

func New(ctx context.Context, coreConnector *connector.Connector, dbConn db.DepositsQ, tendermintClient *http.HTTP, logger *logan.Entry) *Broadcaster {
	return &Broadcaster{
		ctx:              ctx,
		coreConnector:    coreConnector,
		dbConn:           dbConn,
		logger:           logger,
		cache:            sync.Map{},
		wg:               new(sync.WaitGroup),
		tendermintClient: tendermintClient,
	}
}

func (b *Broadcaster) Run(ctx context.Context) {
	b.workersMap = make(map[string]chan containers.WithdrawalContainer)

	for chainID, client := range b.clientsRepo.Clients() {
		handlerChan := make(chan containers.WithdrawalContainer, b.chainTxPoolSize)

		for id := range client.WorkersCount() {
			b.wg.Add(1)
			go b.runNetworkWorker(ctx, chainID, handlerChan, id)
		}

		b.workersMap[chainID] = handlerChan
	}

	b.wg.Add(1)
	go b.runCoreSubmitter(ctx)

	b.wg.Wait()
	for _, ch := range b.workersMap {
		close(ch)
	}

	close(b.submitChan)
}

func (b *Broadcaster) Broadcast(deposit db.Deposit) error {
	err := b.checkExistence(context.Background(), deposit)
	if err != nil {
		if errors.Is(err, internalTypes.ErrAlreadyExists) {
			return errors.Wrap(err, "deposit already exists")
		}

		return errors.Wrap(internalTypes.ErrFailedToBroadcast, err.Error())
	}

	deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PENDING
	err = b.dbConn.Insert(deposit)
	if err != nil {
		return errors.Wrapf(internalTypes.ErrFailedToBroadcast, "%s: error storing deposit %s",
			err.Error(), deposit.String())
	}

	chainClient, err := b.clientsRepo.Client(deposit.WithdrawalChainId)
	if err != nil {
		return errors.Wrapf(internalTypes.ErrFailedToBroadcast, "failed to get the withdrawal chain client, error: %s", err.Error())
	}

	// Store duplicate deposit identifier to cache to avoid spamming db with get queries
	_, ok := b.cache.Load(deposit.TxHash)
	if !ok {
		b.cache.Store(deposit.String(), nil)
	}

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		select {
		case <-b.ctx.Done():
			b.logger.Warnf("stopped broadcastig of deposit %s", deposit.String())
			return

		case b.workersMap[deposit.WithdrawalChainId] <- containers.NewBroadcastContainer(
			chainClient,
			deposit,
			b.dbConn,
			b.coreConnector,
			b.tendermintClient,
			b.logger,
		):
			return
		}

	}()

	return nil
}

func (b *Broadcaster) CatchUp(deposit db.Deposit) error {
	_, ok := b.cache.Load(deposit.String())
	if ok {
		return internalTypes.ErrAlreadyExists
	}

	err := b.dbConn.UpdateStatus(deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PENDING)
	if err != nil {
		return errors.Wrapf(err, "failed to update status to pending: %s", deposit.String())
	}
	b.cache.Store(deposit.String(), nil)

	chainClient, err := b.clientsRepo.Client(deposit.WithdrawalChainId)
	if err != nil {
		return errors.Wrapf(err, "failed to get the withdrawal chain client: %s", deposit.String())
	}

	go func() {
		defer b.wg.Done()
		select {
		case <-b.ctx.Done():
			b.logger.Warnf("stopped broadcastig of deposit %s", deposit.String())
			return

		case b.workersMap[deposit.WithdrawalChainId] <- containers.NewCatchUpContainer(
			chainClient,
			deposit,
			b.dbConn,
			b.coreConnector,
			b.tendermintClient,
			b.logger,
		):
			return
		}

	}()

	return nil
}

// checkExistence checks whether deposit persists in cache,database or chain
func (b *Broadcaster) checkExistence(ctx context.Context, deposit db.Deposit) error {
	_, exists := b.cache.Load(deposit.String())
	if exists {
		return errors.Wrapf(internalTypes.ErrAlreadyExists, "deposit %s already registered in service", deposit.String())
	}

	client, err := b.clientsRepo.Client(deposit.WithdrawalChainId)
	if err != nil {
		return errors.Wrap(err, "failed to get the withdrawal chain client")
	}

	exists, err = client.IsProcessed(ctx, deposit)
	if err != nil {
		return errors.Wrap(err, "failed to check if deposit is processed on-chain")
	}

	if exists {
		b.cache.Store(deposit.String(), nil)
		return errors.Wrapf(internalTypes.ErrAlreadyExists, "deposit %s already processed on-chain", deposit.String())
	}

	depositData, err := b.dbConn.Get(deposit.DepositIdentifier)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve deposit data")
	}

	if depositData != nil {
		return errors.Wrapf(internalTypes.ErrAlreadyExists, "deposit %s already saved", deposit.String())
	}

	return nil
}

func (b *Broadcaster) WithClients(clients chain.Repository) *Broadcaster {
	b.clientsRepo = clients
	return b
}

func (b *Broadcaster) WithChainTxPoolSize(txPoolSize int64) *Broadcaster {
	b.chainTxPoolSize = txPoolSize
	return b
}

func (b *Broadcaster) WithSubmitTxPoolSize(txPoolSize int64) *Broadcaster {
	b.submitChan = make(chan *db.Deposit, txPoolSize)

	return b
}

func (b *Broadcaster) WithSubmitBatchSize(batchSize int64) *Broadcaster {
	b.submitBatchSize = batchSize
	return b
}

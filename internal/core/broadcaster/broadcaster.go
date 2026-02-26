package broadcaster

import (
	"context"
	"math/rand"
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

	coreConnector           *connector.Connector
	clientsRepo             chain.Repository
	withdrawalWorkersMap    map[string]chan containers.WithdrawalContainer
	updateSignersWorkersMap map[string]chan containers.UpdateSignersContainers
	submitChan              chan *db.Deposit

	tendermintClient *http.HTTP

	depositsDbConn db.DepositsQ
	epochsDbConn db.SignaturesQ
	logger *logan.Entry
	cache  sync.Map

	wg *sync.WaitGroup

	chainTxPoolSize int64
	submitBatchSize int64
}

func New(
		ctx context.Context, coreConnector *connector.Connector,
		depositsDbConn db.DepositsQ, epochsDbConn db.SignaturesQ,
		tendermintClient *http.HTTP, logger *logan.Entry,
	) *Broadcaster {
	return &Broadcaster{
		ctx:              ctx,
		coreConnector:    coreConnector,
		depositsDbConn:   depositsDbConn,
		epochsDbConn: 		epochsDbConn,
		logger:           logger,
		cache:            sync.Map{},
		wg:               new(sync.WaitGroup),
		tendermintClient: tendermintClient,
	}
}

func (b *Broadcaster) Run(ctx context.Context) {
	b.withdrawalWorkersMap = make(map[string]chan containers.WithdrawalContainer)
	b.updateSignersWorkersMap = make(map[string]chan containers.UpdateSignersContainers)

	for chainID, client := range b.clientsRepo.Clients() {
		withdrawalHandlerChan := make(chan containers.WithdrawalContainer, b.chainTxPoolSize)
		updateSignersHandlerChan := make(chan containers.UpdateSignersContainers, b.chainTxPoolSize)

		for id := range client.WorkersCount() {
			b.wg.Go(func() {
				b.runWithdrawalNetworkWorker(ctx, chainID, withdrawalHandlerChan, id)
			})
			b.wg.Go(func() {
				b.runUpdateSignersNetworkWorker(ctx, chainID, updateSignersHandlerChan, id)
			})
		}

		b.withdrawalWorkersMap[chainID] = withdrawalHandlerChan
		b.updateSignersWorkersMap[chainID] = updateSignersHandlerChan
	}

	b.wg.Go(func() {
		b.runCoreSubmitter(ctx)
	})
	b.wg.Wait()

	for _, ch := range b.updateSignersWorkersMap {
		close(ch)
	}

	for _, ch := range b.withdrawalWorkersMap {
		close(ch)
	}

	close(b.submitChan)
}

func (b *Broadcaster) BroadcastEpoch(epoch *db.Epoch) error {
	existing, err := b.epochsDbConn.Get(db.SignatureIdentifier{Id: epoch.Id, ChainId: epoch.ChainId, Nonce: epoch.Nonce})
	if err != nil {
		return errors.Wrapf(internalTypes.ErrFailedToBroadcast, "failed to check epoch existence: %s", err.Error())
	}
	if existing != nil {
		return errors.Wrapf(internalTypes.ErrAlreadyExists, "epoch %d already exists", epoch.Id)
	}

	epoch.Status = internalTypes.EpochStatus_EPOCH_STATUS_PENDING
	err = b.epochsDbConn.Insert(*epoch)
	if err != nil {
		return errors.Wrapf(internalTypes.ErrFailedToBroadcast, "failed to insert epoch: %s", err.Error())
	}

	chainClient, err := b.clientsRepo.Client(epoch.ChainId)
	if err != nil {
		return errors.Wrapf(internalTypes.ErrFailedToBroadcast, "failed to get the epoch chain client, error: %s", err.Error())
	}

	client := chainClient.ChildClients()[rand.Intn(len(chainClient.ChildClients()))]
	b.logger.Debug(client)

	b.wg.Go(func() {
		select {
		case <-b.ctx.Done():
			b.logger.Warnf("stopped broadcasting epoch %d", epoch.Id)
			return

		case b.updateSignersWorkersMap[epoch.ChainId] <- containers.NewUpdateSignersBroadcastContainer(
			client,
			epoch,
			b.epochsDbConn,
			b.coreConnector,
			b.tendermintClient,
			b.logger,
		):
			return
		}

	})

	return nil
}

func (b *Broadcaster) BroadcastDeposit(deposit db.Deposit) error {
	err := b.checkExistence(context.Background(), deposit)
	if err != nil {
		if errors.Is(err, internalTypes.ErrAlreadyExists) {
			return errors.Wrap(err, "deposit already exists")
		}

		return errors.Wrap(internalTypes.ErrFailedToBroadcast, err.Error())
	}

	deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PENDING
	err = b.depositsDbConn.Insert(deposit)
	if err != nil {
		return errors.Wrapf(internalTypes.ErrFailedToBroadcast, "%s: error storing deposit %s",
			err.Error(), deposit.String())
	}

	chainClient, err := b.clientsRepo.Client(deposit.WithdrawalChainId)
	if err != nil {
		return errors.Wrapf(internalTypes.ErrFailedToBroadcast, "failed to get the withdrawal chain client, error: %s", err.Error())
	}

	client := chainClient.ChildClients()[rand.Intn(len(chainClient.ChildClients()))]

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

		case b.withdrawalWorkersMap[deposit.WithdrawalChainId] <- containers.NewDepositBroadcastContainer(
			client,
			deposit,
			b.depositsDbConn,
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

	err := b.depositsDbConn.UpdateStatus(deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PENDING)
	if err != nil {
		return errors.Wrapf(err, "failed to update status to pending: %s", deposit.String())
	}
	b.cache.Store(deposit.String(), nil)

	chainClient, err := b.clientsRepo.Client(deposit.WithdrawalChainId)
	if err != nil {
		return errors.Wrapf(err, "failed to get the withdrawal chain client: %s", deposit.String())
	}

	client := chainClient.ChildClients()[rand.Intn(len(chainClient.ChildClients()))]

	go func() {
		defer b.wg.Done()
		select {
		case <-b.ctx.Done():
			b.logger.Warnf("stopped broadcastig of deposit %s", deposit.String())
			return

		case b.withdrawalWorkersMap[deposit.WithdrawalChainId] <- containers.NewDepositCatchUpContainer(
			client,
			deposit,
			b.depositsDbConn,
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

	depositData, err := b.depositsDbConn.Get(deposit.DepositIdentifier)
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

func (b *Broadcaster) CatchUpEpoch(epoch db.Epoch) error {
	chainClient, err := b.clientsRepo.Client(epoch.ChainId)
	if err != nil {
		return errors.Wrapf(err, "failed to get chain client for epoch %d", epoch.Id)
	}

	client := chainClient.ChildClients()[rand.Intn(len(chainClient.ChildClients()))]

	b.wg.Go(func() {
		select {
		case <-b.ctx.Done():
			b.logger.Warnf("stopped catching up epoch %d", epoch.Id)
			return
		case b.updateSignersWorkersMap[epoch.ChainId] <- containers.NewUpdateSignersCatchUpContainer(
			client,
			epoch,
			b.epochsDbConn,
			b.coreConnector,
			b.tendermintClient,
			b.logger,
		):
			return
		}
	})

	return nil
}

package broadcaster

import (
	"context"
	"sync"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/connector"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

type Broadcaster struct {
	coreConnector *connector.Connector
	clientsRepo   chain.Repository
	depositChan   chan db.Deposit

	dbConn db.DepositsQ
	logger *logan.Entry
	cache  sync.Map
}

func New(coreConnector *connector.Connector, dbConn db.DepositsQ, clientsRepo chain.Repository, logger *logan.Entry) *Broadcaster {
	return &Broadcaster{
		coreConnector: coreConnector,
		clientsRepo:   clientsRepo,
		depositChan:   make(chan db.Deposit, bufferChannelSize),
		dbConn:        dbConn,
		logger:        logger,
	}
}

func (b *Broadcaster) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			b.logger.Debug("context canceled. Stopping broadcaster")
			return nil
		case deposit, ok := <-b.depositChan:
			if !ok {
				ctx.Done()
				b.logger.Debug("deposit channel is closed. Stopping broadcaster")
				return nil
			}

			if err := b.processDeposit(ctx, deposit); err != nil {
				b.logger.WithError(err).Error("error processing deposit")
				err = b.dbConn.UpdateStatus(deposit.DepositIdentifier, types.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED)
				if err != nil {
					b.logger.WithError(err).Error("error updating status")
				}
			}
			b.cache.Delete(deposit.String())
		}
	}
}

func (b *Broadcaster) Broadcast(ctx context.Context, deposit db.Deposit) (*string, error) {
	if err := b.checkExistence(deposit); err != nil {
		if errors.Is(err, errAlreadyExists) {
			return nil, errAlreadyExists
		}

		b.logger.WithError(err).Error("failed to check deposit existence")
		return nil, types.ErrFailedToBroadcast
	}

	b.cache.Store(deposit.String(), struct{}{})

	return nil
}

func (b *Broadcaster) processDeposit(ctx context.Context, deposit db.Deposit) error {
	err := b.dbConn.UpdateStatus(deposit.DepositIdentifier, types.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSING)
	if err != nil {
		return errors.Wrap(err, "failed to, update deposit status")
	}

	client, err := b.clientsRepo.Client(deposit.WithdrawalChainId)
	if err != nil {
		return errors.Wrap(err, "failed to get withdrawal chain client")
	}

	var withdrawalTxHash string
	switch deposit.WithdrawalToken {
	case core.DefaultNativeTokenAddress:
		withdrawalTxHash, err = client.WithdrawNative(ctx, deposit)
		if err != nil {
			return errors.Wrap(err, "failed to process native withdrawal")
		}
	default:
		withdrawalTxHash, err = client.WithdrawToken(ctx, deposit)
		if err != nil {
			return errors.Wrap(err, "failed to process token withdrawal")
		}
	}

	err = b.coreConnector.UpdateTxInfo(ctx, deposit)
	if err != nil {
		return errors.Wrap(err, "failed to update tx info")
	}
	b.logger.Debugf("Processed deposit on chain %s, withdrawal tx: %s", deposit.WithdrawalChainId, withdrawalTxHash)

	err = b.dbConn.UpdateWithdrawalTx(deposit.DepositIdentifier, withdrawalTxHash)
	if err != nil {
		return errors.Wrap(err, "failed to update withdrawal tx hash")
	}

	err = b.dbConn.UpdateStatus(deposit.DepositIdentifier, types.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED)
	if err != nil {
		return errors.Wrap(err, "failed to update withdrawal status")
	}

	return nil
}

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

func New(coreConnector *connector.Connector, dbConn db.DepositsQ, clientsRepo chain.Repository,
	depositChan chan db.Deposit, logger *logan.Entry) *Broadcaster {
	return &Broadcaster{
		coreConnector: coreConnector,
		clientsRepo:   clientsRepo,
		depositChan:   depositChan,
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
				b.cache.Delete(deposit.DepositIdentifier.String())
				continue
			}
			b.cache.Delete(deposit.DepositIdentifier.String())
		}
	}
}

func (b *Broadcaster) Broadcast(deposit db.Deposit) error {
	if err := b.validateExistence(deposit); err != nil {
		if !errors.Is(err, errWithdraw) {
			return err
		}

		b.logger.WithError(err).Error("error broadcasting withdrawal")
		return types.ErrFailedToBroadcast
	}

	b.cache.Store(deposit.DepositIdentifier.String(), struct{}{})
	b.depositChan <- deposit

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
		if deposit.IsWrappedToken {
			withdrawalTxHash, err = client.WithdrawWrapped(ctx, deposit)
			if err != nil {
				return errors.Wrap(err, "failed to process wrapped withdrawal")
			}
		} else {
			withdrawalTxHash, err = client.WithdrawToken(ctx, deposit)
			if err != nil {
				return errors.Wrap(err, "failed to process token withdrawal")
			}
		}
	}

	err = b.coreConnector.UpdateTxInfo(deposit)
	if err != nil {
		return errors.Wrap(err, "failed to update tx info")
	}

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

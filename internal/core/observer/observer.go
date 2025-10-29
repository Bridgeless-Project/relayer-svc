package observer

import (
	"context"
	"time"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client/http"
	"github.com/tendermint/tendermint/rpc/core/types"
	"gitlab.com/distributed_lab/logan/v3"
)

type Observer struct {
	client          *http.HTTP
	retries         uint
	retryTimeout    time.Duration
	pollingInterval time.Duration
	logger          *logan.Entry
	clientsRepo     chain.Repository

	depositsDb db.DepositsQ
	blockDb    db.BlocksQ

	broadcaster *broadcaster.Broadcaster
}

func New(client *http.HTTP, retries uint, retryTimeout, pollingInterval time.Duration, blocksDb db.BlocksQ,
	depositsDb db.DepositsQ, brcst *broadcaster.Broadcaster, clientsRepo chain.Repository, logger *logan.Entry) *Observer {

	return &Observer{
		client:          client,
		retries:         retries,
		retryTimeout:    retryTimeout,
		pollingInterval: pollingInterval,
		blockDb:         blocksDb,
		depositsDb:      depositsDb,
		broadcaster:     brcst,
		clientsRepo:     clientsRepo,
		logger:          logger,
	}
}

func (o *Observer) Run(ctx context.Context, startHeight uint64, catchup bool) error {
	// Firstly catch up pending deposits from db
	if catchup {
		if err := o.catchupWithStatus(internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PENDING); err != nil {
			o.logger.WithError(err).Error("catchup with status pending failed")
		}

		if err := o.catchupWithStatus(internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSING); err != nil {
			o.logger.WithError(err).Error("catchup with status processing failed")
		}

		if err := o.catchupWithStatus(internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE); err != nil {
			o.logger.WithError(err).Error("catchup with status submitting failed")
		}
	}

	// Fetch deposits from Bridgeless core
	if err := o.fetchDeposits(ctx, startHeight); err != nil {
		return errors.Wrap(err, "failed to fetch deposits from the core")
	}

	return nil
}

func (o *Observer) fetchDeposits(ctx context.Context, startHeight uint64) error {
	ticker := time.NewTicker(o.pollingInterval)
	defer ticker.Stop()

	if err := o.blockDb.Insert(db.LatestBlock{BlockId: int64(startHeight)}); err != nil {
		return errors.Wrap(err, "failed to insert latest block")
	}

	if startHeight == 0 {
		latestHeight, err := o.getCurrentHeight(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get current height")
		}
		startHeight = latestHeight
	}

	for {
		select {
		case <-ctx.Done():
			o.logger.Debug("fetching deposits stopped")
			return nil

		case <-ticker.C:
			currentHeight, err := o.getCurrentHeight(ctx)
			if err != nil {
				o.logger.WithError(err).Error("failed to get current height")
				continue
			}

			if startHeight > currentHeight {
				o.logger.Debug("Waiting for next block, currentHeight:", currentHeight)
				continue
			}

			if err = o.blockDb.UpdateLatestBlockId(db.LatestBlock{BlockId: int64(startHeight)}); err != nil {
				o.logger.WithError(err).
					WithField("blockNumber", startHeight).
					Error("failed to update latest block height")
				startHeight++
				continue
			}

			deposits, err := o.fetchSubmitDepositEvents(ctx, int64(startHeight))
			if err != nil {
				o.logger.WithError(err).
					WithField("blockNumber", startHeight).
					Error("failed to fetch submit deposit events")
				startHeight++
				continue
			}

			for _, deposit := range deposits {
				if err = o.broadcastDeposit(*deposit); err != nil {
					if errors.Is(err, skippedDeposit) {
						continue
					}
					o.logger.
						WithField("blockNumber", startHeight).
						Warnf("failed to broadcast deposit: %v", err)
					startHeight++
					continue
				}
			}

			startHeight++
		}
	}
}

func (o *Observer) getCurrentHeight(ctx context.Context) (uint64, error) {
	var info *coretypes.ResultABCIInfo

	getCurrentHeight := func() error {
		var err error
		info, err = o.client.ABCIInfo(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get ABCI info")
		}
		return nil
	}

	if err := core.DoWithRetry(ctx, getCurrentHeight); err != nil {
		return 0, errors.Wrap(err, "failed to get current height")
	}

	return uint64(info.Response.LastBlockHeight), nil
}

func (o *Observer) fetchSubmitDepositEvents(ctx context.Context, height int64) ([]*db.Deposit, error) {
	var blockResult *coretypes.ResultBlockResults

	getBlockResult := func() error {
		var err error
		blockResult, err = o.client.BlockResults(ctx, &height)
		if err != nil {
			return errors.Wrap(err, "failed to get block results")
		}

		return nil
	}

	if err := core.DoWithRetry(ctx, getBlockResult); err != nil {
		return nil, errors.Wrap(err, "failed to get block results")
	}

	deposits, err := o.parseDepositsFromTxResults(blockResult.TxsResults)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse deposits from tx results")
	}

	return deposits, nil
}

func (o *Observer) isProcessed(deposit db.Deposit) (bool, error) {
	if deposit.WithdrawalTxHash != nil {
		return true, nil
	}

	return false, nil
}

func (o *Observer) broadcastDeposit(deposit db.Deposit) error {
	processed, err := o.isProcessed(deposit)
	if err != nil {
		return errors.Wrap(err, "failed to check if deposit is processed")
	}

	if processed || !o.clientsRepo.SupportsChain(deposit.WithdrawalChainId) {
		return skippedDeposit
	}

	err = o.broadcaster.Broadcast(deposit)
	if err != nil {
		return errors.Wrap(err, "failed to broadcast deposit")
	}

	return nil
}

func (o *Observer) catchupWithStatus(status internalTypes.WithdrawalStatus) error {
	deposits, err := o.depositsDb.GetWithStatus(status)
	if err != nil {
		return errors.Wrap(err, "failed to get unprocessed deposits")
	}

	for _, deposit := range deposits {
		err = o.broadcaster.CatchUp(deposit)
		if err != nil {
			o.logger.Errorf("failed to broadcast deposit to catchup deposit: %v", err)
			continue
		}
	}

	return nil
}

package observer

import (
	"context"
	"time"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client/http"
	"github.com/tendermint/tendermint/rpc/core/types"
	"gitlab.com/distributed_lab/logan/v3"
)

type Observer struct {
	client          *http.HTTP
	retries         int64
	retryTimeout    time.Duration
	pollingInterval time.Duration
	logger          *logan.Entry
	clientsRepo     chain.Repository

	depositsDb db.DepositsQ
	blockDb    db.BlocksQ

	broadcaster *broadcaster.Broadcaster
}

func New(client *http.HTTP, retries int64, retryTimeout, pollingInterval time.Duration, blocksDb db.BlocksQ,
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

func (o *Observer) Run(ctx context.Context, startHeight int64, catchup bool) error {
	// Firstly catch up pending deposits from db
	if catchup {
		deposits, err := o.depositsDb.GetWithStatus(types.WithdrawalStatus_WITHDRAWAL_STATUS_PENDING)
		if err != nil {
			return errors.Wrap(err, "failed to get unprocessed deposits")
		}

		for _, deposit := range deposits {
			if err := o.broadcaster.Broadcast(ctx, deposit); err != nil {
				o.logger.Errorf("failed to broadcast deposit: %v", err)
				continue
			}
		}
	}

	// Fetch deposits from Bridgeless core
	if err := o.fetchDeposits(ctx, startHeight); err != nil {
		return errors.Wrap(err, "fetch deposits")
	}

	return nil
}

func (o *Observer) fetchDeposits(ctx context.Context, startHeight int64) error {
	ticker := time.NewTicker(o.pollingInterval)
	defer ticker.Stop()

	if err := o.blockDb.Insert(db.LatestBlock{BlockId: startHeight}); err != nil {
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

			if startHeight <= currentHeight {
				if err = o.blockDb.UpdateLatestBlockId(db.LatestBlock{BlockId: startHeight}); err != nil {
					o.logger.WithError(err).Error("failed to update latest block height")
					continue
				}
				deposits, err := o.fetchSubmitDepositEvents(ctx, startHeight)
				if err != nil {

					o.logger.WithError(err).Error("failed to fetch submit deposit events")
					continue
				}

				for _, deposit := range deposits {
					err = o.broadcastDeposit(ctx, *deposit)
					if err != nil {
						if errors.Is(err, skippedDeposit) {
							continue
						}

						o.logger.Warnf("failed to broadcast deposit: %v", err)
						continue
					}
				}
				startHeight++
				continue
			}

			o.logger.Debug("Waiting for next block, currentHeight:", currentHeight)
		}
	}
}

func (o *Observer) getCurrentHeight(ctx context.Context) (int64, error) {
	var info *coretypes.ResultABCIInfo

	getCurrentHeight := func() error {
		var err error
		info, err = o.client.ABCIInfo(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get ABCI info")
		}
		return nil
	}

	if err := o.doWithRetry(ctx, getCurrentHeight); err != nil {
		return 0, errors.Wrap(err, "failed to get current height")
	}

	return info.Response.LastBlockHeight, nil
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

	if err := o.doWithRetry(ctx, getBlockResult); err != nil {
		return nil, errors.Wrap(err, "failed to get block results")
	}

	deposits, err := parseDepositsFromTxResults(blockResult.TxsResults)
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

func (o *Observer) broadcastDeposit(ctx context.Context, deposit db.Deposit) error {
	processed, err := o.isProcessed(deposit)
	if err != nil {
		return errors.Wrap(err, "failed to check if deposit is processed")
	}
	if processed || !o.clientsRepo.SupportsChain(deposit.WithdrawalChainId) {
		return skippedDeposit
	}

	err = o.broadcaster.Broadcast(ctx, deposit)
	if err != nil {
		return errors.Wrap(err, "failed to broadcast deposit")
	}

	return nil
}

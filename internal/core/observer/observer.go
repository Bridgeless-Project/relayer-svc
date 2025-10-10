package observer

import (
	"context"
	"time"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client/http"
	"github.com/tendermint/tendermint/rpc/core/types"
	"gitlab.com/distributed_lab/logan/v3"
)

type Observer struct {
	client      *http.HTTP
	retries     int64
	timeout     time.Duration
	logger      *logan.Entry
	clientsRepo chain.Repository

	depositsDb db.DepositsQ
	blockDb    db.BlocksQ

	depositChannel chan db.Deposit
}

func New(client *http.HTTP, retries int64, timeout time.Duration, blocksDb db.BlocksQ,
	depositsDb db.DepositsQ, logger *logan.Entry, depositChan chan db.Deposit, clientsRepo chain.Repository) *Observer {

	return &Observer{
		client:         client,
		retries:        retries,
		timeout:        timeout,
		blockDb:        blocksDb,
		depositsDb:     depositsDb,
		depositChannel: depositChan,
		clientsRepo:    clientsRepo,
		logger:         logger,
	}
}

func (o *Observer) Run(ctx context.Context, startHeight int64) error {
	select {
	case <-ctx.Done():
		o.logger.Debug("context canceled. Stopping observer")
		return nil

	default:
		if err := o.fetchDeposits(ctx, startHeight); err != nil {
			return errors.Wrap(err, "fetch deposits")
		}

	}

	return nil
}

func (o *Observer) fetchDeposits(ctx context.Context, startHeight int64) error {
	for {
		select {
		case <-ctx.Done():
			o.logger.Debug("fetching deposits stopped")
			return nil

		default:
			currentHeight, err := o.getCurrentHeight(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to get current height")
			}

			if startHeight < currentHeight {
				startHeight++
				deposits, err := o.fetchSubmitDepositEvents(ctx, startHeight)
				if err != nil {
					return errors.Wrap(err, "failed to fetch deposit events")
				}

				for _, deposit := range deposits {
					processed, err := o.IsProcessed(ctx, *deposit)
					if err != nil {
						return errors.Wrap(err, "failed to check if deposit is processed")
					}
					if processed {
						continue
					}

					deposit.WithdrawalStatus = types.WithdrawalStatus_WITHDRAWAL_STATUS_PENDING
					_, err = o.depositsDb.Insert(*deposit)
					if err != nil {
						return errors.Wrap(err, "failed to insert deposit")
					}
				}

				if err = o.blockDb.UpdateLatestBlockId(db.LatestBlock{BlockId: currentHeight}); err != nil {
					return errors.Wrap(err, "failed to update latest block")
				}

				continue
			}

			o.logger.Debugf("At tip (height=%d), waiting for next block...", currentHeight)
			time.Sleep(o.timeout)
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

func (o *Observer) IsProcessed(ctx context.Context, deposit db.Deposit) (bool, error) {
	if deposit.WithdrawalTxHash == nil &&
		deposit.WithdrawalStatus == types.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED {

		client, err := o.clientsRepo.Client(deposit.WithdrawalChainId)
		if err != nil {
			return false, errors.Wrap(err, "failed to get withdrawal chain client")
		}
		processed, err := client.IsProcessed(ctx, deposit)

		return processed, errors.Wrap(err, "failed to check if deposit is processed")
	}

	return true, nil
}

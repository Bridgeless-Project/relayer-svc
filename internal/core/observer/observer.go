package observer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"gitlab.com/distributed_lab/logan/v3"
)

type Observer struct {
	client          *http.HTTP
	pollingInterval time.Duration
	logger          *logan.Entry
	clientsRepo     chain.Repository

	depositsDb db.DepositsQ
	blockDb    db.BlocksQ

	broadcaster   *broadcaster.Broadcaster
	blockDelay    time.Duration
	blockDistance uint64
}

func New(client *http.HTTP, blocksDb db.BlocksQ, depositsDb db.DepositsQ, brcst *broadcaster.Broadcaster, logger *logan.Entry) *Observer {

	return &Observer{
		client:      client,
		blockDb:     blocksDb,
		depositsDb:  depositsDb,
		broadcaster: brcst,
		logger:      logger,
	}
}

func (o *Observer) WithClientsRepo(clientsRepo chain.Repository) *Observer {
	o.clientsRepo = clientsRepo
	return o
}

func (o *Observer) WithPollingInterval(pollingInterval time.Duration) *Observer {
	o.pollingInterval = pollingInterval
	return o
}

func (o *Observer) WithBlockDelay(delay time.Duration) *Observer {
	o.blockDelay = delay
	return o
}

func (o *Observer) WithBlockDistance(distance uint64) *Observer {
	o.blockDistance = distance
	return o
}

func (o *Observer) Run(ctx context.Context, startHeight uint64) error {
	// Fetch deposits from Bridgeless core
	if err := o.fetchEvents(ctx, startHeight); err != nil {
		return errors.Wrap(err, "failed to fetch events from the core")
	}

	return nil
}

func (o *Observer) fetchEvents(ctx context.Context, startHeight uint64) error {
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
			o.logger.Debug("fetching events stopped")
			return nil

		case <-ticker.C:
			currentHeight, err := o.getCurrentHeight(ctx)
			if err != nil {
				o.logger.WithError(err).Error("failed to get current height")
				continue
			}

			o.waitForBlockDistance(ctx, startHeight)

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

			events, err := o.fetchBlockEvents(ctx, int64(startHeight))
			if err != nil {
				o.logger.WithError(err).
					WithField("blockNumber", startHeight).
					Error("failed to fetch submit deposit events")
				startHeight++
				continue
			}

			if len(events.Deposits) != 0 {
				time.Sleep(o.blockDelay)
			}

			for _, deposit := range events.Deposits {
				if err = o.broadcastDeposit(*deposit); err != nil {
					if errors.Is(err, skippedDeposit) {
						continue
					}
					o.logger.
						WithField("blockNumber", startHeight).
						Warnf("failed to broadcast deposit: %v", err)
					continue
				}
			}

			o.logger.Infof("Fetched %d epochs. Epoch chain ids:", len(events.Epochs))
			for _, epoch := range events.Epochs {
				o.logger.Infof("Epoch %d %s: %s %s", epoch.Id, epoch.Nonce, epoch.ChainId, epoch.Signer)
			}

			startHeight++
		}
	}
}

func (o *Observer) fetchBlockEvents(ctx context.Context, height int64)  (*BlockEvents, error) {
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

	epochs, err := o.parseEpochsFromTxResults(blockResult.TxsResults)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse epochs from tx results")
	}

	return &BlockEvents{
		Deposits: deposits,
		Epochs:   epochs,
	}, nil
}

func (o *Observer) parseEpochsFromTxResults(txs []*abciTypes.ResponseDeliverTx) ([]*db.Epoch, error) {
	var epochs []*db.Epoch

	for _, tx := range txs {
		var msgs []MsgEvent

		if tx.Log == "" || !json.Valid([]byte(tx.Log)) {
			o.logger.Warnf("skipping invalid tx log: %s", tx.Log)
			continue
		}

		o.logger.Debug("got log: " + tx.Log)
		if err := json.Unmarshal([]byte(tx.Log), &msgs); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to unmarshal log: %v", tx.Log))
		}
		for _, msg := range msgs {
			for _, event := range msg.Events {
				if event.Type != eventEpochUpdated {
					continue
				}

				epoch, err := parseUpdatedEpochs(event.Attributes)
				if err != nil {
					return nil, errors.Wrap(err, "failed to parse epoch")
				}

				epochs = append(epochs, epoch)
			}
		}
	}

	return epochs, nil
}

func (o *Observer) waitForBlockDistance(ctx context.Context, startHeight uint64) {
	for {
		currentHeight, err := o.getCurrentHeight(ctx)
		if err != nil {
			o.logger.WithError(err).Error("failed to get current height")
			break
		}

		gap := currentHeight - startHeight
		if gap >= o.blockDistance {
			break
		}

		time.Sleep(o.blockDelay)
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

func (o *Observer) parseDepositsFromTxResults(txs []*abciTypes.ResponseDeliverTx) ([]*db.Deposit, error) {
	var deposits []*db.Deposit

	for _, tx := range txs {
		var msgs []MsgEvent

		if tx.Log == "" || !json.Valid([]byte(tx.Log)) {
			o.logger.Warnf("skipping invalid tx log: %s", tx.Log)
			continue
		}

		o.logger.Debug("got log: " + tx.Log)
		if err := json.Unmarshal([]byte(tx.Log), &msgs); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to unmarshal log: %v", tx.Log))
		}
		for _, msg := range msgs {
			for _, event := range msg.Events {
				if event.Type != eventDepositSubmitted {
					continue
				}

				deposit, err := parseSubmittedDeposit(event.Attributes)
				if err != nil {
					return nil, errors.Wrap(err, "failed to parse deposit")
				}

				deposits = append(deposits, deposit)
			}
		}

	}

	return deposits, nil
}

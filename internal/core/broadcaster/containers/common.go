package containers

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/logan/v3"
)

func executeWithdrawal(ctx context.Context, chainClient chain.Client, deposit db.Deposit, tendermintClient *http.HTTP, logger *logan.Entry) (*db.Deposit, error) {
	var (
		txHash      string
		err         error
		blockHeight int64
	)

	switch deposit.WithdrawalToken {
	case core.DefaultNativeTokenAddress:
		txHash, blockHeight, err = chainClient.WithdrawNative(ctx, deposit)
	default:
		txHash, blockHeight, err = chainClient.WithdrawToken(ctx, deposit)
	}
	if err != nil && txHash == "" {
		return nil, errors.Wrap(err, "error processing withdrawal")
	}

	if err != nil {
		logger.WithError(err).Error("error processing withdrawal")
	}

	logger.Infof("Processed deposit %s withdrawal hash %s", deposit.String(), txHash)
	deposit.WithdrawalTxHash = &txHash
	deposit.WithdrawalChainBlock = blockHeight

	abci, err := tendermintClient.ABCIInfo(ctx)
	if err != nil {
		return &deposit, errors.Wrap(err, "error getting ABCI info")
	}

	deposit.WithdrawalCoreBlock = abci.Response.LastBlockHeight

	return &deposit, nil
}

package containers

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/logan/v3"
)

func executeWithdrawal(ctx context.Context, chainClient chain.Client, deposit *db.Deposit, tendermintClient *http.HTTP, logger *logan.Entry) error {
	txHash, blockHeight, err := chainClient.Withdraw(ctx, *deposit)

	deposit.WithdrawalTxHash = &txHash
	deposit.WithdrawalChainBlock = blockHeight

	abci, abciErr := tendermintClient.ABCIInfo(ctx)
	if abciErr != nil {
		return errors.Wrap(err, "error getting ABCI info")
	}

	deposit.WithdrawalCoreBlock = abci.Response.LastBlockHeight

	if err != nil {
		return errors.Wrap(err, "failed to execute withdrawal")
	}

	logger.Infof("Processed deposit %s withdrawal hash %s", deposit.String(), txHash)

	return nil
}

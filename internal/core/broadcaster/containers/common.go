package containers

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/logan/v3"
)

func executeWithdrawal(ctx context.Context, chainClient chain.ChildClient, deposit *db.Deposit, tendermintClient *http.HTTP, logger *logan.Entry) error {
	operator, txHash, blockHeight, err := chainClient.Withdraw(ctx, deposit)

	deposit.WithdrawalTxHash = &txHash
	deposit.WithdrawalChainBlock = blockHeight
	deposit.Operator = operator

	abci, abciErr := tendermintClient.ABCIInfo(ctx)
	if abciErr != nil {
		return errors.Wrap(err, "error getting ABCI info to get withdrawal core block")
	}

	deposit.WithdrawalCoreBlock = abci.Response.LastBlockHeight

	if err != nil {

		return errors.Wrap(err, "failed to execute withdrawal")
	}

	logger.Infof("Processed deposit %s withdrawal hash %s", deposit.String(), txHash)

	return nil
}

func executeUpdateSigners(ctx context.Context, chainClient chain.ChildClient, epoch *db.Epoch, _ *http.HTTP, logger *logan.Entry) error {
	logger.Infof("Executing update signers")
	logger.Infof("Epoch id: %d", epoch.Id)
	logger.Infof("Epoch ChainId: %s", epoch.ChainId)
	logger.Infof("Epoch Nonce: %s", epoch.Nonce)
	logger.Infof("Epoch Signer: %s", epoch.Signer)
	logger.Infof("Epoch Signature: %s", epoch.Signature)
	logger.Infof("Epoch StartTime: %d", epoch.StartTime)
	logger.Infof("Epoch EndTime: %d", epoch.EndTime)
	logger.Infof("Epoch SignatureMode: %v", epoch.SignatureMode)
	chainClient.UpdateSigners(ctx, epoch)
	return nil
}

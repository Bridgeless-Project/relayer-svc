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
	logger.Debugf(
		"Update signers | ID: %d, Chain: %s, Nonce: %s, " +
		"Signer: %s, Sig: %s, Start: %d, End: %d, Mode: %v", 
    epoch.Id,
		epoch.ChainId,
		epoch.Nonce,
		epoch.Signer,
		epoch.Signature,
		epoch.StartTime,
		epoch.EndTime,
		epoch.SignatureMode,
	)

	tx, block, err := chainClient.UpdateSigners(ctx, epoch)
	if err != nil {
		return errors.Wrap(err, "failed to execute update signers")	
	}

	logger.Debugf("Update signers tx: %s; block: %d", tx, block)
	return nil
}

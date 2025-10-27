package solana

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/solana/contract"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
)

func (c *Client) withdrawSPL(ctx context.Context, depositData db.Deposit) (string, int64, error) {
	withdrawalCtx, err := c.getWithdrawalContext(depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal context")
	}

	vault, err := c.getSPLVault(ctx, depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal SPL vault")
	}

	tokenInfo, err := c.chain.Rpc.GetAccountInfo(ctx, withdrawalCtx.Token)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get token info")
	}

	withdrawInstruction := contract.NewWithdrawSplInstruction(
		c.chain.Meta.BridgeId,
		withdrawalCtx.WithdrawalTxHash,
		withdrawalCtx.Amount,
		withdrawalCtx.UID,
		withdrawalCtx.Sig,
		withdrawalCtx.RecID,
		withdrawalCtx.Token,
		*vault,
		withdrawalCtx.Receiver,
		withdrawalCtx.Authority,
		withdrawalCtx.WithdrawalPDA,
		c.chain.OperatorWallet.PublicKey(),
		solana.SystemProgramID,
		tokenInfo.Value.Owner,
	)

	block, err := c.chain.Rpc.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get blockhash")
	}

	txHash, err := c.SendTx(ctx, withdrawInstruction.Build())
	if err != nil {

		if txHash != nil {
			return txHash.String(),
				int64(block.Value.LastValidBlockHeight),
				errors.Wrap(err, "failed to send withdrawal tx")
		}

		return "",
			int64(block.Value.LastValidBlockHeight),
			errors.Wrap(err, "unable to send withdrawal token instruction")
	}

	return txHash.String(), int64(block.Value.LastValidBlockHeight), nil
}

package solana

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/solana/contract"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go"
	"github.com/pkg/errors"
)

func (c *Client) withdrawWrapped(ctx context.Context, depositData db.Deposit) (string, int64, error) {
	withdrawalCtx, err := c.getWithdrawalContext(depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal context")
	}

	tokenData, err := c.getTokenMetadata(withdrawalCtx.Token)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal token metadata")
	}

	withdrawInstruction := contract.NewWithdrawWrappedInstruction(
		c.chain.Meta.BridgeId,
		withdrawalCtx.WithdrawalTxHash,
		tokenData.Nonce,
		tokenData.Symbol,
		withdrawalCtx.Amount,
		withdrawalCtx.UID,
		withdrawalCtx.Sig,
		withdrawalCtx.RecID,
		withdrawalCtx.Token,
		withdrawalCtx.Receiver,
		withdrawalCtx.Authority,
		withdrawalCtx.WithdrawalPDA,
		c.chain.OperatorWallet.PublicKey(),
		solana.SystemProgramID,
		solana.Token2022ProgramID,
	)

	blockNumber, err := c.getLatestBlockWithRetry(ctx)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get latest block number")
	}

	txHash, err := c.SendTx(ctx, withdrawInstruction.Build())
	if err != nil {
		if txHash != nil {
			return txHash.String(),
				blockNumber,
				errors.Wrapf(err, "unable to send withdrawal wrapped instruction")
		}

		return "",
			blockNumber,
			errors.Wrap(err, "unable to send withdrawal wrapped instruction")
	}

	return txHash.String(), blockNumber, nil
}

package solana

import (
	"context"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/solana/contract"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go"
	"github.com/pkg/errors"
)

func (c *Client) WithdrawalAmountValid(amount *big.Int) bool {
	// Solana token amounts are uint64, bigger (or negative) numbers are invalid
	if !amount.IsUint64() {
		return false
	}
	return amount.Cmp(core.ZeroAmount) == 1
}

func (c *Client) WithdrawNative(ctx context.Context, depositData db.Deposit) (string, error) {
	withdrawalCtx, err := c.getWithdrawalContext(depositData)
	if err != nil {
		return "", errors.Wrap(err, "failed to get withdrawal context")
	}

	withdrawInstruction := contract.NewWithdrawNativeInstruction(c.chain.Meta.BridgeId, withdrawalCtx.WithdrawalTxHash,
		withdrawalCtx.Amount, withdrawalCtx.UID, withdrawalCtx.Sig, withdrawalCtx.RecID, withdrawalCtx.Receiver,
		withdrawalCtx.Authority, withdrawalCtx.WithdrawalPDA, c.chain.OperatorWallet.PublicKey(), solana.SystemProgramID)

	hash, err := c.SendTx(ctx, withdrawInstruction.Build())
	if err != nil {
		return "", errors.Wrap(err, "unable to send withdrawal")
	}
	return hash.String(), nil
}

func (c *Client) WithdrawToken(ctx context.Context, depositData db.Deposit) (string, error) {
	if !depositData.IsWrappedToken {
		txHash, err := c.withdrawSPL(ctx, depositData)
		if err != nil {
			return "", errors.Wrap(err, "failed to withdraw SPL token")
		}
		return txHash, nil
	}

	hash, err := c.withdrawWrapped(ctx, depositData)
	if err != nil {
		return "", errors.Wrap(err, "failed to withdraw wrapped token")
	}
	return hash, nil
}

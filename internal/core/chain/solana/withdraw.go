package solana

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/solana/contract"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
)

func (c *Client) WithdrawNative(ctx context.Context, depositData db.Deposit) (string, int64, error) {
	withdrawalCtx, err := c.getWithdrawalContext(depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal context")
	}

	withdrawInstruction := contract.NewWithdrawNativeInstruction(
		c.chain.Meta.BridgeId,
		withdrawalCtx.WithdrawalTxHash,
		withdrawalCtx.Amount,
		withdrawalCtx.UID,
		withdrawalCtx.Sig,
		withdrawalCtx.RecID,
		withdrawalCtx.Receiver,
		withdrawalCtx.Authority,
		withdrawalCtx.WithdrawalPDA,
		c.chain.OperatorWallet.PublicKey(),
		solana.SystemProgramID)

	block, err := c.chain.Rpc.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get blockhash")
	}

	hash, err := c.SendTx(ctx, withdrawInstruction.Build())
	if hash == nil {
		return "",
			int64(block.Value.LastValidBlockHeight),
			errors.Wrap(err, "failed to send withdrawal tx")
	}

	return hash.String(),
		int64(block.Value.LastValidBlockHeight),
		errors.Wrap(err, "unable to send native withdrawal tx")
}

func (c *Client) WithdrawToken(ctx context.Context, depositData db.Deposit) (string, int64, error) {
	receiverPub, err := solana.PublicKeyFromBase58(depositData.Receiver)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to decode receiver public key")
	}
	_, err = c.chain.Rpc.GetAccountInfo(ctx, receiverPub)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get receiver account info")
	}

	if !depositData.IsWrappedToken {
		txHash, blockHeight, err := c.withdrawSPL(ctx, depositData)
		return txHash, blockHeight, errors.Wrap(err, "failed to withdraw SPL token")
	}

	txHash, blockHeight, err := c.withdrawWrapped(ctx, depositData)
	return txHash, blockHeight, errors.Wrap(err, "failed to withdraw wrapped token")

}

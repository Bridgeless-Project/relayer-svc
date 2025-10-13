package solana

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"strconv"

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

func (c *Client) GetSignHash(data db.Deposit) ([]byte, error) {
	amount, err := strconv.ParseUint(data.WithdrawalAmount, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse withdrawal amount")
	}
	amountBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountBytes, amount)

	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, uint64(data.TxNonce))
	// unique id derived from deposit info
	uid := sha256.Sum256(append([]byte(data.TxHash), nonceBytes...))

	receiver, err := solana.PublicKeyFromBase58(data.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse receiver address")
	}

	buffer := []byte("withdraw")
	buffer = append(buffer, []byte(c.chain.Meta.BridgeId)...)
	buffer = append(buffer, amountBytes...)
	buffer = append(buffer, uid[:]...)
	buffer = append(buffer, receiver.Bytes()...)

	if data.WithdrawalToken != core.DefaultNativeTokenAddress {
		token, err := solana.PublicKeyFromBase58(data.WithdrawalToken)
		if err != nil {
			return nil, err
		}
		buffer = append(buffer, token.Bytes()...)
	}

	hash := sha256.Sum256(buffer)
	return hash[:], nil
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
	withdrawalCtx, err := c.getWithdrawalContext(depositData)
	if err != nil {
		return "", errors.Wrap(err, "failed to get withdrawal context")
	}

	ata, err := c.getWithdrawalATA(ctx, depositData)
	if err != nil {
		return "", errors.Wrap(err, "failed to get withdrawal ATA")
	}

	vault, err := c.getSPLVault(ctx, depositData)
	if err != nil {
		return "", errors.Wrap(err, "failed to get withdrawal SPL vault")
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
		*ata,
		withdrawalCtx.Authority,
		withdrawalCtx.WithdrawalPDA,
		c.chain.OperatorWallet.PublicKey(),
		solana.SystemProgramID,
		solana.Token2022ProgramID,
	)

	txHash, err := c.SendTx(ctx, withdrawInstruction.Build())
	if err != nil {
		return "", errors.Wrap(err, "unable to send withdrawal token instruction")
	}
	return txHash.String(), nil
}

func (c *Client) WithdrawWrapped(ctx context.Context, depositData db.Deposit) (string, error) {
	withdrawalCtx, err := c.getWithdrawalContext(depositData)
	if err != nil {
		return "", errors.Wrap(err, "failed to get withdrawal context")
	}

	tokenData, err := c.getTokenMetadata(withdrawalCtx.Token)
	if err != nil {
		return "", errors.Wrap(err, "failed to get withdrawal token metadata")
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
	txHash, err := c.SendTx(ctx, withdrawInstruction.Build())
	if err != nil {
		return "", errors.Wrap(err, "unable to send withdrawal wrapped instruction")
	}

	return txHash.String(), nil
}

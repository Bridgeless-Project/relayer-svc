package evm

import (
	"context"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
)

func (c *Client) WithdrawNative(ctx context.Context, depositData db.Deposit) (txHash string, err error) {
	transactOpts, err := c.prepareTxOpts(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to prepare transact opts")
	}

	amount, ok := new(big.Int).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return "", errors.New("failed to parse withdrawal amount")
	}

	receiverAdress := common.HexToAddress(depositData.Receiver)

	hashBytes, err := hexutil.Decode(depositData.TxHash)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode hash")
	}

	hashBytes32 := to32Bytes(hashBytes)

	signatureBytes, err := hexutil.Decode(depositData.Signature)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode signature")
	}

	tx, err := c.contractClient.WithdrawNative(transactOpts, amount, receiverAdress, hashBytes32,
		big.NewInt(depositData.TxNonce), [][]byte{signatureBytes})
	if err != nil {
		return "", errors.Wrap(err, "failed to withdraw native")
	}

	return tx.Hash().Hex(), nil
}

func (c *Client) WithdrawToken(ctx context.Context, depositData db.Deposit) (txHash string, err error) {
	transactOpts, err := c.prepareTxOpts(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to prepare transact opts")
	}

	amount, ok := new(big.Int).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return "", errors.New("failed to parse withdrawal amount")
	}

	receiverAdress := common.HexToAddress(depositData.Receiver)
	hashBytes, err := hexutil.Decode(depositData.TxHash)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode hash")
	}

	hashBytes32 := to32Bytes(hashBytes)
	signatureBytes, err := hexutil.Decode(depositData.Signature)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode signature")
	}

	tokenAddr := common.HexToAddress(depositData.WithdrawalToken)

	tx, err := c.contractClient.WithdrawERC20(transactOpts, tokenAddr, amount, receiverAdress, hashBytes32,
		big.NewInt(depositData.TxNonce), depositData.IsWrappedToken, [][]byte{signatureBytes})
	if err != nil {
		return "", errors.Wrap(err, "failed to withdraw token")
	}

	return tx.Hash().Hex(), nil
}

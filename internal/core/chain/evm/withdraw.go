package evm

import (
	"context"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
)

func (c *Client) WithdrawNative(ctx context.Context, depositData db.Deposit) (string, int64, error) {
	data, err := c.getWithdrawalTxData(withdrawNative, depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal tx data")
	}

	transactOpts, err := c.prepareTxOpts(ctx, data)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to prepare transact opts")
	}

	amount, ok := new(big.Int).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return "", 0, errors.New("failed to parse withdrawal amount")
	}

	receiverAdress := common.HexToAddress(depositData.Receiver)

	signatureBytes, err := hexutil.Decode(depositData.Signature)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to decode signature")
	}

	hash, err := c.getWithdrawalTxHash(transactOpts, data)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal tx hash")
	}

	block, err := c.getBlockWithRetry(ctx)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get block")
	}

	tx, err := c.contractClient.WithdrawNative(
		transactOpts,
		amount,
		receiverAdress,
		txHashToBytes32(depositData.TxHash),
		big.NewInt(depositData.TxNonce),
		[][]byte{signatureBytes})
	if err != nil {
		return hash, block, errors.Wrap(err, "failed to withdraw native")
	}

	return tx.Hash().Hex(), block, nil
}

func (c *Client) WithdrawToken(ctx context.Context, depositData db.Deposit) (string, int64, error) {
	data, err := c.getWithdrawalTxData(withdrawERC20, depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal tx data")
	}

	transactOpts, err := c.prepareTxOpts(ctx, data)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to prepare transact opts")
	}

	amount, ok := new(big.Int).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return "", 0, errors.New("failed to parse withdrawal amount")
	}

	receiverAdress := common.HexToAddress(depositData.Receiver)
	signatureBytes, err := hexutil.Decode(depositData.Signature)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to decode signature")
	}

	tokenAddr := common.HexToAddress(depositData.WithdrawalToken)

	hash, err := c.getWithdrawalTxHash(transactOpts, data)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal tx hash")
	}

	block, err := c.getBlockWithRetry(ctx)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get block")
	}

	tx, err := c.contractClient.WithdrawERC20(
		transactOpts,
		tokenAddr,
		amount,
		receiverAdress,
		txHashToBytes32(depositData.TxHash),
		big.NewInt(depositData.TxNonce),
		depositData.IsWrappedToken,
		[][]byte{signatureBytes})
	if err != nil {
		return hash, block, errors.Wrap(err, "failed to withdraw token")
	}

	return tx.Hash().Hex(), block, nil
}

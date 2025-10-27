package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
)

func (c *Client) WithdrawNative(ctx context.Context, depositData db.Deposit) (string, int64, error) {
	transactOpts, err := c.prepareTxOpts(ctx)
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

	hash, err := c.getWithdrawalTxHash(withdrawNative, transactOpts, depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal tx hash")
	}

	fmt.Println("PREDICTED HASH: ", hash)
	block, err := c.chain.Rpc.BlockByNumber(ctx, nil)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get block by number")
	}

	tx, err := c.contractClient.WithdrawNative(
		transactOpts,
		amount,
		receiverAdress,
		txHashToBytes32(depositData.TxHash),
		big.NewInt(depositData.TxNonce),
		[][]byte{signatureBytes})
	if err != nil {
		return "", block.Number().Int64(), errors.Wrap(err, "failed to withdraw native")
	}

	return tx.Hash().Hex(), block.Number().Int64(), nil
}

func (c *Client) WithdrawToken(ctx context.Context, depositData db.Deposit) (string, int64, error) {
	transactOpts, err := c.prepareTxOpts(ctx)
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

	hash, err := c.getWithdrawalTxHash(withdrawERC20, transactOpts, depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal tx hash")
	}

	block, err := c.chain.Rpc.BlockByNumber(ctx, nil)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get block by number")
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
		return hash, block.Number().Int64(), errors.Wrap(err, "failed to withdraw token")
	}

	fmt.Println("GOT HASH: ", tx.Hash().Hex())
	fmt.Println("PREDICTED: ", hash)

	return tx.Hash().Hex(), block.Number().Int64(), nil
}

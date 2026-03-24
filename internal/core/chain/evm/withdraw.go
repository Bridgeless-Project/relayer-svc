package evm

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
)

func (c *Client) Withdraw(ctx context.Context, depositData *db.Deposit, signer *signerInfo) (string, string, int64, error) {
	fn := c.getWithdrawFunc(depositData)
	txHash, block, err := fn(ctx, depositData, signer)
	if err != nil {
		return signer.address.String(), txHash, block, err
	}

	return signer.address.String(), txHash, block, nil
}

func (c *Client) withdrawNative(ctx context.Context, depositData *db.Deposit, signer *signerInfo) (string, int64, error) {
	data, err := c.getWithdrawalTxData(withdrawNative, depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal tx data")
	}

	receiverAddress := common.HexToAddress(depositData.Receiver)

	transactOpts, err := c.prepareTxOpts(ctx, data, signer)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to prepare transact opts")
	}

	amount, ok := new(big.Int).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return "", 0, errors.New("failed to parse withdrawal amount")
	}

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
		receiverAddress,
		txHashToBytes32(depositData.TxHash),
		big.NewInt(depositData.TxNonce),
		[][]byte{signatureBytes})
	if err != nil {
		return hash, block, errors.Wrap(err, "failed to withdraw native")
	}

	finalizer := func() error {
		if err := c.finalize(ctx, tx.Hash()); err != nil {
			if errors.Is(err, chain.ErrSkippedFinalization) {
				return nil
			}

			return errors.Wrap(err, "failed to finalize")
		}

		return nil
	}

	err = core.DoWithRetry(ctx, finalizer)
	if err != nil {
		return hash, block, errors.Wrap(err, "failed to perform finalization")
	}

	return tx.Hash().Hex(), block, nil
}

func (c *Client) withdrawToken(ctx context.Context, depositData *db.Deposit, signer *signerInfo) (string, int64, error) {
	data, err := c.getWithdrawalTxData(withdrawERC20, depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal tx data")
	}

	receiverAddress := common.HexToAddress(depositData.Receiver)
	tokenAddr := common.HexToAddress(depositData.WithdrawalToken)

	transactOpts, err := c.prepareTxOpts(ctx, data, signer)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to prepare transact opts")
	}

	amount, ok := new(big.Int).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return "", 0, errors.New("failed to parse withdrawal amount")
	}

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

	tx, err := c.contractClient.WithdrawERC20(
		transactOpts,
		tokenAddr,
		amount,
		receiverAddress,
		txHashToBytes32(depositData.TxHash),
		big.NewInt(depositData.TxNonce),
		depositData.IsWrappedToken,
		[][]byte{signatureBytes})
	if err != nil {
		return hash, block, errors.Wrap(err, "failed to withdraw token")
	}

	finalizer := func() error {
		if err := c.finalize(ctx, tx.Hash()); err != nil {
			if errors.Is(err, chain.ErrSkippedFinalization) {
				return nil
			}

			return errors.Wrap(err, "failed to finalize")
		}

		return nil
	}

	err = core.DoWithRetry(ctx, finalizer)
	if err != nil {

		return hash, block, errors.Wrap(err, "failed to perform finalization")
	}

	return tx.Hash().Hex(), block, nil
}

func (c *Client) withdrawERC20Merkelized(ctx context.Context, depositData *db.Deposit, signer *signerInfo) (string, int64, error) {
	data, err := c.getWithdrawalTxData(withdrawERC20Merkelized, depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal tx data")
	}

	receiverAddress := common.HexToAddress(depositData.Receiver)
	tokenAddr := common.HexToAddress(depositData.WithdrawalToken)

	transactOpts, err := c.prepareTxOpts(ctx, data, signer)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to prepare transact opts")
	}

	amount, ok := new(big.Int).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return "", 0, errors.New("failed to parse withdrawal amount")
	}

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

	proof, err := merkleProofParsing(depositData.MerkleProof)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to parse MerkleProof")
	}

	tx, err := c.contractClient.WithdrawERC20Merkelized(
		transactOpts,
		tokenAddr,
		amount,
		receiverAddress,
		txHashToBytes32(depositData.TxHash),
		big.NewInt(depositData.TxNonce),
		depositData.IsWrappedToken,
		proof,
		[][]byte{signatureBytes})
	if err != nil {
		return hash, block, errors.Wrap(err, "failed to withdraw ERC20Merklized")
	}

	finalizer := func() error {
		if err := c.finalize(ctx, tx.Hash()); err != nil {
			if errors.Is(err, chain.ErrSkippedFinalization) {
				return nil
			}

			return errors.Wrap(err, "failed to finalize")
		}

		return nil
	}

	err = core.DoWithRetry(ctx, finalizer)
	if err != nil {

		return hash, block, errors.Wrap(err, "failed to perform finalization")
	}

	return tx.Hash().Hex(), block, nil

}

func (c *Client) withdrawNativeMerkelized(ctx context.Context, depositData *db.Deposit, signer *signerInfo) (string, int64, error) {
	data, err := c.getWithdrawalTxData(withdrawNativeMerkelized, depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get withdrawal tx data")
	}

	receiverAddress := common.HexToAddress(depositData.Receiver)

	transactOpts, err := c.prepareTxOpts(ctx, data, signer)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to prepare transact opts")
	}

	amount, ok := new(big.Int).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return "", 0, errors.New("failed to parse withdrawal amount")
	}

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
	proof, err := merkleProofParsing(depositData.MerkleProof)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to parse MerkleProof")
	}

	tx, err := c.contractClient.WithdrawNativeMerkelized(
		transactOpts,
		amount,
		receiverAddress,
		txHashToBytes32(depositData.TxHash),
		big.NewInt(depositData.TxNonce),
		proof,
		[][]byte{signatureBytes})
	if err != nil {
		return hash, block, errors.Wrap(err, "failed to withdraw merkelized native")
	}

	finalizer := func() error {
		if err := c.finalize(ctx, tx.Hash()); err != nil {
			if errors.Is(err, chain.ErrSkippedFinalization) {
				return nil
			}

			return errors.Wrap(err, "failed to finalize")
		}

		return nil
	}

	err = core.DoWithRetry(ctx, finalizer)
	if err != nil {
		return hash, block, errors.Wrap(err, "failed to perform finalization")
	}

	return tx.Hash().Hex(), block, nil
}
func (c *Client) getWithdrawFunc(depositData *db.Deposit) func(ctx context.Context, depositData *db.Deposit, signer *signerInfo) (string, int64, error) {
	isNative := depositData.WithdrawalToken == core.DefaultNativeTokenAddress
	isMerkelized := depositData.MerkleProof != ""
	switch {
	case isNative && isMerkelized:
		return c.withdrawNativeMerkelized

	case isNative && !isMerkelized:
		return c.withdrawNative

	case !isNative && isMerkelized:
		return c.withdrawERC20Merkelized
	default:
		return c.withdrawToken
	}
}

func (c *Client) finalize(ctx context.Context, txHash common.Hash) error {
	ctxt, cancel := context.WithDeadline(ctx, time.Now().Add(time.Duration(c.chain.WSTimeout)*time.Second))
	defer cancel()

	headerChan := make(chan *types.Header)
	var (
		sub ethereum.Subscription
		err error
	)

	subscribeToWs := func() error {
		sub, err = c.chain.WSRpc.SubscribeNewHead(ctxt, headerChan)
		if err != nil {
			return errors.Wrap(err, "failed to subscribe to headers")
		}

		return nil
	}
	if err = core.DoWithRetry(ctx, subscribeToWs); err != nil {
		return err
	}

	defer sub.Unsubscribe()
	for {
		select {
		case <-ctxt.Done():
			receipt, err := c.chain.Rpc.TransactionReceipt(ctx, txHash)
			if err != nil {
				if errors.Is(err, ethereum.NotFound) {
					return errors.New("timeout waiting for tx finalize")
				}

				return errors.Wrap(err, "failed to get transaction receipt")
			}

			if receipt.Status == types.ReceiptStatusSuccessful {
				return nil
			}

			return errors.New("transaction failed on network")

		case header, ok := <-headerChan:
			if !ok {
				return errors.New("receipt channel closed")
			}

			blockNumber := rpc.BlockNumber(header.Number.Int64())

			receipts, err := c.chain.Rpc.BlockReceipts(
				ctxt,
				rpc.BlockNumberOrHash{
					BlockNumber: &blockNumber,
				},
			)
			if err != nil {
				if strings.Contains(err.Error(), notAvailableBlockReceipts) {
					return errors.Wrap(chain.ErrSkippedFinalization, "failed to finalize")
				}

				return errors.Wrap(err, "failed to get block receipts")
			}

			for _, receipt := range receipts {
				if receipt.TxHash != txHash {
					continue
				}

				if receipt.Status == types.ReceiptStatusSuccessful {
					return nil
				}

				return errors.New("tx failed on network")
			}
		}

	}
}

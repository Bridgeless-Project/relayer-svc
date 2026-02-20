package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
)

const (
	receiptSuccess = 1
)

func (c *Client) UpdateSigners(ctx context.Context, epochData *db.Epoch, signer *signerInfo) (string, int64, error) {
	signatureBytes, err := hexutil.Decode(epochData.Signature)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to decode signature")
	}

	signer_ := common.HexToAddress(epochData.Signer)
	nonce, ok := new(big.Int).SetString(epochData.Nonce, 10)
	if !ok || nonce == nil {
		return "", 0, errors.New("failed to parse nonce")
	}

	data, err := c.abi.Pack(
		"updateSigner",
		signer_,
		new(big.Int).SetUint64(epochData.StartTime),
		new(big.Int).SetUint64(epochData.EndTime),
		nonce,
		epochData.SignatureMode,
		[][]byte{signatureBytes},
	)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to prepare transact opts")
	}

	transactOpts, err := c.prepareTxOpts(ctx, data, signer)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to prepare transact opts")
	}

	block, err := c.getBlockWithRetry(ctx)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get block")
	}

	tx, err := c.contractClient.UpdateSigner(
		transactOpts,
		signer_,
		new(big.Int).SetUint64(epochData.StartTime),
		new(big.Int).SetUint64(epochData.EndTime),
		nonce,
		epochData.SignatureMode,
		[][]byte{signatureBytes},
	)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to update signers")
	}

	receipt, err := bind.WaitMined(ctx, c.chain.Rpc, tx)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to wait for mining")
	}

	if receipt.Status != receiptSuccess {
    return "", 0, fmt.Errorf("transaction execution failed")
	}

	return tx.Hash().Hex(), block, nil
}

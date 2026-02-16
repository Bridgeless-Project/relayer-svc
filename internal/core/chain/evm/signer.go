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

func (c *Client) UpdateSigners(ctx context.Context, epochData *db.Epoch, signer *signerInfo) (string, int64, error) {
	signatureBytes, err := hexutil.Decode(epochData.Signature)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to decode signature")
	}

	signer_ := common.HexToAddress(epochData.Signer)
	nonce_, _ := new(big.Int).SetString(epochData.Nonce, 10)

	data, err := c.abi.Pack(
		"updateSigner",
		signer_,
		new(big.Int).SetUint64(epochData.StartTime),
		new(big.Int).SetUint64(epochData.EndTime),
		nonce_,
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
		nonce_,
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

	if receipt.Status != 1 {
    return "", 0, fmt.Errorf("transaction execution failed")
	}

	if receipt.GasUsed < 30000 {
		return "", 0, fmt.Errorf("silent failure: gas used (%d) too low for state change", receipt.GasUsed)
  }

	return tx.Hash().Hex(), block, nil
}

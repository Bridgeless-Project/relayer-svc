package evm

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/pkg/errors"
)

func (c *Client) prepareTxOpts(ctx context.Context) (*bind.TransactOpts, error) {
	gasPrice, err := c.chain.Rpc.SuggestGasPrice(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch gas price")
	}

	chainId, err := c.chain.Rpc.ChainID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch chain id")
	}

	tx, err := bind.NewKeyedTransactorWithChainID(c.chain.OperatorPrivKey, chainId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate transactor")
	}
	tx.Nonce = new(big.Int).SetUint64(c.nonce.Load())

	tx.GasLimit = 30000
	tx.GasPrice = gasPrice

	return tx, nil
}

func to32Bytes(data []byte) [32]byte {
	var arr [32]byte
	if len(data) > 32 {
		copy(arr[:], data[:32])
	} else {
		copy(arr[:], data)
	}
	return arr
}

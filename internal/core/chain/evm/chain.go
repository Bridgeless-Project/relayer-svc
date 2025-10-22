package evm

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure/v3"
)

type Chain struct {
	Id              string
	Rpc             *ethclient.Client
	BridgeAddress   common.Address
	OperatorPrivKey *ecdsa.PrivateKey
}

func FromChain(c chain.Chain) Chain {
	if c.Type != chain.TypeEVM {
		panic("chain is not EVM")
	}

	chain := Chain{
		Id: c.Id,
	}

	if err := figure.Out(&chain.Rpc).
		FromInterface(c.Rpc).
		With(figure.EthereumHooks).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain Ethereum clients"))
	}
	if err := figure.Out(&chain.BridgeAddress).
		FromInterface(c.BridgeAddresses).
		With(figure.EthereumHooks).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain bridge addresses"))
	}
	if err := figure.Out(&chain.OperatorPrivKey).
		FromInterface(c.OperatorPrivateKey).
		With(figure.EthereumHooks).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain operator private key"))
	}

	return chain
}

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

	tx.GasPrice = gasPrice

	return tx, nil
}

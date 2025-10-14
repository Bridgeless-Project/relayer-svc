package evm

import (
	"crypto/ecdsa"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
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

	if err := figure.Out(&chain.Rpc).FromInterface(c.Rpc).With(figure.EthereumHooks).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain Ethereum clients"))
	}
	if err := figure.Out(&chain.BridgeAddress).FromInterface(c.BridgeAddresses).With(figure.EthereumHooks).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain bridge addresses"))
	}
	if err := figure.Out(&chain.OperatorPrivKey).FromInterface(c.OperatorPrivateKey).With(figure.EthereumHooks).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain operator private key"))
	}

	return chain
}

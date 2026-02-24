package evm

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure/v3"
)

type Chain struct {
	Id                 string
	Rpc                *ethclient.Client
	BridgeAddress      common.Address
	OperatorsPrivKeys  []*ecdsa.PrivateKey
	Workers            int
	WSRpc              *ethclient.Client
	WSTimeout          int64
	GasPriceMultiplier int64
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

	if err := figure.Out(&chain.OperatorsPrivKeys).
		FromInterface(c.OperatorsPrivateKeys).
		With(EVMHooks).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain operators private keys"))
	}

	if err := figure.Out(&chain.WSTimeout).
		FromInterface(c.WSTimeout).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain ws timeout time"))
	}

	if err := figure.Out(&chain.Workers).FromInterface(c.Workers).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain workers number"))
	}

	if err := figure.Out(&chain.WSRpc).FromInterface(c.WSRpc).With(figure.EthereumHooks).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain ws rpc address"))
	}

	if err := figure.Out(&chain.GasPriceMultiplier).FromInterface(c.GasPriceMultiplier).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain gas multiplier"))
	}

	if chain.Workers > len(chain.OperatorsPrivKeys) {
		panic("number of workers is greater than number of operators private keys")
	}

	return chain
}

func (c *Client) prepareTxOpts(ctx context.Context, data []byte, signer *signerInfo) (*bind.TransactOpts, error) {
	gasPrice, err := c.chain.Rpc.SuggestGasPrice(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch gas price")
	}

	chainId, err := c.chain.Rpc.ChainID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch chain id")
	}

	tx, err := bind.NewKeyedTransactorWithChainID(signer.privateKey, chainId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate transactor")
	}
	nonce, err := c.getNextNonce(ctx, signer.address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch operator account nonce")
	}

	tx.Nonce = big.NewInt(0).SetUint64(nonce)

	mul := big.NewInt(c.chain.GasPriceMultiplier)
	denom := big.NewInt(100)

	txGasPrice := new(big.Int).Mul(gasPrice, mul)
	txGasPrice.Div(txGasPrice, denom)

	callMsg := ethereum.CallMsg{
		From:     signer.address,
		To:       &c.chain.BridgeAddress,
		GasPrice: new(big.Int).Set(txGasPrice),
		Data:     data,
	}

	gasLimit, err := c.chain.Rpc.EstimateGas(ctx, callMsg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to estimate gas limit")
	}

	tx.GasLimit = gasLimit
	return tx, nil
}

func (c *Client) getNextNonce(ctx context.Context, addr common.Address) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	currentNonce, exists := c.nonces[addr]
	if exists {
		c.nonces[addr] = currentNonce + 1
		return currentNonce, nil
	}

	// TODO: consider moving fetching outside of mutex lock
	fetchedNonce, err := c.chain.Rpc.PendingNonceAt(ctx, addr)
	if err != nil {
		return 0, err
	}

	c.nonces[addr] = fetchedNonce + 1
	return fetchedNonce, nil
}
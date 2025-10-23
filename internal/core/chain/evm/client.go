package evm

import (
	"context"
	"math/big"
	"strings"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/evm/contracts"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

const (
	EventDepositedNative = "DepositedNative"
	EventDepositedERC20  = "DepositedERC20"
)

var events = []string{
	EventDepositedNative,
	EventDepositedERC20,
}

type Client struct {
	chain          Chain
	contractClient *contracts.Bridge
	walletAddress  common.Address
}

// NewBridgeClient creates a new bridge Client for the given chain.
func NewBridgeClient(chain Chain) *Client {
	bridgeAbi, err := abi.JSON(strings.NewReader(contracts.BridgeMetaData.ABI))
	if err != nil {
		panic(errors.Wrap(err, "failed to parse bridge ABI"))
	}

	depositEvents := make([]abi.Event, len(events))
	for i, event := range events {
		depositEvent, ok := bridgeAbi.Events[event]
		if !ok {
			panic("wrong bridge ABI events")
		}
		depositEvents[i] = depositEvent
	}

	contractClient, err := contracts.NewBridge(chain.BridgeAddress, chain.Rpc)
	if err != nil {
		panic(errors.Wrap(err, "failed to init bridge client"))
	}

	walletAddress := crypto.PubkeyToAddress(chain.OperatorPrivKey.PublicKey)

	return &Client{
		chain:          chain,
		contractClient: contractClient,
		walletAddress:  walletAddress,
	}
}

func (p *Client) ChainId() string {
	return p.chain.Id
}

func (p *Client) Type() chain.Type {
	return chain.TypeEVM
}

func (p *Client) AddressValid(addr string) bool {
	return common.IsHexAddress(addr)
}

func (p *Client) TransactionHashValid(hash string) bool {
	return core.DefaultTransactionHashPattern.MatchString(hash)
}

func (c *Client) IsProcessed(ctx context.Context, depositData db.Deposit) (bool, error) {
	callOpts := &bind.CallOpts{
		Pending: false,
		Context: ctx,
	}

	used, err := c.contractClient.ContainsHash(callOpts, txHashToBytes32(depositData.TxHash), big.NewInt(depositData.TxNonce))
	if err != nil {
		return false, errors.Wrapf(err, "failed to call contract for used hash")
	}

	return used, nil
}

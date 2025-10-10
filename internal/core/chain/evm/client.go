package evm

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/evm/contracts"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum/accounts/abi"
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
	nonce          atomic.Uint64
}

func (p *Client) IsProcessed(ctx context.Context, depositData db.Deposit) (bool, error) {
	//TODO implement me
	panic("implement me")
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
	nonce, err := chain.Rpc.NonceAt(context.Background(), walletAddress, nil)
	if err != nil {
		panic(errors.Wrapf(err, "failed to get nonce for chain %s", chain.Id))
	}

	atomicNonce := atomic.Uint64{}
	atomicNonce.Store(nonce)

	return &Client{
		chain:          chain,
		contractClient: contractClient,
		walletAddress:  walletAddress,
		nonce:          atomicNonce,
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

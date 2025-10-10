package solana

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go"
)

type Client struct {
	chain Chain
}

func (p *Client) IsProcessed(ctx context.Context, depositData db.Deposit) (bool, error) {
	//TODO implement me
	panic("implement me")
}

// NewBridgeClient creates a new bridge Client for the given chain.
func NewBridgeClient(chain Chain) *Client {
	return &Client{
		chain: chain,
	}
}

func (p *Client) ChainId() string {
	return p.chain.Id
}

func (p *Client) Type() chain.Type {
	return chain.TypeSolana
}

func (p *Client) AddressValid(addr string) bool {
	_, err := solana.PublicKeyFromBase58(addr)
	return err == nil
}

func (p *Client) TransactionHashValid(hash string) bool {
	return core.SolanaTransactionHashPattern.MatchString(hash)
}

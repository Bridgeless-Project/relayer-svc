package solana

import (
	"github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster"
	chain "github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster/chains"
	"github.com/gagliardetto/solana-go"
)

type Client struct {
	chain Chain
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
	return broadcaster.SolanaTransactionHashPattern.MatchString(hash)
}

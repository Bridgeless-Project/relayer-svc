package solana

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
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

func (p *Client) TransactionHashValid(hash string) bool {
	return core.SolanaTransactionHashPattern.MatchString(hash)
}

func (c *Client) IsProcessed(ctx context.Context, depositData db.Deposit) (bool, error) {
	withdrawalHash, err := c.getWithdrawalHash(depositData)
	if err != nil {
		return false, errors.Wrap(err, "failed to get withdrawal hash")
	}

	pda, err := c.getWithdrawalPDA(withdrawalHash)
	if err != nil {
		return false, errors.Wrap(err, "failed to get withdrawal pda")
	}

	_, err = c.chain.Rpc.GetAccountInfo(ctx, *pda)
	if err != nil {
		if errors.Is(err, rpc.ErrNotFound) {
			return false, nil
		}

		return false, errors.Wrap(err, "failed to get account info")
	}

	return true, nil
}

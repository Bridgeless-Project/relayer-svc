package solana

import (
	"context"
	"math/rand"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go"
	"github.com/pkg/errors"
)

type ChildClient struct {
	signers []*solana.Wallet
	parent  *Client
}

func NewChildClient(parent *Client) *ChildClient {
	return &ChildClient{
		parent: parent,
	}
}

func (c *ChildClient) IsProcessed(ctx context.Context, depositData db.Deposit) (bool, error) {
	return c.parent.IsProcessed(ctx, depositData)
}

func (c *ChildClient) Withdraw(ctx context.Context, depositData *db.Deposit) (string, int64, error) {
	if len(c.signers) == 0 {
		return "", 0, errors.New("no signers available")
	}
	signer := c.signers[rand.Intn(len(c.signers))]

	return c.parent.Withdraw(ctx, depositData, signer)
}

func (c *ChildClient) AddSigner(signer *solana.Wallet) {
	c.signers = append(c.signers, signer)
}

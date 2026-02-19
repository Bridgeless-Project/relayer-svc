package ton

import (
	"context"
	"log"
	"math/rand"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type ChildClient struct {
	signers []*wallet.Wallet
	parent  *Client
}

func NewChildClient(parent *Client) *ChildClient {
	return &ChildClient{
		parent: parent,
	}
}

func (c ChildClient) IsProcessed(ctx context.Context, depositData db.Deposit) (bool, error) {
	return c.parent.IsProcessed(ctx, depositData)
}

func (c ChildClient) Withdraw(ctx context.Context, depositData *db.Deposit) (string, string, int64, error) {
	if len(c.signers) == 0 {
		return "", "", 0, errors.New("no signers available")
	}
	signer := c.signers[rand.Intn(len(c.signers))]

	return c.parent.Withdraw(ctx, depositData, signer)
}

func (c *ChildClient) AddSigner(signer *wallet.Wallet) {
	c.signers = append(c.signers, signer)
}

func (c ChildClient) UpdateSigners(ctx context.Context, epochData *db.Epoch) (string, int64, error) {
	log.Default().Printf("TON UPDATE SIGNERS: %d", epochData.Id)
	return "", 0, nil
}

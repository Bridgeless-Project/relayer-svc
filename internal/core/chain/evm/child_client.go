package evm

import (
	"context"
	"crypto/ecdsa"
	"math/rand"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

type ChildClient struct {
	signers []*signerInfo
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

func (c ChildClient) Withdraw(ctx context.Context, depositData *db.Deposit) (string, int64, error) {
	if len(c.signers) == 0 {
		return "", 0, errors.New("no signers available")
	}
	signer := c.signers[rand.Intn(len(c.signers))]

	return c.parent.Withdraw(ctx, depositData, signer)
}

func (c *ChildClient) AddSigner(key *ecdsa.PrivateKey) {
	c.signers = append(c.signers, &signerInfo{
		privateKey: key,
		address:    crypto.PubkeyToAddress(key.PublicKey),
	})
}

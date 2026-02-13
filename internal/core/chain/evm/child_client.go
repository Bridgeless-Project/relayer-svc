package evm

import (
	"context"
	"crypto/ecdsa"
	"log"
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

func (c ChildClient) Withdraw(ctx context.Context, depositData *db.Deposit) (string, string, int64, error) {
	if len(c.signers) == 0 {
		return "", "", 0, errors.New("no signers available")
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

func (c ChildClient) UpdateSigners(ctx context.Context, epochData *db.Epoch) {
	if len(c.signers) == 0 {
		return
	}
	signer := c.signers[rand.Intn(len(c.signers))]
	log.Default().Printf("EVM UPDATE SIGNERS: %d", epochData.Id)

	txHash, block, err := c.parent.UpdateSigners(ctx, epochData, signer)
	log.Default().Printf("tx: %s, block: %d, err: %v", txHash, block, err)
}

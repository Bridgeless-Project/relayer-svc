package solana

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/solana/contract"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/pkg/errors"
)

type Client struct {
	chain Chain
}

// NewBridgeClient creates a new bridge Client for the given chain.
func NewBridgeClient(chain Chain) *Client {
	contract.SetProgramID(chain.BridgeAddress)

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

func (c *Client) SendTx(ctx context.Context, instruction solana.Instruction) (*solana.Signature, error) {
	recent, err := c.chain.Rpc.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get latest blockhash")
	}
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			instruction,
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(c.chain.OperatorWallet.PublicKey()),
	)

	if err != nil {
		return nil, errors.Wrap(err, "unable to create transaction")
	}

	sign, err := tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if c.chain.OperatorWallet.PublicKey().Equals(key) {
				return &c.chain.OperatorWallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign transaction")
	}

	// Send transaction, and wait for confirmation:
	signTx, err := confirm.SendAndConfirmTransaction(ctx, c.chain.Rpc, c.chain.WsRpc, tx)
	if err != nil {
		return &sign[0], errors.Wrap(err, "unable to send transaction")
	}

	return &signTx, nil
}

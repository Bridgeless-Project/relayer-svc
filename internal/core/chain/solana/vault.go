package solana

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/solana/contract"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
)

func (c *Client) getSPLVault(ctx context.Context, deposit db.Deposit) (*solana.PublicKey, error) {
	tokenAccount, err := solana.PublicKeyFromBase58(deposit.WithdrawalToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve withdrawal token account")
	}
	tokenInfo, err := c.chain.Rpc.GetAccountInfo(ctx, tokenAccount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get withdrawal token account info")
	}

	vault, _, err := solana.FindProgramAddress([][]byte{
		[]byte("vault"),
		tokenAccount.Bytes(),
		[]byte(c.chain.Meta.BridgeId),
	}, contract.ProgramID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init splVault")
	}

	vaultInst := contract.NewInitSplVaultInstruction(c.chain.Meta.BridgeId, tokenAccount, vault,
		c.chain.OperatorWallet.PublicKey(), tokenInfo.Value.Owner, solana.SystemProgramID).Build()

	_, err = c.chain.Rpc.GetAccountInfo(ctx, vault)
	if err != nil {
		if errors.Is(err, rpc.ErrNotFound) {
			_, err = c.SendTx(ctx, vaultInst)
			if err != nil {
				return nil, errors.Wrap(err, "unable to init vault")
			}
		}
		return nil, errors.Wrap(err, "unable to get account info")
	}

	return &vault, nil
}

package solana

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
)

func (c *Client) getWithdrawalATA(ctx context.Context, depositData db.Deposit) (*solana.PublicKey, error) {
	ata, err := solana.PublicKeyFromBase58(depositData.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve receiver ATA")
	}
	_, err = c.chain.Rpc.GetAccountInfo(ctx, ata)
	if err != nil {
		if errors.Is(err, rpc.ErrNotFound) {
			err = c.deployATA(ctx, ata, depositData)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create ATA")
			}
		}

		return nil, errors.Wrap(err, "could not retrieve receiver ATA info")
	}

	return &ata, nil
}

func (c *Client) deployATA(ctx context.Context, ata solana.PublicKey, depositData db.Deposit) error {
	receiver, err := solana.PublicKeyFromBase58(depositData.Receiver)
	if err != nil {
		return errors.Wrap(err, "failed to parse receiver address")
	}

	tokenAccount, err := solana.PublicKeyFromBase58(depositData.WithdrawalToken)
	if err != nil {
		return errors.Wrap(err, "failed to parse withdrawal token address")
	}

	tokenInfo, err := c.chain.Rpc.GetAccountInfo(ctx, tokenAccount)
	if err != nil {
		return errors.Wrap(err, "failed to get token account info")
	}

	base := associatedtokenaccount.NewCreateInstruction(
		c.chain.OperatorWallet.PublicKey(),
		receiver,
		tokenAccount,
	)

	dt, _ := base.Build().Data()

	ix := &solana.GenericInstruction{
		ProgID: solana.SPLAssociatedTokenAccountProgramID,
		AccountValues: []*solana.AccountMeta{
			solana.NewAccountMeta(c.chain.OperatorWallet.PublicKey(), true, true),
			solana.NewAccountMeta(ata, true, false),
			solana.NewAccountMeta(receiver, false, false),
			solana.NewAccountMeta(tokenAccount, false, false),
			solana.NewAccountMeta(solana.SystemProgramID, false, false),
			solana.NewAccountMeta(tokenInfo.Value.Owner, false, false),
			solana.NewAccountMeta(solana.SysVarRentPubkey, false, false),
		},
		DataBytes: dt,
	}

	_, err = c.SendTx(ctx, ix)
	if err != nil {
		return errors.Wrap(err, "failed to deploy ATA")
	}

	return nil
}

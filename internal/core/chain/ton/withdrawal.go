package ton

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func (c *Client) withdraw(ctx context.Context, body *cell.Cell) (string, error) {
	bytes, err := c.OperatorWallet.SendManyWaitTxHash(ctx, []*wallet.Message{
		{
			Mode: 1,
			InternalMessage: &tlb.InternalMessage{
				IHRDisabled: true,
				Bounce:      false,
				DstAddr:     c.Chain.BridgeContractAddress,
				Amount:      tlb.FromNanoTONU(1500000000),
				Body:        body,
			},
		},
	})
	if err != nil {
		return "", errors.Wrap(err, "error sending withdrawal transaction")
	}

	return hex.EncodeToString(bytes), nil
}

func (c *Client) getStoreAddress(ctx context.Context, depositData db.Deposit) (*address.Address, error) {
	var (
		withdrawalHash []byte
		err            error
	)
	if depositData.WithdrawalToken == core.DefaultNativeTokenAddress {
		withdrawalHash, err = c.getWithdrawalNativeHash(ctx, depositData)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to get withdrawal native hash")
		}
	} else {
		withdrawalHash, err = c.getWithdrawalJettonHash(ctx, depositData)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to get withdrawal jetton hash")
		}
	}

	b, err := c.Chain.Client.CurrentMasterchainInfo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current master chain info")
	}

	res, err := c.Chain.Client.RunGetMethod(ctx, b, c.Chain.BridgeContractAddress,
		storeAddressMethod, big.NewInt(0).SetBytes(withdrawalHash))
	if err != nil {
		return nil, errors.Wrap(err, "failed to call contract to get store address")
	}

	resSlice, err := res.Slice(0)
	if err != nil {
		return nil, errors.Wrap(err, "error getting result slice")
	}
	val, err := resSlice.LoadAddr()
	if err != nil {
		return nil, errors.Wrap(err, "error loading store address")
	}

	return val, nil
}

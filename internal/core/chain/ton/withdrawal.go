package ton

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func (c *Client) WithdrawalAmountValid(amount *big.Int) bool {
	return amount.Cmp(core.ZeroAmount) == 1
}

func (c *Client) withdraw(ctx context.Context, body *cell.Cell) (string, error) {
	bytes, err := c.OperatorWallet.SendManyWaitTxHash(ctx, []*wallet.Message{
		{
			Mode: 1,
			InternalMessage: &tlb.InternalMessage{
				IHRDisabled: true,
				Bounce:      false,
				DstAddr:     c.Chain.BridgeContractAddress,
				Amount:      tlb.FromNanoTONU(1000000000),
				Body:        body,
			},
		},
	})
	if err != nil {
		return "", errors.Wrap(err, "error sending withdrawal transaction")
	}

	return hex.EncodeToString(bytes), nil
}

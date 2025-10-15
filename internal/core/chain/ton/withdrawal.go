package ton

import (
	"context"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
)

func (c *Client) WithdrawalAmountValid(amount *big.Int) bool {
	return amount.Cmp(core.ZeroAmount) == 1
}

func (c *Client) WithdrawToken(ctx context.Context, depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) WithdrawNative(ctx context.Context, depositData db.Deposit) (txHash string, err error) {
	return "", nil
}

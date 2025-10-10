package ton

import (
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
)

func (c *Client) WithdrawalAmountValid(amount *big.Int) bool {
	return amount.Cmp(broadcaster.ZeroAmount) == 1
}

func (c *Client) WithdrawNative(depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) WithdrawToken(depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) WithdrawWrapped(depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}

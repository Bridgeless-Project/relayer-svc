package evm

import (
	"context"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
)

type Operation interface {
	CalculateHash() []byte
}

func (p *Client) WithdrawalAmountValid(amount *big.Int) bool {
	if amount.Cmp(core.ZeroAmount) != 1 {
		return false
	}

	return true
}

func (p *Client) WithdrawNative(ctx context.Context, depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *Client) WithdrawToken(ctx context.Context, depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *Client) WithdrawWrapped(ctx context.Context, depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}

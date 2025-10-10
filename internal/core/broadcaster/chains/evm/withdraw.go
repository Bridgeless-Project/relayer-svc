package evm

import (
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
)

type Operation interface {
	CalculateHash() []byte
}

func (p *Client) WithdrawalAmountValid(amount *big.Int) bool {
	if amount.Cmp(broadcaster.ZeroAmount) != 1 {
		return false
	}

	return true
}

func (p *Client) WithdrawNative(depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *Client) WithdrawToken(depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *Client) WithdrawWrapped(depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}

// Code generated; DO NOT EDIT.
package contract

import (
	"math/big"

	"github.com/xssnick/tonutils-go/address"
)

const (
	OpJettonDeposited       = 0x2ddcbe3
	OpNativeDeposited       = 0xe858a993
	OpWithdrawError         = 0xc01d9831
	OpExcessed              = 0x1389787
	OpAfterDuplicationCheck = 0x1111112
	OpBeforeSignatureCheck  = 0x11112222
	OpBeforeWithdraw        = 0x33333333
)

type JettonDeposited struct {
	Sender     address.Address `bit_size:"-"`
	Receiver   []byte          `bit_size:"-"`
	Amount     *big.Int        `bit_size:"-"`
	Network    []byte          `bit_size:"-"`
	IsWrapped  bool            `bit_size:"-"`
	Token      address.Address `bit_size:"-"`
	ReferralId *big.Int        `bit_size:"16"`
}

type NativeDeposited struct {
	Sender     address.Address `bit_size:"-"`
	Receiver   []byte          `bit_size:"-"`
	Network    []byte          `bit_size:"-"`
	Amount     *big.Int        `bit_size:"-"`
	ReferralId *big.Int        `bit_size:"16"`
}

type WithdrawError struct {
	Sender      address.Address `bit_size:"-"`
	Receiver    address.Address `bit_size:"-"`
	Network     []byte          `bit_size:"-"`
	Amount      *big.Int        `bit_size:"-"`
	MessageType *big.Int        `bit_size:"-"`
	ErrorCode   *big.Int        `bit_size:"-"`
	Hash        *big.Int        `bit_size:"-"`
}

type Excessed struct {
	QueryId *big.Int `bit_size:"64"`
}

type AfterDuplicationCheck struct {
	Hash *big.Int `bit_size:"-"`
}

type BeforeSignatureCheck struct {
	Hash *big.Int `bit_size:"-"`
}

type BeforeWithdraw struct {
	Hash *big.Int `bit_size:"-"`
}

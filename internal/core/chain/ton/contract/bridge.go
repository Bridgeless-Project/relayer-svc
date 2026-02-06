// Code generated; DO NOT EDIT.
package contract

import (
	"math/big"

	"github.com/xssnick/tonutils-go/address"
)

const (
	OpMsgDepositNative              = 0x5386a723
	OpMsgWithdrawJetton             = 0x0255667d
	OpMsgWithdrawNative             = 0x4cdec437
	OpMsgUpdateSignerKeyBySignature = 0x12312324
	OpMsgRemoveSignerKey            = 0x14141414
	OpMsgSetSignerKey               = 0x12312323
	OpMsgUpdateOwner                = 0x12121212
	OpMsgUpdateBridgeJettonWallets  = 0x13131444
	OpMsgUpdateWrappedJetton        = 0x13131555
	OpMsgChangeJettonMasterOwner    = 0x7187b179
)

const (
	AddressSize = 267
	BitSize     = 1
	IntSize     = 257
	CellSize    = 1023
)

type Msg interface {
	MsgDepositNative | MsgWithdrawJetton | MsgWithdrawNative | MsgUpdateSignerKeyBySignature | MsgRemoveSignerKey | MsgSetSignerKey | MsgUpdateOwner | MsgUpdateBridgeJettonWallets | MsgUpdateWrappedJetton | MsgChangeJettonMasterOwner | JettonDeposited | NativeDeposited | WithdrawError | Excessed | AfterDuplicationCheck | BeforeSignatureCheck | BeforeWithdraw | MsgIsUsed | MsgIsUsedResponse
}

type MsgDepositNative struct {
	Receiver   []byte   `bit_size:"-"`
	Network    []byte   `bit_size:"-"`
	ReferralId *big.Int `bit_size:"16"`
}

type MsgWithdrawJetton struct {
	Receiver            address.Address `bit_size:"-"`
	TxHash              *big.Int        `bit_size:"-"`
	TxNonce             *big.Int        `bit_size:"-"`
	Network             []byte          `bit_size:"-"`
	IsWrapped           bool            `bit_size:"-"`
	Signature           []byte          `bit_size:"-"`
	Token               address.Address `bit_size:"-"`
	BridgeJettonAddress address.Address `bit_size:"-"`
	ForwardTonAmount    *big.Int        `bit_size:"-"`
	TotalTonAmount      *big.Int        `bit_size:"-"`
}

type MsgWithdrawNative struct {
	Amount    *big.Int        `bit_size:"-"`
	Receiver  address.Address `bit_size:"-"`
	TxHash    *big.Int        `bit_size:"257"`
	TxNonce   *big.Int        `bit_size:"-"`
	Signature []byte          `bit_size:"-"`
	Network   []byte          `bit_size:"-"`
}

type MsgUpdateSignerKeyBySignature struct {
	H            *big.Int `bit_size:"8"`
	X            *big.Int `bit_size:"256"`
	Y            *big.Int `bit_size:"256"`
	Starttime    *big.Int `bit_size:"32"`
	Deadline     *big.Int `bit_size:"32"`
	AdditionMode bool     `bit_size:"-"`
	Signature    []byte   `bit_size:"-"`
}

type MsgRemoveSignerKey struct {
	H *big.Int `bit_size:"8"`
	X *big.Int `bit_size:"256"`
	Y *big.Int `bit_size:"256"`
}

type MsgSetSignerKey struct {
	H *big.Int `bit_size:"8"`
	X *big.Int `bit_size:"256"`
	Y *big.Int `bit_size:"256"`
}

type MsgUpdateOwner struct {
	NewOwner address.Address `bit_size:"-"`
}

type MsgUpdateBridgeJettonWallets struct {
}

type MsgUpdateWrappedJetton struct {
}

type MsgChangeJettonMasterOwner struct {
	NewOwner address.Address `bit_size:"-"`
}

type fieldInfo struct {
	typeName  string
	fieldName string
	bitSize   int
	value     interface{}
}

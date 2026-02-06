// Code generated; DO NOT EDIT.
package contract

import (
	"math/big"

	"github.com/xssnick/tonutils-go/address"
)

const (
	OpMsgIsUsed         = 0x12dd3d08
	OpMsgIsUsedResponse = 0x1c3b2f8d
)

type MsgIsUsed struct {
	Destination        address.Address `bit_size:"-"`
	DestinationPayload []byte          `bit_size:"-"`
	Hash               *big.Int        `bit_size:"-"`
	MessageType        *big.Int        `bit_size:"-"`
	Sender             address.Address `bit_size:"-"`
}

type MsgIsUsedResponse struct {
	DestinationPayload []byte          `bit_size:"-"`
	IsUsed             bool            `bit_size:"-"`
	RequestHash        *big.Int        `bit_size:"-"`
	MessageType        *big.Int        `bit_size:"-"`
	Sender             address.Address `bit_size:"-"`
}

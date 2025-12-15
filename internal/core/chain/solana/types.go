package solana

import (
	"github.com/gagliardetto/solana-go"
)

type withdrawalContext struct {
	Receiver         solana.PublicKey
	Token            solana.PublicKey
	WithdrawalPDA    solana.PublicKey
	Authority        solana.PublicKey
	BridgeID         string
	UID              [32]byte
	Sig              [64]uint8
	RecID            uint8
	WithdrawalTxHash [32]byte
	TxNonce          uint64
	Amount           uint64
}

type tokenMetadata struct {
	Name   string
	Symbol string
	Uri    string
	Nonce  uint64
}

type signerInfo struct {
}

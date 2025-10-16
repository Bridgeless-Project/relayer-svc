package chain

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
)

var (
	ErrChainNotSupported      = errors.New("chain not supported")
	ErrTxPending              = errors.New("transaction is pending")
	ErrTxFailed               = errors.New("transaction failed")
	ErrTxNotFound             = errors.New("transaction not found")
	ErrDepositNotFound        = errors.New("deposit not found")
	ErrTxNotConfirmed         = errors.New("transaction not confirmed")
	ErrInvalidReceiverAddress = errors.New("invalid receiver address")
	ErrInvalidDepositedAmount = errors.New("invalid deposited amount")
	ErrInvalidScriptPubKey    = errors.New("invalid script pub key")
	ErrInvalidTxNonce         = errors.New("invalid tx nonce")
	ErrFailedUnpackLogs       = errors.New("failed to unpack logs")
	ErrUnsupportedEvent       = errors.New("unsupported event")
	ErrUnsupportedContract    = errors.New("unsupported contract")
)

type Client interface {
	Type() Type
	ChainId() string

	TransactionHashValid(hash string) bool
	IsProcessed(ctx context.Context, depositData db.Deposit) (bool, error)

	WithdrawNative(ctx context.Context, depositData db.Deposit) (string, error)
	WithdrawToken(ctx context.Context, depositData db.Deposit) (string, error)
}

type Repository interface {
	Client(chainId string) (Client, error)
	SupportsChain(chainId string) bool
}

type Chain struct {
	Id                 string `fig:"id,required"`
	Type               Type   `fig:"type,required"`
	Rpc                any    `fig:"rpc,required"`
	BridgeAddresses    any    `fig:"bridge_address,required"`
	OperatorPrivateKey string `fig:"operator_private_key,required"`

	Meta any `fig:"meta"`
}

type Type string

const (
	TypeEVM    Type = "EVM"
	TypeTON    Type = "TON"
	TypeSolana Type = "SOL"
	TypeOther  Type = "other"
)

var typesMap = map[Type]struct{}{
	TypeEVM:    {},
	TypeOther:  {},
	TypeTON:    {},
	TypeSolana: {},
}

func (c Type) Validate() error {
	if _, ok := typesMap[c]; !ok {
		return errors.New("invalid chain type")
	}

	return nil
}

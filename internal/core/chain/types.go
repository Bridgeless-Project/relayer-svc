package chain

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
)

var (
	ErrChainNotSupported = errors.New("chain not supported")
)

type Client interface {
	Type() Type
	ChainId() string
	Workers() int

	TransactionHashValid(hash string) bool
	IsProcessed(ctx context.Context, depositData db.Deposit) (bool, error)

	Withdraw(ctx context.Context, depositData *db.Deposit) (string, int64, error)
}

type Repository interface {
	Client(chainId string) (Client, error)
	SupportsChain(chainId string) bool
	Clients() map[string]Client
}

type Chain struct {
	Id                 string `fig:"id,required"`
	Type               Type   `fig:"type,required"`
	Rpc                any    `fig:"rpc,required"`
	BridgeAddresses    any    `fig:"bridge_address,required"`
	OperatorPrivateKey string `fig:"operator_private_key,required"`
	WSTimeout          int64  `fig:"ws_timeout"`
	WSRpc              any    `fig:"ws_rpc"`
	Workers            int    `fig:"workers,required"`

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

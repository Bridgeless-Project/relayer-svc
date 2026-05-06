package solana

import (
	"context"
	"reflect"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure/v3"
)

type Chain struct {
	Id               string
	Rpc              *rpc.Client
	WsRpc            *ws.Client
	BridgeAddress    solana.PublicKey
	OperatorsWallets []*solana.Wallet
	Workers          int
	WSTimeout        int64

	Meta Meta
}

type Meta struct {
	BridgeId  string `fig:"bridge_id,required"`
	IsTestnet bool   `fig:"is_testnet,required"`
}

var SolanaHooks = figure.Hooks{
	"*rpc.Client": func(value interface{}) (reflect.Value, error) {
		switch v := value.(type) {
		case string:
			client := rpc.New(v)
			return reflect.ValueOf(client), nil
		default:
			return reflect.Value{}, errors.Errorf("unsupported conversion from %T", value)
		}
	},
	"solana.PublicKey": func(value interface{}) (reflect.Value, error) {
		switch v := value.(type) {
		case string:
			pubKey, err := solana.PublicKeyFromBase58(v)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(pubKey), nil
		default:
			return reflect.Value{}, errors.Errorf("unsupported conversion from %T", value)
		}
	},
	"[]*solana.Wallet": func(value interface{}) (reflect.Value, error) {
		switch v := value.(type) {
		case []string:
			wallets := make([]*solana.Wallet, len(v))
			for i, str := range v {
				wallet, err := solana.WalletFromPrivateKeyBase58(str)
				if err != nil {
					return reflect.Value{}, err
				}

				wallets[i] = wallet
			}

			return reflect.ValueOf(wallets), nil
		default:
			return reflect.Value{}, errors.Errorf("unsupported conversion from %T", value)
		}
	},
}

func FromChain(c chain.Chain) Chain {
	if c.Type != chain.TypeSolana {
		panic("chain is not Solana")
	}
	chain := Chain{
		Id: c.Id,
	}

	if err := figure.Out(&chain.Meta).
		FromInterface(c.Meta).Please(); err != nil {
		panic(errors.Wrap(err, "failed to decode chain meta"))
	}
	if err := figure.Out(&chain.Rpc).
		FromInterface(c.Rpc).
		With(SolanaHooks).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain Solana clients"))
	}
	if err := figure.Out(&chain.BridgeAddress).
		FromInterface(c.BridgeAddresses).
		With(SolanaHooks).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain bridge addresses"))
	}
	if err := figure.Out(&chain.OperatorsWallets).
		FromInterface(c.OperatorsPrivateKeys).
		With(SolanaHooks).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain operator wallet"))
	}
	if err := figure.Out(&chain.WSTimeout).FromInterface(c.WSTimeout).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain timeout"))
	}

	if err := figure.Out(&chain.Workers).FromInterface(c.Workers).Please(); err != nil {
		panic(errors.Wrap(err, "failed to obtain workers number"))
	}

	wsEndpoint := rpc.MainNetBeta_WS
	if chain.Meta.IsTestnet {
		wsEndpoint = rpc.TestNet_WS
	}

	wsRpc, err := ws.Connect(context.Background(), wsEndpoint)
	if err != nil {
		panic(errors.Wrap(err, "failed to connect to websocket"))
	}

	chain.WsRpc = wsRpc

	if chain.Workers > len(chain.OperatorsWallets) {
		panic("number of workers is greater than number of operators private keys")
	}

	return chain
}

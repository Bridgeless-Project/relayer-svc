package config

import (
	"context"
	"reflect"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/evm"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/solana"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/ton"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type Chainer interface {
	Chains() []chain.Chain
	Clients(ctx context.Context) []chain.Client
}

type chainer struct {
	chainsOnce  comfig.Once
	clientsOnce comfig.Once
	getter      kv.Getter
}

func NewChainer(getter kv.Getter) Chainer {
	return &chainer{
		getter: getter,
	}
}

func (c *chainer) Clients(ctx context.Context) []chain.Client {
	return c.clientsOnce.Do(func() interface{} {
		chains := c.Chains()
		clients := make([]chain.Client, len(chains))

		for i, ch := range chains {
			switch ch.Type {
			case chain.TypeSolana:
				clients[i] = solana.NewBridgeClient(solana.FromChain(ctx, ch))
			case chain.TypeTON:
				clients[i] = ton.NewBridgeClient(ton.FromChain(ch))
			case chain.TypeEVM:
				clients[i] = evm.NewBridgeClient(evm.FromChain(ch))
			default:
				panic(errors.Errorf("unsupported chain type: %s", ch.Type))
			}
		}

		return clients
	}).([]chain.Client)
}

func (c *chainer) Chains() []chain.Chain {
	return c.chainsOnce.Do(func() interface{} {
		var cfg struct {
			Chains []chain.Chain `fig:"list,required"`
		}

		if err := figure.
			Out(&cfg).
			With(
				figure.BaseHooks,
				figure.EthereumHooks,
				solana.SolanaHooks,
				interfaceHook,
			).
			From(kv.MustGetStringMap(c.getter, "chain")).
			Please(); err != nil {
			panic(errors.Wrap(err, "failed to figure out chain"))
		}

		if len(cfg.Chains) == 0 {
			panic(errors.New("no chain were configured"))
		}

		return cfg.Chains
	}).([]chain.Chain)
}

// simple hook to delay parsing interface details
var interfaceHook = figure.Hooks{
	"interface {}": func(value interface{}) (reflect.Value, error) {
		return reflect.ValueOf(value), nil
	},
}

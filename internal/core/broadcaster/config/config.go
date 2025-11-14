package config

import (
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

const yamlKey = "broadcaster"

type BroadcasterConfigurer interface {
	TxPoolSize() int64
}

type BroadcasterConfig struct {
	TxPoolSize int64 `fig:"tx_pool_size,required"`
}

type configurer struct {
	once   comfig.Once
	getter kv.Getter
}

func NewBroadcasterConfigurer(getter kv.Getter) BroadcasterConfigurer {
	return &configurer{
		getter: getter,
	}
}

func (b *configurer) TxPoolSize() int64 {
	return b.Config().TxPoolSize
}

func (c *configurer) Config() BroadcasterConfig {
	return c.once.Do(func() interface{} {
		cfg := &BroadcasterConfig{}

		if err := figure.
			Out(&cfg).
			From(kv.MustGetStringMap(c.getter, yamlKey)).
			Please(); err != nil {
			panic(errors.Wrap(err, "failed to configure broadcaster"))
		}

		return cfg
	}).(BroadcasterConfig)
}

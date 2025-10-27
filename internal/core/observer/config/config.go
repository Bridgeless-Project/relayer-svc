package config

import (
	"time"

	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

const observerConfigKey = "observer"

type ObserverConfigurator interface {
	ObserverPollingInterval() time.Duration
}

type observer struct {
	once   comfig.Once
	getter kv.Getter
}

func (sc *observer) ObserverPollingInterval() time.Duration {
	return time.Duration(sc.Config().PollingInterval) * time.Second
}

type config struct {
	PollingInterval int64 `fig:"polling_interval_sec,required"`
}

func NewConfigurator(getter kv.Getter) ObserverConfigurator {
	return &observer{
		getter: getter,
	}
}

func (sc *observer) Config() *config {
	return sc.once.Do(func() interface{} {
		var cfg config
		if err := figure.Out(&cfg).From(kv.MustGetStringMap(sc.getter, observerConfigKey)).Please(); err != nil {
			panic(errors.Wrap(err, "failed to figure out core observer config"))
		}
		return &cfg
	}).(*config)
}

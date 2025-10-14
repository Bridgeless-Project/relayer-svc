package config

import (
	"time"

	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

const observerConfigKey = "observer"

type ObserverConfigurator interface {
	TendermintHttpClient() *http.HTTP
	ObserverRetries() int64
	ObserverRetryTimeout() time.Duration
	ObserverPollingInterval() time.Duration
}

type observer struct {
	once   comfig.Once
	getter kv.Getter
}

func (sc *observer) TendermintHttpClient() *http.HTTP {
	cfg := sc.Config()
	client, err := http.New(cfg.Addr, "/websocket")
	if err != nil {
		panic(errors.Wrap(err, "failed to create tendermint http client"))
	}

	if err = client.Start(); err != nil {
		panic(errors.Wrap(err, "failed to start tendermint http client"))
	}

	return client
}

func (sc *observer) ObserverRetries() int64 {
	return sc.Config().Retries
}

func (sc *observer) ObserverRetryTimeout() time.Duration {
	return time.Duration(sc.Config().RetryTimeout) * time.Second
}

func (sc *observer) ObserverPollingInterval() time.Duration {
	return time.Duration(sc.Config().PollingInterval) * time.Second
}

type config struct {
	Addr            string `fig:"tendermint_rpc,required"`
	Retries         int64  `fig:"retry_attempts,required"`
	RetryTimeout    int64  `fig:"retry_timeout_sec,required"`
	PollingInterval int64  `fig:"polling_interval_sec,required"`
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
			panic(errors.Wrap(err, "failed to figure out core subscriber config"))
		}
		return &cfg
	}).(*config)
}

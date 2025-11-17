package config

import (
	"time"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

const (
	retrierKey = "retry"
)

type retrier struct {
	getter kv.Getter
	once   comfig.Once
}

func (r *retrier) RetryAttempts() uint {
	return r.config().Retries
}

func (r *retrier) RetryTimeout() time.Duration {
	return time.Duration(r.config().RetryTimeout) * time.Second
}

type Retrier interface {
	RetryAttempts() uint
	RetryTimeout() time.Duration
}

type retrierCfg struct {
	Retries      uint  `fig:"retry_attempts,required"`
	RetryTimeout int64 `fig:"retry_timeout_sec,required"`
}

func NewRetrier(getter kv.Getter) Retrier {
	return &retrier{
		getter: getter,
	}
}

func (r *retrier) config() *retrierCfg {
	return r.once.Do(func() interface{} {
		var cfg retrierCfg
		if err := figure.Out(&cfg).From(kv.MustGetStringMap(r.getter, retrierKey)).Please(); err != nil {
			panic(errors.Wrap(err, "failed to figure out retry config"))
		}

		core.Retries = cfg.Retries
		core.RetryTimeout = time.Duration(cfg.RetryTimeout) * time.Second

		return &cfg
	}).(*retrierCfg)
}

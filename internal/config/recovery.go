package config

import (
	"time"

	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

const (
	recoveryKey = "recovery"
)

type recovery struct {
	getter kv.Getter
	once   comfig.Once
}

func (r *recovery) RecoveryParams() (uint, time.Duration) {
	return r.config().RecoveryAttempts, time.Duration(r.config().RecoveryTimeout) * time.Second
}

type Recovery interface {
	RecoveryParams() (uint, time.Duration)
}

type recoveryCfg struct {
	RecoveryAttempts uint  `fig:"recovery_attempts,required"`
	RecoveryTimeout  int64 `fig:"recovery_timeout,required"`
}

func NewRecovery(getter kv.Getter) Recovery {
	return &recovery{
		getter: getter,
	}
}

func (r *recovery) config() *recoveryCfg {
	return r.once.Do(func() interface{} {
		var cfg recoveryCfg
		if err := figure.Out(&cfg).From(kv.MustGetStringMap(r.getter, recoveryKey)).Please(); err != nil {
			panic(errors.Wrap(err, "failed to figure out retry config"))
		}

		return &cfg
	}).(*recoveryCfg)
}

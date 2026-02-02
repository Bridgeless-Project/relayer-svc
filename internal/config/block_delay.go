package config

import (
	"time"

	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

const blockDelayerKey = "block_delay"

type BlockDelay struct {
	BlockDelay int64 `fig:"time,required"`
}

type BlockDelaySetter interface {
	BlockDelay() time.Duration
}

type blockDelaySetter struct {
	getter kv.Getter
	once   comfig.Once
}

func (s *blockDelaySetter) BlockDelay() time.Duration {
	return time.Duration(s.config().BlockDelay) * time.Second
}

func NewBlockDelaySetter(getter kv.Getter) BlockDelaySetter {
	return &blockDelaySetter{
		getter: getter,
	}
}

func (s *blockDelaySetter) config() *BlockDelay {
	return s.once.Do(func() interface{} {
		var cfg BlockDelay

		if err := figure.Out(&cfg).From(kv.MustGetStringMap(s.getter, blockDelayerKey)).Please(); err != nil {
			panic(errors.Wrap(err, "failed to figure out block delay config"))
		}
		return &cfg
	}).(*BlockDelay)
}

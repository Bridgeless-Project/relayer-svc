package config

import (
	observer "github.com/Bridgeless-Project/relayer-svc/internal/core/observer/config"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
)

type Config interface {
	comfig.Logger
	pgdb.Databaser
	Listenerer
	observer.ObserverConfigurator
}

type config struct {
	getter kv.Getter

	comfig.Logger
	pgdb.Databaser
	Listenerer
	observer.ObserverConfigurator
}

func New(getter kv.Getter) Config {
	return &config{
		getter:               getter,
		Logger:               comfig.NewLogger(getter, comfig.LoggerOpts{}),
		Databaser:            pgdb.NewDatabaser(getter),
		Listenerer:           NewListenerer(getter),
		ObserverConfigurator: observer.NewConfigurator(getter),
	}
}

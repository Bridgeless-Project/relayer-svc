package ctx

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/connector"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	dbKey          ctxKey = iota
	loggerKey      ctxKey = iota
	clientsRepoKey ctxKey = iota
	broadcasterKey ctxKey = iota
	connectorKey   ctxKey = iota
)

func DBProvider(q db.DepositsQ) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {

		return context.WithValue(ctx, dbKey, q)
	}
}

// DB always returns unique connection
func DB(ctx context.Context) db.DepositsQ {
	return ctx.Value(dbKey).(db.DepositsQ).New()
}

func LoggerProvider(l *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {

		return context.WithValue(ctx, loggerKey, l)
	}
}
func Logger(ctx context.Context) *logan.Entry {

	return ctx.Value(loggerKey).(*logan.Entry)
}

func ClientsProvider(c chain.Repository) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, clientsRepoKey, c)
	}
}

func Clients(ctx context.Context) chain.Repository {
	return ctx.Value(clientsRepoKey).(chain.Repository)
}

func BroadcasterProvider(c *broadcaster.Broadcaster) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, broadcasterKey, c)
	}
}

func Broadcaster(ctx context.Context) *broadcaster.Broadcaster {
	return ctx.Value(broadcasterKey).(*broadcaster.Broadcaster)
}

func ConnectorProvider(c *connector.Connector) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, connectorKey, c)
	}
}

func Connector(ctx context.Context) *connector.Connector {
	return ctx.Value(connectorKey).(*connector.Connector)
}

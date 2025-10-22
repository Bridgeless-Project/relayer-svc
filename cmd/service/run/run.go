package run

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Bridgeless-Project/relayer-svc/cmd/utils"
	"github.com/Bridgeless-Project/relayer-svc/internal/api"
	"github.com/Bridgeless-Project/relayer-svc/internal/config"
	withdrawalBroadcaster "github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/repository"
	coreConnector "github.com/Bridgeless-Project/relayer-svc/internal/core/connector"
	coreObserver "github.com/Bridgeless-Project/relayer-svc/internal/core/observer"
	pg "github.com/Bridgeless-Project/relayer-svc/internal/db/postgres"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func init() {
	utils.RegisterCatchUpFlag(Cmd)
	utils.RegisterConfigFlag(Cmd)
	utils.RegisterStartHeightFlag(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := utils.ConfigFromFlags(cmd)
		if err != nil {
			return errors.Wrap(err, "failed to get config from flags")
		}

		catchUp, err := utils.CatchUpFromFlags(cmd)
		if err != nil {
			return errors.Wrap(err, "failed to catch up from flags")
		}

		startHeight, err := utils.StartHeightFromFlags(cmd)
		if err != nil {
			return errors.Wrap(err, "failed to start height")
		}

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
		defer cancel()

		err = runService(ctx, cfg, catchUp, startHeight)

		return errors.Wrap(err, "failed to run relayer service")
	},
}

func runService(ctx context.Context, cfg config.Config, catchUp bool, startHeight uint64) error {
	wg := new(sync.WaitGroup)
	eg, ctx := errgroup.WithContext(ctx)
	logger := cfg.Log()
	clients := cfg.Clients()
	clientsRepo := repository.NewClientsRepository(clients)
	dtb := pg.NewDepositsQ(cfg.DB())
	blocksQ := pg.NewBlocksQ(cfg.DB())

	connector, err := coreConnector.NewConnector(cfg.CoreConnectorConfig().Account, cfg.CoreConnectorConfig().Connection,
		cfg.CoreConnectorConfig().Settings)
	if err != nil {
		return errors.Wrap(err, "failed to create connector")
	}

	broadcaster := withdrawalBroadcaster.New(connector, dtb, clientsRepo, logger.WithField("component", "broadcaster"))

	observer := coreObserver.New(cfg.TendermintHttpClient(), cfg.ObserverRetries(), cfg.ObserverRetryTimeout(),
		cfg.ObserverPollingInterval(), blocksQ, dtb, broadcaster, clientsRepo, logger)

	apiServer := api.NewServer(cfg.ApiGrpcListener(), cfg.ApiHttpListener(), dtb, connector, broadcaster, clientsRepo,
		logger.WithField("component", "api-server"))

	wg.Add(2)
	eg.Go(func() error {
		defer wg.Done()
		return errors.Wrap(apiServer.RunHTTP(ctx), "error while running API HTTP gateway")
	})
	eg.Go(func() error {
		defer wg.Done()
		return errors.Wrap(apiServer.RunGRPC(ctx), "error while running API GRPC server")
	})

	wg.Add(2)
	eg.Go(func() error {
		defer wg.Done()
		broadcaster.Run(ctx)
		return nil
	})
	eg.Go(func() error {
		defer wg.Done()
		return errors.Wrap(observer.Run(ctx, startHeight, catchUp), "error while running observer")
	})

	err = eg.Wait()
	wg.Wait()

	return err
}

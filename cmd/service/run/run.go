package run

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Bridgeless-Project/relayer-svc/cmd/utils"
	"github.com/Bridgeless-Project/relayer-svc/internal/api"
	"github.com/Bridgeless-Project/relayer-svc/internal/config"
	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	withdrawalBroadcaster "github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster"
	catch_upper "github.com/Bridgeless-Project/relayer-svc/internal/core/catch-upper"
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
	utils.RegisterBlockDistanceFlag(Cmd)
	utils.RegisterObserverFlag(Cmd)
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

		blockDistance, err := utils.BlockDistanceFromFlags(cmd)
		if err != nil {
			return errors.Wrap(err, "failed to get block distance")
		}

		observer, err := utils.ObserverFromFlags(cmd)
		if err != nil {
			return errors.Wrap(err, "failed to observer from flags")
		}

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
		defer cancel()

		err = runService(ctx, cfg, catchUp, observer, startHeight, blockDistance)

		return errors.Wrap(err, "failed to run relayer service")
	},
}

func runService(ctx context.Context, cfg config.Config, catchUp, observerNeeded bool, startHeight, blockDistance uint64) error {
	wg := new(sync.WaitGroup)
	eg, ctx := errgroup.WithContext(ctx)
	logger := cfg.Log()
	clients := cfg.Clients()
	clientsRepo := repository.NewClientsRepository(clients)
	dtb := pg.NewDepositsQ(cfg.DB())
	blocksQ := pg.NewBlocksQ(cfg.DB())

	core.Logger = logger.WithField("component", "retrier")
	core.Retries = cfg.RetryAttempts()
	core.RetryTimeout = cfg.RetryTimeout()

	connector, err := coreConnector.NewConnector(cfg.CoreConnectorConfig().Account, cfg.CoreConnectorConfig().Connection,
		cfg.CoreConnectorConfig().Settings)
	if err != nil {
		return errors.Wrap(err, "failed to create connector")
	}

	broadcaster := withdrawalBroadcaster.New(ctx, connector, dtb, cfg.TendermintHttpClient(), logger.WithField("component", "broadcaster"))

	observer := coreObserver.New(cfg.TendermintHttpClient(), blocksQ, dtb, broadcaster, logger.WithField("component", "observer"))

	apiServer := api.NewServer(cfg.ApiGrpcListener(), cfg.ApiHttpListener(), dtb, connector, broadcaster, clientsRepo,
		logger.WithField("component", "api-server"))

	catchUpper := catch_upper.NewCatchUpper(ctx, broadcaster, dtb, logger.WithField("component", "catch-upper"))

	wg.Add(2)
	eg.Go(func() error {
		defer wg.Done()
		return errors.Wrap(apiServer.RunHTTP(ctx), "error while running API HTTP gateway")
	})
	eg.Go(func() error {
		defer wg.Done()
		return errors.Wrap(apiServer.RunGRPC(ctx), "error while running API GRPC server")
	})

	wg.Add(1)
	eg.Go(func() error {
		defer wg.Done()
		broadcaster.
			WithClients(clientsRepo).
			WithChainTxPoolSize(cfg.ChainTxPoolSize()).
			WithSubmitTxPoolSize(cfg.SubmitTxPoolSize()).
			WithSubmitBatchSize(cfg.SubmitBatchSize()).
			Run(ctx)

		return nil
	})

	if observerNeeded {
		wg.Add(1)
		eg.Go(func() error {
			defer wg.Done()
			return errors.Wrap(observer.
				WithClientsRepo(clientsRepo).
				WithPollingInterval(cfg.ObserverPollingInterval()).
				WithBlockDelay(cfg.BlockDelay()).
				WithBlockDistance(blockDistance).
				Run(ctx, startHeight), "error while running observer")
		})
	}

	if catchUp {
		wg.Add(1)
		eg.Go(func() error {
			defer wg.Done()
			return errors.Wrap(catchUpper.Start(), "error while running observer")
		})
	}

	err = eg.Wait()
	wg.Wait()

	return err
}

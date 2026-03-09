package broadcaster

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster/containers"
)

func (b *Broadcaster) runNetworkWorker(ctx context.Context, chainID string, ch <-chan containers.WithdrawalContainer, workerId int) {
	defer b.wg.Done()
	log := b.logger.WithField("chain_id", chainID).WithField("worker_id", workerId)
	log.Debug("started broadcaster worker")

	for {
		select {
		case <-ctx.Done():
			log.Debug("context canceled, stopping network worker")
			return
		case container, ok := <-ch:
			if !ok {
				log.Debug("channel closed, stopping network worker")
				return
			}

			deposit, err := container.Run(ctx)
			if err != nil {
				b.cache.Delete(deposit.String())
				log.WithError(err).Errorf("error processing withdrawal, container ID: %s", container.ID())
				continue
			}

			b.submitChan <- deposit
		}
	}
}

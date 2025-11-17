package broadcaster

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
)

func (b *Broadcaster) runCoreSubmitter(ctx context.Context) {
	defer b.wg.Done()
	logger := b.logger.WithField("component", "core-submitter")

	for {
		select {
		case <-ctx.Done():
			logger.Debug("context stopped, stopping core submitter")
			return

		case deposit, ok := <-b.submitChan:
			if !ok {
				logger.Debug("submit channel closed, stopping core submitter")
				return
			}

			updateTx := func() error {
				return b.coreConnector.UpdateTxInfo(ctx, *deposit)
			}

			err := core.DoWithRetry(ctx, updateTx)
			if err != nil {
				logger.WithError(err).Errorf("error updating withdrawal info for deposit: %s", deposit.String())
				continue
			}

			if deposit.WithdrawalStatus != internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE {
				continue
			}

			err = b.dbConn.UpdateStatus(deposit.DepositIdentifier,
				internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED)
			if err != nil {
				logger.WithError(err).Errorf("error updating withdrawal status to processed for deposit: %s",
					deposit.String())
			}
		}
	}
}

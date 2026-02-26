package broadcaster

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
)

func (b *Broadcaster) runCoreSubmitter(ctx context.Context) {
	defer b.wg.Done()
	logger := b.logger.WithField("component", "core-submitter")

	submitTxPool := make([]*db.Deposit, 0, b.submitBatchSize)

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

			submitTxPool = append(submitTxPool, deposit)
			if len(submitTxPool) < int(b.submitBatchSize) {
				continue
			}

			updateTx := func() error {
				return b.coreConnector.UpdateTxInfo(ctx, submitTxPool)
			}

			err := core.DoWithRetry(ctx, updateTx)
			if err != nil {
				logger.WithError(err).Errorf("error updating withdrawal info for deposit: %s", deposit.String())
				continue
			}

			for _, d := range submitTxPool {
				if d.WithdrawalStatus != internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE {
					continue
				}

				err = b.depositsDbConn.UpdateStatus(d.DepositIdentifier,
					internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED)
				if err != nil {
					logger.WithError(err).Errorf("error updating withdrawal status to processed for deposit: %s",
						d.String())
				}
			}

			submitTxPool = submitTxPool[:0]
		}
	}
}

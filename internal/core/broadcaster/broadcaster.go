package broadcaster

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/connector"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

type Broadcaster struct {
	coreConnector *connector.Connector
	clientsRepo   chain.Repository
	depositChan   chan db.Deposit

	dbConn db.DepositsQ
	logger *logan.Entry
}

func New(coreConnector *connector.Connector, dbConn db.DepositsQ, clientsRepo chain.Repository,
	depositChan chan db.Deposit, logger *logan.Entry) *Broadcaster {
	return &Broadcaster{
		coreConnector: coreConnector,
		clientsRepo:   clientsRepo,
		depositChan:   depositChan,
		dbConn:        dbConn,
	}
}

func (b *Broadcaster) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			b.logger.Debug("context canceled. Stopping broadcaster")
			return nil
		case deposit := <-b.depositChan:
			var withdrawalTxHash string
			chainClient, err := b.clientsRepo.Client(deposit.WithdrawalChainId)
			if err != nil {
				return errors.Wrapf(err, "error getting chain client for chain id: %s", deposit.WithdrawalChainId)
			}
			switch deposit.WithdrawalToken {
			case core.DefaultNativeTokenAddress:
				withdrawalTxHash, err = chainClient.WithdrawNative(deposit)
				if err != nil {
					return errors.Wrapf(err, "failed to withdraw deposit, identifier: %v ", deposit.DepositIdentifier)
				}

				err = b.dbConn.Transaction(func() error {
					dbErr := b.dbConn.UpdateWithdrawalTx(deposit.DepositIdentifier, withdrawalTxHash)
					if dbErr != nil {
						return errors.Wrap(err, "failed to update withdrawal tx hash")
					}

					dbErr = b.dbConn.UpdateStatus(deposit.DepositIdentifier, types.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED)
					if dbErr != nil {
						return errors.Wrap(err, "failed to update deposit status")
					}

					return nil
				})

				if err != nil {
					return errors.Wrap(err, "failed to update deposit withdrawal details")
				}

			}
		}
	}
}

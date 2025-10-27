package containers

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/connector"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

type catchupContainer struct {
	id string

	dbQ           db.DepositsQ
	deposit       db.Deposit
	chainClient   chain.Client
	coreConnector *connector.Connector

	logger *logan.Entry
}

func NewCatchUpContainer(chainClient chain.Client, deposit db.Deposit, dbQ db.DepositsQ,
	connector *connector.Connector, logger *logan.Entry) WithdrawalContainer {
	return &catchupContainer{
		id:            deposit.String(),
		chainClient:   chainClient,
		deposit:       deposit,
		dbQ:           dbQ,
		coreConnector: connector,
		logger:        logger.WithField("catchup_container", deposit.String()),
	}
}

func (c *catchupContainer) ID() string {
	return c.id
}

func (c *catchupContainer) Run(ctx context.Context) (*db.Deposit, error) {
	switch c.deposit.WithdrawalStatus {
	case internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSING:
		return c.ProcessWithdraw(ctx)
	case internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PENDING:
		return c.ProcessWithdraw(ctx)
	case internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE:
		submitted, err := c.isSubmitted(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check if deposit is submitted")
		}

		if submitted {
			err = c.dbQ.UpdateStatus(c.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED)
		}

		return &c.deposit, nil
	}

	return &c.deposit, nil
}

func (c *catchupContainer) ProcessWithdraw(ctx context.Context) (*db.Deposit, error) {
	if err := c.dbQ.UpdateStatus(c.deposit.DepositIdentifier,
		internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSING); err != nil {
		return nil, errors.Wrap(err, "failed to update deposit status processing")
	}

	processed, err := c.chainClient.IsProcessed(ctx, c.deposit)
	if err != nil {
		return nil, errors.Wrap(err, "error checking if deposit exists on chain")
	}

	// TODO: Investigate the ways to retrieve the hash of processed tx
	// if deposit is already processed just skip it for now
	if processed {
		err = c.dbQ.UpdateStatus(c.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED)
		if err != nil {
			c.logger.WithError(err).Error("failed to update deposit status to processed")
		}

		return nil, internalTypes.ErrWithdrawalProcessed
	}

	err = executeWithdrawal(ctx, c.chainClient, c.deposit, c.logger)
	if err != nil {

		updateErr := c.dbQ.UpdateStatus(c.deposit.DepositIdentifier,
			internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED)
		if updateErr != nil {
			c.logger.WithError(updateErr).Error("error updating deposit status to failed")
		}

		return nil, errors.Wrap(err, "error checking if deposit exists on chain")
	}

	if err = c.dbQ.UpdateWithdrawalTx(c.deposit.DepositIdentifier, *c.deposit.WithdrawalTxHash); err != nil {

		updateErr := c.dbQ.UpdateStatus(c.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED)
		if updateErr != nil {
			c.logger.WithError(updateErr).Error("failed to update deposit status to FAILED")
		}

		return &c.deposit, errors.Wrap(err, "failed to update deposit withdrawal tx")
	}

	c.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE
	err = c.dbQ.UpdateStatus(c.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE)
	if err != nil {
		return &c.deposit, errors.Wrap(err, "failed to update deposit withdrawal status to submit")
	}

	return &c.deposit, nil

}

func (c *catchupContainer) isSubmitted(ctx context.Context) (bool, error) {
	deposit, err := c.coreConnector.GetDeposit(ctx, c.deposit.DepositIdentifier)
	if err != nil {
		return false, errors.Wrap(err, "error getting deposit from core")
	}

	return deposit.WithdrawalTxHash != nil, nil
}

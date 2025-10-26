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
	dbQ           db.DepositsQ
	deposit       db.Deposit
	chainClient   chain.Client
	coreConnector *connector.Connector

	logger *logan.Entry
}

func NewCatchUpContainer(chainClient chain.Client, deposit db.Deposit, dbQ db.DepositsQ,
	connector *connector.Connector, logger *logan.Entry) WithdrawalContainer {
	return &catchupContainer{
		chainClient:   chainClient,
		deposit:       deposit,
		dbQ:           dbQ,
		coreConnector: connector,
		logger:        logger.WithField("catchup_container", deposit.String()),
	}
}

func (c *catchupContainer) ID() string {
	return c.ID()
}

func (c *catchupContainer) Run(ctx context.Context) (*db.Deposit, error) {
	switch c.deposit.WithdrawalStatus {
	case internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSING:
		return c.Process(ctx)
	case internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PENDING:
		return c.Process(ctx)
	case internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE:
		submitted, err := c.isSubmitted(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check if deposit is submitted")
		}

		if submitted {
			c.logger.Warnf("deposit: %s is already submitted", c.deposit.String())
			return nil, nil
		}

		return &c.deposit, nil
	}

	return &c.deposit, nil
}

func (c *catchupContainer) Process(ctx context.Context) (*db.Deposit, error) {
	processed, err := c.chainClient.IsProcessed(ctx, c.deposit)
	if err != nil {
		return nil, errors.Wrap(err, "error checking if deposit exists on chain")
	}

	// TODO: Investigate the ways to retrieve the hash of processed tx
	// if deposit is already processed just skip it for now
	if processed {
		c.logger.Warnf("deposit %s is already processed", c.deposit.String())
		return nil, nil
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

	return nil, nil
}

func (c *catchupContainer) isSubmitted(ctx context.Context) (bool, error) {
	deposit, err := c.coreConnector.GetDeposit(ctx, c.deposit.DepositIdentifier)
	if err != nil {
		return false, errors.Wrap(err, "error getting deposit from core")
	}

	return deposit.WithdrawalTxHash != nil, nil
}

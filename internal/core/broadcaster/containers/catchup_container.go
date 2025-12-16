package containers

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/connector"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/logan/v3"
)

type catchupContainer struct {
	id string

	dbQ              db.DepositsQ
	deposit          *db.Deposit
	chainClient      chain.ChildClient
	coreConnector    *connector.Connector
	tendermintClient *http.HTTP

	logger *logan.Entry
}

func NewCatchUpContainer(chainClient chain.ChildClient, deposit db.Deposit, dbQ db.DepositsQ,
	connector *connector.Connector, tendermintClient *http.HTTP, logger *logan.Entry) WithdrawalContainer {
	return &catchupContainer{
		id:               deposit.String(),
		chainClient:      chainClient,
		deposit:          &deposit,
		dbQ:              dbQ,
		tendermintClient: tendermintClient,
		coreConnector:    connector,
		logger:           logger.WithField("catchup_container", deposit.String()),
	}
}

func (c *catchupContainer) ID() string {
	return c.id
}

func (c *catchupContainer) Run(ctx context.Context) (*db.Deposit, error) {
	c.logger.Debug("catching up deposit")

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
			c.logger.Warn("withdrawal is already submitted")

			err = c.dbQ.UpdateStatus(c.deposit.DepositIdentifier, internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED)
			if err != nil {
				return nil, errors.Wrap(err, "failed to update deposit status")
			}

			return nil, internalTypes.ErrAlreadySubmitted
		}

		return c.deposit, nil
	}

	return c.deposit, nil
}

func (c *catchupContainer) ProcessWithdraw(ctx context.Context) (*db.Deposit, error) {

	c.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSING
	if err := c.dbQ.UpdateStatus(c.deposit.DepositIdentifier, c.deposit.WithdrawalStatus); err != nil {
		return nil, errors.Wrap(err, "failed to update deposit status processing")
	}

	processed, err := c.chainClient.IsProcessed(ctx, *c.deposit)
	if err != nil {
		c.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED
		updateErr := c.dbQ.UpdateStatus(c.deposit.DepositIdentifier, c.deposit.WithdrawalStatus)
		if updateErr != nil {
			c.logger.WithError(updateErr).Error("error updating deposit status to failed")
		}

		return c.deposit, errors.Wrap(err, "error checking if deposit exists on chain")
	}

	// if deposit is already processed just skip
	if processed {
		c.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED
		err = c.dbQ.UpdateStatus(c.deposit.DepositIdentifier, c.deposit.WithdrawalStatus)
		if err != nil {
			c.logger.WithError(err).Error("failed to update deposit status to processed")
		}

		if c.deposit.WithdrawalTxHash == nil {
			c.deposit.WithdrawalTxHash = ptr(defaultWithdrawalHash)
		}
		err = c.dbQ.UpdateWithdrawalDetails(*c.deposit)
		if err != nil {
			c.logger.WithError(err).Error("failed to update deposit withdrawal details")
		}

		return c.deposit, internalTypes.ErrWithdrawalProcessed
	}

	err = executeWithdrawal(ctx, c.chainClient, c.deposit, c.tendermintClient, c.logger)
	if err != nil {

		c.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED
		updateErr := c.dbQ.UpdateStatus(c.deposit.DepositIdentifier, c.deposit.WithdrawalStatus)
		if updateErr != nil {
			c.logger.WithError(updateErr).Error("error updating deposit status to failed")
		}

		if c.deposit.WithdrawalTxHash == nil {
			c.deposit.WithdrawalTxHash = ptr(defaultWithdrawalHash)
		}

		updateErr = c.dbQ.UpdateWithdrawalDetails(*c.deposit)
		if updateErr != nil {
			c.logger.WithError(updateErr).Error("error updating withdrawal details")
		}

		return c.deposit, errors.Wrap(err, "error processing the withdrawal")
	}

	if err = c.dbQ.UpdateWithdrawalDetails(*c.deposit); err != nil {

		c.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED
		updateErr := c.dbQ.UpdateStatus(c.deposit.DepositIdentifier, c.deposit.WithdrawalStatus)
		if updateErr != nil {
			c.logger.WithError(updateErr).Error("failed to update deposit status to FAILED")
		}

		return c.deposit, errors.Wrap(err, "failed to update deposit withdrawal details")
	}

	c.deposit.WithdrawalStatus = internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE
	err = c.dbQ.UpdateStatus(c.deposit.DepositIdentifier, c.deposit.WithdrawalStatus)
	if err != nil {
		return c.deposit, errors.Wrap(err, "failed to update deposit withdrawal status to submit")
	}

	return c.deposit, nil

}

func (c *catchupContainer) isSubmitted(ctx context.Context) (bool, error) {
	deposit, err := c.coreConnector.GetDeposit(ctx, c.deposit.DepositIdentifier)
	if err != nil {
		return false, errors.Wrap(err, "error getting deposit from core")
	}

	return deposit.WithdrawalTxHash != nil, nil
}

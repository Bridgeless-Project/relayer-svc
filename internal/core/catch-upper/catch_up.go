package catch_upper

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

type CatchUpper struct {
	ctx         context.Context
	broadcaster *broadcaster.Broadcaster

	depositsDb db.DepositsQ
	epochsDb   db.EpochsQ

	logger *logan.Entry
}

func NewCatchUpper(ctx context.Context, broadcaster *broadcaster.Broadcaster, depositsDb db.DepositsQ, epochsDb db.EpochsQ, log *logan.Entry) *CatchUpper {
	return &CatchUpper{
		ctx:         ctx,
		broadcaster: broadcaster,
		depositsDb:  depositsDb,
		epochsDb:    epochsDb,
		logger:      log,
	}
}

func (с *CatchUpper) Start() error {
	// Firstly catch up pending deposits from db
	if err := с.catchupWithStatus(internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PENDING); err != nil {
		return errors.Wrap(err, "catchup with status pending failed")
	}

	if err := с.catchupWithStatus(internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSING); err != nil {
		return errors.Wrap(err, "catchup with status processing failed")
	}

	if err := с.catchupWithStatus(internalTypes.WithdrawalStatus_WITHDRAWAL_STATUS_SUBMITTING_TO_CORE); err != nil {
		return errors.Wrap(err, "catchup with status submitting failed")
	}

	if err := с.catchupEpochsWithStatus(internalTypes.EpochStatus_EPOCH_STATUS_PENDING); err != nil {
		return errors.Wrap(err, "catchup epochs with status pending failed")
	}

	if err := с.catchupEpochsWithStatus(internalTypes.EpochStatus_EPOCH_STATUS_FAILED); err != nil {
		return errors.Wrap(err, "catchup epochs with status failed failed")
	}

	return nil
}

func (c *CatchUpper) catchupWithStatus(status internalTypes.WithdrawalStatus) error {
	deposits, err := c.depositsDb.GetWithStatus(status)
	if err != nil {
		return errors.Wrap(err, "failed to get unprocessed deposits")
	}

	for _, deposit := range deposits {
		err = c.broadcaster.CatchUp(deposit)
		if err != nil {
			c.logger.Errorf("failed to broadcast deposit to catchup deposit: %v", err)
			continue
		}
	}

	return nil
}

func (c *CatchUpper) catchupEpochsWithStatus(status internalTypes.EpochStatus) error {
	epochs, err := c.epochsDb.GetWithStatus(status)
	if err != nil {
		return errors.Wrap(err, "failed to get unprocessed epochs")
	}

	for _, epoch := range epochs {
		err = c.broadcaster.CatchUpEpoch(epoch)
		if err != nil {
			c.logger.Errorf("failed to catchup epoch: %v", err)
			continue
		}
	}

	return nil
}

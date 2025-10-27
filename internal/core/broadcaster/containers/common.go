package containers

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

func executeWithdrawal(ctx context.Context, chainClient chain.Client, deposit db.Deposit, logger *logan.Entry) error {
	var (
		txHash string
		err    error
	)

	switch deposit.WithdrawalToken {
	case core.DefaultNativeTokenAddress:
		txHash, err = chainClient.WithdrawNative(ctx, deposit)
	default:
		txHash, err = chainClient.WithdrawToken(ctx, deposit)
	}
	if err != nil && txHash == "" {
		return errors.Wrap(err, "error processing withdrawal")
	}

	if err != nil {
		logger.WithError(err).Error("error processing withdrawal")
	}

	logger.Infof("Processed deposit %s withdrawal hash %s", deposit.String(), txHash)
	deposit.WithdrawalTxHash = &txHash

	return nil
}

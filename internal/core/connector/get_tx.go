package connector

import (
	"context"

	bridgetypes "github.com/Bridgeless-Project/bridgeless-core/v12/x/bridge/types"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
)

func (c *Connector) GetDeposit(ctx context.Context, depositId db.DepositIdentifier) (*db.Deposit, error) {

	msg := &bridgetypes.QueryTransactionByIdRequest{
		ChainId: depositId.ChainId,
		TxHash:  depositId.TxHash,
		TxNonce: uint64(depositId.TxNonce),
	}

	tx, err := c.querier.TransactionById(ctx, msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find specified transaction")
	}

	return txToDeposit(tx.Transaction), nil
}

func txToDeposit(tx bridgetypes.Transaction) *db.Deposit {

	return &db.Deposit{
		DepositIdentifier: db.DepositIdentifier{
			TxHash:  tx.DepositTxHash,
			TxNonce: int64(tx.DepositTxIndex),
			ChainId: tx.DepositChainId,
		},
		Depositor:         tx.Depositor,
		DepositAmount:     tx.DepositAmount,
		DepositToken:      tx.DepositToken,
		Receiver:          tx.Receiver,
		WithdrawalToken:   tx.WithdrawalToken,
		DepositBlock:      int64(tx.DepositBlock),
		CommissionAmount:  tx.CommissionAmount,
		ReferralId:        uint16(tx.ReferralId),
		WithdrawalStatus:  types.WithdrawalStatus_WITHDRAWAL_STATUS_UNSPECIFIED,
		WithdrawalTxHash:  nilOrNotEmpty(tx.WithdrawalTxHash),
		WithdrawalChainId: tx.WithdrawalChainId,
		WithdrawalAmount:  tx.WithdrawalAmount,
		IsWrappedToken:    tx.IsWrapped,
		Signature:         tx.Signature,
	}
}

func nilOrNotEmpty(str string) *string {
	if str == "" {
		return nil
	}

	return &str
}

package connector

import (
	"context"
	"strings"

	bridgetypes "github.com/Bridgeless-Project/bridgeless-core/v12/x/bridge/types"
	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
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

}

func txToDeposit(tx *bridgetypes.Transaction) *db.Deposit {
	depositor := tx.Depositor
	signature := tx.Signature
	withdrawal

	return &db.Deposit{
		DepositIdentifier: db.DepositIdentifier{
			TxHash:  tx.DepositTxHash,
			TxNonce: int64(tx.DepositTxIndex),
			ChainId: tx.DepositChainId,
		},
		Depositor:         tx.Depositor,
		DepositAmount:     "",
		DepositToken:      "",
		Receiver:          "",
		WithdrawalToken:   "",
		DepositBlock:      0,
		CommissionAmount:  "",
		ReferralId:        0,
		WithdrawalStatus:  0,
		WithdrawalTxHash:  nil,
		WithdrawalChainId: "",
		WithdrawalAmount:  "",
		IsWrappedToken:    false,
		Signature:         nil,
	}
}

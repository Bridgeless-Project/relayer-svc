package connector

import (
	"context"

	bridgeTypes "github.com/Bridgeless-Project/bridgeless-core/v12/x/bridge/types"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

func (c *Connector) UpdateTxInfo(ctx context.Context, deposits []*db.Deposit) error {
	messages := make([]types.Msg, len(deposits))

	for i, d := range deposits {
		depositData, err := c.GetDeposit(ctx, d.DepositIdentifier)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve deposit data")
		}

		// If withdrawal tx hash is already present on Core skip the deposit
		if depositData.WithdrawalTxHash != nil {
			continue
		}

		msg := bridgeTypes.NewMsgUpdateTransaction(c.account.CosmosAddress(), toTransaction(*d))
		messages[i] = msg
	}

	err := c.submitMsgs(ctx, messages...)
	if err != nil {
		return errors.Wrap(err, "failed to update tx info")
	}

	return nil
}

func toTransaction(deposit db.Deposit) bridgeTypes.Transaction {
	return bridgeTypes.Transaction{
		DepositChainId:    deposit.ChainId,
		DepositTxHash:     deposit.TxHash,
		DepositTxIndex:    uint64(deposit.TxNonce),
		DepositBlock:      uint64(deposit.DepositBlock),
		DepositToken:      deposit.DepositToken,
		DepositAmount:     deposit.DepositAmount,
		Depositor:         deposit.Depositor,
		Receiver:          deposit.Receiver,
		WithdrawalChainId: deposit.WithdrawalChainId,
		WithdrawalTxHash:  emptyOrString(deposit.WithdrawalTxHash),
		WithdrawalToken:   deposit.WithdrawalToken,
		Signature:         deposit.Signature,
		IsWrapped:         deposit.IsWrappedToken,
		WithdrawalAmount:  deposit.WithdrawalAmount,
		CommissionAmount:  deposit.CommissionAmount,
		TxData:            deposit.TxData,
		ReferralId:        uint32(deposit.ReferralId),
	}
}

func emptyOrString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

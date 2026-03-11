package common

import (
	apiTypes "github.com/Bridgeless-Project/relayer-svc/internal/api/types"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	database "github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/Bridgeless-Project/relayer-svc/internal/types"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

func ValidateIdentifier(identifier *types.DepositIdentifier) error {
	if identifier == nil {
		return errors.New("identifier is required")
	}

	return validation.Errors{
		"tx_hash":  validation.Validate(identifier.TxHash, validation.Required),
		"chain_id": validation.Validate(identifier.ChainId, validation.Required),
		"tx_nonce": validation.Validate(identifier.TxNonce, validation.Min(0)),
	}.Filter()
}

func ValidateTxHash(identifier *types.DepositIdentifier, client chain.Client) error {
	if !client.TransactionHashValid(identifier.TxHash) {
		return errors.New("invalid transaction hash")
	}

	return nil
}

func ToStatusResponse(d *database.Deposit) *apiTypes.CheckWithdrawalResponse {
	result := &apiTypes.CheckWithdrawalResponse{
		DepositIdentifier: &types.DepositIdentifier{
			TxHash:  d.TxHash,
			TxNonce: d.TxNonce,
			ChainId: d.ChainId,
		},
		WithdrawalStatus: d.WithdrawalStatus,
	}

	result.TransferData = &types.TransferData{
		Sender:           &d.Depositor,
		Receiver:         d.Receiver,
		DepositAmount:    d.DepositAmount,
		WithdrawalAmount: d.WithdrawalAmount,
		CommissionAmount: d.CommissionAmount,
		DepositAsset:     d.DepositToken,
		WithdrawalAsset:  d.WithdrawalToken,
		IsWrappedAsset:   d.IsWrappedToken,
		DepositBlock:     d.DepositBlock,
		Signature:        &d.Signature,
		MerkleProof:      d.MerkleProof,
		TxData:           d.TxData,
	}
	result.WithdrawalIdentifier = &types.WithdrawalIdentifier{
		TxHash:               d.WithdrawalTxHash,
		ChainId:              d.WithdrawalChainId,
		WithdrawalChainBlock: &d.WithdrawalChainBlock,
		WithdrawalCoreBlock:  &d.WithdrawalCoreBlock,
		Operator:             d.Operator,
	}

	return result
}

func ToWithdrawalByStatusResponse(withdrawals []database.Deposit) *apiTypes.GetWithdrawalsByStatusResponse {
	mapped := make([]*apiTypes.CheckWithdrawalResponse, len(withdrawals))
	for i := range withdrawals {
		mapped[i] = ToStatusResponse(&withdrawals[i])
	}
	return &apiTypes.GetWithdrawalsByStatusResponse{
		Withdrawal: mapped,
	}
}

func ToDbIdentifier(identifier *types.DepositIdentifier) database.DepositIdentifier {
	return database.DepositIdentifier{
		TxHash:  identifier.TxHash,
		TxNonce: identifier.TxNonce,
		ChainId: identifier.ChainId,
	}
}

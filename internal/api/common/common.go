package common

import (
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

func ToDbIdentifier(identifier *types.DepositIdentifier) database.DepositIdentifier {
	return database.DepositIdentifier{
		TxHash:  identifier.TxHash,
		TxNonce: identifier.TxNonce,
		ChainId: identifier.ChainId,
	}
}

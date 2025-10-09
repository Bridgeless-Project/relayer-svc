package common

import (
	database "github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/Bridgeless-Project/relayer-svc/internal/types"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	ParamChainId = "chain_id"
	ParamTxHash  = "tx_hash"
	ParamTxNonce = "tx_nonce"
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

func FromDbIdentifier(identifier database.DepositIdentifier) *types.DepositIdentifier {
	return &types.DepositIdentifier{
		TxHash:  identifier.TxHash,
		TxNonce: identifier.TxNonce,
		ChainId: identifier.ChainId,
	}
}

func ProtoJsonMustMarshal(msg proto.Message) []byte {
	raw, _ := protojson.Marshal(msg)
	return raw
}

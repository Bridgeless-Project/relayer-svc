package observer

import (
	"fmt"
	"strconv"

	bridgeTypes "github.com/Bridgeless-Project/bridgeless-core/v12/x/bridge/types"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
)

func parseUpdatedEpochs(attributes []Attribute) (*db.Epoch, error) {
	epoch := &db.Epoch{}

	for _, attribute := range attributes {
		switch attribute.Key {
		case db.AttributeEpochId:
			epochId, err := strconv.ParseUint(attribute.Value, 10, 32)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse epoch id")
			}
			epoch.Id = uint32(epochId)
		case db.AttributeChainId:
			epoch.ChainId = attribute.Value
		case db.AttributeEpochSignature:
			epoch.Signature = attribute.Value
		case db.AttributeEpochSigner:
			epoch.Signer = attribute.Value
		case db.AttributeEpochStartTime:
			startTime, err := strconv.ParseUint(attribute.Value, 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse start time")
			}
			epoch.StartTime = startTime
		case db.AttributeEpochEndTime:
			endTime, err := strconv.ParseUint(attribute.Value, 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse end time")
			}
			epoch.EndTime = endTime
		case db.AttributeEpochNonce:
			epoch.Nonce = attribute.Value
		}
	}

	return epoch, nil
}

func parseSubmittedDeposit(attributes []Attribute) (*db.Deposit, error) {
	deposit := &db.Deposit{}

	for _, attribute := range attributes {

		switch attribute.Key {
		case bridgeTypes.AttributeKeyDepositTxHash:
			deposit.TxHash = attribute.Value
		case bridgeTypes.AttributeKeyDepositNonce:
			n, err := strconv.ParseInt(attribute.Value, 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse deposit nonce")
			}
			deposit.TxNonce = n
		case bridgeTypes.AttributeKeyDepositChainId:
			deposit.ChainId = attribute.Value
		case bridgeTypes.AttributeKeyDepositAmount:
			deposit.DepositAmount = attribute.Value
		case bridgeTypes.AttributeKeyDepositToken:
			deposit.DepositToken = attribute.Value
		case bridgeTypes.AttributeKeyDepositBlock:
			b, err := strconv.ParseInt(attribute.Value, 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse deposit block")
			}
			deposit.DepositBlock = b
		case bridgeTypes.AttributeKeyWithdrawalAmount:
			deposit.WithdrawalAmount = attribute.Value
		case bridgeTypes.AttributeKeyDepositor:
			deposit.Depositor = attribute.Value
		case bridgeTypes.AttributeKeyReceiver:
			deposit.Receiver = attribute.Value
		case bridgeTypes.AttributeKeyWithdrawalChainID:
			deposit.WithdrawalChainId = attribute.Value
		case bridgeTypes.AttributeKeyWithdrawalTxHash:
			if attribute.Value != "" {
				deposit.WithdrawalTxHash = &attribute.Value
			}
		case bridgeTypes.AttributeKeyWithdrawalToken:
			deposit.WithdrawalToken = attribute.Value
		case bridgeTypes.AttributeKeySignature:
			if attribute.Value != "" {
				deposit.Signature = attribute.Value
			}
		case bridgeTypes.AttributeKeyIsWrapped:
			isWrapped, err := strconv.ParseBool(attribute.Value)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse isWrapped attribute")
			}
			deposit.IsWrappedToken = isWrapped
		case bridgeTypes.AttributeKeyCommissionAmount:
			deposit.CommissionAmount = attribute.Value
		case bridgeTypes.AttributeKeyMerkleProof:
			deposit.MerkleProof = attribute.Value
		default:
			return nil, errors.Wrap(errors.New(fmt.Sprintf("unknown attribute key: %s", attribute.Key)), "failed to parse attribute")
		}
	}

	return deposit, nil
}

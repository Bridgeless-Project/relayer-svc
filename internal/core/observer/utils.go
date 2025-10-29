package observer

import (
	"encoding/json"
	"fmt"
	"strconv"

	bridgeTypes "github.com/Bridgeless-Project/bridgeless-core/v12/x/bridge/types"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

func (o *Observer) parseDepositsFromTxResults(txs []*abciTypes.ResponseDeliverTx) ([]*db.Deposit, error) {
	var deposits []*db.Deposit

	for _, tx := range txs {
		var msgs []MsgEvent

		if tx.Log == "" || !json.Valid([]byte(tx.Log)) {
			o.logger.Warnf("skipping invalid tx log: %s", tx.Log)
			continue
		}

		if err := json.Unmarshal([]byte(tx.Log), &msgs); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to unmarshal log: %v", tx.Log))
		}
		for _, msg := range msgs {
			for _, event := range msg.Events {
				if event.Type != eventDepositSubmitted {
					continue
				}

				deposit, err := parseSubmittedDeposit(event.Attributes)
				if err != nil {
					return nil, errors.Wrap(err, "failed to parse deposit")
				}

				deposits = append(deposits, deposit)
			}
		}

	}

	return deposits, nil
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
		default:
			return nil, errors.Wrap(errors.New(fmt.Sprintf("unknown attribute key: %s", attribute.Key)), "failed to parse attribute")
		}
	}

	return deposit, nil
}

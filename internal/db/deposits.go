package db

import (
	"fmt"
	"time"

	"github.com/Bridgeless-Project/relayer-svc/internal/types"
)

type DepositsQ interface {
	New() DepositsQ
	Insert(Deposit) (err error)
	Get(identifier DepositIdentifier) (*Deposit, error)
	GetWithStatus(status types.WithdrawalStatus) ([]Deposit, error)

	UpdateStatus(DepositIdentifier, types.WithdrawalStatus) error
	UpdateWithdrawalDetails(Deposit) error

	UpdateWithdrawalTx(identifier DepositIdentifier, hash string) error

	Transaction(f func() error) error
}

type DepositIdentifier struct {
	TxHash  string `structs:"tx_hash" db:"tx_hash"`
	TxNonce int64  `structs:"tx_nonce" db:"tx_nonce"`
	ChainId string `structs:"chain_id" db:"chain_id"`
}

func (di *DepositIdentifier) String() string {
	return fmt.Sprintf("%s/%s/%d", di.ChainId, di.TxHash, di.TxNonce)
}

type Deposit struct {
	DepositIdentifier

	Depositor            string                 `structs:"depositor" db:"depositor"`
	DepositAmount        string                 `structs:"deposit_amount" db:"deposit_amount"`
	DepositToken         string                 `structs:"deposit_token" db:"deposit_token"`
	Receiver             string                 `structs:"receiver" db:"receiver"`
	DepositBlock         int64                  `structs:"deposit_block" db:"deposit_block"`
	CommissionAmount     string                 `structs:"commission_amount" db:"commission_amount"`
	ReferralId           uint16                 `structs:"referral_id" db:"referral_id"`
	IsWrappedToken       bool                   `structs:"is_wrapped_token" db:"is_wrapped_token"`
	WithdrawalStatus     types.WithdrawalStatus `structs:"withdrawal_status" db:"withdrawal_status"`
	WithdrawalToken      string                 `structs:"withdrawal_token" db:"withdrawal_token"`
	WithdrawalChainBlock int64                  `structs:"withdrawal_chain_block" db:"withdrawal_chain_block"`
	WithdrawalCoreBlock  int64                  `structs:"withdrawal_core_block" db:"withdrawal_core_block"`
	WithdrawalTxHash     *string                `structs:"withdrawal_tx_hash" db:"withdrawal_tx_hash"`
	WithdrawalChainId    string                 `structs:"withdrawal_chain_id" db:"withdrawal_chain_id"`
	WithdrawalAmount     string                 `structs:"withdrawal_amount" db:"withdrawal_amount"`
	MerkleProof          string                 `structs:"merkle_proof" db:"merkle_proof"`

	TxData    string `structs:"tx_data" db:"tx_data"`
	Signature string `structs:"signature" db:"signature"`
	Operator  string `structs:"operator" db:"operator"`

	// Fields for retry logic
	RecoveryAttempts  int       `structs:"recovery_attempts" db:"recovery_attempts"`
	RecoveryTimestamp time.Time `structs:"recovery_timestamp" db:"recovery_timestamp"`
}

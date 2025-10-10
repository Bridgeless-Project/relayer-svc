package db

import (
	"fmt"

	"github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
)

type DepositsQ interface {
	New() DepositsQ
	Insert(Deposit) (id int64, err error)
	Get(identifier DepositIdentifier) (*Deposit, error)

	UpdateWithdrawalTx(DepositIdentifier, string) error
	UpdateSignature(DepositIdentifier, string) error
	UpdateStatus(DepositIdentifier, types.WithdrawalStatus) error

	Transaction(f func() error) error
}

var ErrAlreadySubmitted = errors.New("transaction already submitted")

type DepositIdentifier struct {
	TxHash  string `structs:"tx_hash" db:"tx_hash"`
	TxNonce int64  `structs:"tx_nonce" db:"tx_nonce"`
	ChainId string `structs:"chain_id" db:"chain_id"`
}

func (di *DepositIdentifier) String() string {
	return fmt.Sprintf("%s/%s/%d", di.ChainId, di.TxHash, di.TxNonce)
}

type Deposit struct {
	Id int64 `structs:"-" db:"id"`
	DepositIdentifier

	Depositor        string `structs:"depositor" db:"depositor"`
	DepositAmount    string `structs:"deposit_amount" db:"deposit_amount"`
	DepositToken     string `structs:"deposit_token" db:"deposit_token"`
	Receiver         string `structs:"receiver" db:"receiver"`
	WithdrawalToken  string `structs:"withdrawal_token" db:"withdrawal_token"`
	DepositBlock     int64  `structs:"deposit_block" db:"deposit_block"`
	CommissionAmount string `structs:"commission_amount" db:"commission_amount"`
	ReferralId       uint16 `structs:"referral_id" db:"referral_id"`

	WithdrawalStatus types.WithdrawalStatus `structs:"withdrawal_status" db:"withdrawal_status"`

	WithdrawalTxHash  *string `structs:"withdrawal_tx_hash" db:"withdrawal_tx_hash"`
	WithdrawalChainId string  `structs:"withdrawal_chain_id" db:"withdrawal_chain_id"`
	WithdrawalAmount  string  `structs:"withdrawal_amount" db:"withdrawal_amount"`

	IsWrappedToken bool `structs:"is_wrapped_token" db:"is_wrapped_token"`

	Signature string `structs:"signature" db:"signature"`
}

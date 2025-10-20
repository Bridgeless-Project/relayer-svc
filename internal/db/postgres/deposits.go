package pg

import (
	"database/sql"
	"strings"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	depositsTable   = "deposits"
	depositsTxHash  = "tx_hash"
	depositsTxNonce = "tx_nonce"
	depositsChainId = "chain_id"

	depositsDepositor        = "depositor"
	depositsDepositAmount    = "deposit_amount"
	depositsWithdrawalAmount = "withdrawal_amount"
	depositsDepositToken     = "deposit_token"
	depositsReceiver         = "receiver"
	depositsWithdrawalToken  = "withdrawal_token"
	depositsDepositBlock     = "deposit_block"
	depositsReferralId       = "referral_id"
	depositsTxData           = "tx_data"

	depositsWithdrawalChainId = "withdrawal_chain_id"
	depositsWithdrawalTxHash  = "withdrawal_tx_hash"

	depositsWithdrawalStatus = "withdrawal_status"

	depositsIsWrappedToken   = "is_wrapped_token"
	depositsCommissionAmount = "commission_amount"

	depositsSignature = "signature"
)

type depositsQ struct {
	db       *pgdb.DB
	selector squirrel.SelectBuilder
}

func (d *depositsQ) GetDefault() (*db.Deposit, error) {
	var deposit db.Deposit

	err := d.db.Get(&deposit, d.selector)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &deposit, errors.Wrap(err, "error getting deposit")
}

func (d *depositsQ) FilterById(id string) db.DepositsQ {
	d.selector = d.selector.Where(squirrel.Eq{idField: id})
	return d
}

func (d *depositsQ) GetWithStatus(status types.WithdrawalStatus) ([]db.Deposit, error) {
	stmt := d.selector.Where(squirrel.Eq{
		depositsWithdrawalStatus: status,
	})

	var deposits []db.Deposit
	if err := d.db.Select(&deposits, stmt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to get deposits")
	}

	return deposits, nil
}

func (d *depositsQ) New() db.DepositsQ {
	return NewDepositsQ(d.db.Clone())
}

func (d *depositsQ) Insert(deposit db.Deposit) (int64, error) {
	stmt := squirrel.
		Insert(depositsTable).
		SetMap(map[string]interface{}{
			depositsTxHash:           deposit.TxHash,
			depositsTxNonce:          deposit.TxNonce,
			depositsChainId:          deposit.ChainId,
			depositsWithdrawalStatus: deposit.WithdrawalStatus,
			depositsDepositAmount:    deposit.DepositAmount,
			depositsWithdrawalAmount: deposit.WithdrawalAmount,
			depositsReceiver:         deposit.Receiver,
			depositsDepositBlock:     deposit.DepositBlock,
			depositsIsWrappedToken:   deposit.IsWrappedToken,
			// can be 0x00... in case of native ones
			depositsDepositToken: deposit.DepositToken,
			depositsDepositor:    deposit.Depositor,
			// can be 0x00... in case of native ones
			depositsWithdrawalToken:   deposit.WithdrawalToken,
			depositsSignature:         deposit.Signature,
			depositsWithdrawalChainId: deposit.WithdrawalChainId,
			depositsCommissionAmount:  deposit.CommissionAmount,
			depositsReferralId:        deposit.ReferralId,
			depositsTxData:            deposit.TxData,
		}).
		Suffix("RETURNING id")

	var id int64
	if err := d.db.Get(&id, stmt); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			err = db.ErrAlreadySubmitted
		}

		return id, err
	}

	return id, nil
}

func (d *depositsQ) Get(identifier db.DepositIdentifier) (*db.Deposit, error) {
	var deposit db.Deposit
	err := d.db.Get(&deposit, d.selector.Where(identifierToPredicate(identifier)))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &deposit, err
}

func identifierToPredicate(identifier db.DepositIdentifier) squirrel.Eq {
	return squirrel.Eq{
		depositsTxHash:  identifier.TxHash,
		depositsTxNonce: identifier.TxNonce,
		depositsChainId: identifier.ChainId,
	}
}

func (d *depositsQ) UpdateWithdrawalDetails(identifier db.DepositIdentifier, hash *string, signature *string) error {
	query := squirrel.Update(depositsTable).
		Set(depositsWithdrawalTxHash, hash).
		Set(depositsSignature, signature).
		Set(depositsWithdrawalStatus, types.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED).
		Where(identifierToPredicate(identifier))

	return d.db.Exec(query)
}

func (d *depositsQ) UpdateSignature(identifier db.DepositIdentifier, sig string) error {
	query := squirrel.Update(depositsTable).
		Set(depositsWithdrawalStatus, types.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED).
		Set(depositsSignature, sig).
		Where(identifierToPredicate(identifier))

	return d.db.Exec(query)
}

func (d *depositsQ) UpdateStatus(identifier db.DepositIdentifier, status types.WithdrawalStatus) error {
	query := squirrel.Update(depositsTable).
		Set(depositsWithdrawalStatus, status).
		Where(identifierToPredicate(identifier))

	return d.db.Exec(query)
}

func (d *depositsQ) UpdateWithdrawalTx(identifier db.DepositIdentifier, hash string) error {
	query := squirrel.Update(depositsTable).
		Set(depositsWithdrawalTxHash, hash).
		Set(depositsWithdrawalStatus, types.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED).
		Where(identifierToPredicate(identifier))

	return d.db.Exec(query)
}

func NewDepositsQ(db *pgdb.DB) db.DepositsQ {
	return &depositsQ{
		db:       db.Clone(),
		selector: squirrel.Select("*").From(depositsTable),
	}
}

func (d *depositsQ) Transaction(f func() error) error {
	return d.db.Transaction(f)
}

func (d *depositsQ) InsertProcessedDeposit(deposit db.Deposit) (int64, error) {
	stmt := squirrel.
		Insert(depositsTable).
		SetMap(map[string]interface{}{
			depositsTxHash:           deposit.TxHash,
			depositsTxNonce:          deposit.TxNonce,
			depositsChainId:          deposit.ChainId,
			depositsDepositAmount:    deposit.DepositAmount,
			depositsWithdrawalAmount: deposit.WithdrawalAmount,
			depositsCommissionAmount: deposit.CommissionAmount,
			depositsReceiver:         strings.ToLower(deposit.Receiver),
			depositsDepositBlock:     deposit.DepositBlock,
			depositsIsWrappedToken:   deposit.IsWrappedToken,
			// can be 0x00... in case of native ones
			depositsDepositToken: strings.ToLower(deposit.DepositToken),
			depositsDepositor:    deposit.Depositor,
			// can be 0x00... in case of native ones
			depositsWithdrawalToken:   strings.ToLower(deposit.WithdrawalToken),
			depositsWithdrawalChainId: deposit.WithdrawalChainId,
			depositsWithdrawalTxHash:  deposit.WithdrawalTxHash,
			depositsSignature:         deposit.Signature,
			depositsWithdrawalStatus:  types.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED,
			depositsReferralId:        deposit.ReferralId,
		}).
		Suffix("RETURNING id")

	var id int64
	if err := d.db.Get(&id, stmt); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			err = db.ErrAlreadySubmitted
		}

		return id, err
	}

	return id, nil
}

package pg

import (
	"database/sql"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	depositsTable        = "deposits"
	depositsTxHash       = "tx_hash"
	depositsTxNonce      = "tx_nonce"
	depositsChainId      = "chain_id"
	withdrawalCoreBlock  = "withdrawal_core_block"
	withdrawalChainBlock = "withdrawal_chain_block"

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

	depositsSignature   = "signature"
	depositsMerkleProof = "merkle_proof"
)

type depositsQ struct {
	db       *pgdb.DB
	selector squirrel.SelectBuilder
}

func (d *depositsQ) UpdateWithdrawalCoreBlock(identifier db.DepositIdentifier, i int64) error {
	query := squirrel.Update(depositsTable).
		Set(withdrawalCoreBlock, i).
		Where(identifierToPredicate(identifier))

	return d.db.Exec(query)
}

func (d *depositsQ) UpdateWithdrawalChainBlock(identifier db.DepositIdentifier, i int64) error {
	query := squirrel.Update(depositsTable).
		Set(withdrawalChainBlock, i).
		Where(identifierToPredicate(identifier))

	return d.db.Exec(query)
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

func (d *depositsQ) Insert(deposit db.Deposit) error {
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
			withdrawalCoreBlock:      deposit.WithdrawalCoreBlock,
			withdrawalChainBlock:     deposit.WithdrawalChainBlock,
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
			depositsMerkleProof:       deposit.MerkleProof,
		})

	if err := d.db.Exec(stmt); err != nil {
		return err
	}

	return nil
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

func (d *depositsQ) UpdateWithdrawalDetails(deposit db.Deposit) error {
	query := squirrel.Update(depositsTable).
		Set(depositsWithdrawalTxHash, deposit.WithdrawalTxHash).
		Set(withdrawalChainBlock, deposit.WithdrawalChainBlock).
		Set(withdrawalCoreBlock, deposit.WithdrawalCoreBlock).
		Where(identifierToPredicate(deposit.DepositIdentifier))

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

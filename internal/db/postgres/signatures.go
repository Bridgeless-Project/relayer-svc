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
	signaturesTable     = "signatures"
	epochsId            = "id"
	epochsChainId       = "chain_id"
	epochsSignature     = "signature"
	epochsSigner        = "signer"
	epochsStartTime     = "start_time"
	epochsEndTime       = "end_time"
	epochsNonce         = "nonce"
	epochsSignatureMode = "signature_mode"
	epochsStatus        = "status"
)

type signaturesQ struct {
	db       *pgdb.DB
	selector squirrel.SelectBuilder
}

func NewSignaturesQ(db *pgdb.DB) db.SignaturesQ {
	return &signaturesQ{
		db:       db.Clone(),
		selector: squirrel.Select("*").From(signaturesTable),
	}
}

func (q *signaturesQ) New() db.SignaturesQ {
	return NewSignaturesQ(q.db.Clone())
}

func (q *signaturesQ) Insert(epoch db.Epoch) error {
	stmt := squirrel.
		Insert(signaturesTable).
		SetMap(map[string]interface{}{
			epochsId:            epoch.Id,
			epochsChainId:       epoch.ChainId,
			epochsSignature:     epoch.Signature,
			epochsSigner:        epoch.Signer,
			epochsStartTime:     epoch.StartTime,
			epochsEndTime:       epoch.EndTime,
			epochsNonce:         epoch.Nonce,
			epochsSignatureMode: epoch.SignatureMode,
			epochsStatus:        epoch.Status,
		})

	if err := q.db.Exec(stmt); err != nil {
		return err
	}

	return nil
}

func (q *signaturesQ) Get(identifier db.SignatureIdentifier) (*db.Epoch, error) {
	var epoch db.Epoch
	err := q.db.Get(&epoch, q.selector.Where(epochIdentifierToPredicate(identifier)))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &epoch, err
}

func (q *signaturesQ) GetWithStatus(status types.EpochStatus) ([]db.Epoch, error) {
	stmt := q.selector.Where(squirrel.Eq{
		epochsStatus: status,
	})

	var epochs []db.Epoch
	if err := q.db.Select(&epochs, stmt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to get epochs")
	}

	return epochs, nil
}

func (q *signaturesQ) UpdateStatus(identifier db.SignatureIdentifier, status types.EpochStatus) error {
	query := squirrel.Update(signaturesTable).
		Set(epochsStatus, status).
		Where(epochIdentifierToPredicate(identifier))

	return q.db.Exec(query)
}

func (q *signaturesQ) Transaction(f func() error) error {
	return q.db.Transaction(f)
}

func epochIdentifierToPredicate(identifier db.SignatureIdentifier) squirrel.Eq {
	return squirrel.Eq{
		epochsId:      identifier.Id,
		epochsChainId: identifier.ChainId,
		epochsNonce:   identifier.Nonce,
	}
}
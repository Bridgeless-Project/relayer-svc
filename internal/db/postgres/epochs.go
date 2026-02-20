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
	epochsTable         = "epochs"
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

type epochsQ struct {
	db       *pgdb.DB
	selector squirrel.SelectBuilder
}

func NewEpochsQ(db *pgdb.DB) db.EpochsQ {
	return &epochsQ{
		db:       db.Clone(),
		selector: squirrel.Select("*").From(epochsTable),
	}
}

func (q *epochsQ) New() db.EpochsQ {
	return NewEpochsQ(q.db.Clone())
}

func (q *epochsQ) Insert(epoch db.Epoch) error {
	stmt := squirrel.
		Insert(epochsTable).
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

func (q *epochsQ) Get(identifier db.EpochIdentifier) (*db.Epoch, error) {
	var epoch db.Epoch
	err := q.db.Get(&epoch, q.selector.Where(epochIdentifierToPredicate(identifier)))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &epoch, err
}

func (q *epochsQ) GetWithStatus(status types.EpochStatus) ([]db.Epoch, error) {
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

func (q *epochsQ) UpdateStatus(identifier db.EpochIdentifier, status types.EpochStatus) error {
	query := squirrel.Update(epochsTable).
		Set(epochsStatus, status).
		Where(epochIdentifierToPredicate(identifier))

	return q.db.Exec(query)
}

func (q *epochsQ) Transaction(f func() error) error {
	return q.db.Transaction(f)
}

func epochIdentifierToPredicate(identifier db.EpochIdentifier) squirrel.Eq {
	return squirrel.Eq{
		epochsId:      identifier.Id,
		epochsChainId: identifier.ChainId,
		epochsNonce:   identifier.Nonce,
	}
}
-- +migrate Up

ALTER TABLE deposits ADD COLUMN merkle_proof TEXT DEFAULT '';

-- +migrate Down

ALTER TABLE deposits DROP COLUMN merkle_proof;
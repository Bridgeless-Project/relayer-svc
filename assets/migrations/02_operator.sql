-- +migrate Up

ALTER TABLE deposits ADD COLUMN operator TEXT DEFAULT '';

-- +migrate Down

ALTER TABLE deposits DROP COLUMN operator;
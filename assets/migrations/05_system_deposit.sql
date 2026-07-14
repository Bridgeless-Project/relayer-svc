-- +migrate Up

ALTER TABLE deposits
    ADD COLUMN is_system BOOLEAN NOT NULL DEFAULT FALSE;

-- +migrate Down

ALTER TABLE deposits
    DROP COLUMN is_system;

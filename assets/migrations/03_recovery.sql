-- +migrate Up

ALTER TABLE deposits ADD COLUMN recovery_timestamp TIMESTAMP DEFAULT NOW();
ALTER TABLE deposits ADD COLUMN recovery_attempts INTEGER DEFAULT 0;

-- +migrate Down

ALTER TABLE deposits DROP COLUMN recovery_timestamp;
ALTER TABLE deposits DROP COLUMN recovery_attempts;
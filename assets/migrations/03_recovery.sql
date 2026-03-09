-- +migrate Up

ALTER TABLE deposits ADD COLUMN recoveryTimestamp TIMESTAMP DEFAULT NOW();
ALTER TABLE deposits ADD COLUMN recoveryAttempts INTEGER DEFAULT 0;

-- +migrate Down

ALTER TABLE deposits DROP COLUMN recoveryTimestamp;
ALTER TABLE deposits DROP COLUMN recoveryAttempts;
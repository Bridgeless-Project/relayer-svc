-- +migrate Up
CREATE DOMAIN uint16 AS integer
    CHECK (VALUE BETWEEN 0 AND 65535);

CREATE TABLE IF NOT EXISTS deposits
(
    tx_hash             VARCHAR(100) NOT NULL,
    tx_nonce            BIGINT          NOT NULL,
    chain_id            VARCHAR(50)  NOT NULL,

    depositor           VARCHAR(100) NOT NULL,
    receiver            VARCHAR(100) NOT NULL,
    deposit_amount      TEXT        NOT NULL,
    withdrawal_amount   TEXT       NOT NULL,
    commission_amount   TEXT NOT NULL,
    deposit_token       VARCHAR(100) NOT NULL,
    withdrawal_token    VARCHAR(100) NOT NULL,
    is_wrapped_token    BOOLEAN DEFAULT false,
    deposit_block       BIGINT      NOT NULL,
    withdrawal_core_block BIGINT NOT NULL,
    withdrawal_chain_block BIGINT NOT NULL,
    signature           TEXT NOT NULL,

    withdrawal_status   int          NOT NULL,

    withdrawal_tx_hash  VARCHAR(100),
    tx_data TEXT,
    referral_id uint16 NOT NULL DEFAULT 0,
    withdrawal_chain_id VARCHAR(50) NOT NULL,

    PRIMARY KEY (tx_hash,tx_nonce,chain_id));

CREATE TABLE IF NOT EXISTS latest_block
(
    id INT NOT NULL,
    latest_block_id INT NOT NULL
);

-- +migrate Down

DROP TABLE deposits;
DROP TABLE latest_block;
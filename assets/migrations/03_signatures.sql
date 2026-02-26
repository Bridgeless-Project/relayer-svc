-- +migrate Up

CREATE TABLE IF NOT EXISTS signatures (
  id bigint NOT NULL,
  chain_id text NOT NULL,
  nonce text NOT NULL,
  signature text NOT NULL,
  signer text NOT NULL,
  start_time bigint NOT NULL,
  end_time bigint NOT NULL,
  signature_mode boolean NOT NULL,
  status smallint NOT NULL,

  PRIMARY KEY (chain_id, nonce)
);

-- +migrate Down

DROP TABLE IF EXISTS signatures;


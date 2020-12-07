BEGIN;

-- restore migration #5
ALTER TABLE triggers ADD COLUMN network text NOT NULL DEFAULT 'mainnet';

ALTER TABLE triggers DROP COLUMN network_id;

ALTER TABLE state DROP COLUMN network_id;

DROP TABLE IF EXISTS networks;

COMMIT;
BEGIN;

ALTER TABLE triggers ADD COLUMN network text NOT NULL DEFAULT 'mainnet';

COMMIT;
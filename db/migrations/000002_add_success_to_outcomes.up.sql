BEGIN;

ALTER TABLE outcomes ADD COLUMN success BOOLEAN NOT NULL DEFAULT true;

COMMIT;
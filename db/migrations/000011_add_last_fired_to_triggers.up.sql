BEGIN;

ALTER TABLE triggers ADD COLUMN last_fired timestamp with time zone;

COMMIT;
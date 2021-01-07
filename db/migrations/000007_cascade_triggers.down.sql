BEGIN;

ALTER TABLE actions DROP CONSTRAINT actions_trigger_uuid_fkey,
    ADD CONSTRAINT actions_trigger_uuid_fkey
    FOREIGN KEY (trigger_uuid)
    REFERENCES triggers(uuid);

ALTER TABLE matches DROP CONSTRAINT matches_trigger_uuid_fkey,
    ADD CONSTRAINT matches_trigger_uuid_fkey
    FOREIGN KEY (trigger_uuid)
    REFERENCES triggers(uuid);

ALTER TABLE outcomes DROP CONSTRAINT outcomes_match_uuid_fkey,
     ADD CONSTRAINT outcomes_match_uuid_fkey
     FOREIGN KEY (match_uuid)
     REFERENCES matches(uuid);

COMMIT;
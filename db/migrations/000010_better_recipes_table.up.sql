BEGIN;

-- drop old recipes table from migration #05
ALTER TABLE triggers DROP COLUMN recipe_uuid;
DROP TABLE IF EXISTS recipes;

CREATE TABLE IF NOT EXISTS recipes (
    id serial PRIMARY KEY,
    dapp_name text NOT NULL,
    recipe_name text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE triggers ADD COLUMN recipe_id integer REFERENCES recipes (id);

COMMIT;


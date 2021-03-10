BEGIN;

CREATE TYPE action_type AS ENUM (
    'get_triggers',
    'get_recipes',
    'create_trigger',
    'create_recipe',
    'delete_trigger',
    'delete_recipe',
    'get_matches',
    'get_outcomes',
    'update_trigger'
);

CREATE TABLE IF NOT EXISTS user_stats (
    id serial PRIMARY KEY,
    user_uuid uuid NOT NULL REFERENCES users (uuid),
    created_at timestamp with time zone NOT NULL DEFAULT NOW(),
    action_type action_type NOT NULL
);

COMMIT;
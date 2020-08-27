BEGIN;

CREATE TABLE IF NOT EXISTS public.recipes (
      uuid uuid DEFAULT public.uuid_generate_v4() NOT NULL,
      recipe_id text NOT NULL,
      user_uuid uuid NOT NULL REFERENCES users (uuid),
      created_at timestamp with time zone NOT NULL,
      updated_at timestamp with time zone NOT NULL,
      is_deleted boolean NOT NULL DEFAULT false
);

ALTER TABLE ONLY public.recipes ADD CONSTRAINT recipes_pkey PRIMARY KEY (uuid);

ALTER TABLE triggers ADD COLUMN recipe_uuid uuid REFERENCES recipes (uuid);

COMMIT;
/* USERS */

CREATE TABLE IF NOT EXISTS public.users (
      uuid uuid DEFAULT public.uuid_generate_v4() NOT NULL,
      display_name text NOT NULL,
      email text NOT NULL,
      actions_monthly_cap integer NOT NULL,
      user_type text NOT NULL,
      created_at timestamp with time zone NOT NULL,
      counter_current_month integer NOT NULL,
      avatar text
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (uuid);


/* TRIGGERS */

CREATE TABLE IF NOT EXISTS public.triggers (
    uuid uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    trigger_data jsonb NOT NULL,
    is_active boolean NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    triggered boolean NOT NULL,
    user_uuid uuid NOT NULL REFERENCES users (uuid),
    match_c integer DEFAULT 0 NOT NULL
);

ALTER TABLE ONLY public.triggers
    ADD CONSTRAINT triggers_pkey PRIMARY KEY (uuid);

CREATE INDEX user_index ON public.triggers USING btree (user_uuid);


/* ACTIONS */

CREATE TABLE IF NOT EXISTS public.actions (
    uuid uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    action_data jsonb NOT NULL,
    is_active boolean NOT NULL,
    trigger_uuid uuid NOT NULL REFERENCES public.triggers (uuid),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);

ALTER TABLE ONLY public.actions
    ADD CONSTRAINT actions_pkey PRIMARY KEY (uuid);

/* ANALYTICS */

CREATE TABLE IF NOT EXISTS public.analytics (
    id integer NOT NULL,
    type text NOT NULL,
    no_triggers integer NOT NULL,
    block_no integer NOT NULL,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone NOT NULL,
    duration interval NOT NULL,
    block_time timestamp with time zone NOT NULL
);

CREATE SEQUENCE public.analytics_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.analytics_id_seq OWNED BY public.analytics.id;

ALTER TABLE ONLY public.analytics ALTER COLUMN id SET DEFAULT nextval('public.analytics_id_seq'::regclass);

ALTER TABLE ONLY public.analytics
    ADD CONSTRAINT analytics_pkey PRIMARY KEY (id);


/* MATCHES */

CREATE TABLE IF NOT EXISTS public.matches (
    uuid uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    trigger_uuid uuid NOT NULL REFERENCES triggers (uuid),
    match_data jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL
);

ALTER TABLE ONLY public.matches
    ADD CONSTRAINT matches_pkey PRIMARY KEY (uuid);

CREATE INDEX trigger_index ON public.matches USING btree (trigger_uuid);


/* OUTCOMES */

CREATE TABLE IF NOT EXISTS public.outcomes (
    uuid uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    match_uuid uuid NOT NULL REFERENCES matches (uuid),
    payload_data jsonb NOT NULL,
    outcome_data jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL
);

ALTER TABLE ONLY public.outcomes
    ADD CONSTRAINT outcomes_pkey PRIMARY KEY (uuid);

CREATE INDEX match_index ON public.outcomes USING btree (match_uuid);


/* STATE */

CREATE TABLE IF NOT EXISTS public.state (
    id integer NOT NULL,
    wat_last_block_processed integer,
    wat_date timestamp with time zone,
    wac_last_block_processed integer,
    wac_date timestamp with time zone,
    wae_last_block_processed integer,
    wae_date timestamp with time zone
);

ALTER TABLE ONLY public.state
    ADD CONSTRAINT state_pkey PRIMARY KEY (id);
    
-- populate state
INSERT INTO state (
    id,
    wat_last_block_processed,
    wac_last_block_processed,
    wae_last_block_processed)
VALUES (1, 0, 0, 0);

/* Create SQL Trigger to update match_c */

START TRANSACTION;
CREATE OR REPLACE FUNCTION update_match_count() RETURNS trigger
AS $update_match_count$
BEGIN
UPDATE triggers SET match_c = match_c + 1 WHERE uuid = NEW.trigger_uuid;
RETURN NEW;
END;
$update_match_count$ LANGUAGE plpgsql;
 
CREATE CONSTRAINT TRIGGER t1
   AFTER INSERT ON matches
   DEFERRABLE INITIALLY DEFERRED
   FOR EACH ROW EXECUTE PROCEDURE update_match_count();
   
COMMIT;



BEGIN;

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

COMMIT;
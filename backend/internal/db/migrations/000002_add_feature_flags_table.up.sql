CREATE TABLE IF NOT EXISTS public.feature_flags (
    key character varying(100) NOT NULL,
    value boolean DEFAULT false,
    updated_at timestamp with time zone,
    CONSTRAINT feature_flags_pkey PRIMARY KEY (key)
);

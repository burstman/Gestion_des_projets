CREATE TABLE  IF NOT EXISTS authentification (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL UNIQUE,
    email text NOT NULL UNIQUE,
    password text NOT NULL UNIQUE
);
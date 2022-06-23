CREATE TABLE IF NOT EXISTS tokens (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    hash bytea PRIMARY KEY,
    expiry timestamp(0) with time zone NOT NULL,
    scope text NOT NULL
);
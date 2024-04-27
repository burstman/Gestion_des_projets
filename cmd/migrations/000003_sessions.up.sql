CREATE TABLE IF NOT EXISTS sessions (
    token CHAR(43) PRIMARY KEY,
    data BLOB NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
    );
CREATE INDEX sessions_expiry_idx ON sessions (expiry);
CREATE TABLE IF NOT EXISTS users (
    id          INTEGER NOT NULL PRIMARY KEY,
    tg_id       INTEGER UNIQUE,
    tg_username VARCHAR(255),
    status      VARCHAR(32) NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_tokens (
    user_id    VARCHAR(16) NOT NULL,
    token      VARCHAR(16) NOT NULL UNIQUE,
    created_at DATETIME NOT NULL,
    CONSTRAINT UniqPair UNIQUE (user_id, token)
);

CREATE TABLE IF NOT EXISTS tg_one_time_tokens (
    token   VARCHAR(16) NOT NULL UNIQUE,
    user_id VARCHAR(16),
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_logs (
    user_id    VARCHAR(16) NOT NULL,
    action     VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL
);

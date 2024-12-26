CREATE TABLE IF NOT EXISTS users (
    id         VARCHAR(16)  NOT NULL PRIMARY KEY,
    tg_id      VARCHAR(255) NOT NULL UNIQUE,
    auth_token VARCHAR(16)  UNIQUE
);

CREATE TABLE IF NOT EXISTS one_time_tokens (
    user_id  VARCHAR(16) NOT NULL UNIQUE,
    token    VARCHAR(16) NOT NULL UNIQUE,
    CONSTRAINT UniqPair UNIQUE (user_id, token)
);

CREATE TABLE IF NOT EXISTS audit_logs (
    user_id    VARCHAR(16) NOT NULL,
    action     VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL
);

INSERT INTO users (id, tg_id) VALUES ("user-id-1", "tg-user-id-1");
INSERT INTO one_time_tokens (user_id, token) VALUES ("user-id-1", "random-token");

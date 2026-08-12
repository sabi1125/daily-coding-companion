CREATE TABLE oauth_credentials (
    oauth_id      CHAR(36) NOT NULL PRIMARY KEY,
    user_id       CHAR(36) NOT NULL,
    refresh_token TEXT     NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expiry_at     DATETIME NOT NULL,
    CONSTRAINT fk_oauth_credentials_user
        FOREIGN KEY (user_id) REFERENCES users (user_id)
        ON DELETE CASCADE,
    UNIQUE KEY uq_oauth_credentials_user_id (user_id)
);

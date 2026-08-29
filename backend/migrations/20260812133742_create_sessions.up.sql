CREATE TABLE sessions (
    session_id CHAR(36)  NOT NULL PRIMARY KEY,
    user_id    CHAR(36)  NOT NULL,
    created_at DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME  NOT NULL,
    CONSTRAINT fk_sessions_user
        FOREIGN KEY (user_id) REFERENCES users (user_id)
        ON DELETE CASCADE
);

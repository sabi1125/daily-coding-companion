CREATE TABLE settings (
    setting_id           CHAR(36) NOT NULL PRIMARY KEY,
    user_id              CHAR(36) NOT NULL,
    get_help_preferences TEXT     NULL,
    CONSTRAINT fk_settings_user
        FOREIGN KEY (user_id) REFERENCES users (user_id)
        ON DELETE CASCADE,
    UNIQUE KEY uq_settings_user_id (user_id)
);

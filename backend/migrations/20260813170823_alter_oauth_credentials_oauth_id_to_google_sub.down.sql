ALTER TABLE oauth_credentials
    MODIFY COLUMN oauth_id CHAR(36) NOT NULL;

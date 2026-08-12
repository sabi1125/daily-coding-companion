CREATE TABLE users (
    user_id    CHAR(36)     NOT NULL PRIMARY KEY,
    email      VARCHAR(255) NOT NULL,
    firstname  VARCHAR(60)  NOT NULL,
    lastname   VARCHAR(60)  NOT NULL
);

CREATE TABLE problems (
    problem_id         CHAR(36)  NOT NULL PRIMARY KEY,
    user_id            CHAR(36)  NOT NULL,
    raw_problem        TEXT      NOT NULL,
    title              TEXT      NULL,
    problem_text       TEXT      NULL,
    algorithm_tag      TEXT      NULL,
    difficulty         TEXT      NULL,
    ai_help            TEXT      NULL,
    needs_review_flag  BOOLEAN   NOT NULL DEFAULT FALSE,
    created_at         DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_problems_user
        FOREIGN KEY (user_id) REFERENCES users (user_id)
        ON DELETE CASCADE
);

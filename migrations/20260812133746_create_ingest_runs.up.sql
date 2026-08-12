CREATE TABLE ingest_runs (
    ingest_run_id  CHAR(36)    NOT NULL PRIMARY KEY,
    user_id        CHAR(36)    NOT NULL,
    problem_id     CHAR(36)    NULL,
    status         VARCHAR(50) NOT NULL,
    error          VARCHAR(255) NULL,
    retried        BOOLEAN     NOT NULL DEFAULT FALSE,
    ingest_date    DATE        NOT NULL,
    created_at     DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_ingest_runs_user
        FOREIGN KEY (user_id) REFERENCES users (user_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_ingest_runs_problem
        FOREIGN KEY (problem_id) REFERENCES problems (problem_id)
        ON DELETE SET NULL,
    UNIQUE KEY uq_ingest_runs_user_date_retried (user_id, ingest_date, retried)
);

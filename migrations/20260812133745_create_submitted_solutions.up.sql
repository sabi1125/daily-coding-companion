CREATE TABLE submitted_solutions (
    solution_id   CHAR(36)    NOT NULL PRIMARY KEY,
    problem_id    CHAR(36)    NOT NULL,
    solution      TEXT        NOT NULL,
    status        VARCHAR(50) NOT NULL,
    submitted_at  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_submitted_solutions_problem
        FOREIGN KEY (problem_id) REFERENCES problems (problem_id)
        ON DELETE CASCADE
);

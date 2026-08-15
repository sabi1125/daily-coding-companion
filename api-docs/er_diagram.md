# Daily coding problems ER Diagram

```mermaid
erDiagram
users {
    user_id uuid PK
    email varchar(255) "not null"
    firstname varchar(60) "not null"
    lastname varchar(60) "not null"
}

settings {
    setting_id uuid pk
    user_id uuid fk "not null"
    get_help_preferences text "nullable"
}

sessions {
    session_id uuid PK
    user_id uuid FK
    created_at datetime
    expires_at datetime
}

oauth_credentials {
    oauth_id varchar(255) PK
    user_id uuid FK
    refresh_token text "not null"
    created_at datetime
    updated_at datetime
    expiry_at datetime
}

problems {
    problem_id uuid pk
    user_id uuid fk
    raw_problem text "not null"
    title text "nullable"
    problem_text text "nullable"
    algorithm_tag text "nullable"
    difficulty text "nullable"
    ai_help text "nullable"
    needs_review_flag boolean
    created_at datetime
    updated_at datetime
}

submitted_solutions {
    solution_id uuid pk
    problem_id uuid fk
    solution text "not null"
    status string "not null"
    submitted_at datetime
}

ingest_runs {
    ingest_run_id uuid pk
    user_id uuid fk "not null"
    problem_id uuid fk "nullable"
    status string "not null"
    error string "nullable"
    retried boolean "not null, true if this row is the retry attempt"
    ingest_date date "not null, unique with (user_id, retried) — see note below"
    created_at datetime
}

users ||--o{ sessions : "has none or many"
users ||--|| oauth_credentials : "has"
users ||--|{ problems: "has"
users ||--|| settings: "has"
users ||--o{ ingest_runs : "has none or many"
ingest_runs }|--o| problems : "has none or many"
problems ||--o{ submitted_solutions: "has none or many"
```

**`ingest_runs` unique constraint:** `(user_id, ingest_date, retried)` is unique. Prevents
two concurrent `/problems/today` requests (two tabs, a reload race) from both invoking
ingest at once — the second insert fails at the DB level instead of silently duplicating
the fetch or the `problems` row. Still allows the legitimate pair of rows per user per
day: one `retried = false` (cron) and one `retried = true` (the retry).

**`oauth_credentials.oauth_id`:** deliberately *not* a self-generated uuid like every other
PK in this schema — it's Google's own `sub` claim from the verified `id_token` (OIDC's
stable, permanent per-account identifier), stored as `varchar(255)` to match what the spec
actually guarantees (a string, not necessarily a fixed-width integer). Existence checks on
sign-in go through this table by `oauth_id`, not through `users` by email — email can
change, `sub` can't. A first-time sign-in creates both the `users` row and this
`oauth_credentials` row together; a returning user's `oauth_id` lookup here is what finds
their existing `users.user_id` via the FK.

# Ingest job

Not an API endpoint — background job, no request/response, no auth model. Triggered two
ways, described below.

Both triggers end up calling the same function, `ingest.RunForUser(user_id, retried)` —
no duplicated fetch/parse/write logic, just a different caller and a different loop (or no
loop) around it.

**Trigger 1 — cron**

1. Fires once daily (schedule below), runs as its own subcommand on the same binary as the
   web server: `./backend ingest` vs `./backend serve` — cron can't reach into the
   already-running server process, Railway starts a fresh one on schedule.
2. On start, queries `oauth_credentials` for every `user_id` — that's the full list of
   connected users.
3. For each `user_id` in that list, calls `ingest.RunForUser(user_id, retried = false)`.

So the loop is over `user_id`s fetched from `oauth_credentials` first, then one
`RunForUser` call per `user_id` — the function itself only ever handles one user per call,
cron is what repeats it.

**Trigger 2 — retry**

`/problems/today` already has the calling user's `user_id` from their session — no
enumeration needed. It calls `ingest.RunForUser(user_id, retried = true)` once, in-process,
as a plain function call (no subprocess, unlike cron).

**Schedule** — 2am JST (17:00 UTC previous day). Checked the actual Daily Coding Problem
emails, they land between 12am-1am JST, so 2am gives an hour of buffer.

**Behavior**

1. Check for an existing `ingest_runs` row for `(user_id, ingest_date, retried)`.
    - If one already exists, this is a duplicate invocation — no-op, exit (cron skips to
      the next user per step 6, doesn't touch users who already have a row today).
2. Get `refresh_token` from `oauth_credentials`, exchange for a Gmail access token.
    - If invalid/revoked (7-day testing-mode expiry, or user revoked access), log the
      error and exit to step 5 as a failure (cron then moves to the next user, per step 6).
      User needs to reconnect Gmail before ingest works for them again.
3. Query Gmail for today's Daily Coding Problem email.
    - If unreachable/rate-limited, or no email found, exit to step 5 as a failure.
    - If more than one match, take the most recently received — not a failure.
4. Parse the email with Claude into `title` / `problem_text` / `algorithm_tag` /
   `difficulty`.
    - If parsing errors or `title`/`problem_text` come back missing, that's **not** a
      failure — set `needs_review_flag = true`, keep going. `raw_problem` is always
      stored regardless.
5. Write `problems` + `ingest_runs` in one transaction — both or neither.
    - Failure (step 2/3): `ingest_runs` only, `status = failed`, `error` set, no
      `problems` row.
    - Success (step 4, flagged or not): `problems` row created, `ingest_runs.status =
      success`, `problem_id` set.
    - If even the `ingest_runs` write fails (DB down, transaction can't commit), this user
      ends up with no row at all — falls to the lazy retry same as an unattempted user.
6. (Cron only) one user's outcome doesn't stop the loop — move to the next
   `oauth_credentials` row. Same applies if the whole job process crashes mid-loop: users
   not yet reached that run get no row, and pick up via the lazy retry on next login.

`ingest_runs.status` is only ever `success` or `failed`. A Claude parse failure is still
`success` — the row got created, it's just flagged. Only a Gmail/auth failure is `failed`.

**Idempotency**
- Same `(user_id, ingest_date, retried)` twice (cron re-triggered, or the race between two
  concurrent `/problems/today` calls) — the unique constraint on `ingest_runs`
  (`er_diagram.md`) rejects the second write. Cron just moves on. The retry path re-reads
  for a problem instead of erroring (see `api-docs.md`).
- One `retried = false` row (cron) and one `retried = true` row (the retry) coexisting per
  user per day is expected, not a collision.

Ingest job runs at 2am JST every day.

**Errors** — every way `RunForUser` can end with `ingest_runs.status = "failed"`, same
Expected/Operational/Unexpected categories `api-template.md` uses for HTTP endpoints (no
status code here, this isn't a response):

| Category | When | `ingest_runs.error` |
|---|---|---|
| Operational | Refresh token invalid/revoked (7-day testing-mode expiry, or user revoked access) | `"refresh token invalid"` |
| Operational | Gmail API unreachable or rate-limited | `"gmail unreachable"` |
| Expected | No Daily Coding Problem email found for today | `"no email found"` |
| Unexpected | `problems`/`ingest_runs` write fails (DB down, transaction can't commit) | `"write failed"` |

A Claude parse failure is **not** in this list — it doesn't produce `status = "failed"`,
see step 4.

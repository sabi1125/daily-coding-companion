# Design — Daily Coding Companion

> How the app is actually built, satisfying `docs/REQUIREMENTS.md`. The structural pieces
> (data model, lifecycles, moving parts) are their own diagrams — this doc ties them
> together, summarizes the screen/API surface, and covers the behavior and mechanics that
> don't fit into a diagram.

## Structural design (diagrams)
- [`api-docs/er_diagram.md`](../api-docs/er_diagram.md) — the data model
- [`api-docs/state.md`](../api-docs/state.md) — problem, auth/session, and ingest-run
  lifecycles
- [`api-docs/architectural-diagram.md`](../api-docs/architectural-diagram.md) — the moving
  parts and how they talk

---

## Screens

Summary only — detailed wireframes/mockups are their own separate deliverable.

| Route | Screen | What it shows |
|---|---|---|
| `/` | Today's Problem | Today's parsed problem, an attempt-submission form (solution/notes + self-reported solved-or-not), "Get Help" button. |
| `/history` | History | List of past problems, filterable by status (`Open` / `Failed` / `Solved`). |
| `/settings` | Settings | Gmail connection status, "Reconnect Gmail," sign out, basic stats, Get Help preferences (free text). |

Sign-in is a conditional state at `/`, not its own route.

Problem detail is a popup, not a route — opened from a click on a row in History, closed
back to History. Same content/behavior for any past problem (full problem, full attempt
history, Get Help), just not addressable by its own URL.

## API surface

Summary only — full request/response/error specs are their own separate deliverable, one
file per resource in `api-docs/`.

| Resource | Endpoints (rough) |
|---|---|
| Auth | sign in, OAuth callback, sign out |
| Problems | get today's problem, get history, get one by id |
| Submissions | submit an attempt against a problem |
| Settings | get settings, update settings (incl. Get Help preferences) |
| Ingest | internal — cron-triggered fetch, retry check on page load |

## Backend structure

Go, `controller → interactor → repository`, dependency inversion — each layer depends on
an interface (`inputport`), not the layer below's concrete type.

```
cmd/backend/main.go
internal/
  controller/            HTTP handlers (Echo)
  domain/
    entities/             plain structs
    interactor/            business logic
      inputport/             interface the controller depends on
        mock/                   generated mocks (go.uber.org/mock)
    repository/            GORM data access
      inputport/             interface the interactor depends on
        mock/
  infrastructure/        database.go (GORM + MySQL connection), transaction.go
  tx/                    transaction Manager interface — kept separate from
                         infrastructure so the domain layer never imports it directly
  log/                   structured logging (zap)
  response/              shared HTTP response helpers
  util/
```

Each resource (Problems, Submissions, Settings, Ingest) is one consistent set of files
across `controller`/`domain` — no per-resource layout decisions made ad hoc.

Frontend is Vite + React + TypeScript (see `STACK.md`) — standard layout, no separate
structure doc needed.

---

## Problem status

No stored status column. A problem's status (`Open` / `Failed` / `Solved`) is derived at
read time from its submissions — see `state.md`.

Grading is self-report — the user declares whether a submission solved the problem as part
of submitting it, not a separate step, and not automated/AI-checked. See `REQUIREMENTS.md`
Non-goals for why (no sandboxed execution engine, no test cases available from the source
emails — an automated checker is explicitly deferred, not rejected forever).

## Ingest

**Idempotent.** Running the ingest job twice for the same day never creates a duplicate
`problems` row — a second run against a day that already succeeded is a no-op.

**Parse-quality flag, set deterministically.** `needs_review_flag` is `true` when parsing
throws an error or required fields (`title`, `problem_text`) come back missing — no AI
confidence score involved. It's a badge, not a blocker: `raw_problem` is always stored and
readable regardless, so a flagged problem can still be attempted and solved normally.

**Retry is lazy, not scheduled.** If the day's ingest fails, there's no background retry
loop — the check happens when the user next opens the app, and it retries exactly once
then. If that retry also fails, the day is given up on for good — see `state.md`'s
ingest-run diagram for the full shape.

## Get Help

Single button, single AI call, single cached response. Clicking "Get Help" generates one
structured response (nudge → approach → walkthrough → full solution, all sections in one
response) — never a separate call per hint level. Every response also explains the
underlying concept/technique, not just the specific problem — that's a fixed part of the
base prompt, not optional. Nothing is generated unless the button is clicked, and once
generated it's cached on the `problems` row — reopening the problem later costs nothing.

**Preferences are additive, not a replacement.** A user can save free-text Get Help
preferences in Settings (e.g. "explain more simply," "skip the walkthrough") — appended to
the base prompt on every request, on top of it. They can steer tone/depth/format; they
can't remove the required concept explanation or the fixed response structure above.

## Streak

Computed from submission timestamps, not from a problem's overall status. A problem shows
`Solved` regardless of *when* it was solved, but the streak only counts a day if a solving
submission's timestamp actually falls on that problem's due date. Solving a problem late
doesn't retroactively fix an already-broken streak day.

A genuinely failed ingest (no `problems` row ever created that day — see `state.md`) does
**not** break the streak; a problem that existed but was never solved does.

## Auth

Sign-in is currently restricted to an allowlist of approved users (Google OAuth Testing
mode) — see D1 and D5 in `DECISIONS.md`. The login screen states this explicitly so a
visitor who can't sign in sees why, rather than hitting a dead end.

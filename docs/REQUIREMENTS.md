# Requirements — Daily Coding Companion

## Overview
A web app that turns the daily "Daily Coding Problem" email into a practice habit —
fetch the day's problem automatically, work it, get AI help if stuck, and track
progress over time.

## Goals
The daily loop works end to end without manual intervention: a problem is there every
morning, can be attempted and marked solved, AI help is available on demand, and
unsolved problems stay visible until dealt with. The app is deployed and usable daily,
not just running locally.

## Requirements

### Must-have

**Auth**
- User can sign in with Google, which also grants read-only access to their Gmail.
- User can connect and disconnect their Gmail account.
- The login screen states that sign-in is currently limited to approved users (see D5 in
  `DECISIONS.md`) — a visitor who can't sign in sees why, not just a dead end.

**Getting the daily problem**
- The day's Daily Coding Problem email is fetched automatically, without manual action.
- The raw email is turned into a structured problem (title, problem text, algorithm tag,
  difficulty).
- If a fetch fails to produce a usable problem, it's retried automatically the next time
  the user opens the app — never silently dropped.
- User can see whether the day's fetch succeeded, is still pending, or failed.

**Working a problem**
- User can view today's problem, and a history of every past problem — every problem
  stays viewable permanently, regardless of status. Nothing is ever removed from history.
- User submits an attempt (notes and/or a pasted solution), indicating whether it solved
  the problem — solving and submitting aren't separate steps.
- A problem is `Solved` once a submission solves it. Otherwise, at the end of its day it's
  fixed as either `Attempted` (at least one submission that day, none solved it) or
  `Untouched` (no submission at all that day) — whichever applies at that moment, set once
  and not re-evaluated afterward, even if a later submission on that problem fails again.
- Unsolved (`Attempted`/`Untouched`) problems don't need any extra action to stay
  reachable — they're already in history like everything else, and can be submitted
  against again at any time until solved.

**AI help**
- User can request AI-generated help on a problem they're stuck on.
- Help is only generated when requested, and only once per problem — re-opening it later
  doesn't regenerate it.
- Every Get Help response explains the underlying concept/technique, not just this
  specific problem — that's a fixed part of every response, not a separate optional
  feature you turn on. (Get Help itself is still only ever triggered by the user.)

**Tracking**
- User can see basic stats — e.g. how many problems are solved, how many are still
  unsolved.

### Nice-to-have

- Streak / calendar heatmap — daily activity view.
- Filter / group problems by algorithm tag.
- Weak-area stats — e.g. performance broken down by topic.
- Complexity analysis of a submitted solution.
- Daily reminder when the day's problem is ready.

### Later / out of scope for now

- Bookmark / favorite, search, free-text tags.
- "Practice more like this" — related problems.
- AI usage / cost dashboard.
- Configurable settings — sender pattern, schedule time.
- Dark mode.
- Automated code-checking (run a submitted solution against test cases for a real
  pass/fail verdict).
- Support for other email providers (Hotmail/Outlook, etc.) beyond Gmail — still only
  ever Daily Coding Problem content, just not limited to a Gmail inbox.

## Non-goals
- Automated code execution/grading — self-report (solved / not solved) is the mechanism,
  not a sandboxed test runner.
- Multi-user support beyond a single account — no other user's mailbox is ever ingested
  or shown.
- Monetization or paid features.
- General-purpose email tracking (any newsletter, any content type) — scoped to Daily
  Coding Problem only.

## Glossary
- **Problem** — one day's Daily Coding Problem, stored as a single `problems` row. Not a
  bug, not an issue — always refers to the coding problem entity.
- **Attempt** — one submission (notes and/or a pasted solution) against a problem. A
  problem can have many attempts over time.
- **Attempted / Untouched / Solved** — the three states a problem's status resolves to
  (see "Working a problem" above). Not stored — derived from attempts and the problem's
  date. The full state diagram (including the transient pre-day-end state) will live in
  the design doc once it's added to the repo.
- **Ingest / ingest run** — the process that fetches the day's email and turns it into a
  problem. One ingest run per fetch attempt, not per day (a failed attempt can retry).
- **Get Help** — the specific one-click, one-AI-call assistance feature. Capitalized as a
  feature name, not generic help/support.
- **Streak** — consecutive days with at least one problem solved. Broken by a day with a
  problem that went `Untouched`; not broken by a genuinely failed ingest — full reasoning
  will live in the design doc once it's added to the repo.
- **Session (auth)** — the signed-in/signed-out state of the Google connection. Not a
  browsing/HTTP session.
- **needs_review_flag** — a badge on a problem meaning the parse was low-confidence. Not a
  workflow state, doesn't block attempting the problem.


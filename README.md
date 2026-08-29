# Daily Coding Companion

![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)
![Status](https://img.shields.io/badge/status-v0.0.1%20MVP-brightgreen.svg)

Subscribed to daily coding problems but don't try most of it and you're not smart enough to
solve it in your head? Give this a try.

**Live at [dailycodingcompanion.quest](https://dailycodingcompanion.quest)** — sign-in is
currently limited to approved test users (see [Prerequisite](#prerequisite) and
[Access](#access) below).

## What it is

A small web app that, every morning, pulls your **Daily Coding Problem**
(dailycodingproblem.com) email, shows it to you in a clean interface, lets you actually run
your solution against a real sandboxed execution engine, and — if you get stuck — gives you
AI-generated help instead of just handing you the answer.

Those emails usually just pile up unread. This turns them into an actual daily practice
habit: see the problem → attempt it → run it, get help if stuck → track progress → revisit
the ones you didn't get.

## Prerequisite

You must already be subscribed to the Daily Coding Problem newsletter at
dailycodingproblem.com — this app only reads and organizes emails that are already arriving
in your inbox, it doesn't source problems on its own. Without an active subscription there's
nothing for it to pull, and the app has no way to function.

## How it works

1. Sign in with Google once — grants read-only Gmail access.
2. A scheduled job checks your inbox each morning for the day's email and parses it into a
   clean problem (title, problem text, algorithm tag, difficulty). If no email shows up that
   day (or it doesn't actually contain a problem), Claude invents an original one in the same
   style instead, so there's still something to solve — flagged in the UI so you can tell it
   apart from a real one.
3. You open the app and see today's problem, plus your history.
4. Write your solution and hit **Run** — executes against a real sandboxed engine (Python,
   JavaScript, Go, C++) and shows stdout/stderr/timing, same as running it locally.
5. Still stuck? One click on **Get Help** gets a single AI-generated response covering a
   nudge, an approach, a walkthrough, and the full solution — generated once, cached from
   then on.
6. Mark where you landed — solved, stuck, or untouched. Anything left unsolved when the day
   ends sits in an unsolved view until you come back and finish it.

## Access

Google's OAuth app is in Testing mode (see [`docs/DECISIONS.md`](docs/DECISIONS.md) D5) —
sign-in is limited to an approved allowlist while this stays a personal project, not because
of anything wrong with your account.

## Stack

- **Backend:** Go + Echo, MySQL, GORM, Gmail API, Claude API, Piston (self-hosted code execution)
- **Frontend:** React + TypeScript, Vite, Tailwind, shadcn/ui
- **Hosting:** Railway (backend, frontend, database, scheduled job) + a small self-hosted
  instance for code execution (see D6 in Decisions)

## More

[`docs/DECISIONS.md`](docs/DECISIONS.md) — the engineering decisions behind the stack, the
auth approach, and cost.

## License

[MIT](LICENSE)

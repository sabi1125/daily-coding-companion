# Daily Coding Companion

Subscribed to daily coding problems but don't try most of it and you're not smart enough to
solve it in your head? Give this a try.

## What it is

A small web app that, every morning, pulls your **Daily Coding Problem**
(dailycodingproblem.com) email, shows it to you in a clean interface, and — if you get
stuck — gives you AI-generated help instead of just handing you the answer.

Those emails usually just pile up unread. This turns them into an actual daily practice
habit: see the problem → attempt it → get help if stuck → track progress → revisit the ones
you didn't get.

## Prerequisite

You must already be subscribed to the Daily Coding Problem newsletter at
dailycodingproblem.com — this app only reads and organizes emails that are already arriving
in your inbox, it doesn't source problems on its own. Without an active subscription there's
nothing for it to pull, and the app has no way to function.

## How it works

1. Sign in with Google once — grants read-only Gmail access.
2. A scheduled job checks your inbox each morning for the day's email and parses it into a
   clean problem (title, problem text, algorithm tag, difficulty).
3. You open the app and see today's problem, plus your history.
4. Attempt it. Stuck? One click on **Get Help** gets a single AI-generated response covering
   a nudge, an approach, a walkthrough, and the full solution — generated once, cached from
   then on.
5. Mark where you landed — solved, stuck, or untouched. Anything left unsolved when the day
   ends sits in an unsolved view until you come back and finish it.

## Stack

- **Backend:** Go + Echo, MySQL, GORM, Gmail API, Claude API
- **Frontend:** React + TypeScript, Vite, Tailwind, shadcn/ui
- **Hosting:** Railway (backend, frontend, database, scheduled job)

## More

[`docs/DECISIONS.md`](docs/DECISIONS.md) — the engineering decisions behind the stack, the
auth approach, and cost.

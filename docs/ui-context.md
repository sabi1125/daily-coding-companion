# UI Context — Daily Coding Companion

> Briefing document for designing the 4 screens (issue #7). Every state, element, and piece
> of copy below is pulled from `REQUIREMENTS.md`, `design.md`, `api-docs/state.md`,
> `api-docs/er_diagram.md`, and `DECISIONS.md` — nothing here is invented, it's the existing
> decisions made concrete per screen. Source doc noted in parens where it matters.

## Design direction

**Aesthetic, minimalist.** Not busy, not dashboard-dense — one problem, one action at a time.
Stack is React + TypeScript + Tailwind + shadcn/ui (`STACK.md`) — favor shadcn's default
visual language over custom components; whitespace over borders/dividers; typography and
spacing doing the hierarchy work rather than color/decoration.

## Global states that touch every screen

- **Signed-out** — conditional state at `/`, not its own route. Login screen states sign-in
  is currently limited to approved users (D5) — a visitor who can't sign in sees why, not a
  dead end.
- **NeedsReAuth** — distinct from signed-out. Surfaces via a dedicated reconnect prompt, not
  a banner (D1) — explains what expired (~7-day Google Testing-mode token), why, and the
  "Reconnect Gmail" action. Also conditional at `/`, not its own route.

---

## Screen 1 — Today's Problem (`/`)

Primary screen, default landing point once signed in.

**States:**
- *Signed-out* — see Global states above.
- *NeedsReAuth* — see Global states above.
- *Ingest pending* — user can see the day's fetch is "still pending" (`REQUIREMENTS.md`,
  Getting the daily problem) — shown briefly on load before ingest/retry resolves.
- *Ingest failed, retry pending* — fetch failed once; retries automatically next time the
  user opens the app (`state.md` ingest diagram). Message should not imply the user needs to
  do anything — it's automatic.
- *Ingest failed permanently* — retry also failed, no `problems` row exists for today
  (`state.md`: `RetryIngest --> [*]`). Empty/error state — "no problem today," not a broken
  page.
- *Problem loaded* — normal case. Shows:
  - Title, difficulty, algorithm tag, problem text (`er_diagram.md: problems`)
  - `needs_review_flag` badge if true — badge only, never blocks interaction
    (`design.md`: Ingest)
  - Current derived status (`Open` / `Failed` / `Solved`, `state.md`) — relevant because a
    `Failed` problem can be resubmitted from this same screen
  - Attempt-submission form: notes text area + solution field as a syntax-highlighted code
    editor (`@uiw/react-codemirror`, `STACK.md`) with auto-detected language — plus
    self-reported solved-or-not control, submitted together as one action, not two steps
    (`REQUIREMENTS.md`, Working a problem)
  - "Get Help" button

**Get Help interaction (issue #7 calls this out specifically):**
- Single button. First click generates one structured response (nudge → approach →
  walkthrough → full solution, all in one) — never a separate call per hint level
  (`design.md`: Get Help).
- Once generated, cached — reopening costs nothing, button reflects "already generated"
  state rather than re-triggering a call.
- Response always includes the underlying concept/technique explanation — not a toggle, not
  optional (`REQUIREMENTS.md`, AI help).

## Screen 2 — History (`/history`)

- Filter control by status: `Open` / `Failed` / `Solved` (`design.md`: Screens).
- List items show: title, date, status, `needs_review_flag` badge if applicable.
- Empty state: no problems ingested yet (first-ever use, before the first ingest run).
- Every problem stays visible permanently regardless of status — nothing is ever removed
  (`REQUIREMENTS.md`, Working a problem) — so this list only grows, never needs an "archived"
  distinction.
- Clicking a row navigates to Screen 3 (Problem detail) — see below.

## Screen 3 — Problem detail (`/history/:id`)

Opened from a click on a History row, a "Back to History" link returns to `/history`. A
full page, not a popup — a popup made writing a full solution attempt cramped/awkward.
Addressable by its own URL — same content reused for today's problem and any past problem
(`design.md`: Screens).

- Full problem content (same fields as Screen 1).
- Full attempt history for this problem — every submission (solution, self-reported
  outcome, timestamp), not just the most recent (`REQUIREMENTS.md`: Attempt = "one submission
  ... a problem can have many attempts over time"). Past solutions render read-only — plain
  text is fine, syntax highlighting isn't required since nothing is being edited.
- Get Help button / cached response, same behavior as Screen 1.
- Can be submitted against again at any time until solved — no extra action needed to
  "reopen" it (`REQUIREMENTS.md`, Working a problem).
- `needs_review_flag` badge if applicable.

## Screen 4 — Settings (`/settings`)

- Gmail connection status — reflects `SignedIn` / `NeedsReAuth` / `SignedOut`
  (`api-docs/state.md`, Auth/Session diagram).
- "Reconnect Gmail" button (dedicated action, ties to NeedsReAuth global state above).
- Sign out button.
- Basic stats — count solved, count unsolved (`REQUIREMENTS.md`, Tracking).
- **Get Help preferences** — free-text field, saved once, appended to every future Get Help
  request on top of the fixed base prompt. Additive only: copy near this field should make
  clear it extends the response, it doesn't replace the required concept explanation
  (`REQUIREMENTS.md`, AI help — added this session).

---

## Explicitly not in scope for these wireframes

Per `REQUIREMENTS.md` Non-goals / Later: no dark mode, no bookmark/favorite/search/tags, no
"practice more like this," no AI usage/cost dashboard, no complexity analysis. Don't design
UI hooks for any of these — they're not deferred-but-present, they're absent.

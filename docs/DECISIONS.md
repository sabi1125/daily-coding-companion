# Decisions & Cost

## D1 — Gmail access: OAuth, not IMAP + app password

The app needs to read the daily "Daily Coding Problem" email via a scheduled background job —
no person present to click anything if a login/approval is needed in the moment.

### Options considered
| Option | What it is | Merits | Demerits / trade-offs |
|---|---|---|---|
| **A. OAuth + Gmail API** (External · Testing · single test user) | Google's OAuth flow, `gmail.readonly` scope | Scoped **read-only**; revocable; never holds the password; modern platform-standard flow | Refresh token **expires every ~7 days** in Testing mode → needs manual re-consent |
| **B. OAuth, published to production** | Same, but published to lift the 7-day limit | No re-consent; permanent tokens | Restricted scope ⇒ brand verification **+ paid annual CASA security audit** — weeks + money |
| **C. `Internal` user type** (Workspace) | OAuth with no verification, no 7-day limit | Free, no re-consent, no audit | Requires a Google Workspace org — our inbox is personal `gmail.com` → **not available** |
| **D. IMAP + app password** | Skip OAuth; connect via IMAP with a 16-char app password | Long-lived; runs unattended forever; no consent screen | We'd hold a **full mailbox master key** (read *and* write/delete, not scopeable to read-only). Fine for one trusting user, but a multi-user version turns our DB into a vault of everyone's master keys — catastrophic if breached |

### Decision & reasoning
**Chosen: A** — OAuth + Gmail API, `gmail.readonly`, External/Testing, single test user.

1. **D's risk is conditional on scale — that's the real reason to avoid it.** For solo use,
   holding one master key (our own) is a risk we take on ourselves alone. It becomes a real
   problem the moment the app might hold *other* people's mailboxes too — a single point of
   failure for everyone's data. Avoided now rather than "fixed later."
2. **Disproportionate vs. B** — a paid annual audit to read one inbox doesn't make sense for a
   solo project; the 7-day re-consent is a cheap price to skip it.
3. **C isn't available** — no Workspace org behind a personal Gmail.
4. OAuth is the properly-scoped, platform-standard way in, and leaves room to support more
   users later without a rewrite.

Trade-off accepted: refresh token expires ~7 days in Testing mode. Handled via a
`needs-reauth` state + "Reconnect Gmail" button — a missed day is visible and recoverable,
never a silent failure. Reauth surfaces via a dedicated reconnect screen, not a banner —
enough happens there (explaining what expired, why, and the reconnect action) that a
banner would be too easy to miss or dismiss.

## D2 — Storing/deploying Daily Coding Problem content

[Daily Coding Problem's Terms of Service](https://dailycodingproblem.com/terms-of-service),
load-bearing clause:

> "You will use protected content **solely for your personal use**, and will make no other
> use of the content without the express written permission of Daily Coding Problem..."

Alongside prohibitions on modifying, publishing, transmitting, selling, or "in any way
exploiting" the content — "not for resale."

### The key distinction — hosting location vs. content exposure
Worked through two deployment scenarios, both free, both on the same public URL:

| | Scenario A | Scenario B |
|---|---|---|
| Setup | URL requires login (Google sign-in) before showing anything | URL shows today's problem on the homepage, no login |
| Who can see a stored problem | Only the account holder, and only content parsed from *their own* Gmail | Anyone who visits the URL |
| Is this exploit? | No | **Yes — this is the actual line** |

Both are equally "on the internet." What matters isn't where the server sits — it's whether
someone who never subscribed and never received the email can reach the content anyway.
Scenario B does that; Scenario A doesn't.

### Decision & reasoning
Storing the parsed content is within "personal use," and deploying to a public URL is fine —
provided the engineering rule below is followed.

1. Storage is squarely personal use — saving an email you received, in your own private
   database, to read back to yourself, doesn't touch any prohibited verb (publish, transmit,
   sell, distribute).
2. Deployment location is a separate axis from exposure — auth, not hosting, decides "exploit."
3. This is already the auth model in D1 — "Sign in with Google" *is* Scenario A. Nothing new
   needed, just never accidentally add Scenario B (a public/unauthenticated leak).
4. Not touching their monetized content — DCP sells in-depth solutions as a paid feature; the
   AI-generated help here is independently written, not a copy of their paid answer key.
5. Monetizing this app would be the real line ("not for resale") — not a concern for a free
   tool, worth remembering if that ever changes.

**Engineering rule:** every route or API endpoint that returns problem content must require
the requester to be authenticated as the account that owns that data. No public demo page, no
unauthenticated endpoint, nothing a search engine could index.

## D3 — Go (Echo) + React, not single-language TypeScript
Chose two languages over one (e.g. Next.js) because: matches existing fluency; Go's
goroutines/channels are a real fit for idempotent, concurrent ingest; keeps the backend a
clean, standalone service rather than bundled into a frontend framework. Accepted the
two-toolchain cost knowingly for a solo one-month build.

## D4 — Personal GitHub account, not an organization
Chose a personal account over creating a GitHub organization for this repo. Orgs exist to
manage multi-person teams with different permission levels — a solo one-month project gets
zero benefit from that, pure overhead. One repo, personal account.

## D5 — Restricting sign-in to approved users (MVP)

Since the app is deployed on a public URL (D2) and Get Help costs real Claude API money
per call, there's a real risk: a stranger finds the URL, signs in, and uses the app's
Claude API budget for free — cost with no control on our end.

### Decision & reasoning
Already solved by D1's OAuth choice, not new infrastructure. Google's OAuth **Testing**
publishing status (chosen in D1 to avoid the CASA audit cost) has a hard side effect:
only emails explicitly added to a test-user allowlist in Google Cloud Console can
complete sign-in at all. Anyone else hits Google's own "this app hasn't completed
verification" error before ever reaching the backend — they never get far enough to call
Get Help, let alone burn API budget.

Deployment stays public (per D2) — Testing mode restricts *who can sign in*, not *where
the app can be hosted*. The two aren't in tension.

**Decided:** surface this restriction explicitly on the login screen — a line stating
that sign-in is currently limited to approved users. Not required for the mechanism to
work (Google already enforces it), but matches the project's "never silently fail"
principle: a visitor who can't sign in should see why, not just a dead-end button.

**Still open:** this protection only holds while the app stays in Testing mode. Moving to
Production publishing status (lifting the 100-user cap) reopens this risk and would need
its own solution — rate limiting, a paid tier, or bring-your-own-API-key — at that point,
not before.

## AI cost
Two calls touch the API: **ingest parse** (once a day at ingest) and **Get Help** (at most
once per problem, ever — cached after the first generation).

### Per-token pricing (as of 2026-07-16)
| Model | Input $ / 1M tokens | Output $ / 1M tokens |
|---|---|---|
| Claude Haiku 4.5 | $1.00 | $5.00 |
| Claude Sonnet 5 | $3.00 (**$2.00 intro**, through 2026-08-31) | $15.00 (**$10.00 intro**) |
| Claude Opus 4.8 | $5.00 | $25.00 |

### Per-call cost
| Call | Model | Input tokens (~) | Output tokens (~) | Cost per call |
|---|---|---|---|---|
| Ingest parse | Haiku 4.5 | 800 | 250 | **$0.0021** |
| Get Help | Sonnet 5 (intro) | 1,000 | 2,048 (cap) | **$0.023** |
| Get Help | Opus 4.8 (if used instead) | 1,000 | 2,048 (cap) | **$0.056** |

`MaxTokens` for Get Help is 2,048 (raised from 1,024 — the old cap was too tight for longer
preference-driven responses, e.g. "very detailed" or code examples in the walkthrough, and
was truncating replies). Cost below assumes the worst case of every response hitting the cap
— real usage is typically well under it.

### Monthly estimate (30 days)
| Scenario | Parse (30 days, Haiku) | Get Help (Sonnet 5) | Total / month |
|---|---|---|---|
| Rarely need help (~5 days/month) | $0.06 | $0.12 | **~$0.18** |
| Need help most days (~20 days/month) | $0.06 | $0.46 | **~$0.52** |
| Need help every single day | $0.06 | $0.69 | **~$0.75** |

Even worst-case (help every day, Sonnet 5, every response hitting the cap) stays **under
$1/month**. Opus 4.8 instead of Sonnet roughly doubles that. Model choice for Get Help
(Sonnet vs. Opus) still open — a quality call, not cost. Sonnet 5's intro pricing reverts to
$3/$15 after **2026-08-31** — re-check this estimate if the app is still running past that
date.

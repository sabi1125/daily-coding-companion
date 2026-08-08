# UI Mockups — Daily Coding Companion

Issue #7 deliverable. Built from `docs/ui-context.md` — the briefing doc listing every
state/element per screen these mockups needed to cover.

- **Web app UI mockups** (all 4 screens): https://claude.ai/code/artifact/2a01d114-9269-4327-ad87-4a91b726bfd1
- **UI kit** (colors, type scale, spacing/radius, component specs — reference values pulled
  directly from the built screens): https://claude.ai/code/artifact/f5201845-2250-4f21-9903-2b908b5339b9

## Coverage

- **Today's Problem (`/`)** — includes ingest states, `needs_review_flag` badge, attempt form
  with the CodeMirror solution editor, Get Help interaction.
- **History (`/history`)** — status filter, list rows.
- **Problem detail (`/problems/:id`)** — added as a detail popup/modal (not a separate route)
  rather than the originally-scoped standalone screen — intentional change, not scope drift.
- **Settings (`/settings`)** — Gmail connection status, Get Help preferences field, stats.


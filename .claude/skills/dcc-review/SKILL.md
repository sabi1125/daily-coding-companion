---
name: dcc-review
description: Review a daily-coding-companion PR's implementation against its linked issue and the matching api-docs.md/ingest.md spec — status codes, error categories, response bodies, layering conventions.
---

# dcc-review

Takes a GitHub PR URL. Checks whether the PR's implementation actually matches what its
linked issue and the project's own docs say it should do — same procedure used manually
for #41 (Google sign in API) this session.

## Args

`$1` — a PR URL (e.g. `https://github.com/sabi1125/daily-coding-companion/pull/54`). If
missing, ask for one.

## Procedure

1. **Extract the PR number** from the URL. `gh pr view <number> --json title,body,headRefName,baseRefName,url,files`.

2. **Find the linked issue.** Try in order:
   - An explicit reference in the PR body (`Closes #N`, `Fixes #N`, `Resolves #N`).
   - The branch-naming convention this repo uses: `feature/<issue-number>` (e.g.
     `feature/41` → issue #41). `headRefName` from step 1 gives you this.
   - If neither resolves cleanly, ask the user which issue this PR is for rather than
     guessing.

3. **Read the issue.** `gh issue view <N> --json title,body,labels`. This tells you which
   resource/endpoint(s) are in scope, and any issue-specific "Done when" criteria beyond
   what the doc alone says.

4. **Find the matching doc section.** Most issues are one endpoint from
   `api-docs/api-docs.md` (title usually names it directly, e.g. "Implement Google sign in
   API" → `## Google sign in API` section, `### GET /auth/google`). The Ingest job issue
   (#40 and its lineage) maps to `api-docs/ingest.md` instead — it has no HTTP
   request/response shape, check its Behavior/Errors sections instead of a status-code
   table. Read the full matching section, not just a grep snippet — need the whole
   Description/Auth/Headers/Responses/Errors picture.

5. **Get the actual diff and read the changed files in full.** `gh pr diff <number>` for
   an overview, then `Read` each changed `.go` file completely (not just the diff hunk) —
   need full context to judge whether layering/error-handling is right, not just what
   changed.

6. **Compare implementation against the doc, point by point:**
   - **Route**: method + path matches `### METHOD /path` exactly.
   - **Auth**: if doc says `Required`, is a session actually checked? If `None`, is
     nothing gating it?
   - **Request shape**: path/query params and request body match the doc's tables.
   - **Success response**: status code and body shape match the doc's example.
   - **Errors table**: every row — status code, Category (Expected/Operational/
     Unexpected — should map to `response.Status`'s `Category` field), and exact message
     text (`{ "message": "..." }` bodies are a wire contract, not to be reworded) all
     present and correctly categorized in the code.
   - **Ingest-specific**: if reviewing ingest work, check against `ingest.md`'s numbered
     Behavior steps and its Errors table instead — same rigor, different shape (no HTTP
     status codes, `ingest_runs.status`/`error` instead).

7. **Check repo-wide conventions** (established this session, apply regardless of which
   endpoint):
   - `controller → interactor → repository` layering — `net/http`/`echo.Context` never
     imported below the controller.
   - Errors from repository/interactor are `*response.AppError` (via `response.NewX(err)`
     constructors in `internal/response/apperror.go`), not bare errors or manual
     `c.JSON(status, err)` in the controller — controllers should `return err` and let
     `response.ErrorHandler` (wired as `e.HTTPErrorHandler`) handle status code + logging.
   - No double logging — `response.ErrorHandler` is the single place that logs
     non-`Expected` errors; a repository/interactor manually calling `logger.Error`/
     `logger.Errorw` on an error it's also returning as an `*AppError` is a bug.
   - Dependencies (config, repository) injected once at constructor time, not re-passed
     on every method call.
   - New resources register their routes by appending a `Registered<Name>Routes` function
     to `internal/infrastructure/router.go` (via `codeseed create`), wired into `Router(e,
     db)` — check the call is actually there, not just the function existing unused.

8. **Report findings to the user** ranked by severity — file:line, what's wrong, what the
   doc/issue actually says, and the concrete failure it causes (wrong status code, wrong
   JSON body, double log, bypassed error handler, etc.). If everything matches, say so
   plainly rather than padding with nitpicks — #41's final review was "matches the doc
   exactly," not every review needs to find something.

9. **Post a PR comment — always, no confirmation needed.** This skill is standing
   authorization to comment; don't ask "should I post this?" first, just post it.
   - Must be clearly attributed as an AI review — open the comment with a line like
     `🤖 **Automated review by Claude (dcc-review)**` so it's unambiguous this wasn't a
     human reviewer, both in the comment body and (if the CLI session's identity is used
     to post it) not relied on alone — the body text itself must say it.
   - If findings survived: post them as the comment body (same content as step 8, ranked
     by severity, file:line references).
   - If nothing survived (implementation matches doc + conventions cleanly): post just
     `🤖 **Automated review by Claude (dcc-review)**\n\nLGTM!` — short, no padding.
   - Use `gh pr comment <number> --body "$(cat <<'EOF' ... EOF)"` (heredoc, per this
     repo's own commit/PR conventions) so multi-line formatting survives correctly.

## Notes

- Don't just diff-eyeball — actually re-derive what the doc requires first, then check the
  code against that, the same order used manually this session (read doc section in full,
  then compare).
- If the PR only partially implements its issue (e.g. #35's migrations were split into a
  separate #52), don't flag the unimplemented part as a bug — check the issue's own body
  for whether scope was intentionally split, same as `design.md`/`ingest.md` cross-references
  throughout this repo.

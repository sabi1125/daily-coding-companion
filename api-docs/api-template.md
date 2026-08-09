# API Documentation Template

One file, `api-docs/api.md`. Each endpoint gets a `##` title (short, human-readable, e.g.
"Google sign in") followed by `### METHOD /path`, using the structure below. Ingest is
internal (cron-triggered, not a user-facing endpoint) and gets its own doc,
`api-docs/ingest.md` — different auth model, doesn't fit this template.

Deliberately shaped after OpenAPI's own model (`paths.<path>.<method>` — summary,
parameters, requestBody, responses keyed by status code) even though this is Markdown, not
YAML. When this eventually moves to a real OpenAPI spec, it should be a reformat, not a
redesign — every field below has a direct OpenAPI equivalent.

---

## Per-endpoint structure

## Endpoint title

### `METHOD /path`

**Summary** — one line, what this endpoint does.
**Description** — Description of what we are creating.

**Auth** — required / none. If required, note what identifies the caller (session cookie,
etc).

**Path parameters** *(omit section if none)*

| Name | Type | Description |
|---|---|---|

**Query parameters** *(omit section if none)*

| Name | Type | Required | Description |
|---|---|---|---|

**Request body** *(omit section if none — GET endpoints usually have none)*

| Field | Type | Required | Description |
|---|---|---|---|

```json
// example request body
```

**Responses**

`STATUS` — description

```json
// example response body
```

**Errors** — every non-success response, grouped by cause (this grouping is ours, not
OpenAPI's — each one is still just a normal `responses` entry keyed by status code):

- **Expected** — normal, client-caused (bad input, not found, not authorized). Routine,
  not logged as a problem.
- **Operational** — an external dependency failed (Gmail unreachable, Claude API error,
  DB timeout). Not our bug, but something to potentially retry/surface.
- **Unexpected** — anything else. Should be rare; if it's not rare, it belongs in one of
  the categories above instead.

| Status | Category | When | Body |
|---|---|---|---|

---

## Worked example — Auth: sign in

### `POST /auth/signin`

**Summary** — Starts the Google OAuth flow; redirects the browser to Google's consent
screen.

**Auth** — none (this endpoint is how a session begins).

**Request body** — none.

**Responses**

`302 Found` — redirect to Google's OAuth consent URL.

```
Location: https://accounts.google.com/o/oauth2/v2/auth?...
```

**Errors**

| Status | Category | When | Body |
|---|---|---|---|
| `500` | Unexpected | OAuth client misconfigured (missing client ID/secret) | `{"message": "internal server error"}` |

---

## Worked example — generic, with a request body

### `POST /users`

**Summary** — Creates a new user.

**Auth** — none.

**Request body**

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Full name |
| `email` | string | yes | Email address |

```json
{
  "name": "Ada Lovelace",
  "email": "ada@example.com"
}
```

**Responses**

`201 Created` — the created user.

```json
{
  "id": "1",
  "name": "Ada Lovelace",
  "email": "ada@example.com"
}
```

**Errors**

| Status | Category | When | Body |
|---|---|---|---|
| `400` | Expected | Missing or invalid field | `{"message": "email is required"}` |
| `409` | Expected | Email already registered | `{"message": "email already exists"}` |
| `500` | Unexpected | Unhandled server error | `{"message": "internal server error"}` |

`400 Bad Request` example:

```json
{
  "message": "email is required"
}
```

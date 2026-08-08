# Stack — The Actual Technical Decisions

> Every concrete language, framework, and library decision for the app, in one place.

## Backend (Go)

| Concern | Choice |
|---|---|
| Language / framework | **Go + Echo** |
| Calling Claude | `github.com/anthropics/anthropic-sdk-go` |
| Gmail API | `google.golang.org/api/gmail/v1` + `golang.org/x/oauth2` |
| Database | **MySQL** |
| Database access | **GORM** |
| Migrations | **golang-migrate** |
| Logging | **zap** (`go.uber.org/zap`) |
| Testing | `testing` (stdlib) + **testify** for assertions + **`go.uber.org/mock`** for mocking the Gmail/Claude clients and the repository layer. No `testcontainers-go` — spinning up real Docker databases per test run is more overhead (time, resource usage) than this project's scale justifies; mocked repository tests already cover the logic that matters. |
| Scheduler | **Railway's built-in cron.** No library needed. |
| Local dev | **Docker + docker-compose** — a local MySQL container to develop and test against, closer to what Railway actually runs. Different from the rejected `testcontainers-go`: this is one long-running container you start once, not a fresh one spun up per test run. |
| Scaffolding | **codeseed** (own tool) — generates the controller/interactor/repository + inputport layout consistently per resource (see `design.md`), one command instead of hand-rolling each new resource's files. |

---

## Frontend (React)

| Concern | Choice |
|---|---|
| Language | **TypeScript** |
| Build tool | **Vite** |
| Styling | **Tailwind CSS** |
| UI components | **shadcn/ui** |
| Routing | **React Router** |
| Data fetching / server state | **Axios** — plain HTTP client, no caching layer (TanStack Query considered and skipped as unnecessary at this scale) |
| Code editor (submitted solution) | **`@uiw/react-codemirror`** — syntax highlighting for the pasted solution field. Chosen over Monaco: Monaco is the full VS Code engine (autocomplete/IntelliSense) which is overkill for displaying/editing an already-written paste, not writing fresh code in-browser. Language auto-detected, not user-selected — nothing in the data model tracks solution language, and this is just for the user's own readability, not system-level processing. |

---

## Hosting

**Railway** — backend, database, the daily scheduled job, and the frontend all hosted there.
Chosen over AWS (disproportionate setup complexity for a solo one-month app) and over
Render/Fly (Render's free-tier idle spin-down is a real risk for a job that must reliably
fire every morning; Fly asks more infra-thinking than this project needs). Railway bundles
everything this app needs under one dashboard and deploys straight from the GitHub repo.

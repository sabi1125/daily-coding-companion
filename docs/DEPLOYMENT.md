# Deployment

> How the production deployment is put together, and how to rebuild it from nothing.

Railway is configured through its dashboard, not through code in this repo. There is no
Terraform, CloudFormation or `railway.json`. This file is the record of what was set up
there, so the deployment can be rebuilt without the original account.

It lists names and sources of settings, never values. Why Railway was chosen is in
`STACK.md`. What the pieces are is in `api-docs/architectural-diagram.md`.

## Overview

Four Railway services and one server outside Railway.

| Service | Where | Built from | Public URL |
|---|---|---|---|
| `backend` | Railway | `backend/docker/Dockerfile` | `api.dailycodingcompanion.quest` |
| `frontend` | Railway | `frontend/docker/Dockerfile` | `dailycodingcompanion.quest` |
| `ingest-engine` | Railway (cron) | `backend/docker/Dockerfile` | none |
| `MySQL` | Railway (database) | Railway's MySQL service | none |
| Piston | DigitalOcean server | `piston/docker-compose.yml` | none (port 2000, called by the backend) |

`backend`, `frontend` and `ingest-engine` are deployed from this GitHub repository.
`ingest-engine` runs the same image as `backend` with a different start command.
Auto-deploy is on, so a push to the GitHub repo redeploys the affected services.

Piston is outside Railway because it needs a `privileged` container, which Railway does not
allow (see D6 in `DECISIONS.md`).

## backend

The Go API server. Serves on port 8080, which is hard-coded in `backend/cmd/backend/main.go`.
The `PORT` and `ENVIRONMENT` variables in `.env` are not read by the server.

- **Build:** builder set to Dockerfile, with the Dockerfile path `/backend/docker/Dockerfile`.
  No root directory is set. The Dockerfile copies paths like `backend/go.mod`, so the build
  context is the repository root, not `backend/`.
- **Watch paths:** `/backend/**`, so only changes under `backend/` trigger a deploy.
- **Image contents:** the compiled `backend` binary, the `migrate` CLI
  (golang-migrate v4.19.0, MySQL driver), the `migrations/` folder and `entrypoint.sh`.
- **Start:** the image's own entrypoint and default command, no override. `entrypoint.sh`
  runs `migrate up` against the database, then starts `./backend`. Every deploy therefore
  applies any new migrations before serving.
- **Health check:** `GET /health` exists. No Railway health check is configured.

Environment variables. Every one except `LOG_LEVEL` is required, and the process exits at
startup if it is missing.

| Variable | What it is for | Where the value comes from |
|---|---|---|
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | MySQL connection | The Railway `MySQL` service's variables |
| `LOG_LEVEL` | Optional. `DEBUG` or `INFO`. Anything else, or unset, is treated as `INFO` | Chosen by hand |
| `FRONTEND_ORIGINS` | Comma-separated origins allowed by CORS | The frontend's public origin, `https://dailycodingcompanion.quest` |
| `REDIRECT_URL` | The OAuth redirect URI Google sends the user back to | `https://api.dailycodingcompanion.quest/auth/google/callback`. Must match the authorized redirect URI on the Google OAuth client |
| `CALLBACK_REDIRECT_URL` | Where the backend sends the browser after sign-in finishes | The frontend's public origin |
| `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` | Google OAuth client | Google Cloud Console |
| `ANTHROPIC_API_KEY` | Get Help and problem parsing | Anthropic console |
| `PISTON_BASE_API` | Where the backend reaches Piston | The Piston server. Shape: `http://<server-address>:2000/api`. The backend appends `/v2/execute` |
| `PISTON_SHARED_SECRET` | Sent as the `X-Auth` header on every Piston call | Chosen by hand. Must equal the same variable on the Piston server |

Session and OAuth-state cookies are set with `Secure`, and the session cookie uses
`SameSite=None` because the frontend and API live on different subdomains. The API must be
served over HTTPS.

## frontend

A static React build served by nginx.

- **Build:** builder set to Dockerfile, with the Dockerfile path `/frontend/docker/Dockerfile`.
  No root directory is set. Same rule as the backend: the Dockerfile copies `frontend/...`
  paths, so the build context is the repository root.
- **Watch paths:** `/frontend/**`.
- **Build steps inside the image:** `npm ci`, then `npm run build`. The output in `dist/` is
  copied into an nginx image.
- **Serving:** nginx listens on `$PORT` (default 8080) using
  `frontend/docker/nginx.conf.template`. Unknown paths fall back to `index.html` so
  client-side routes work.

| Variable | What it is for | Where the value comes from |
|---|---|---|
| `VITE_API_BASE_URL` | The API origin the browser calls | The backend's public URL, `https://api.dailycodingcompanion.quest`. Read at build time and baked into the bundle, so changing it needs a redeploy. Set as a Railway variable |
| `PORT` | nginx listen port | Injected by Railway |

## ingest-engine

The daily job. It fetches each connected user's Daily Coding Problem email and turns it into
a problem. What the job does is in `api-docs/ingest.md`.

- **Service type:** a Railway cron service, separate from `backend`.
- **Build:** the same Dockerfile as `backend`.
- **Start command:** `./backend ingest`. Without the `ingest` argument the binary starts the
  web server. The image entrypoint still runs first, so `migrate up` also runs at the start
  of each job.
- **Schedule:** 2am JST, which is `0 17 * * *` in UTC (Railway cron runs in UTC). The reason
  for 2am is in `api-docs/ingest.md`.
- **Variables:** the same values as `backend` for `DB_HOST`, `DB_PORT`, `DB_NAME`,
  `DB_USER`, `DB_PASSWORD`, `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `REDIRECT_URL`,
  `CALLBACK_REDIRECT_URL` and `ANTHROPIC_API_KEY`, plus optionally `LOG_LEVEL`. The job does
  not need `FRONTEND_ORIGINS`, `PISTON_BASE_API` or `PISTON_SHARED_SECRET`, because those
  are only read on the server path.

## MySQL

- **Engine:** MySQL 9.7.2, from Railway's MySQL service, with a volume named `mysql-volume`
  for data.
- **Provisioning:** added to the Railway project as a MySQL service. The schema is created
  by the migrations, not by hand.
- **Variables:** the MySQL service defines its own connection variables in its Variables tab
  in Railway. The `DB_*` variables on `backend` and `ingest-engine` take their values from
  those.
- **Migrations:** `backend/migrations/`, applied by `entrypoint.sh` when `backend` or
  `ingest-engine` starts.
- **Local development differs.** `docker-compose.yml` runs `mysql:8.4`. The schema has been
  developed against 8.4 and runs in production on 9.7.2.

## Piston

The code-execution engine, run with Docker on a single DigitalOcean droplet named
`daily-coding-companion-piston`: region NYC1, 1 GB RAM, 25 GB disk, tagged `piston`. The
operating system image is not recorded here.

- **What runs:** `piston/docker-compose.yml`. It starts three containers: the Piston engine
  (`ghcr.io/engineer-man/piston`, `privileged`), an nginx proxy on port 2000, and a one-shot
  installer that adds the language runtimes.
- **Runtimes installed:** python 3.12.0, node 20.11.1, go 1.16.2, gcc 10.2.0
  (`piston/scripts/piston_runtimes_install.sh`).
- **How it is protected:** nginx rejects requests without an `X-Auth` header (401) and
  requests where the header does not equal `PISTON_SHARED_SECRET` (403). Only then does it
  pass the request to Piston.
- **Variables:** `PISTON_SHARED_SECRET` in `piston/.env` on the server. That file is not in
  the repo. It is read by the nginx template.
- **Start:** `docker compose up -d` in the `piston/` folder. Do not add
  `docker-compose.local.yml`, which only joins the local development network.
- **How the backend reaches it:** a POST to `<PISTON_BASE_API>/v2/execute` with the
  `X-Auth` header.
- **Known limitation:** the connection between Railway and this server is plain HTTP, not
  HTTPS. The shared secret and the submitted code cross the network unencrypted. Putting TLS
  in front of the proxy would fix it.

## Domains

The frontend is served at `dailycodingcompanion.quest` and the API at
`api.dailycodingcompanion.quest`, both as custom domains on their Railway services.

The domain is managed in Railway. Railway also manages its DNS and issues the HTTPS
certificates, so there is no certificate or renewal step to do by hand.

The individual DNS records and the Google Cloud Console settings are deliberately not
written down here. They are account-level details kept out of the repo on purpose.

## Rebuilding from scratch

1. Create a Railway project. Add a MySQL service and note its variables.
2. Add the `backend` service from this GitHub repo. Set the builder to Dockerfile, the
   Dockerfile path to `/backend/docker/Dockerfile`, and the watch path to `/backend/**`.
   Leave the root directory empty. Set the backend variables from the table above.
3. Add the `frontend` service from the same repo. Set the builder to Dockerfile, the
   Dockerfile path to `/frontend/docker/Dockerfile`, and the watch path to `/frontend/**`.
   Set `VITE_API_BASE_URL` before the first build.
4. Add the `ingest-engine` service from the same repo, built with the backend Dockerfile.
   Set the start command to `./backend ingest`, the cron schedule to `0 17 * * *`, and the
   variables listed above.
5. Create a droplet (or any server that allows privileged containers; 1 GB RAM is what runs
   today), install Docker, copy the
   `piston/` folder, create `piston/.env` with `PISTON_SHARED_SECRET`, and run
   `docker compose up -d`. Wait for the installer to finish, then check that
   `<address>:2000/api/v2/runtimes` answers when the `X-Auth` header is sent.
6. Set `PISTON_BASE_API` and `PISTON_SHARED_SECRET` on `backend` to match the Piston server.
7. Register a Google OAuth client. Add the redirect URI from `REDIRECT_URL`, then put the
   client ID and secret on `backend` and `ingest-engine`. Sign-in stays limited to approved
   test users (D5 in `DECISIONS.md`).
8. Attach the two custom domains to the frontend and backend services. Railway shows the
   DNS records to add, and issues the certificates once they resolve.
9. Deploy the backend first so the migrations create the schema. Then deploy the frontend
   and confirm sign-in works end to end.

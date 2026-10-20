# URL Shortener

[![Go](https://github.com/CosmoS1X/url-shortener/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/CosmoS1X/url-shortener/actions/workflows/go.yml)

A self-hosted URL shortener: Go API (Gin + PostgreSQL) with a react-admin SPA frontend, served behind Caddy in a single Docker container.

**Live demo:** https://url-shortener-gv4f.onrender.com

## Features

- Shorten URLs with an auto-generated or custom short name
- Redirect via `GET /r/:code`
- Visit statistics per link (IP, referer, user agent, timestamp)
- Admin UI (react-admin) for managing links and viewing analytics
- Pagination via inclusive `range=[start,end]` (ra-data-simple-rest compatible)
- Consistent JSON error format and request timeouts

## Tech stack

- **Backend:** Go 1.26, Gin, sqlc (SQL-first codegen), pgx/v5, goose migrations
- **Frontend:** prebuilt SPA (`@hexlet/project-url-shortener-frontend`) — react-admin + Vite
- **Infra:** Docker (Caddy as reverse proxy + static file server), deploy on Render

## Quick start (local)

Requirements: Go 1.26+, Node 22+, PostgreSQL.

```bash
cp .env.example .env          # adjust DATABASE_URL / BASE_URL if needed
createdb url_shortener
make migrate-up
make dev                      # backend (air, :8080) + frontend (vite preview, :5173)
```

Open http://localhost:5173 — the frontend proxies `/api` to the backend, so everything is same-origin.

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | yes | — | PostgreSQL connection string |
| `BASE_URL` | yes | — | Public origin **without** `/r` (e.g. `https://example.com`); `short_url` is built as `BASE_URL + "/r/" + short_name` |
| `SERVER_PORT` | no | `8080` | Port for the Go HTTP server |
| `CORS_ORIGIN` | no | `http://localhost:5173` | Allowed CORS origin (frontend origin in your deployment) |

## API

| Method | Path | Description |
|---|---|---|
| `GET` | `/ping` | Health check |
| `GET` | `/r/:code` | Redirect to the original URL (records a visit) |
| `GET` | `/api/links?range=[start,end]` | List links (default 10, max 100 per page) |
| `POST` | `/api/links` | Create link — `{"original_url": "...", "short_name": "optional"}` |
| `GET` | `/api/links/:id` | Get link |
| `PUT` | `/api/links/:id` | Update link |
| `DELETE` | `/api/links/:id` | Delete link (cascades visits) |
| `GET` | `/api/link_visits?range=[start,end]` | List visit records |

List responses are JSON arrays with a `Content-Range` header (`links {start}-{end}/{total}`). Validation errors return `422 {"errors": {"<field>": "<message>"}}`; missing resources return `404 {"error": "..."}`.

## Deployment

The project ships a multi-stage `Dockerfile` (frontend build → Go build + goose → Alpine runtime with Caddy). At container start, `bin/run.sh` applies migrations, launches Caddy (serves the SPA and proxies `/api/*`, `/ping`, `/r/*` to the Go server on `:8080`), then execs the Go server.

On [Render](https://render.com) deploy from a Docker runtime, set `DATABASE_URL` (managed PostgreSQL), `BASE_URL` to your service URL and `CORS_ORIGIN` if the frontend is served from a different origin.

## Development

| Command | Description |
|---|---|
| `make dev` | Backend + frontend with hot reload |
| `make test` / `make test-race` | Run tests |
| `make lint` / `make lint-fix` | golangci-lint |
| `make migrate-up` / `make migrate-down` | Apply / roll back migrations |
| `make sqlc-generate` | Regenerate sqlc code |
| `make build` | Build binary to `./bin/server` |

See `Makefile` for the full list (`test-coverage`, `vuln`, `tidy`, ...).

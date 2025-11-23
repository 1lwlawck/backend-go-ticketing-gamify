# backend-go-ticketing-gamify

Backend API for the Vue ticketing + gamification app (Go + Gin + pgx).
All business endpoints are under `/api/v1` and protected by JWT; optional API key via `X-API-Key` if `API_KEY` is set in `.env`.

## Quick start
1. Copy `.env.example` to `.env` and fill `DATABASE_URL`, `JWT_SECRET`, (optional) `API_KEY`.
2. Install Go 1.22+ and deps: `go mod tidy`.
3. Run: `go run ./cmd/server` (or `air`).
4. Health checks: `/healthz`, `/version`.

## Auth endpoints
- `POST /api/v1/auth/login` — `{ username, password } -> { token, user }`
- `POST /api/v1/auth/register` — creates user and returns token
- `POST /api/v1/auth/change-password` — auth required, body `{ oldPassword, newPassword }`, returns 204

## Tickets & XP
- `PATCH /api/v1/tickets/:id/status` awards XP when moving into `done`; moving out of `done` rolls XP back.
- Comments can be edited/deleted by the author: `PATCH /comments/:commentId`, `DELETE /comments/:commentId`.

## Seeding with faker
We ship a faker seeder for dev data.
```
SEED_USERS=20 SEED_PROJECTS=5 SEED_TICKETS=20 SEED_COMMENTS=20 \
DATABASE_URL=postgres://... go run ./cmd/seed
```
Defaults (if not set): users=10, projects=3, tickets=25, comments=40. Password for seeded users: `password`.

## Docker (API only)
```
docker build -t ticketing-backend:latest .
docker run --env-file .env -p 8080:8080 ticketing-backend:latest
```
For local Postgres, add a `db` service in docker-compose and set `DATABASE_URL=postgres://...@db:5432/...`.

## Project layout
- `cmd/server` — HTTP server bootstrap
- `cmd/seed` — faker seeder runner
- `cmd/dbcheck` — quick DB connectivity check
- `internal/*` — domain modules (auth, users, projects, tickets, gamification, audit), middleware, config
- `migrations/` — SQL migrations (schema + seed snapshots)

## Notes
- JWT expires in 24h; send `Authorization: Bearer <token>`.
- Rate limiting per IP/API key configurable via env (`RATE_LIMIT_*`).
- API key header: `X-API-Key` (if `API_KEY` set).

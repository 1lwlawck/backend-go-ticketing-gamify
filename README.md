# backend-go-ticketing-gamify

Backend API for the Vue ticketing + gamification app (Go + Gin + pgx).
All business endpoints are under `/api/v1` and protected by JWT; optional API key via `X-API-Key` if `API_KEY` is set in `.env`.

## Quick start
1) Prasyarat: Go 1.22+, Postgres, `psql` (atau GUI DB).
2) Salin env: `cp .env.example .env`, isi `DATABASE_URL`, `JWT_SECRET`, (opsional) `API_KEY`.
3) Install deps: `go mod tidy`.

4) Jalankan server:
   - `go run ./cmd/server`
   - atau hot reload: `air`
5) Cek: `/healthz`, `/version` di `http://localhost:8080`.

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

## Jalankan frontend
1) Buka repo `taskmgr-ticketing-gamification`.
2) Salin `.env.example` ke `.env`, set `VITE_API_BASE_URL=http://localhost:8080/api/v1` dan `VITE_API_KEY` jika backend memakai API key.
3) `npm install` lalu `npm run dev` (Vite).
4) Login dengan user yang sudah ada atau register via UI.

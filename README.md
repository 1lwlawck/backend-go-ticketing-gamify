# backend-go-ticketing-gamify

Backend API for the Vue ticketing + gamification app (Go + Gin + pgx).
All business endpoints are under `/api/v1` and protected by JWT; optional API key via `X-API-Key` if `API_KEY` is set in `.env`.

## Quick start (Docker)
1) Salin env: `cp .env.example .env`, isi `JWT_SECRET`, opsional `API_KEY`, dan set `DATABASE_URL=postgres://ticket:ticket@db:5432/ticket?sslmode=disable`.
2) Start stack:
   ```
   docker compose up -d db
   docker compose up -d api
   ```
   - Kalau port 5432 di host bentrok, ubah mapping di `docker-compose.yml` (mis. `"5433:5432"`). `DATABASE_URL` di API tetap `db:5432`.
3) Skema otomatis diinit dari `database/schema.sql` saat volume `db-data` pertama kali dibuat. Jika volume lama sudah ada dan butuh reset, jalankan `docker compose down -v` lalu start ulang.
4) (Opsional) Seed data dummy:
   ```
   docker compose run --rm seed
   ```
5) Cek: `/healthz`, `/version` di `http://localhost:8080`; API routes di `/api/v1/...`.

## Dev native (opsional)
Masih bisa jalan tanpa Docker:
1) Prasyarat: Go 1.22+, Postgres.
2) `cp .env.example .env`, set `DATABASE_URL`, `JWT_SECRET`.
3) `go mod tidy`, lalu `go run ./cmd/server` (atau `air` untuk hot reload).
4) Import skema ke Postgres lokal pakai `database/schema.sql`.

## Auth endpoints
- `POST /api/v1/auth/login` — `{ username, password } -> { token, user }`
- `POST /api/v1/auth/register` — creates user and returns token
- `POST /api/v1/auth/change-password` — auth required, body `{ oldPassword, newPassword }`, returns 204

## Tickets & XP
- `PATCH /api/v1/tickets/:id/status` awards XP when moving into `done`; moving out of `done` rolls XP back.
- Comments can be edited/deleted by the author: `PATCH /comments/:commentId`, `DELETE /comments/:commentId`.

## Seeding with faker (manual)
```
SEED_USERS=20 SEED_PROJECTS=5 SEED_TICKETS=20 SEED_COMMENTS=20 \
DATABASE_URL=postgres://... go run ./cmd/seed
```
Defaults (if not set): users=10, projects=3, tickets=25, comments=40. Password untuk user seeded: `password`.

## Seeding demo dataset (lebih realistis)
```
SEED_PRESET=demo DATABASE_URL=postgres://... go run ./cmd/seed
```
Preset ini mengisi user (PM, dev, QA, admin), dua project aktif, epics, tiket dengan status beragam, komentar, history, serta poin gamifikasi agar papan terlihat seperti sprint nyata.

## Project layout
- `cmd/server` - HTTP server bootstrap
- `cmd/seed` - faker seeder runner
- `cmd/dbcheck` - quick DB connectivity check
- `internal/*` - domain modules (auth, users, projects, tickets, gamification, audit), middleware, config
- `database/schema.sql` - base schema untuk init DB
- `migrations/` - SQL migrations (snapshot lanjutan)

## Notes
- JWT expires in 24h; send `Authorization: Bearer <token>`.
- Rate limiting per IP/API key configurable via env (`RATE_LIMIT_*`).
- API key header: `X-API-Key` (if `API_KEY` set).

## Jalankan frontend
1) Buka repo `taskmgr-ticketing-gamification`.
2) Salin `.env.example` ke `.env`, set `VITE_API_BASE_URL=http://localhost:8080/api/v1` dan `VITE_API_KEY` jika backend memakai API key.
3) `npm install` lalu `npm run dev` (Vite).
4) Login dengan user yang sudah ada atau register via UI.

# backend-go-ticketing-gamify

Backend API for the Vue ticketing + gamification app. Built with Go + Gin, pgx, and Supabase Postgres. All business endpoints are versioned under /api/v1.

## Requirements

- Go 1.23+
- Supabase (or Postgres) connection string

## Setup

1. Copy the environment template and edit credentials:
   `ash
   cp .env.example .env
   # update DATABASE_URL, JWT_SECRET, PORT, etc.
   `
2. Install dependencies:
   `ash
   go mod tidy
   `
3. Run the server:
   `ash
   go run ./cmd/server
   `
   The API listens on http://localhost:8080 (or the PORT you configure).

### DB connectivity helper

`ash
go run ./cmd/dbcheck
`

The command reads DATABASE_URL, opens a pgx connection, and prints SELECT version() so you can confirm Supabase connectivity quickly.

## API Overview

Authentication uses JWT (24h expiry). Send Authorization: Bearer <token> for every /api/v1/* request except /api/v1/auth/*.

### Auth

| Method | Route | Description |
| ------ | ----- | ----------- |
| POST | /api/v1/auth/login | Body { "username": "...", "password": "..." } ? { token, user } |
| POST | /api/v1/auth/register | Body { "name","username","password","avatarUrl?" } registers a user and returns { token, user }. |

### Users

| Method | Route | Description |
| ------ | ----- | ----------- |
| GET | /api/v1/users | Admin / project_manager only. List all users. |
| GET | /api/v1/users/me | Return the authenticated user profile. |
| PATCH | /api/v1/users/me | Update own profile { name, bio, avatarUrl }. |
| GET | /api/v1/users/:id | Admin / project_manager: fetch another user profile. |
| PATCH | /api/v1/users/:id/role | Admin only: update a member’s role. |

### Projects

| Method | Route | Notes |
| ------ | ----- | ----- |
| GET | /api/v1/projects | List projects. |
| GET | /api/v1/projects/:id | Detail with members. |
| POST | /api/v1/projects | Admin / project_manager only. Body { name, description, members[] }. |
| POST | /api/v1/projects/:id/members | Admin / project_manager: add member { userId, role? }. |
| POST | /api/v1/projects/:id/invites | Admin / project_manager: generate invite { maxUses, expiryDays }. |
| POST | /api/v1/projects/join | Body { code } to join by invite. |

### Tickets

| Method | Route | Notes |
| ------ | ----- | ----- |
| GET | /api/v1/tickets | Filters: projectId, ssigneeId, status, limit. |
| POST | /api/v1/tickets | Create ticket; body includes projectId/title/description/priority/type/etc. |
| GET | /api/v1/tickets/:id | Ticket detail (history + comments included). |
| PATCH | /api/v1/tickets/:id/status | Body { "status": "done" }; awards XP when moving to done. |
| POST | /api/v1/tickets/:id/comments | Body { "text": "..." } adds a comment. |
| DELETE | /api/v1/tickets/:id | Remove ticket (history/comments cascade). |

### Gamification

| Method | Route | Description |
| ------ | ----- | ----------- |
| GET | /api/v1/gamification/stats/:userId | Aggregated XP + level info. |
| GET | /api/v1/gamification/events | XP feed (?userId=&limit=). |

### Audit (admin / project_manager only)

| Method | Route | Description |
| ------ | ----- | ----------- |
| GET | /api/v1/audit | Paginated audit trail (?limit=). |

### Public utility endpoints

- GET / – service ping
- GET /healthz – health probe
- GET /version – running service version

## Architecture

- cmd/server – bootstrap, config loading, DB pool init, graceful shutdown.
- cmd/dbcheck – Supabase connectivity validator.
- internal/config – environment loader (supports .env).
- internal/database – pgx pool helper.
- internal/server – Gin router + dependency injection.
- Domain modules (internal/auth, internal/users, internal/projects, internal/tickets, internal/gamification, internal/audit) expose repository ? service ? handler layers.
- internal/middleware – request ID + JWT auth/role middleware.

Extend functionality by creating new repo/service/handler modules and registering them inside internal/server.

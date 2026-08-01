# TEDx Universitas Airlangga — Backend

REST API for the TEDx Universitas Airlangga web platform. Built with Go, Gin,
GORM (PostgreSQL), JWT auth, and a modular per-feature layout wired together
with `samber/do` dependency injection.

## Stack

- **Language:** Go 1.25
- **HTTP:** [Gin](https://github.com/gin-gonic/gin)
- **ORM / DB:** [GORM](https://gorm.io) + PostgreSQL (uses the `uuid-ossp` extension)
- **Auth:** JWT access tokens ([golang-jwt v4](https://github.com/golang-jwt/jwt)) + opaque refresh tokens
- **DI:** [samber/do](https://github.com/samber/do)
- **Config:** [godotenv](https://github.com/joho/godotenv) + [viper](https://github.com/spf13/viper)
- **Email:** [Brevo](https://www.brevo.com) transactional API
- **Live reload (dev):** [air](https://github.com/air-verse/air)

## Project layout

```
cmd/                     Entry point (main.go): DI bootstrap + route registration
config/                  DB connection, GORM logger, email config
database/
  entities/              GORM models (User, Todo, RefreshToken, common Timestamp)
  migrations.go          AutoMigrate
middlewares/             CORS, JWT authentication guard
modules/                 Feature modules — each: dto / handler / service / repository / routes
  auth/                  Register, login, refresh, logout, email verify, password reset
  bundle/                Bundle catalog (public read, admin write) + image gallery
  todo/                  CRUD example resource (auth-protected)
  user/                  User DTOs + repository (shared by auth)
pkg/
  constants/             Roles, pagination defaults, DI keys
  helpers/               Password hashing (bcrypt)
  utils/                 Standard API response envelope, email sender + templates
providers/core.go        RegisterDependencies — wires repos, services, handlers
API_Test/                API collection (OpenCollection format) for Auth, Todo & Bundle
```

Each module follows the same flow: `routes → handler → service → repository → entities`.

## Getting started

### Prerequisites
- Go 1.25+
- PostgreSQL (a database matching `DB_NAME`; the app creates the `uuid-ossp` extension on startup)

### Setup

```bash
cp .env.example .env      # then fill in the values below
go mod download
go run ./cmd              # or: air   (live reload, see .air.toml)
```

Server listens on `GOLANG_PORT` (default `8888`). On boot it runs AutoMigrate
and prints a TEDx banner.

### Environment variables

| Var | Description |
|-----|-------------|
| `APP_NAME` | Application name |
| `APP_ENV` | `localhost` binds `0.0.0.0`, otherwise binds `:PORT` |
| `GOLANG_PORT` | HTTP port (default `8888`) |
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASS` / `DB_NAME` | PostgreSQL connection |
| `JWT_SECRET` | HMAC secret for access tokens |
| `BREVO_API_KEY` | Brevo transactional email API key |
| `BREVO_SENDER_EMAIL` / `BREVO_SENDER_NAME` | Email sender identity |

> `.env` is git-ignored. Email sending (verification, password reset) requires a
> valid `BREVO_API_KEY`; without it those endpoints return an error.

## API

Base path: `/api/v1`. All responses use a shared envelope:

```json
{ "status": true, "message": "...", "data": {}, "error": "...", "meta": {} }
```

### Auth — `/api/v1/auth`

| Method | Path | Auth | Body |
|--------|------|------|------|
| POST | `/register` | — | `name`, `email`, `password` (min 8), `telp_number?` |
| POST | `/login` | — | `email`, `password` → access + refresh token |
| POST | `/refresh` | — | `refresh_token` → new token pair (rotates) |
| POST | `/logout` | Bearer | — (revokes user's refresh tokens) |
| POST | `/send-verification-email` | — | `email` (sends 6-digit OTP, 15 min) |
| POST | `/verify-email` | — | `email`, `code` (6 digits) |
| POST | `/send-password-reset` | — | `email` (emails a reset token) |
| POST | `/reset-password` | — | `token`, `new_password` (min 8) |

Register also fires a verification email asynchronously.

### Todo — `/api/v1/todos` (all require `Authorization: Bearer <access_token>`)

| Method | Path | Body / Query |
|--------|------|------|
| POST | `` | `name`, `category` |
| GET | `` | `page`, `per_page` (paginated) |
| GET | `/:id` | — |
| PATCH | `/:id` | `name?`, `category?`, `is_done?` |
| DELETE | `/:id` | — |

### Bundle — `/api/v1/bundles`

Bundle catalog — display only; checkout happens through a Google Form.

| Method | Path | Auth | Body / Query |
|--------|------|------|------|
| GET | `` | — | `page`, `per_page`, `is_active?` |
| GET | `/:id` | — | — (detail incl. `images`) |
| POST | `` | Bearer (admin) | `name`, `description`, `price` |
| PATCH | `/:id` | Bearer (admin) | `name?`, `description?`, `price?`, `is_active?` |
| DELETE | `/:id` | Bearer (admin) | — |
| POST | `/:id/images` | Bearer (admin) | `image_url` |
| DELETE | `/:id/images/:imageId` | Bearer (admin) | — |

Notes:

- `price` is sent and returned as a **string** (e.g. `"150000.00"`) so money keeps its
  precision on both sides. Valid range: `0` – `99999999.99` (column is `numeric(10,2)`).
- `GET /bundles` without an `is_active` parameter returns **active bundles only**, since
  the endpoint is public. Pass `?is_active=false` to list hidden ones.
- Bundles are always created active — deactivate them with `PATCH`.
- Deleting a bundle also deletes its images (`ON DELETE CASCADE`).
- ⚠️ Admin endpoints are currently guarded by `Authenticate` (login check) only. The
  role check waits on the `AuthorizeAdmin` middleware, which is a separate task.

### Auth flow

1. Access tokens are HS256 JWTs, 15 min TTL, carry `user_id` + `role`.
2. Refresh tokens are random opaque strings, 7 day TTL, stored in DB and rotated on `/refresh`.
3. Protected routes send `Authorization: Bearer <access_token>`; the middleware validates and injects `user_id` into the context.

### 🧪 API Testing

Project memakai **Bruno** (collection file-based di folder `API_Test/` pada repo backend).
Sudah berisi request untuk Auth & Todo. Setiap modul API baru **wajib menambahkan
folder request-nya** ke `API_Test/`, mengikuti pola folder `Todo/`. Collection ikut
di-commit dan di-review di branch `be`.

```bash
go test ./...
```

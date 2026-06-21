# goph scaffolded web app

## Stack

- **Router:** chi (`github.com/go-chi/chi/v5`)
- **ORM:** bun (`github.com/uptrace/bun`) on pgx
- **Templates:** templ (`.templ` files, run `templ generate`)
- **Frontend:** HTMX 2.x + Bootstrap 5 + Catppuccin themes
- **Auth:** JWT session cookies (golang-jwt), bcrypt passwords
- **Middleware:** gorilla/csrf, go-chi/httprate

## Structure

| Path | Purpose |
|------|---------|
| `cmd/server/main.go` | Entry point; creates App, listens on `:8080` |
| `db/db.go` | bun DB open + goose migration runner |
| `internal/server/app.go` | Config, middleware wiring, route mounting |
| `internal/server/auth/` | Login, register, forgot/reset password |
| `internal/server/dashboard/` | Dashboard page |
| `internal/server/settings/` | Profile, email, password, theme, delete account |
| `internal/infrastructure/` | Email (Resend), security (bcrypt/HMAC/JWT), validation |
| `internal/models/` | bun model definitions |
| `internal/views/` | templ page components |
| `middleware/` | JWT auth, CSRF, rate limiting |
| `migrations/` | goose SQL migrations |

## Conventions

- Use `infrastructure.EmailService` interface everywhere (not re-declare it)
- Auth handlers live in `auth/handler.go`, business logic in `auth/service.go`
- HTMX error responses return body with 200 OK (client ignores 4xx bodies)

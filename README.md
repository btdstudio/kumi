# goph

Scaffold production-ready Go + HTMX + Bootstrap 5 + Catppuccin web applications.

## Install

```bash
go install github.com/briantimmer/goph/cmd/goph@latest
```

After installing, run `goph doctor` to check and install required tools ([templ](https://github.com/a-h/templ), [air](https://github.com/air-verse/air), [goose](https://github.com/pressly/goose)):

```bash
goph doctor --install
```

## Usage

### Scaffold a new project

```bash
goph new myapp
goph new myapp github.com/user/myapp
```

This creates a project directory with auth (email/password, JWT sessions, password reset), dashboard, settings (profile/email/theme/delete account), and a full Catppuccin-themed UI with HTMX.

### Environment variables

Generated projects require these environment variables:

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes | Postgres connection string |
| `SECRET_KEY` | Yes | JWT signing key (any random string) |
| `RESEND_API_KEY` | No | For password reset emails |
| `MAIL_FROM_ADDRESS` | No | Sender address for emails |

### Run database migrations

```bash
cd myapp
export DATABASE_URL="postgres://user:pass@localhost:5432/myapp?sslmode=disable"
export SECRET_KEY="dev-secret-key-change-in-production"
goph db migrate
```

### Start development server

```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/myapp?sslmode=disable"
export SECRET_KEY="dev-secret-key-change-in-production"
goph dev
```

Starts air with hot-reload on port 8080.

### Inside a generated project

```bash
goph db migrate  # goose up
goph dev         # air hot-reload server
```

## Generated Stack

- **Language:** Go 1.22+
- **Router:** chi/v5
- **ORM:** bun
- **Migrations:** goose
- **Auth:** JWT (golang-jwt/jwt/v5)
- **Templating:** templ
- **Frontend:** HTMX + Bootstrap 5 + Catppuccin
- **Rate limiting:** httprate
- **CSRF:** gorilla/csrf

## Project Structure

```
myapp/
├── cmd/server/main.go    # entry point
├── internal/
│   ├── server/           # handlers (auth, dashboard, settings)
│   ├── infrastructure/   # security, validation, email
│   ├── models/           # data models
│   └── views/            # templ components
├── middleware/            # JWT auth, CSRF
├── migrations/            # goose migrations
├── db/                   # database setup
└── static/               # CSS, JS, images
```

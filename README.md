# goph

Scaffold production-ready Go + HTMX + Bootstrap 5 + Catppuccin (plus other themes) web applications.

## Install

```bash
go install github.com/btdstudio/goph/cmd/goph@latest
```

### Keeping it updated
To get the latest bugfixes and features, simply re-run:
```bash
go install github.com/btdstudio/goph/cmd/goph@latest
```

### Shell Path Setup
If the `goph` command is not recognized after running the installation, make sure Go's binary directory is added to your shell's `$PATH` variable:

```bash
# Add Go binary path to Zsh profile (macOS default):
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc && source ~/.zshrc
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

Generated projects include `.env.example`. Copy it and fill in your values:

```bash
cp .env.example .env
```

The app automatically loads `.env` at startup. Required vars:

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes | Postgres connection string |
| `SECRET_KEY` | Yes | Signing key, at least 32 characters. Generate with `openssl rand -hex 32` |
| `BASE_URL` | Yes | Public URL of the app (e.g. `https://example.com`). Used for links in emails; `https` also marks cookies `Secure` |
| `TRUST_PROXY` | No | Set to `true` only behind a reverse proxy that overwrites `X-Forwarded-For`/`X-Real-IP`, so rate limits see the real client IP |
| `RESEND_API_KEY` | No | For password reset emails |
| `MAIL_FROM_ADDRESS` | No | Sender address for emails |

### Production checklist

- Generate a unique `SECRET_KEY` (`openssl rand -hex 32`); the app refuses to start with a short or placeholder key.
- Set `BASE_URL` to your public `https://` URL so session and CSRF cookies are `Secure` and email links point at your host.
- Terminate TLS in front of the app. If it sits behind a reverse proxy, set `TRUST_PROXY=true` and make sure the proxy overwrites `X-Forwarded-For`; never enable it when the app is directly reachable.

### Run database migrations

```bash
cd myapp
goph db migrate
```

### Start development server

```bash
goph dev
```

Starts air with hot-reload on port 8080. Both `goph db migrate` and `goph dev` read the same environment — set vars in `.env` or export them.

### Inside a generated project

```bash
goph db migrate  # goose up
goph dev         # air hot-reload server
```

### Git hooks

Generated projects include `.githooks/pre-commit` that formats code and runs `go vet` before each commit. Enable it:

```bash
git config core.hooksPath .githooks
```

## Generated Stack

- **Language:** Go 1.26+
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

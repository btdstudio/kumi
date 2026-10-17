# goph CLI

## Commands

- `cmd/goph/main.go` — root cobra command
- `cmd/goph/new.go` — `goph new <name> [module-path]`: scaffolds project from `_template/`
- `cmd/goph/dev.go` — `goph dev [dir]`: starts air hot-reload
- `cmd/goph/doctor.go` — `goph doctor [--install]`: checks/installs templ, air, goose
- `cmd/goph/db.go` — `goph db migrate [dir]`: runs goose migrations

## Scaffolding

- `embed.go` embeds `_template/` via `//go:embed`
- `internal/scaffold/scaffold.go` walks embedded FS, substitutes module name in `.tmpl`/`.go`/`.templ` files
- `internal/scaffold/scaffold_test.go` tests module substitution, import rewriting, verbatim copy

## Template

- `_template/` is a production-ready web app skeleton — edits to it become the default for new projects
- Module is placeholder `goph`; scaffold rewrites `goph/` -> user's module path
- `.tmpl` files are processed through `text/template` for dynamic content; `.go`/`.templ` files get string replacement

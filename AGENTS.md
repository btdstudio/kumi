# kumi CLI

## Commands

- `cmd/kumi/main.go` — root cobra command
- `cmd/kumi/new.go` — `kumi new <name> [module-path]`: scaffolds project from `_template/`
- `cmd/kumi/dev.go` — `kumi dev [dir]`: starts air hot-reload
- `cmd/kumi/doctor.go` — `kumi doctor [--install]`: checks/installs templ, air, goose
- `cmd/kumi/db.go` — `kumi db migrate [dir]`: runs goose migrations

## Scaffolding

- `embed.go` embeds `_template/` via `//go:embed`
- `internal/scaffold/scaffold.go` walks embedded FS, substitutes module name in `.tmpl`/`.go`/`.templ` files
- `internal/scaffold/scaffold_test.go` tests module substitution, import rewriting, verbatim copy

## Template

- `_template/` is a production-ready web app skeleton — edits to it become the default for new projects
- Module is placeholder `kumi`; scaffold rewrites `kumi/` -> user's module path
- `.tmpl` files are processed through `text/template` for dynamic content; `.go`/`.templ` files get string replacement

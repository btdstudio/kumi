package scaffold

import (
	"os"
	"os/exec"
	"testing"

	goph "github.com/briantimmer/goph"
)

func skipOrFail(t *testing.T, format string, args ...any) {
	t.Helper()
	if os.Getenv("CI") != "" {
		t.Fatalf(format, args...)
	}
	t.Skipf(format, args...)
}

func TestScaffold_TemplateBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping template build in short mode")
	}

	dir := t.TempDir()
	if err := Scaffold(goph.TemplateFS, dir, "example.com/app"); err != nil {
		t.Fatalf("Scaffold failed: %v", err)
	}

	if err := TemplGenerate(dir); err != nil {
		skipOrFail(t, "templ generate unavailable: %v", err)
	}
	if err := GoModTidy(dir); err != nil {
		skipOrFail(t, "go mod tidy failed (offline?): %v", err)
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build ./... failed: %v\n%s", err, out)
	}
}

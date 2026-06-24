package scaffold

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed all:testdata
var testFS embed.FS

func TestScaffold_NoModuleSubstitution(t *testing.T) {
	dir := t.TempDir()
	err := ScaffoldRoot(testFS, "testdata", dir, "example.com/app")
	if err != nil {
		t.Fatalf("Scaffold failed: %v", err)
	}

	// Verify go.mod was processed from .tmpl
	b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatalf("go.mod not found: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "module example.com/app") {
		t.Errorf("expected 'module example.com/app' in go.mod, got:\n%s", content)
	}
	if strings.Contains(content, "goph") {
		t.Errorf("go.mod should not contain 'goph':\n%s", content)
	}
}

func TestScaffold_ImportReplacement(t *testing.T) {
	dir := t.TempDir()
	err := ScaffoldRoot(testFS, "testdata", dir, "example.com/app")
	if err != nil {
		t.Fatalf("Scaffold failed: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatalf("main.go not found: %v", err)
	}
	content := string(b)
	if strings.Contains(content, `"goph/`) {
		t.Errorf("main.go should not contain 'goph/' import:\n%s", content)
	}
	if !strings.Contains(content, `"example.com/app/internal/server"`) {
		t.Errorf("expected updated import in main.go:\n%s", content)
	}
}

func TestScaffold_CopyVerbatim(t *testing.T) {
	dir := t.TempDir()
	err := ScaffoldRoot(testFS, "testdata", dir, "example.com/app")
	if err != nil {
		t.Fatalf("Scaffold failed: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(dir, "static", "test.txt"))
	if err != nil {
		t.Fatalf("test.txt not found: %v", err)
	}
	if string(b) != "hello\n" {
		t.Errorf("unexpected content: %q", string(b))
	}
}

func TestScaffold_TmplStripped(t *testing.T) {
	dir := t.TempDir()
	err := ScaffoldRoot(testFS, "testdata", dir, "example.com/app")
	if err != nil {
		t.Fatalf("Scaffold failed: %v", err)
	}

	// .tmpl file should NOT exist after scaffold
	if _, err := os.Stat(filepath.Join(dir, "go.mod.tmpl")); !os.IsNotExist(err) {
		t.Error("go.mod.tmpl should have been removed")
	}
}

func TestScaffold_DirectoryCreated(t *testing.T) {
	dir := t.TempDir()
	err := ScaffoldRoot(testFS, "testdata", dir, "test.app")
	if err != nil {
		t.Fatalf("Scaffold failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	expected := []string{"go.mod", "main.go", "static", ".testfile"}
	for _, e := range expected {
		found := false
		for _, n := range names {
			if n == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q in output, got %v", e, names)
		}
	}
}

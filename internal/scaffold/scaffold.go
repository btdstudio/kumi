package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type Data struct {
	ModuleName string
}

func Scaffold(fsys embed.FS, targetDir, moduleName string) error {
	return ScaffoldRoot(fsys, "_template", targetDir, moduleName)
}

func ScaffoldRoot(fsys embed.FS, root, targetDir, moduleName string) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, root+"/")
		if relPath == path {
			relPath = strings.TrimPrefix(path, root)
		}
		if relPath == "" {
			return nil
		}
		targetPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		content, err := fsys.ReadFile(path)
		if err != nil {
			return err
		}

		if strings.HasSuffix(relPath, ".tmpl") {
			targetPath = strings.TrimSuffix(targetPath, ".tmpl")
			tmpl, err := template.New("").Parse(string(content))
			if err != nil {
				return err
			}
			var buf strings.Builder
			if err := tmpl.Execute(&buf, Data{ModuleName: moduleName}); err != nil {
				return err
			}
			processed := buf.String()
			processed = strings.ReplaceAll(processed, "goph/", moduleName+"/")
			processed = strings.ReplaceAll(processed, "module goph", "module "+moduleName)
			return os.WriteFile(targetPath, []byte(processed), 0644)
		}

		ext := filepath.Ext(relPath)
		if ext == ".go" || ext == ".templ" {
			contentStr := string(content)
			contentStr = strings.ReplaceAll(contentStr, "\"goph/", "\""+moduleName+"/")
			return os.WriteFile(targetPath, []byte(contentStr), 0644)
		}

		mode := os.FileMode(0644)
		if ext == ".sh" {
			mode = 0755
		}
		return os.WriteFile(targetPath, content, mode)
	})
}

func GoModTidy(targetDir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func TemplGenerate(targetDir string) error {
	binary, err := exec.LookPath("templ")
	if err != nil {
		binary = filepath.Join(os.Getenv("HOME"), "go", "bin", "templ")
		if _, statErr := os.Stat(binary); statErr != nil {
			return fmt.Errorf("templ not found in PATH or ~/go/bin/templ: %w", err)
		}
	}
	cmd := exec.Command(binary, "generate")
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GitInit(targetDir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/btdstudio/goph"
	"github.com/btdstudio/goph/internal/scaffold"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new <name> [module-path]",
	Short: "Scaffold a new Go + HTMX web application",
	Long: `Creates a new project directory and scaffolds a production-ready
Go + HTMX + Bootstrap 5 + Catppuccin web application into it.

<name> is the directory name (e.g. "myapp").
[module-path] is the Go module path (default: <name>).

Examples:
  goph new myapp
  goph new myapp github.com/user/myapp`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		moduleName := name
		if len(args) > 1 {
			moduleName = args[1]
		}

		targetDir, err := filepath.Abs(name)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}

		if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
			return fmt.Errorf("directory %q already exists", targetDir)
		}

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}

		fmt.Printf("  Scaffolding %q into %s ...\n", moduleName, targetDir)
		if err := scaffold.Scaffold(goph.TemplateFS, targetDir, moduleName); err != nil {
			return fmt.Errorf("scaffold: %w", err)
		}

		if err := scaffold.TemplGenerate(targetDir); err != nil {
			return fmt.Errorf("templ generate: %w (install with 'go install github.com/a-h/templ/cmd/templ@latest')", err)
		}

		fmt.Println("  Running go mod tidy ...")
		if err := scaffold.GoModTidy(targetDir); err != nil {
			return fmt.Errorf("go mod tidy: %w", err)
		}

		if _, err := os.Stat(filepath.Join(targetDir, ".git")); os.IsNotExist(err) {
			fmt.Println("  Initializing git repository ...")
			if err := scaffold.GitInit(targetDir); err != nil {
				return fmt.Errorf("git init: %w", err)
			}
		}

		fmt.Println()
		fmt.Println("  Done! To get started:")
		fmt.Println()
		fmt.Printf("    cd %s\n", name)
		fmt.Println("    goph db migrate    # run database migrations")
		fmt.Println("    goph dev           # start development server")
		fmt.Println()
		return nil
	},
}

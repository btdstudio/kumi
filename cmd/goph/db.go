package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(migrateCmd)
}

var dbCmd = &cobra.Command{
	Use:   "db <command>",
	Short: "Manage database",
}

var migrateCmd = &cobra.Command{
	Use:   "migrate [directory]",
	Short: "Run database migrations",
	Long: `Runs goose database migrations (goose up) in the specified directory
or the current directory.

This command should be run inside a goph-generated project after
setting up the database.

Examples:
  goph db migrate
  goph db migrate /path/to/project`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		absDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("invalid directory: %w", err)
		}

		migrationsDir := filepath.Join(absDir, "migrations")
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			return fmt.Errorf("no migrations directory found at %s", migrationsDir)
		}

		driver := os.Getenv("GOOSE_DRIVER")
		if driver == "" {
			driver = "postgres"
		}
		dbstring := os.Getenv("GOOSE_DBSTRING")
		if dbstring == "" {
			dbstring = os.Getenv("DATABASE_URL")
		}
		if dbstring == "" {
			return fmt.Errorf("neither GOOSE_DBSTRING nor DATABASE_URL is set")
		}

		if !findBinary("goose") {
			return fmt.Errorf("goose not found — run 'goph doctor --install' or 'go install github.com/pressly/goose/v3/cmd/goose@latest'")
		}

		fmt.Printf("  Running goose %s up in %s ...\n", driver, migrationsDir)

		goose := exec.Command("goose", "-dir", migrationsDir, "up")
		goose.Dir = absDir
		goose.Stdout = os.Stdout
		goose.Stderr = os.Stderr
		goose.Env = append(os.Environ(), "GOOSE_DRIVER="+driver, "GOOSE_DBSTRING="+dbstring)

		return goose.Run()
	},
}

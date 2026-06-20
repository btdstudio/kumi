package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(devCmd)
}

var devCmd = &cobra.Command{
	Use:   "dev [directory]",
	Short: "Start the development server with hot-reload",
	Long: `Starts the air hot-reload server in the specified directory
or the current directory.

This command should be run inside a goph-generated project.
It requires air to be installed (go install github.com/air-verse/air@latest).

Examples:
  goph dev
  goph dev /path/to/project`,
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

		if _, err := os.Stat(filepath.Join(absDir, ".air.toml")); os.IsNotExist(err) {
			return fmt.Errorf("no .air.toml found at %s (not a goph project?)", absDir)
		}

		if !findBinary("air") {
			return fmt.Errorf("air not found — run 'goph doctor --install' or 'go install github.com/air-verse/air@latest'")
		}

		fmt.Printf("  Starting dev server in %s ...\n", absDir)

		air := exec.Command("air")
		air.Dir = absDir
		air.Stdout = os.Stdout
		air.Stderr = os.Stderr
		air.Stdin = os.Stdin

		return air.Run()
	},
}

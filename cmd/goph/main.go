package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goph",
	Short: "Scaffold Go + HTMX + Bootstrap 5 + Catppuccin web applications",
	Long: `goph is a CLI tool that scaffolds production-ready Go web applications
with HTMX, Bootstrap 5, and Catppuccin theming.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

type tool struct {
	Name   string
	Pkg    string
	Binary string
}

var requiredTools = []tool{
	{"templ", "github.com/a-h/templ/cmd/templ@latest", "templ"},
	{"air", "github.com/air-verse/air@latest", "air"},
	{"goose", "github.com/pressly/goose/v3/cmd/goose@latest", "goose"},
}

func init() {
	doctorCmd.Flags().BoolP("install", "i", false, "Install missing tools")
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check and install required tools",
	Long: `Checks that all tools required by goph projects are installed.

If --install is given, missing tools are automatically installed
via 'go install'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		install, _ := cmd.Flags().GetBool("install")

		allFound := true
		for _, t := range requiredTools {
			found := findBinary(t.Binary)
			if found {
				fmt.Printf("  ✓ %s found\n", t.Name)
			} else if install {
				fmt.Printf("  Installing %s ...\n", t.Name)
				if err := installTool(t); err != nil {
					fmt.Fprintf(os.Stderr, "  ✗ %s install failed: %v\n", t.Name, err)
					allFound = false
				} else {
					fmt.Printf("  ✓ %s installed\n", t.Name)
				}
			} else {
				fmt.Fprintf(os.Stderr, "  ✗ %s not found — run: go install %s\n", t.Name, t.Pkg)
				allFound = false
			}
		}

		if !allFound {
			if !install {
				fmt.Println()
				fmt.Println("  Run 'goph doctor --install' to install missing tools.")
			}
			os.Exit(1)
		}
		return nil
	},
}

func findBinary(name string) bool {
	if _, err := exec.LookPath(name); err == nil {
		return true
	}
	home, err := os.UserHomeDir()
	if err == nil {
		if _, err := os.Stat(filepath.Join(home, "go", "bin", name)); err == nil {
			return true
		}
	}
	return false
}

func installTool(t tool) error {
	cmd := exec.Command("go", "install", t.Pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

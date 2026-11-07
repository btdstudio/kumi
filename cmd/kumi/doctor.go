package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/btdstudio/kumi"
	"github.com/btdstudio/kumi/internal/scaffold"
	"github.com/spf13/cobra"
)

type tool struct {
	Name   string
	Pkg    string
	Binary string
	// Version is the exact version the tool must report, or empty if any
	// version will do.
	Version string
}

// requiredTools lists the tools kumi projects need. templ is pinned to the
// template's version because its generated code only compiles against a
// matching runtime.
func requiredTools() ([]tool, error) {
	templVersion, err := scaffold.TemplVersion(kumi.TemplateFS)
	if err != nil {
		return nil, err
	}
	return []tool{
		{"templ", "github.com/a-h/templ/cmd/templ@" + templVersion, "templ", templVersion},
		{"air", "github.com/air-verse/air@latest", "air", ""},
		{"goose", "github.com/pressly/goose/v3/cmd/goose@latest", "goose", ""},
	}, nil
}

func init() {
	doctorCmd.Flags().BoolP("install", "i", false, "Install missing or wrong-version tools")
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check and install required tools",
	Long: `Checks that all tools required by kumi projects are installed, and that
templ matches the version kumi's template uses.

If --install is given, missing or wrong-version tools are automatically
installed via 'go install'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		install, _ := cmd.Flags().GetBool("install")

		tools, err := requiredTools()
		if err != nil {
			return err
		}

		allOK := true
		for _, t := range tools {
			problem := checkTool(t)
			if problem == "" {
				fmt.Printf("  ✓ %s found\n", t.Name)
				continue
			}
			if !install {
				fmt.Fprintf(os.Stderr, "  ✗ %s %s — run: go install %s\n", t.Name, problem, t.Pkg)
				allOK = false
				continue
			}

			fmt.Printf("  Installing %s ...\n", t.Name)
			if err := installTool(t); err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ %s install failed: %v\n", t.Name, err)
				allOK = false
				continue
			}
			// Another copy earlier on PATH can shadow the one just installed.
			if problem := checkTool(t); problem != "" {
				fmt.Fprintf(os.Stderr, "  ✗ %s %s after install — check your PATH\n", t.Name, problem)
				allOK = false
				continue
			}
			fmt.Printf("  ✓ %s installed\n", t.Name)
		}

		if !allOK {
			if !install {
				fmt.Println()
				fmt.Println("  Run 'kumi doctor --install' to install missing or wrong-version tools.")
			}
			os.Exit(1)
		}
		return nil
	},
}

// checkTool describes what is wrong with an installed tool, or returns an empty
// string if it is usable.
func checkTool(t tool) string {
	path := findBinary(t.Binary)
	if path == "" {
		return "not found"
	}
	if t.Version == "" {
		return ""
	}
	out, err := exec.Command(path, "version").Output()
	if err != nil {
		return fmt.Sprintf("version unknown (need %s)", t.Version)
	}
	if got := strings.TrimSpace(string(out)); got != t.Version {
		return fmt.Sprintf("is %s (need %s)", got, t.Version)
	}
	return ""
}

func findBinary(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	home, err := os.UserHomeDir()
	if err == nil {
		path := filepath.Join(home, "go", "bin", name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func installTool(t tool) error {
	cmd := exec.Command("go", "install", t.Pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

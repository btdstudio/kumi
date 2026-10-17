package main

import (
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is the current version of the CLI, set during build time with ldflags,
// or dynamically fetched via debug.ReadBuildInfo() at runtime if installed via go install.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "goph",
	Short: "Scaffold Go + HTMX + Bootstrap 5 + Catppuccin web applications",
	Long: `goph is a CLI tool that scaffolds production-ready Go web applications
with HTMX, Bootstrap 5, and Catppuccin theming.`,
}

func init() {
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				Version = info.Main.Version
			} else {
				// Fallback to VCS build settings (revision, dirty status)
				var revision string
				var modified bool
				for _, setting := range info.Settings {
					switch setting.Key {
					case "vcs.revision":
						revision = setting.Value
					case "vcs.modified":
						modified = setting.Value == "true"
					}
				}
				if revision != "" {
					shortSHA := revision
					if len(shortSHA) > 7 {
						shortSHA = shortSHA[:7]
					}
					Version = "dev-" + shortSHA
					if modified {
						Version += "+dirty"
					}
				}
			}
		}
	}
	rootCmd.Version = Version
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

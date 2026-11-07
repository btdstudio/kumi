package scaffold

import (
	"fmt"
	"io/fs"
	"strings"
)

const templModule = "github.com/a-h/templ"

// TemplVersion returns the templ version pinned in the template's go.mod.tmpl.
// Generated templ code only compiles against a matching runtime, so anything
// that installs the templ generator must use this version.
func TemplVersion(fsys fs.FS) (string, error) {
	content, err := fs.ReadFile(fsys, "_template/go.mod.tmpl")
	if err != nil {
		return "", fmt.Errorf("read template go.mod: %w", err)
	}
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		for i := 0; i+1 < len(fields); i++ {
			if fields[i] == templModule {
				return fields[i+1], nil
			}
		}
	}
	return "", fmt.Errorf("%s not found in template go.mod", templModule)
}

package scaffold

import (
	"strings"
	"testing"
	"testing/fstest"

	kumi "github.com/btdstudio/kumi"
)

func TestTemplVersion_EmbeddedTemplate(t *testing.T) {
	version, err := TemplVersion(kumi.TemplateFS)
	if err != nil {
		t.Fatalf("TemplVersion failed: %v", err)
	}
	if !strings.HasPrefix(version, "v") {
		t.Errorf("version = %q, want a v-prefixed version", version)
	}
}

func TestTemplVersion_Forms(t *testing.T) {
	tests := []struct {
		name    string
		gomod   string
		want    string
		wantErr bool
	}{
		{"require block", "module kumi\n\nrequire (\n\tgithub.com/a-h/templ v1.2.3\n)\n", "v1.2.3", false},
		{"single line", "module kumi\n\nrequire github.com/a-h/templ v1.2.3\n", "v1.2.3", false},
		{"similar module", "require (\n\tgithub.com/a-h/templ-extra v9.9.9\n)\n", "", true},
		{"missing", "module kumi\n", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fstest.MapFS{"_template/go.mod.tmpl": {Data: []byte(tt.gomod)}}
			got, err := TemplVersion(fsys)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("version = %q, want %q", got, tt.want)
			}
		})
	}
}

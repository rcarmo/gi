package docs

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

func TestTUISelfDocumentationIsEmbeddedAndLinksResolve(t *testing.T) {
	reference := InternalReferenceFS()
	index, err := fs.ReadFile(reference, "README.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"tui.md", "tui-pi-1.5-alignment.md", "keybindings.md", "prompt-caching.md"} {
		if !strings.Contains(string(index), "("+name+")") {
			t.Errorf("reference index does not expose %s", name)
		}
		if _, err := fs.ReadFile(reference, name); err != nil {
			t.Error(err)
		}
	}
	guide, err := fs.ReadFile(reference, "tui.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, match := range regexp.MustCompile(`\]\(([^)#]+)(?:#[^)]*)?\)`).FindAllStringSubmatch(string(guide), -1) {
		if _, err := fs.Stat(reference, match[1]); err != nil {
			t.Errorf("shipped TUI guide has unresolved reference %s: %v", match[1], err)
		}
	}
}

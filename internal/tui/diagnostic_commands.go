package tui

import (
	"encoding/json"
	"io"
	"runtime"
	"strings"

	gidocs "github.com/rcarmo/gi/docs"
	"github.com/rcarmo/gi/internal/version"
)

const diagnosticLimit = 64 * 1024

func changelogLines(args []string) []string {
	if len(args) > 1 {
		return []string{"usage: /changelog [version]"}
	}
	name := "unreleased"
	if len(args) == 1 {
		name = strings.TrimPrefix(args[0], "v")
	}
	// Only bundled names are accepted; never form a path from user input.
	files := map[string]string{"unreleased": "changelog-unreleased.md"}
	file, ok := files[name]
	if !ok {
		return []string{"changelog: no bundled notes for that version; available: unreleased"}
	}
	data, err := gidocs.InternalReferenceFS().Open(file)
	if err != nil {
		return []string{"changelog: bundled notes unavailable"}
	}
	defer data.Close()
	// Embedded assets are bounded independently of filesystem or network input.
	body, err := io.ReadAll(io.LimitReader(data, diagnosticLimit+1))
	if len(body) > diagnosticLimit {
		return []string{"changelog: bundled notes exceed 64 KiB"}
	}
	if err != nil || len(body) == 0 {
		return []string{"changelog: notes unavailable"}
	}
	return strings.Split(strings.TrimSpace(string(body)), "\n")
}

func (c *chatTUI) bugPreviewLines(args []string) []string {
	if len(args) > 1 || len(args) == 1 && args[0] != "preview" {
		return []string{"usage: /bug [preview] (local preview only; no save or upload)"}
	}
	mode := "fullscreen"
	if c.regularMode {
		mode = "regular"
	}
	// Do not marshal RuntimeConfig: it contains paths, provider endpoints and
	// user-defined values. This schema is deliberately small and allowlisted.
	report := struct {
		Version    string `json:"version"`
		Go         string `json:"go"`
		OS         string `json:"os"`
		Arch       string `json:"arch"`
		TUI        string `json:"tui_mode"`
		Running    bool   `json:"running"`
		Compaction bool   `json:"compacting"`
	}{version.String(), runtime.Version(), runtime.GOOS, runtime.GOARCH, mode, c.running, c.compaction.active}
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil || len(body) > diagnosticLimit {
		return []string{"bug: diagnostics unavailable"}
	}
	lines := []string{"bug: local preview only; nothing saved or uploaded", "Review before copying to https://github.com/rcarmo/gi/issues/new", "No credentials, paths, endpoints, prompts or session identifiers are included."}
	return append(lines, strings.Split(string(body), "\n")...)
}

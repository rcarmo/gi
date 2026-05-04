package turn

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/routing"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestDiscoverExtensionScripts(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".gi", "extensions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "smart.joke"), []byte("nil"), 0o644); err != nil {
		t.Fatal(err)
	}
	ext := discoverExtensionScripts(root)
	if len(ext) != 1 || ext[0].Engine != "joker" || ext[0].Path != filepath.Join(".gi", "extensions", "smart.joke") {
		t.Fatalf("extensions: %#v", ext)
	}
}

func TestJokerSmartCompactionExtension(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".gi", "extensions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := `(gi-register-event-hook {:name "session_before_compact" :engine "joker" :script "(json/write-string {:payload {:summary (str \"joker smart summary: \" (get-in *gi-hook* [:payload :preparation :messages_to_summarize]))}})"})`
	if err := os.WriteFile(filepath.Join(dir, "smart-compaction.joke"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cfg := config.RuntimeConfig{WorkspaceRoot: root, DefaultModel: "bootstrap", MaxIterations: 64, Compaction: config.CompactionSettings{Enabled: true, ContextWindow: 1000, ThresholdTokens: 50, KeepRecentTokens: 20, ReserveTokens: 10}, Agents: routing.AgentsConfig{List: []routing.AgentConfig{{ID: "agent", Default: true, Model: "bootstrap"}}}}
	e := NewWithRuntimeConfig(s, cfg, "")
	r := &sessionRunner{store: s, engine: e}
	conv := &goai.Context{Messages: []goai.Message{
		goai.UserMessage(strings.Repeat("older1 ", 80)),
		goai.UserMessage(strings.Repeat("older2 ", 80)),
		goai.UserMessage(strings.Repeat("older3 ", 80)),
		goai.UserMessage(strings.Repeat("older4 ", 80)),
		goai.UserMessage("recent question"),
		goai.UserMessage("recent answer"),
	}}
	r.maybeCompactContext(context.Background(), "missing-session", "turn_ext", "agent", "bootstrap", conv)
	got := goai.GetTextContent(&conv.Messages[0])
	if !strings.Contains(got, "joker smart summary:") {
		t.Fatalf("missing joker smart summary: %q", got)
	}
}

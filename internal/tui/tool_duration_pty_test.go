package tui

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

// Test-only terminal entrypoint: the shell harness owns the isolated directory.
func TestRecordedDurationPTYFixture(t *testing.T) {
	dir := os.Getenv("GI_DURATION_PTY_DIR")
	if dir == "" {
		t.Skip("make test-tool-duration-pty")
	}
	cfg := config.Load(dir)
	cfg.DefaultModel = "test-model"
	cfg.QuietStartup = "true"
	s, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateDefaultSession(defaultTUIAgentID, "gi", "default", "duration-session")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "duration-session", Title: "timing", State: map[string]any{"model": "test-model"}, Allocation: alloc})
	if err != nil {
		t.Fatal(err)
	}
	sid := sess.ID
	if err := s.AddMessage(ctx, "timed-result", sid, "tool_result", "recorded output", map[string]any{"tool_name": "shell", "duration_ms": 1500, "arguments": map[string]any{"command": "echo timed"}}); err != nil {
		t.Fatal(err)
	}
	engine := turn.NewWithRuntimeConfig(s, cfg, cfg.SystemPrompt)
	defer engine.Close()
	if err := runWithEngineMode(s, engine, cfg, false); err != nil {
		t.Fatal(err)
	}
}

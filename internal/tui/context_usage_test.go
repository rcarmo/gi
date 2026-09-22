package tui

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
)

func TestTerminalContextUsesMeasuredRequestWithoutAdditionalRows(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(filepath.Join(t.TempDir(), "context.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.CreateSession(ctx, "A", "A", map[string]any{"model": "opencode-zen/minimax-m2.5-free"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "turn", "A", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	c := &chatTUI{store: s, sessionID: "A", cfg: config.RuntimeConfig{DefaultModel: "opencode-zen/minimax-m2.5-free"}}
	baseline := len(c.footerLines(100))
	c.updateUsageFromPayload(map[string]any{"usage": map[string]any{"input": 96000, "output": 12000, "total": 108000}})
	if c.contextSummaryData().contextTokens != 0 {
		t.Fatal("cumulative billing became context")
	}
	if err := s.AppendTurnEvent(ctx, "turn", "A", "context.measured", map[string]any{"tokens": 16000, "input": 14000, "cache_read": 2000, "model": "opencode-zen/minimax-m2.5-free"}); err != nil {
		t.Fatal(err)
	}
	if c.contextSummaryData().contextTokens != 16000 {
		t.Fatal("measurement not shown")
	}
	if len(c.footerLines(100)) != baseline {
		t.Fatal("measurement added idle rows")
	}
}

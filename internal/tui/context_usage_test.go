package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	gotui "github.com/grindlemire/go-tui"

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
	baseline := map[int]int{}
	for _, width := range []int{60, 100, 140} {
		baseline[width] = len(c.footerLines(width))
	}
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
	for i, tokens := range []int{96000, 115200, 160000, 0} {
		if err := s.AppendTurnEvent(ctx, "turn", "A", "context.measured", map[string]any{"tokens": tokens, "input": tokens, "model": "opencode-zen/minimax-m2.5-free", "iteration": i + 2}); err != nil {
			t.Fatal(err)
		}
		data := c.contextSummaryData()
		if data.contextTokens != tokens || data.contextWindow != 128000 {
			t.Fatalf("measurement/capacity: %+v", data)
		}
		if tokens == 160000 && formatContextUsage(data.contextTokens, data.contextWindow) != "ctx 160K/128K 125%" {
			t.Fatal("overflow percentage was clamped")
		}
		for _, width := range []int{60, 100, 140} {
			t.Run(fmt.Sprintf("tokens-%d-width-%d", tokens, width), func(t *testing.T) {
				lines := c.footerLines(width)
				if len(lines) != baseline[width] {
					t.Fatal("context added idle footer rows", lines)
				}
				for _, line := range lines {
					if gotui.StringWidth(line) > width {
						t.Fatal("context footer wrapped", line)
					}
				}
			})
		}
	}
}

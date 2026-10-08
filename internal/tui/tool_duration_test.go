package tui

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
)

// A recorded execution duration must win over delayed event delivery and the
// timestamp of the saved result. Zero is a valid measurement; absence is not.
func TestRecordedToolDurationLiveAndRestored(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value any
		want  string
	}{
		{"measured", float64(1500), "Took 1.5s"},
		{"zero", float64(0), "Took 0.0s"},
		{"legacy", nil, ""},
		{"negative", float64(-1), ""},
		{"fractional", float64(1.5), ""},
		{"nan", math.NaN(), ""},
		{"infinite", math.Inf(1), ""},
		{"overflow", float64(math.MaxInt64), ""},
		{"string", "1500", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			payload := map[string]any{"tool_name": "shell", "duration_ms": tc.value}
			live := &chatTUI{cfg: config.RuntimeConfig{}, transcriptExpanded: map[string]bool{}}
			live.renderToolEvent(map[string]any{"type": "tool_started", "tool": "shell", "turn_id": "t", "tool_call_id": "call"}, start)
			live.renderToolEvent(map[string]any{"type": "tool_finished", "tool": "shell", "turn_id": "t", "tool_call_id": "call", "duration_ms": tc.value, "output": "ok"}, start.Add(9*time.Second))
			restored := &chatTUI{cfg: config.RuntimeConfig{}, transcriptExpanded: map[string]bool{}}
			restored.transcript = restored.renderToolResultWithArguments(store.Message{ID: "result", Role: "tool_result", Content: "ok", CreatedAt: start.Add(time.Hour).Format(time.RFC3339Nano), Payload: payload}, nil)
			for name, c := range map[string]*chatTUI{"live": live, "restored": restored} {
				t.Run(name, func(t *testing.T) {
					meta := transcriptLastBlockMeta(t, c.transcript)
					_, screen, _ := renderToolForTest(t, c, meta, []string{"ok"}, 60)
					if tc.want == "" {
						if strings.Contains(screen, "Took") || strings.Contains(screen, "Elapsed") {
							t.Fatalf("%s fabricated final timing: %q", name, screen)
						}
					} else if !strings.Contains(screen, tc.want) {
						t.Fatalf("%s: expected %q, got %q", name, tc.want, screen)
					}
				})
			}
		})
	}
}

func TestRecordedToolDurationOccurrenceAndTerminalFencing(t *testing.T) {
	c := &chatTUI{transcriptExpanded: map[string]bool{}}
	at := time.Now()
	event := func(kind, occurrence string, ms any) map[string]any {
		return map[string]any{"type": kind, "tool": "shell", "turn_id": "t", "tool_call_id": "same", "occurrence_id": occurrence, "duration_ms": ms}
	}
	c.renderToolEvent(event("tool_started", "old", nil), at)
	c.renderToolEvent(event("tool_finished", "old", int64(1500)), at.Add(time.Second))
	c.renderToolEvent(event("tool_started", "new", nil), at)
	c.renderToolEvent(event("tool_cancelled", "new", int64(2500)), at.Add(9*time.Second))
	key := c.transcriptToolBlocks[c.toolRuntimeBlockKey(event("", "new", nil), "shell")]
	before := strings.Join(c.transcript, "\n")
	c.renderToolEvent(event("tool_finished", "new", int64(9000)), at.Add(time.Minute))
	if strings.Join(c.transcript, "\n") != before {
		t.Fatal("late event changed terminal occurrence")
	}
	meta, _ := parseTranscriptBlockMarker(c.transcript[c.transcriptBlockSpans[key].HeaderIndex])
	if meta.Status != "cancelled" || meta.DurationMS == nil || *meta.DurationMS != 2500 {
		t.Fatal(meta)
	}
	blocks := c.buildTranscriptRenderableBlocks(c.transcript)
	if len(blocks) != 2 {
		t.Fatalf("reused call merged occurrences: %d", len(blocks))
	}
	a := blockHeightKey(blocks[1], "", 60)
	ms := int64(1)
	blocks[1].DurationMS = &ms
	if blockHeightKey(blocks[1], "", 60) == a {
		t.Fatal("timing omitted from render cache key")
	}
}

func TestRecordedStoppedDurationReload(t *testing.T) {
	c := sessionTestChat(t)
	ctx := context.Background()
	if _, err := c.store.CreateTurn(ctx, "timing-turn", "A", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"tool.started", "tool.cancelled"} {
		if err := c.store.AppendTurnEvent(ctx, "timing-turn", "A", kind, map[string]any{"tool": "shell", "tool_call_id": "call", "occurrence_id": "occ", "duration_ms": 1500}); err != nil {
			t.Fatal(err)
		}
	}
	c.transcript = c.loadTranscript()
	meta := transcriptLastBlockMeta(t, c.transcript)
	if meta.Status != "cancelled" || meta.DurationMS == nil || *meta.DurationMS != 1500 {
		t.Fatal(meta)
	}
	_, screen, _ := renderToolForTest(t, c, meta, []string{"stopped"}, 60)
	if !strings.Contains(screen, "Took 1.5s") {
		t.Fatal(screen)
	}
}

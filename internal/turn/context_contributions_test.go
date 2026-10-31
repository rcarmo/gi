package turn

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestContextContributionStopsAcrossReopenForkAndLegacy(t *testing.T) {
	for _, format := range []string{"direct", "envelope", "legacy"} {
		t.Run(format, func(t *testing.T) {
			ctx := context.Background()
			path := filepath.Join(t.TempDir(), "contributions.db")
			s, err := store.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { s.Close() }()
			if _, err = s.CreateSession(ctx, "source", "source", nil); err != nil {
				t.Fatal(err)
			}
			stops := []goai.StopReason{goai.StopReasonPending, goai.StopReasonAborted, goai.StopReasonError, goai.StopReasonDeferred, goai.StopReasonStop}
			for i, stop := range stops {
				message := replayAnswer(string(stop))
				message.StopReason = stop
				payload := map[string]any{}
				switch format {
				case "direct":
					payload["provider_message"] = message
				case "envelope":
					payload["provider_message"] = &inference.StreamResult{Message: &message, Text: string(stop)}
				case "legacy":
					payload["stop_reason"] = string(stop)
				}
				id := fmt.Sprintf("receipt-%d", i)
				if err = s.AddMessage(ctx, id, "source", "assistant", string(stop), payload); err != nil {
					t.Fatal(err)
				}
				if _, err = s.DB().Exec(`update messages set created_at=? where id=?`, fmt.Sprintf("2026-01-01T00:00:0%dZ", i), id); err != nil {
					t.Fatal(err)
				}
			}
			if _, err = s.CloneSessionBefore(ctx, "source", "fork", "fork", "", ""); err != nil {
				t.Fatal(err)
			}
			if err = s.Close(); err != nil {
				t.Fatal(err)
			}
			s, err = store.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			r := &sessionRunner{}
			for _, id := range []string{"source", "fork"} {
				snapshot, err := s.ContextSnapshot(ctx, id)
				if err != nil {
					t.Fatal(err)
				}
				projected := r.projectContextSnapshot(ctx, id, snapshot)
				if len(projected) != 2 || projected[0].StopReason != goai.StopReasonPending || projected[1].StopReason != goai.StopReasonStop {
					t.Fatalf("%s contributes excluded receipts: %+v", id, projected)
				}
				raw, err := s.ListMessages(ctx, id)
				if err != nil || len(raw) != 5 {
					t.Fatalf("raw receipts lost: %+v %v", raw, err)
				}
				// Compaction counts projected messages, but durable coverage must include
				// the three excluded native receipts as well.
				boundary, err := r.compactionBoundary(ctx, id, &goai.Context{Messages: projected}, snapshot, map[string]any{"messages_to_summarize": 2})
				if err != nil || boundary == nil || len(boundary.Covered) != 5 {
					t.Fatalf("boundary: %+v %v", boundary, err)
				}
			}
		})
	}
}

func TestContextReducerPairsWithinBatchesAndSynthesizesMissingResults(t *testing.T) {
	call := func(ids ...string) goai.Message {
		m := goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse}
		for _, id := range ids {
			m.Content = append(m.Content, goai.ContentBlock{Type: "toolCall", ID: id, Name: "shell", Arguments: map[string]any{}})
		}
		return m
	}
	result := func(id, text string) goai.Message {
		return goai.Message{Role: goai.RoleToolResult, ToolCallID: id, ToolName: "shell", Content: []goai.ContentBlock{{Type: "text", Text: text}}}
	}
	aborted := call("excluded")
	aborted.StopReason = goai.StopReasonAborted
	source := []goai.Message{aborted, result("excluded", "must be dropped"), call("a", "b"), result("b", "B"), result("a", "A"), result("a", "duplicate"), result("unknown", "orphan"), goai.UserMessage("boundary"), call("a"), goai.UserMessage("another boundary"), call("a"), result("a", "later batch")}
	got, boundaries := reduceContextContributions(source)
	if len(got) != 9 {
		t.Fatalf("wrong context: %+v", got)
	}
	if goai.GetTextContent(&got[1]) != "A" || goai.GetTextContent(&got[2]) != "B" {
		t.Fatal("tool results not ordered", got)
	}
	if !got[5].IsError || got[5].Details.(map[string]any)["reason"] != "missing_result" {
		t.Fatalf("missing receipt not repaired: %+v", got[5])
	}
	if goai.GetTextContent(&got[8]) != "later batch" {
		t.Fatal("reused ID stolen across boundary", got)
	}
	if _, ok := boundaries[1]; ok {
		t.Fatal("compaction split first batch")
	}
	if boundaries[3] != 7 || boundaries[6] != 9 {
		t.Fatal("source boundaries lost", boundaries)
	}
}

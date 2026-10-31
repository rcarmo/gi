package turn

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func testProviderPartial() *goai.Message {
	return &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonPending, Usage: &goai.Usage{Input: 17}, Content: []goai.ContentBlock{{Type: "text", Text: "streamed answer"}, {Type: "thinking", Thinking: "streamed reasoning", ThinkingSignature: "signature"}, {Type: "toolCall", ID: "prefix", Name: "danger", Arguments: map[string]any{"prefix": "unfinished"}, ArgumentsError: "incomplete"}}}
}
func TestProviderPartialFailureAbortAndIncomplete(t *testing.T) {
	for _, mode := range []string{"error", "abort", "incomplete"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			s := openTestStore(t)
			defer s.Close()
			e := New(s)
			defer e.Close()
			if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			var executed, requests atomic.Int32
			if err := e.RegisterTool(tools.RegisteredTool{Name: "danger", Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
				executed.Add(1)
				return "bad", nil
			}}); err != nil {
				t.Fatal(err)
			}
			original := streamWithToolsWithHooks
			defer func() { streamWithToolsWithHooks = original }()
			streamWithToolsWithHooks = func(requestCtx context.Context, _ string, _ *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
				requests.Add(1)
				partial := testProviderPartial()
				if err := hooks.OnProgress(partial); err != nil {
					return nil, err
				}
				// The provider is allowed to mutate its own buffers after the callback.
				partial.Content[0].Text = "mutated after observation"
				partial.Content[2].Arguments["prefix"] = "mutated"
				switch mode {
				case "abort":
					if err := e.CancelActiveTurn(ctx, "A", e.LiveActiveTurn("A")); err != nil {
						return nil, err
					}
					return nil, requestCtx.Err()
				case "incomplete":
					latest := testProviderPartial()
					latest.Content[0].Text = "streamed latest"
					return &inference.StreamResult{Partial: latest, Incomplete: true}, fmt.Errorf("stream missing terminal")
				default:
					failed := testProviderPartial()
					failed.StopReason = goai.StopReasonError
					failed.ErrorMessage = "provider failed"
					return &inference.StreamResult{Message: failed}, fmt.Errorf("provider failed")
				}
			}
			run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "test partial", Model: "provider/model"})
			if err != nil {
				t.Fatal(err)
			}
			status := "failed"
			if mode == "abort" {
				status = "cancelled"
			}
			waitForCondition(t, 3*time.Second, func() bool { rec, err := s.GetTurn(ctx, run.TurnID); return err == nil && rec.Status == status }, "partial terminal settlement")
			if executed.Load() != 0 || requests.Load() != 1 {
				t.Fatal("partial authorised tool or retry", executed.Load(), requests.Load())
			}
			messages, err := s.ListMessages(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			receipts := 0
			for _, m := range messages {
				if m.Role != "assistant" {
					continue
				}
				receipts++
				want := "streamed answer"
				if mode == "incomplete" {
					want = "streamed latest"
				}
				saved, ok := decodeAssistantReplay(m.Payload["provider_message"])
				if !ok || m.Content != want || len(saved.Content) != 3 || saved.Content[1].ThinkingSignature != "signature" || saved.Content[2].Arguments["prefix"] != "unfinished" || m.Payload["adopted"] != false {
					t.Fatal("structured partial lost or mutated", m)
				}
				if mode == "error" && saved.ErrorMessage != "provider failed" {
					t.Fatal("error response lost", saved)
				}
			}
			if receipts != 1 {
				t.Fatal("duplicate or missing partial row", receipts)
			}
			snapshot, err := s.ContextSnapshot(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			for _, m := range e.runner("A").projectContextSnapshot(ctx, "A", snapshot) {
				if strings.Contains(goai.GetTextContent(&m), "streamed") || len(goai.GetToolCalls(&m)) > 0 {
					t.Fatal("partial entered model context", m)
				}
			}
			accounting, err := s.TurnAccounting(ctx, run.TurnID)
			wantUnknown := 1
			if mode == "error" {
				wantUnknown = 0
			}
			if err != nil || accounting.Usage.Input != 17 || accounting.Unknown != wantUnknown {
				t.Fatal("partial usage lost or falsely final", accounting, err)
			}
		})
	}
}
func TestProviderPartialBoundsCoalescingAndFence(t *testing.T) {
	s, _ := intentFixture(t)
	ctx := context.Background()
	if err := s.AppendTurnEvent(ctx, "t", "A", "inference.started", map[string]any{"request_id": "request"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`create table partial_writes(n integer); create trigger count_partial after update on messages begin insert into partial_writes values(1); end`); err != nil {
		t.Fatal(err)
	}
	p := &providerPartialRecorder{store: s, ctx: ctx, turnID: "t", sessionID: "A", occurrenceID: "request", model: "model"}
	if err := p.Observe(testProviderPartial()); err != nil {
		t.Fatal(err)
	}
	p.last = time.Now().Add(time.Hour) // exercise coalescing without scheduler timing
	for i := 0; i < 100; i++ {
		if err := p.Observe(testProviderPartial()); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := s.DB().QueryRow(`select count(*) from partial_writes`).Scan(&count); err != nil || count != 0 {
		t.Fatal("raw deltas were persisted", count, err)
	}
	latest := testProviderPartial()
	latest.Content[0].Text = "final owned partial"
	if err := p.Seal(latest); err != nil {
		t.Fatal(err)
	}
	late := testProviderPartial()
	late.Content[0].Text = "late callback"
	if err := p.Observe(late); err != nil {
		t.Fatal(err)
	}
	record, err := s.ProviderPartial(ctx, "t", "A", "request")
	if err != nil || record == nil || record.Content != "final owned partial" {
		t.Fatal("post-seal callback overwrote receipt", record, err)
	}
	huge := testProviderPartial()
	huge.Content[0].Text = strings.Repeat("<&", 1<<18)
	huge.Content = append(huge.Content, goai.ContentBlock{Type: "image", Data: strings.Repeat("x", 1<<20)})
	bounded, truncated := boundedProviderPartial(huge)
	raw, err := json.Marshal(bounded)
	if err != nil || !truncated || len(raw) > 256<<10 || len(bounded.Content[0].Text) > providerPartialStringBytes {
		t.Fatal("snapshot unbounded", len(raw), truncated, err)
	}
	cyclic := map[string]any{}
	cyclic["cycle"] = cyclic
	huge = testProviderPartial()
	huge.Content[2].Arguments = cyclic
	bounded, truncated = boundedProviderPartial(huge)
	if _, err = json.Marshal(bounded); err != nil || !truncated {
		t.Fatal("recursive input escaped bounds", err, truncated)
	}
}

func TestProviderPartialMalformedErrorRoleCannotInjectContext(t *testing.T) {
	s, e := intentFixture(t)
	ctx := context.Background()
	message := goai.UserMessage("provider-injected user text")
	if err := e.runner("A").settleProviderFailure(context.Background(), s, "t", "A", "model", 1, &inference.StreamResult{OccurrenceID: "bad-role", Message: &message}, fmt.Errorf("provider error")); err != nil {
		t.Fatal(err)
	}
	snapshot, err := s.ContextSnapshot(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	if projected := e.runner("A").projectContextSnapshot(ctx, "A", snapshot); len(projected) != 0 {
		t.Fatal("error receipt injected a user contribution", projected)
	}
}

func TestProviderPartialUnencodableErrorStillRetainsContent(t *testing.T) {
	s, e := intentFixture(t)
	ctx := context.Background()
	message := testProviderPartial()
	message.StopReason = goai.StopReasonError
	message.Usage.Cost.Total = math.NaN()
	if err := e.runner("A").settleProviderFailure(context.Background(), s, "t", "A", "model", 1, &inference.StreamResult{OccurrenceID: "invalid-json", Message: message}, fmt.Errorf("provider error")); err != nil {
		t.Fatal(err)
	}
	messages, err := s.ListMessages(ctx, "A")
	if err != nil || len(messages) != 1 || messages[0].Content != "streamed answer" || messages[0].Payload["provider_message_omissions"] == nil || messages[0].Payload["partial"] != true {
		t.Fatal("unencodable usage discarded content", messages, err)
	}
	saved, ok := decodeAssistantReplay(messages[0].Payload["provider_message"])
	if !ok || saved.Usage != nil || saved.Content[1].ThinkingSignature != "signature" {
		t.Fatal("bounded structured fallback lost", saved)
	}
	a, err := s.TurnAccounting(ctx, "t")
	if err != nil || a.Unknown != 1 {
		t.Fatal("invalid usage claimed as known", a, err)
	}
}

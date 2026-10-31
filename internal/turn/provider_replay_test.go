package turn

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// Exercise real writes, then reopen SQLite and send the persisted context to
// the next provider request. Hand-built snapshot fixtures miss envelope bugs.
func TestProductionProviderReplayAfterReopenAndFork(t *testing.T) {
	for _, tool := range []bool{false, true} {
		for _, legacy := range []bool{false, true} {
			t.Run(fmt.Sprintf("tool=%v/legacy=%v", tool, legacy), func(t *testing.T) {
				ctx := context.Background()
				path := filepath.Join(t.TempDir(), "replay.db")
				s, err := store.Open(path)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { s.Close() }()
				if _, err = s.CreateSession(ctx, "source", "source", nil); err != nil {
					t.Fatal(err)
				}
				inference.Init()
				model := "production-replay/fixture"
				goai.RegisterModel(&goai.Model{ID: "fixture", Provider: "production-replay", Api: goai.ApiOpenAICompletions, ContextWindow: 32000, MaxTokens: 1024})
				saved := goai.Message{Role: goai.RoleAssistant, Provider: "production-replay", Api: goai.ApiOpenAICompletions, Model: "fixture", ResponseID: "response-1", ResponseModel: "resolved-fixture", StopReason: goai.StopReasonStop, Usage: &goai.Usage{Input: 12, Output: 7}, Content: []goai.ContentBlock{{Type: "thinking", Thinking: "reasoning", ThinkingSignature: "thinking-signed"}, {Type: "text", Text: "first answer", TextSignature: "text-signed"}}}
				if tool {
					saved.StopReason = goai.StopReasonToolUse
					saved.Content = append(saved.Content, goai.ContentBlock{Type: "toolCall", ID: "replay-call", Name: "missing-fixture-tool", Arguments: map[string]any{"value": "fixture"}, ThoughtSignature: "call-signed"})
				}
				original := streamWithToolsWithHooks
				defer func() { streamWithToolsWithHooks = original }()
				calls := 0
				streamWithToolsWithHooks = func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error) {
					calls++
					if calls == 1 {
						return &inference.StreamResult{Message: &saved, Text: "first answer", Usage: saved.Usage}, nil
					}
					end := replayAnswer("after tool")
					end.StopReason = goai.StopReasonStop
					return &inference.StreamResult{Message: &end, Text: "after tool"}, nil
				}
				e := New(s)
				run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "source", Prompt: "first", Model: model})
				if err != nil {
					t.Fatal(err)
				}
				waitForCondition(t, 3*time.Second, func() bool { r, err := s.GetTurn(ctx, run.TurnID); return err == nil && r.Status == "completed" }, "initial completion")
				e.Close()
				rows, err := s.ListMessages(ctx, "source")
				if err != nil {
					t.Fatal(err)
				}
				for _, row := range rows {
					if row.Role != "assistant" {
						continue
					}
					if legacy {
						// Old releases wrote the entire StreamResult; preserve read compatibility.
						message := saved
						if row.Payload["kind"] != "tool_calls" && tool {
							message = replayAnswer("after tool")
							message.StopReason = goai.StopReasonStop
						}
						envelope, err := json.Marshal(&inference.StreamResult{Message: &message, Text: row.Content, Usage: message.Usage})
						if err != nil {
							t.Fatal(err)
						}
						if _, err = s.DB().Exec(`update messages set payload_json=json_set(payload_json,'$.provider_message',json(?)) where id=?`, string(envelope), row.ID); err != nil {
							t.Fatal(err)
						}
					} else {
						// New writes must contain the message itself, never its stream envelope.
						m, ok := decodeReplayValue[goai.Message](row.Payload["provider_message"])
						if !ok || m.Role != goai.RoleAssistant {
							t.Fatalf("non-message replay payload: %+v", row.Payload["provider_message"])
						}
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
				// Compare protocol values after JSON round-trip, including signatures and
				// metadata, without imposing byte identity on generated text.
				raw, _ := json.Marshal(saved)
				var expected goai.Message
				if err = json.Unmarshal(raw, &expected); err != nil {
					t.Fatal(err)
				}
				observed := make(chan bool, 2)
				streamWithToolsWithHooks = func(_ context.Context, _ string, conv *goai.Context, _ func(map[string]any), _ *inference.StreamHooks) (*inference.StreamResult, error) {
					found := false
					for _, m := range conv.Messages {
						if m.Role == goai.RoleAssistant && m.ResponseID == "response-1" {
							found = reflect.DeepEqual(m, expected)
						}
					}
					observed <- found
					end := replayAnswer("next answer")
					end.StopReason = goai.StopReasonStop
					return &inference.StreamResult{Message: &end, Text: "next answer"}, nil
				}
				e = New(s)
				defer e.Close()
				for _, id := range []string{"source", "fork"} {
					next, err := e.SubmitPrompt(ctx, RunInput{SessionID: id, Prompt: "next", Model: model})
					if err != nil {
						t.Fatal(err)
					}
					waitForCondition(t, 3*time.Second, func() bool { r, err := s.GetTurn(ctx, next.TurnID); return err == nil && r.Status == "completed" }, "next completion")
					select {
					case ok := <-observed:
						if !ok {
							t.Errorf("%s: provider message lost on next request", id)
						}
					case <-time.After(time.Second):
						t.Fatal("provider not called")
					}
				}
			})
		}
	}
}

func replayAnswer(text string) goai.Message {
	return goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: text}}}
}

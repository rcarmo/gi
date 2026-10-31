package turn

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func intentFixture(t *testing.T) (*store.Store, *Engine) {
	t.Helper()
	s := openTestStore(t)
	t.Cleanup(func() { s.Close() })
	e := New(s)
	t.Cleanup(func() { e.Close() })
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "t", "A", "running", "fixture", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "t", "owner", "claim"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	return s, e
}

func TestProviderHostEntryRequiresCommittedIntent(t *testing.T) {
	for _, eventType := range []string{"inference.started", "inference.request_prepared"} {
		t.Run(eventType, func(t *testing.T) {
			s, e := intentFixture(t)
			ctx := context.Background()
			effects := 0
			if _, err := s.DB().Exec(`create trigger reject_provider_intent before insert on turn_events when new.event_type='` + eventType + `' begin select raise(abort,'provider intent rejected'); end`); err != nil {
				t.Fatal(err)
			}
			original := streamWithToolsWithHooks
			defer func() { streamWithToolsWithHooks = original }()
			streamWithToolsWithHooks = func(_ context.Context, _ string, _ *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
				if _, err := hooks.OnPayload(map[string]any{"model": "fixture"}, nil); err != nil {
					return nil, err
				}
				effects++
				answer := replayAnswer("done")
				answer.StopReason = goai.StopReasonStop
				return &inference.StreamResult{Message: &answer}, nil
			}
			if _, err := e.runner("A").runProviderIteration(ctx, s, "t", "A", "fixture", "agent", 1, 1, &goai.Context{}); err == nil {
				t.Fatal("intent failure hidden")
			}
			if effects != 0 {
				t.Fatal("provider entered without intent", effects)
			}
		})
	}
}

func TestProviderPreparedIntentPinsFinalHookPayloadWithoutCredentials(t *testing.T) {
	s, e := intentFixture(t)
	ctx := context.Background()
	if _, err := e.RegisterHook(HookBeforeProviderRequest, "replace-payload", func(_ context.Context, request HookRequest) (HookResponse, error) {
		if request.Payload["stage"] == "payload" {
			return HookResponse{Payload: map[string]any{"request": map[string]any{"model": "final", "headers": map[string]any{"Authorization": "secret", "X-Trace": "kept"}, "access_token": "secret"}}}, nil
		}
		return HookResponse{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	original := streamWithToolsWithHooks
	defer func() { streamWithToolsWithHooks = original }()
	streamWithToolsWithHooks = func(_ context.Context, _ string, _ *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
		if _, err := hooks.OnPayload(map[string]any{"model": "original"}, nil); err != nil {
			return nil, err
		}
		answer := replayAnswer("done")
		answer.StopReason = goai.StopReasonStop
		return &inference.StreamResult{Message: &answer}, nil
	}
	if _, err := e.runner("A").runProviderIteration(ctx, s, "t", "A", "fixture", "agent", 1, 1, &goai.Context{}); err != nil {
		t.Fatal(err)
	}
	events, err := s.ListTurnEvents(ctx, "t")
	if err != nil {
		t.Fatal(err)
	}
	var prepared map[string]any
	for _, event := range events {
		if event.Type == "inference.request_prepared" {
			prepared = event.Payload
		}
	}
	raw, _ := json.Marshal(prepared)
	if prepared == nil || strings.Contains(string(raw), "secret") || !strings.Contains(string(raw), "final") || !strings.Contains(string(raw), "kept") {
		t.Fatalf("unsafe or stale prepared intent: %s", raw)
	}
}

func TestToolHostEntryRequiresCommittedPreparedIntent(t *testing.T) {
	for _, reject := range []bool{false, true} {
		t.Run(map[bool]string{true: "rejected", false: "committed"}[reject], func(t *testing.T) {
			s, e := intentFixture(t)
			effects := 0
			ctx := context.Background()
			if err := e.RegisterTool(tools.RegisteredTool{Name: "intent-fixture", ReplaySafe: true, Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) { effects++; return "ok", nil }}); err != nil {
				t.Fatal(err)
			}
			if _, err := e.RegisterHook(HookApproveTool, "prepare", func(_ context.Context, req HookRequest) (HookResponse, error) {
				req.ToolCall.Arguments["value"] = "prepared"
				return HookResponse{ToolCall: req.ToolCall}, nil
			}); err != nil {
				t.Fatal(err)
			}
			if reject {
				if _, err := s.DB().Exec(`create trigger reject_tool_intent before insert on turn_events when new.event_type='tool.started' begin select raise(abort,'tool intent rejected'); end`); err != nil {
					t.Fatal(err)
				}
			}
			subscription := e.Subscribe("A")
			defer e.Unsubscribe("A", subscription)
			conv := &goai.Context{}
			out := e.runner("A").executeToolCallsPhase(ctx, s, "t", "A", "bootstrap", "agent", 1, conv, []goai.ToolCall{{ID: "c", Name: "intent-fixture", Arguments: map[string]any{"value": "original"}}}, nil, "", 0, &goai.Usage{})
			if reject {
				if effects != 0 || !out.terminated || len(conv.Messages) != 0 {
					t.Fatal("uncommitted tool entered", effects, out)
				}
				for len(subscription) > 0 {
					if ev := <-subscription; ev["type"] == "tool_activity_changed" {
						t.Fatal("uncommitted start published", ev)
					}
				}
			} else {
				if effects != 1 || out.terminated {
					t.Fatal(effects, out)
				}
				events, err := s.ListTurnEvents(ctx, "t")
				if err != nil {
					t.Fatal(err)
				}
				var start map[string]any
				for _, ev := range events {
					if ev.Type == "tool.started" {
						start = ev.Payload
					}
				}
				original := start["original_call"].(map[string]any)["arguments"].(map[string]any)
				prepared := start["prepared_call"].(map[string]any)["arguments"].(map[string]any)
				if original["value"] != "original" || prepared["value"] != "prepared" || start["schema"] == nil || start["replay_safe"] != true {
					t.Fatal("provenance/intent not separated", start)
				}
			}
		})
	}
}

func TestDeniedAndSkippedResultsCommitBeforeContinuation(t *testing.T) {
	for _, disposition := range []string{HookToolCall, HookApproveTool, "skip", "hook-response"} {
		for _, table := range []string{"messages", "turn_events"} {
			t.Run(disposition+"/"+table, func(t *testing.T) {
				s, e := intentFixture(t)
				ctx := context.Background()
				effects := 0
				if err := e.RegisterTool(tools.RegisteredTool{Name: "denied-fixture", Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) { effects++; return "ok", nil }}); err != nil {
					t.Fatal(err)
				}
				if disposition != "skip" {
					hook := disposition
					if disposition == "hook-response" {
						hook = HookToolCall
					}
					if _, err := e.RegisterHook(hook, "deny", func(context.Context, HookRequest) (HookResponse, error) {
						if disposition == "hook-response" {
							text := "injected"
							return HookResponse{Action: "respond", Handled: true, ToolResult: &text}, nil
						}
						return HookResponse{Block: true, Reason: "denied"}, nil
					}); err != nil {
						t.Fatal(err)
					}
				}
				when := ""
				if table == "turn_events" {
					when = " when new.event_type in ('tool.skipped','tool.finished')"
				}
				if _, err := s.DB().Exec(`create trigger reject_result before insert on ` + table + when + ` begin select raise(abort,'result rejected'); end`); err != nil {
					t.Fatal(err)
				}
				subscription := e.Subscribe("A")
				defer e.Unsubscribe("A", subscription)
				conv := &goai.Context{}
				calls := []goai.ToolCall{{ID: "c", Name: "denied-fixture", Arguments: map[string]any{}}}
				if disposition == "skip" {
					if err := e.runner("A").skipRemainingToolCalls(ctx, "A", "t", conv, calls, 0); err == nil {
						t.Fatal("skip settlement failure hidden")
					}
				} else {
					out := e.runner("A").executeToolCallsPhase(ctx, s, "t", "A", "bootstrap", "agent", 1, conv, calls, nil, "", 0, &goai.Usage{})
					if !out.terminated {
						t.Fatal("continued after result write failure")
					}
				}
				if effects != 0 || len(conv.Messages) != 0 {
					t.Fatal("uncommitted result contributed", effects, conv.Messages)
				}
				for len(subscription) > 0 {
					ev := <-subscription
					if ev["type"] == "tool_skipped" || ev["type"] == "tool_finished" {
						t.Fatal("uncommitted result published", ev)
					}
				}
			})
		}
	}
}

func TestProviderIntentJSONFastPathPreservesOwnershipNumbersAndEscapedSecrets(t *testing.T) {
	for _, raw := range []string{
		`{"model":"fixture","n":9007199254740993,"text":"\\\"Authorization\\\": value","schema":{"properties":{"plain":{"type":"string"}}}}`,
		`{"n":9007199254740993,"headers":{"\u0041uthorization":"fixture-secret","X-Trace":"kept"},"nested":[{"CoOkIe":"fixture-secret","safe":"kept"}]}`,
		`[{"api_Key":"fixture-secret"},{"proxy-authorization":"fixture-secret"},{"api_key":"fixture-secret"},{"apikey":"fixture-secret"},{"api-key":"fixture-secret"},{"x-api-key":"fixture-secret"},{"access_token":"fixture-secret"},{"refresh_token":"fixture-secret"},{"set-cookie":"fixture-secret"}]`,
	} {
		intent, err := sanitizedProviderIntent(json.RawMessage(raw))
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(intent)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(encoded), "fixture-secret") {
			t.Fatal("secret retained", string(encoded))
		}
		if strings.Contains(raw, "9007199254740993") && !strings.Contains(string(encoded), "9007199254740993") {
			t.Fatal("intent number rounded", string(encoded))
		}
		if strings.Contains(raw, "kept") && !strings.Contains(string(encoded), "kept") {
			t.Fatal("non-secret removed", string(encoded))
		}
	}
	source := map[string]any{"messages": []any{map[string]any{"text": "original"}}}
	pinned, err := sanitizedProviderIntent(source)
	if err != nil {
		t.Fatal(err)
	}
	source["messages"].([]any)[0].(map[string]any)["text"] = "mutated"
	encoded, err := json.Marshal(pinned)
	if err != nil || strings.Contains(string(encoded), "mutated") {
		t.Fatal("intent aliases caller", string(encoded), err)
	}
}

func FuzzProviderIntentSanitizer(f *testing.F) {
	for _, seed := range []string{`{"model":"fixture","n":9007199254740993}`, `{"\u0041uthorization":"secret","text":"\\\"cookie\\\": example"}`, `[{"api_Key":"secret"},{"safe":[{"X-API-Key":"secret","text":"kept"}]}]`} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 4096 || !json.Valid([]byte(raw)) {
			t.Skip()
		}
		decode := func(raw []byte) any {
			var value any
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err != nil {
				t.Fatal(err)
			}
			return value
		}
		expected := decode([]byte(raw))
		var scrub func(any)
		scrub = func(value any) {
			switch v := value.(type) {
			case map[string]any:
				for key, child := range v {
					if providerIntentSecretKey(key) {
						delete(v, key)
					} else {
						scrub(child)
					}
				}
			case []any:
				for _, child := range v {
					scrub(child)
				}
			}
		}
		scrub(expected)
		pinned, err := sanitizedProviderIntent(json.RawMessage(raw))
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(pinned)
		if err != nil {
			t.Fatal(err)
		}
		if got := decode(encoded); !reflect.DeepEqual(got, expected) {
			t.Fatal("JSON intent sanitizer disagrees with decoded reference")
		}
	})
}

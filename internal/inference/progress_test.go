package inference

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

func TestStreamProgressPreservesStructuredPrefixAndExcludesTerminal(t *testing.T) {
	partial := &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonPending, Content: []goai.ContentBlock{{Type: "text", Text: "prefix"}, {Type: "thinking", Thinking: "reason", ThinkingSignature: "signature"}, {Type: "toolCall", ID: "c", Name: "shell", Arguments: map[string]any{"command": "echo"}}}}
	terminal := &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonError, ErrorMessage: "failed"}
	events := make(chan goai.Event, 2)
	events <- &goai.ToolCallDeltaEvent{Partial: partial}
	events <- &goai.ErrorEvent{Reason: goai.StopReasonError, Error: terminal, Err: errors.New("failed")}
	close(events)
	seen := 0
	result, err := collectProviderStream(context.Background(), "fixture", events, nil, func(message *goai.Message) error {
		seen++
		if message != partial {
			t.Fatal("terminal used as progress")
		}
		return nil
	})
	if err == nil || seen != 1 || result.Partial != partial || result.Message != terminal || result.TerminalReason != goai.StopReasonError || result.Partial.Content[1].ThinkingSignature != "signature" {
		t.Fatal(result, seen, err)
	}
}
func TestStreamProgressFallbackAndObserverFailure(t *testing.T) {
	events := make(chan goai.Event, 5)
	events <- &goai.TextDeltaEvent{ContentIndex: 0, Delta: "prefix"}
	events <- &goai.ThinkingDeltaEvent{ContentIndex: 1, Delta: "reason"}
	events <- &goai.ToolCallDeltaEvent{ContentIndex: 2, Delta: `{"command":`}
	events <- &goai.ErrorEvent{Err: errors.New("transport failed")}
	close(events)
	failure := errors.New("checkpoint failed")
	calls := 0
	result, err := collectProviderStream(context.Background(), "fixture", events, nil, func(*goai.Message) error { calls++; return failure })
	if !errors.Is(err, failure) || calls != 1 || !result.Incomplete || result.Partial == nil || len(result.Partial.Content) != 3 {
		t.Fatal(result, calls, err)
	}
	if result.Partial.Content[0].Text != "prefix" || result.Partial.Content[1].Thinking != "reason" || result.Partial.Content[2].Arguments["_partial_json"] != `{"command":` {
		t.Fatal("provider fallback prefix lost", result.Partial)
	}
}
func TestCompatProgressFailureStopsAndRetainsUsageAndPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, `data: {"choices":[{"delta":{"content":"prefix","tool_calls":[{"index":0,"id":"c","function":{"name":"shell","arguments":"{\"command\":"}}]}}],"usage":{"prompt_tokens":17,"completion_tokens":1,"total_tokens":18}}`+"\n\n")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	failure := errors.New("checkpoint failed")
	result, err := streamOpenCodeZen(context.Background(), &goai.Model{ID: "fixture", Provider: "fixture", Api: goai.ApiOpenAICompletions, BaseURL: server.URL}, &goai.Context{}, nil, &StreamHooks{OnProgress: func(message *goai.Message) error {
		if len(message.Content) != 2 || message.Usage.Input != 17 || message.Content[1].ID != "c" {
			t.Error("compat progress lost structure", message)
		}
		return failure
	}})
	if !errors.Is(err, failure) || result == nil || !result.Incomplete || result.Usage.Input != 17 || result.Text != "prefix" || result.Partial == nil {
		t.Fatal(result, err)
	}
}

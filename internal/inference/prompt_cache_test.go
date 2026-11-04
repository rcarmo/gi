package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

// Capture actual go-ai wire requests, not a reconstruction of its builder.
// This proves cache routing and prefix stability without spending provider tokens.
func TestPromptCacheStableSessionIdentityAndPrefix(t *testing.T) {
	Init()
	agent := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", agent)
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(agent, "absent"))
	if err := os.WriteFile(filepath.Join(agent, "auth.json"), []byte(`{"openai":{"type":"api_key","key":"fixture-key"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	var requests []map[string]json.RawMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			http.Error(w, "bad request", 400)
			return
		}
		requests = append(requests, body)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"done\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":1500,\"completion_tokens\":1,\"total_tokens\":1501,\"prompt_tokens_details\":{\"cached_tokens\":1200}}}\n\ndata: [DONE]\n\n")
	}))
	defer server.Close()
	id := "gi-prompt-cache-" + filepath.Base(agent)
	goai.RegisterModel(&goai.Model{ID: id, Provider: "openai", Api: goai.ApiOpenAICompletions, BaseURL: server.URL + "/api.openai.com/v1", Input: []string{"text"}, MaxTokens: 1024})
	conv := &goai.Context{SystemPrompt: "Stable system instructions", Messages: []goai.Message{goai.UserMessage("first")}, Tools: []goai.Tool{{Name: "read", Description: "Read a file", Parameters: json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}}}`)}}}
	for i, session := range []string{"session-A", "session-A", "session-B", "session-A"} {
		if i == 1 {
			conv.Messages = append(conv.Messages, goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}, goai.UserMessage("second"))
		}
		hooks := &StreamHooks{SessionID: session}
		if i == 3 {
			hooks.CacheRetention = goai.CacheRetentionNone
		}
		result, err := StreamWithToolsWithHooks(context.Background(), "openai/"+id, conv, nil, hooks)
		if err != nil {
			t.Fatal(err)
		}
		if result.Usage == nil || result.Usage.CacheRead != 1200 || result.Usage.Input != 300 {
			t.Fatalf("cache counters not normalized: %+v", result.Usage)
		}
	}
	if len(requests) != 4 {
		t.Fatal("request count", len(requests))
	}
	for i, want := range []string{"session-A", "session-A", "session-B"} {
		var got string
		if err := json.Unmarshal(requests[i]["prompt_cache_key"], &got); err != nil || got != want {
			t.Fatalf("request %d cache key=%q, want %q (%v)", i, got, want, err)
		}
	}
	if _, ok := requests[3]["prompt_cache_key"]; ok {
		t.Fatal("cacheRetention none emitted a cache key")
	}
	if !reflect.DeepEqual(requests[0]["tools"], requests[1]["tools"]) {
		t.Fatal("tool prefix changed")
	}
	var first, second []json.RawMessage
	if err := json.Unmarshal(requests[0]["messages"], &first); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(requests[1]["messages"], &second); err != nil {
		t.Fatal(err)
	}
	if len(second) <= len(first) || !reflect.DeepEqual(first, second[:len(first)]) {
		t.Fatal("conversation no longer has the previous request as a prefix")
	}
}

func TestPromptCacheAnthropicMarkersStayEnabledAndDeterministic(t *testing.T) {
	Init()
	agent := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", agent)
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(agent, "absent"))
	if err := os.WriteFile(filepath.Join(agent, "auth.json"), []byte(`{"anthropic":{"type":"api_key","key":"fixture-key"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	id := "gi-anthropic-cache-" + filepath.Base(agent)
	goai.RegisterModel(&goai.Model{ID: id, Provider: "anthropic", Api: goai.ApiAnthropicMessages, BaseURL: "http://127.0.0.1:1", Input: []string{"text"}, MaxTokens: 1024})
	conv := &goai.Context{SystemPrompt: "Stable system", Messages: []goai.Message{goai.UserMessage("stable user")}, Tools: []goai.Tool{{Name: "read", Description: "Read", Parameters: json.RawMessage(`{"type":"object","properties":{}}`)}}}
	var payloads [][]byte
	stop := errors.New("fixture stops before network send")
	for range 2 {
		_, _ = StreamWithToolsWithHooks(context.Background(), "anthropic/"+id, conv, nil, &StreamHooks{SessionID: "session-A", OnPayload: func(payload any, _ *goai.Model) (any, error) {
			raw, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			payloads = append(payloads, raw)
			return nil, stop
		}})
	}
	if len(payloads) != 2 || !bytes.Equal(payloads[0], payloads[1]) {
		t.Fatal("unchanged Anthropic context changes its wire payload")
	}
	if bytes.Count(payloads[0], []byte(`"cache_control"`)) < 3 {
		t.Fatal("missing system/tool/message cache breakpoints")
	}
}

func TestPromptCacheOpenCodeAffinityAndUsage(t *testing.T) {
	var session string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session = r.Header.Get("x-opencode-session")
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"done\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":1500,\"completion_tokens\":1,\"total_tokens\":1501,\"prompt_tokens_details\":{\"cached_tokens\":1200}}}\n\ndata: [DONE]\n\n")
	}))
	defer server.Close()
	result, err := streamOpenCodeZen(context.Background(), &goai.Model{ID: "fixture", Provider: "opencode-zen", BaseURL: server.URL}, &goai.Context{Messages: []goai.Message{goai.UserMessage("hello")}}, nil, &StreamHooks{SessionID: "stable-session"})
	if err != nil {
		t.Fatal(err)
	}
	if session != "stable-session" {
		t.Fatalf("OpenCode affinity lost: %q", session)
	}
	if result.Usage == nil || result.Usage.Input != 300 || result.Usage.CacheRead != 1200 || result.Usage.TotalTokens != 1501 {
		t.Fatalf("overlapping or discarded cache counters: %+v", result.Usage)
	}
}

package inference

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	goai "github.com/rcarmo/go-ai"
)

func TestAnthropicEmptyToolInputAndOutputRoundTrip(t *testing.T) {
	Init()
	home := t.TempDir()
	t.Setenv("HOME", home)
	os.MkdirAll(filepath.Join(home, ".pi", "agent"), 0700)
	os.WriteFile(filepath.Join(home, ".pi", "agent", "auth.json"), []byte(`{"anthropic":{"type":"api_key","apiKey":"fixture"}}`), 0600)
	var requests atomic.Int32
	var secondInput any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		n := requests.Add(1)
		if n >= 2 {
			for _, msg := range payload["messages"].([]any) {
				m := msg.(map[string]any)
				blocks, _ := m["content"].([]any)
				for _, value := range blocks {
					b := value.(map[string]any)
					if b["type"] == "tool_result" {
						if content, ok := b["content"].([]any); ok {
							for _, item := range content {
								text, _ := item.(map[string]any)
								if text["type"] == "text" && (text["text"] == nil || text["text"] == "") {
									w.WriteHeader(400)
									fmt.Fprint(w, `{"type":"error","error":{"type":"invalid_request_error","message":"tool_result text: Field required"}}`)
									return
								}
							}
						}
					}
					if b["type"] == "tool_use" {
						secondInput = b["input"]
						if _, ok := secondInput.(map[string]any); !ok {
							w.WriteHeader(400)
							fmt.Fprint(w, `{"type":"error","error":{"type":"invalid_request_error","message":"messages.1.content.0.tool_use.input: Input should be an object"}}`)
							return
						}
					}
				}
			}
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"fixture\",\"type\":\"message\",\"role\":\"assistant\",\"model\":\"fixture\",\"content\":[],\"usage\":{\"input_tokens\":2,\"output_tokens\":0}}}\n\n")
		if n == 1 {
			fmt.Fprint(w, "event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"tool_use\",\"id\":\"call-empty\",\"name\":\"noop\",\"input\":{}}}\n\nevent: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"tool_use\"},\"usage\":{\"output_tokens\":1}}\n\n")
		} else {
			fmt.Fprint(w, "event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\nevent: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"done\"}}\n\nevent: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":1}}\n\n")
		}
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	}))
	defer srv.Close()
	id := fmt.Sprintf("gi-empty-input-fixture-%d", time.Now().UnixNano())
	goai.RegisterModel(&goai.Model{ID: id, Name: id, Provider: "anthropic", Api: goai.ApiAnthropicMessages, BaseURL: srv.URL, MaxTokens: 64})
	conversation := &goai.Context{Messages: []goai.Message{{Role: goai.RoleUser, Content: []goai.ContentBlock{{Type: "text", Text: "noop"}}}}, Tools: []goai.Tool{{Name: "noop", Description: "No arguments", Parameters: json.RawMessage(`{"type":"object","properties":{}}`)}}}
	first, err := StreamWithTools(context.Background(), "anthropic/"+id, conversation, nil)
	if err != nil {
		t.Fatal(err)
	}
	conversation.Messages = append(conversation.Messages, *first.Message)
	goai.AppendToolResult(conversation, "call-empty", "noop", "", false)
	second, err := StreamWithTools(context.Background(), "anthropic/"+id, conversation, nil)
	if err != nil {
		t.Fatal(err)
	}
	if second.Text != "done" || requests.Load() != 2 {
		t.Fatal(second, requests.Load())
	}
	if args, ok := secondInput.(map[string]any); !ok || len(args) != 0 {
		t.Fatal(secondInput)
	}
	// The next turn replays the empty result from persisted history too.
	conversation.Messages = append(conversation.Messages, *second.Message, goai.UserMessage("continue"))
	raw, err := json.Marshal(conversation)
	if err != nil {
		t.Fatal(err)
	}
	var replay goai.Context
	if err := json.Unmarshal(raw, &replay); err != nil {
		t.Fatal(err)
	}
	third, err := StreamWithTools(context.Background(), "anthropic/"+id, &replay, nil)
	if err != nil || third == nil || third.Text != "done" || requests.Load() != 3 {
		t.Fatalf("replay: result=%+v err=%v requests=%d", third, err, requests.Load())
	}
	if conversation.Messages[2].Content[0].Text != "" {
		t.Fatal("request preparation mutated empty tool output in history")
	}
}

func TestToolInputNormalisationCopiesOnlyMissingMaps(t *testing.T) {
	params := map[string]any{"query": "literal", "nested": map[string]any{"x": 1}}
	original := &goai.Context{Messages: []goai.Message{{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: "prose"}, {Type: "toolCall", ID: "empty", Name: "noop"}, {Type: "toolCall", ID: "full", Name: "find", Arguments: params}}}}}
	fixed := contextWithObjectToolInputs(original)
	if fixed == original || fixed.Messages[0].Content[1].Arguments == nil || original.Messages[0].Content[1].Arguments != nil {
		t.Fatal("copy contract")
	}
	if fixed.Messages[0].Content[0].Text != "prose" || fixed.Messages[0].Content[2].Arguments["query"] != "literal" {
		t.Fatal(fixed)
	}
	if contextWithObjectToolInputs(fixed) != fixed {
		t.Fatal("unnecessary copy")
	}
	if objectToolInputs(nil) != nil || contextWithObjectToolInputs(nil) != nil {
		t.Fatal("nil")
	}
}

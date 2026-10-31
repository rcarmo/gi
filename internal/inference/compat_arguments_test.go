package inference

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

func TestCompatStreamAccumulatesArgumentsAndMarksInvalidCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		frame := func(calls []any) {
			body, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"delta": map[string]any{"tool_calls": calls}}}})
			fmt.Fprintf(w, "data: %s\n\n", body)
		}
		call := func(index int, id, name, args string) any {
			return map[string]any{"index": index, "id": id, "function": map[string]any{"name": name, "arguments": args}}
		}
		frame([]any{call(1, "array", "shell", "["), call(0, "valid", "shell", `{"command":`)})
		frame([]any{call(0, "", "", `"echo ok"}`), call(1, "", "", "1]")})
		frame([]any{call(2, "broken", "shell", `{"command":`), call(3, "scalar", "shell", "null"), call(4, "empty", "shell", "")})
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"tool_calls\"}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()
	model := &goai.Model{ID: "fixture", Provider: "fixture", Api: goai.ApiOpenAICompletions, BaseURL: server.URL}
	result, err := streamOpenCodeZen(context.Background(), model, &goai.Context{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	calls := goai.GetToolCalls(result.Message)
	if len(calls) != 5 {
		t.Fatalf("calls: %+v", calls)
	}
	if calls[0].ID != "valid" || calls[0].Arguments["command"] != "echo ok" || calls[0].ArgumentsError != "" {
		t.Fatalf("fragment loss: %+v", calls[0])
	}
	for _, index := range []int{1, 2, 3} {
		if calls[index].ArgumentsError == "" {
			t.Fatalf("invalid call became runnable: %+v", calls[index])
		}
	}
	if calls[4].ArgumentsError != "" || calls[4].Arguments == nil {
		t.Fatalf("empty valid call: %+v", calls[4])
	}
}

func TestCompatCompletedArgumentRepair(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{`{"path":"C:\Users\docs"}`, `C:\Users\docs`},
		{"{\"text\":\"line\nbreak\"}", "line\nbreak"},
	} {
		parsed, marker := completedCompatToolArguments(tc.raw)
		if marker != "" {
			t.Fatalf("repair rejected %q: %s", tc.raw, marker)
		}
		value := parsed["path"]
		if value == nil {
			value = parsed["text"]
		}
		if value != tc.want {
			t.Fatalf("repair changed value: %#v want %q", value, tc.want)
		}
	}
	for _, raw := range []string{`{"x":`, `[]`, `null`, `{"x":1} trailing`} {
		if _, marker := completedCompatToolArguments(raw); marker == "" {
			t.Fatalf("invalid accepted %q", raw)
		}
	}
}

func TestCompatMessagesPreserveToolExchangeAndFailClosed(t *testing.T) {
	conv := &goai.Context{SystemPrompt: "system", Messages: []goai.Message{
		goai.UserMessage("request"),
		{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "toolCall", ID: "call", Name: "shell", Arguments: map[string]any{"command": "echo safe"}}}},
		{Role: goai.RoleToolResult, ToolCallID: "call", Content: []goai.ContentBlock{{Type: "text", Text: "safe"}}},
	}}
	messages, err := openAICompatMessages(conv)
	if err != nil {
		t.Fatal(err)
	}
	calls := messages[2]["tool_calls"].([]any)
	wire := calls[0].(map[string]any)
	if wire["id"] != "call" || wire["function"].(map[string]any)["arguments"] != `{"command":"echo safe"}` || messages[2]["content"] != nil {
		t.Fatalf("call missing: %+v", messages)
	}
	if messages[3]["tool_call_id"] != "call" || messages[3]["role"] != "tool" {
		t.Fatalf("result unpaired: %+v", messages)
	}
	conv.Messages[1].Content[0].Arguments["bad"] = make(chan int)
	if _, err := openAICompatMessages(conv); err == nil {
		t.Fatal("unserializable arguments silently replaced")
	}
}

func TestCompatMultimodalUserAndToolResult(t *testing.T) {
	image := goai.ContentBlock{Type: "image", MimeType: "image/png", Data: "ZmFrZQ=="}
	conv := &goai.Context{Messages: []goai.Message{
		{Role: goai.RoleUser, Content: []goai.ContentBlock{{Type: "text", Text: "before"}, image, {Type: "text", Text: "after"}}},
		{Role: goai.RoleToolResult, ToolCallID: "image-call", Content: []goai.ContentBlock{image}},
	}}
	messages, err := openAICompatMessages(conv)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 3 {
		t.Fatalf("image followup missing: %+v", messages)
	}
	parts := messages[0]["content"].([]any)
	if len(parts) != 3 || parts[0].(map[string]any)["text"] != "before" || parts[2].(map[string]any)["text"] != "after" {
		t.Fatalf("ordered content lost: %+v", parts)
	}
	if parts[1].(map[string]any)["image_url"].(map[string]any)["url"] != "data:image/png;base64,ZmFrZQ==" {
		t.Fatalf("image lost: %+v", parts)
	}
	if messages[1]["role"] != "tool" || messages[1]["tool_call_id"] != "image-call" || messages[1]["content"] != "(no tool output)" {
		t.Fatalf("tool pairing lost: %+v", messages)
	}
	follow := messages[2]["content"].([]any)
	if messages[2]["role"] != "user" || follow[0].(map[string]any)["text"] != "Tool result image:" || len(follow) != 2 {
		t.Fatalf("image followup wrong: %+v", messages)
	}
	conv.Messages = []goai.Message{{Role: goai.RoleAssistant, Content: []goai.ContentBlock{image}}}
	if _, err := openAICompatMessages(conv); err == nil {
		t.Fatal("unsupported assistant image silently dropped")
	}
}

func TestCompatNilArgumentsNormalizeToObject(t *testing.T) {
	conv := &goai.Context{Messages: []goai.Message{{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "toolCall", ID: "empty", Name: "fixture"}}}}}
	messages, err := openAICompatMessages(conv)
	if err != nil {
		t.Fatal(err)
	}
	calls := messages[0]["tool_calls"].([]any)
	if got := calls[0].(map[string]any)["function"].(map[string]any)["arguments"]; got != "{}" {
		t.Fatalf("nil arguments emitted %v", got)
	}
}

func TestCompatTerminalFramingRejectsMissingDuplicateAndContradictoryStops(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"missing", "data: {\"choices\":[{\"delta\":{\"content\":\"partial\"}}]}\n\ndata: [DONE]\n\n"},
		{"duplicate", "data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"},{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n"},
		{"contradictory", "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"c\",\"function\":{\"name\":\"shell\",\"arguments\":\"{}\"}}]},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			done := false
			_, err := streamOpenCodeZen(context.Background(), &goai.Model{ID: "fixture", Provider: "fixture", Api: goai.ApiOpenAICompletions, BaseURL: server.URL}, &goai.Context{}, func(event map[string]any) {
				if event["type"] == "done" {
					done = true
				}
			}, nil)
			if err == nil || done {
				t.Fatal("invalid terminal adopted", err, done)
			}
		})
	}
}

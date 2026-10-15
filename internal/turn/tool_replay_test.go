package turn

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/png"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestForkToolReplayAfterReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "replay.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSession(ctx, "source", "source", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "turn", "source", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	calls := []map[string]any{{"id": "call", "name": "shell", "arguments": map[string]any{"command": "exit 1"}}}
	if err := s.AddMessage(ctx, "calls", "source", "assistant", "[tool_call: shell]", map[string]any{"kind": "tool_calls", "display_text": "Checking", "tool_calls": calls}); err != nil {
		t.Fatal(err)
	}
	if err := s.SettleToolResult(ctx, "turn", "source", "tool.failed", map[string]any{"tool": "shell", "tool_call_id": "call", "occurrence_id": "o"}, "result", "exit status 1", map[string]any{"kind": "tool_result", "tool_name": "shell", "tool_call_id": "call", "occurrence_id": "o", "turn_id": "turn", "is_error": true, "details": map[string]any{"exit": 1}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CloneSessionBefore(ctx, "source", "fork", "fork", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	runner := &sessionRunner{}
	for _, id := range []string{"source", "fork"} {
		snapshot, err := s.ContextSnapshot(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		messages := runner.projectContextSnapshot(ctx, id, snapshot)
		if len(messages) != 2 {
			t.Fatalf("%s replay: %+v", id, messages)
		}
		restored := goai.GetToolCalls(&messages[0])
		if len(restored) != 1 || restored[0].ID != "call" || restored[0].Arguments["command"] != "exit 1" {
			t.Fatalf("%s calls: %+v", id, restored)
		}
		if messages[1].Role != goai.RoleToolResult || messages[1].ToolCallID != "call" || !messages[1].IsError || messages[1].Details == nil {
			t.Fatalf("%s result: %+v", id, messages[1])
		}
	}
	fork, err := s.ListMessages(ctx, "fork")
	if err != nil {
		t.Fatal(err)
	}
	if fork[0].ID == "calls" || fork[1].ID == "result" {
		t.Fatal("fork reused row IDs")
	}
	terminal, err := s.ListTerminalToolResults(ctx, "fork")
	if err != nil || len(terminal) != 0 {
		t.Fatalf("source terminal leaked: %+v %v", terminal, err)
	}
}

func TestCompactionCannotSplitReplayedToolExchange(t *testing.T) {
	r := &sessionRunner{}
	snapshot := store.ContextSnapshot{Messages: []store.Message{
		{ID: "call", Role: "assistant", Payload: map[string]any{"kind": "tool_calls", "tool_calls": []map[string]any{{"id": "c", "name": "shell", "arguments": map[string]any{}}}}},
		{ID: "result", Role: "tool_result", Payload: map[string]any{"kind": "tool_result", "tool_call_id": "c", "tool_name": "shell"}, Content: "done"},
	}}
	conv := &goai.Context{Messages: r.projectContextSnapshot(context.Background(), "s", snapshot)}
	split, err := r.compactionBoundary(context.Background(), "s", conv, snapshot, map[string]any{"messages_to_summarize": 1})
	if err != nil || split != nil {
		t.Fatalf("split accepted: %+v %v", split, err)
	}
	complete, err := r.compactionBoundary(context.Background(), "s", conv, snapshot, map[string]any{"messages_to_summarize": 2})
	if err != nil || complete == nil {
		t.Fatalf("complete exchange rejected: %+v %v", complete, err)
	}
}

func TestReplayPreservesProviderBlocksAndToolLoadout(t *testing.T) {
	r := &sessionRunner{}
	assistant := goai.Message{Role: goai.RoleAssistant, Provider: "anthropic", Model: "fixture", Content: []goai.ContentBlock{
		{Type: "thinking", Thinking: "private reasoning", ThinkingSignature: "signed"},
		{Type: "toolCall", ID: "c", Name: "shell", Arguments: map[string]any{}, ThoughtSignature: "call-signed"},
	}}
	extras := &toolExtras{blocks: []goai.ContentBlock{{Type: "image", Data: "aW1hZ2U=", MimeType: "image/png"}}, added: []string{"discovered"}}
	payload := extras.withReplay(map[string]any{"tool_call_id": "c", "tool_name": "shell"}, "output", false)
	snapshot := store.ContextSnapshot{Messages: []store.Message{
		{Role: "assistant", Payload: map[string]any{"provider_message": assistant}},
		{Role: "tool_result", Content: "display notes", Payload: payload},
	}}
	messages := r.projectContextSnapshot(context.Background(), "s", snapshot)
	if len(messages) != 2 || messages[0].Content[0].ThinkingSignature != "signed" || messages[0].Content[1].ThoughtSignature != "call-signed" {
		t.Fatalf("signed content lost: %+v", messages)
	}
	result := messages[1]
	if len(result.Content) != 2 || result.Content[1].Type != "image" || result.Content[1].Data != "aW1hZ2U=" || len(result.AddedToolNames) != 1 {
		t.Fatalf("result extras lost: %+v", result)
	}
	boundary, err := r.compactionBoundary(context.Background(), "s", &goai.Context{Messages: messages}, snapshot, map[string]any{"messages_to_summarize": 2})
	if err != nil || boundary != nil {
		t.Fatalf("image checkpoint accepted: %+v %v", boundary, err)
	}
	failed := extras.withReplay(map[string]any{}, "failure", true)
	if _, ok := failed["added_tool_names"]; ok {
		t.Fatal("failed result activated tools")
	}
}

func TestInterruptedReplayUsesProviderOrphanRepair(t *testing.T) {
	model := &goai.Model{ID: "fixture", Provider: "fixture", Api: goai.ApiOpenAICompletions}
	messages := []goai.Message{
		{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "toolCall", ID: "reused", Name: "shell", Arguments: map[string]any{}}}},
		goai.UserMessage("continue"),
		{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "toolCall", ID: "reused", Name: "shell", Arguments: map[string]any{}}}},
		{Role: goai.RoleToolResult, ToolCallID: "reused", ToolName: "shell", Content: []goai.ContentBlock{{Type: "text", Text: "later result"}}},
	}
	repaired := goai.TransformMessages(messages, model)
	if len(repaired) != 5 || repaired[1].Role != goai.RoleToolResult || !repaired[1].IsError {
		t.Fatalf("orphan not fenced at user boundary: %+v", repaired)
	}
	if repaired[4].IsError {
		t.Fatal("later reused ID result was replaced")
	}
}

func TestLargeFailedToolImageReplayAfterReopenAndFork(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "image-replay.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSession(ctx, "images", "images", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "image-turn", "images", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	// Valid, poorly compressible image exceeds the display-text truncation bound.
	img := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	rng := rand.New(rand.NewSource(42))
	if _, err := rng.Read(img.Pix); err != nil {
		t.Fatal(err)
	}
	var pngBytes bytes.Buffer
	if err := png.Encode(&pngBytes, img); err != nil {
		t.Fatal(err)
	}
	if pngBytes.Len() < 100000 {
		t.Fatal("image fixture too small")
	}
	extras := &toolExtras{}
	extras.attachFunc()("image/png", pngBytes.Bytes())
	payload := extras.withReplay(map[string]any{"kind": "tool_result", "tool_name": "shell", "tool_call_id": "image-call", "is_error": true, "turn_id": "image-turn", "occurrence_id": "image-o"}, "failed with image", true)
	call := goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "toolCall", ID: "image-call", Name: "shell", Arguments: map[string]any{}}}}
	if err := s.AddMessage(ctx, "image-call-row", "images", "assistant", "call", map[string]any{"kind": "tool_calls", "provider_message": call}); err != nil {
		t.Fatal(err)
	}
	if err := s.SettleToolResult(ctx, "image-turn", "images", "tool.failed", map[string]any{"tool": "shell", "occurrence_id": "image-o"}, "image-result", "failed with image", payload); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CloneSessionBefore(ctx, "images", "image-fork", "fork", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	r := &sessionRunner{}
	for _, id := range []string{"images", "image-fork"} {
		snapshot, err := s.ContextSnapshot(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		replay := r.projectContextSnapshot(ctx, id, snapshot)
		if len(replay) != 2 || !replay[1].IsError || len(replay[1].Content) != 2 {
			t.Fatalf("%s replay shape: %+v", id, replay)
		}
		block := replay[1].Content[1]
		decoded, err := base64.StdEncoding.DecodeString(block.Data)
		if err != nil {
			t.Fatal(err)
		}
		restored, err := png.Decode(bytes.NewReader(decoded))
		if err != nil {
			t.Fatal(err)
		}
		if restored.Bounds() != img.Bounds() {
			t.Fatal("image dimensions changed")
		}
		// Lossless pixels have an exact numerical contract; no output hash gate.
		for _, point := range []image.Point{{0, 0}, {100, 200}, {511, 511}} {
			if restored.At(point.X, point.Y) != img.At(point.X, point.Y) {
				t.Fatalf("pixel changed at %v", point)
			}
		}
		page, err := s.PageConversationMessages(ctx, id, "", "", 100)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range page.Messages {
			if m.Payload["provider_content_blocks"] != nil {
				t.Fatal("image bytes exposed in conversation")
			}
		}
	}
}

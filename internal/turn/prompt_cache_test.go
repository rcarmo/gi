package turn

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	goai "github.com/rcarmo/go-ai"
)

func TestPromptCacheIdentitySurvivesTurnsAndPrefixReplay(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	defer e.Close()
	ctx := context.Background()
	const session = "cache-conversation"
	if _, err := s.CreateSession(ctx, session, "Cache fixture", nil); err != nil {
		t.Fatal(err)
	}
	original := streamWithToolsWithHooks
	defer func() { streamWithToolsWithHooks = original }()
	var mu sync.Mutex
	var requests []*goai.Context
	streamWithToolsWithHooks = func(_ context.Context, _ string, request *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
		if hooks == nil || hooks.SessionID != session {
			t.Errorf("provider cache identity was not the conversation: %+v", hooks)
		}
		mu.Lock()
		requests = append(requests, &goai.Context{SystemPrompt: request.SystemPrompt, Tools: append([]goai.Tool(nil), request.Tools...), Messages: append([]goai.Message(nil), request.Messages...)})
		mu.Unlock()
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}, StopReason: goai.StopReasonStop}}, nil
	}
	for _, text := range []string{"first", "second"} {
		result, err := e.SubmitPrompt(ctx, RunInput{SessionID: session, Prompt: text, Model: "provider/model"})
		if err != nil {
			t.Fatal(err)
		}
		waitForCondition(t, 3*time.Second, func() bool {
			turn, err := s.GetTurn(ctx, result.TurnID)
			return err == nil && turn.Status == "completed" && e.LiveActiveTurn(session) == ""
		}, "cache fixture completion")
	}
	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 2 {
		t.Fatalf("requests=%d", len(requests))
	}
	if requests[0].SystemPrompt != requests[1].SystemPrompt || !reflect.DeepEqual(requests[0].Tools, requests[1].Tools) {
		t.Fatal("stable tools/system prompt changed between turns")
	}
	n := len(requests[0].Messages)
	if len(requests[1].Messages) <= n || !reflect.DeepEqual(requests[0].Messages, requests[1].Messages[:n]) {
		t.Fatal("replay changed the previous request prefix")
	}
}

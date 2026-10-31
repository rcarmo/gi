package turn

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// Hold a real automatic before-compact hook while a second steering message
// arrives. The first was already selected at the prior reply boundary. Both
// must remain owned by the running turn and reach distinct provider requests.
func containsUser(users []string, text string) bool {
	for _, user := range users {
		if user == text {
			return true
		}
	}
	return false
}

func TestNativeSteeringDuringAutomaticCompaction(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	defer s.Close()
	const sessionID = "steering-during-compaction"
	if _, err := s.CreateSession(ctx, sessionID, sessionID, map[string]any{"model": "mock-tool", "status": "idle"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if err := s.AddMessage(ctx, store.NowID("msg"), sessionID, "user", strings.Repeat("history ", 60), nil); err != nil {
			t.Fatal(err)
		}
	}
	firstStarted, releaseFirst := make(chan struct{}), make(chan struct{})
	compacting, releaseCompaction := make(chan struct{}), make(chan struct{})
	requests := make(chan []string, 3)
	var count atomic.Int32
	withStreamWithToolsStub(t, func(ctx context.Context, _ string, conv *goai.Context, _ func(map[string]any)) (*inference.StreamResult, error) {
		index := int(count.Add(1))
		var users []string
		for _, message := range conv.Messages {
			if message.Role == goai.RoleUser {
				users = append(users, goai.GetTextContent(&message))
			}
		}
		requests <- users
		if index == 1 {
			close(firstStarted)
			select {
			case <-releaseFirst:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "answer"}}}}, nil
	})
	e := New(s)
	defer e.Close()
	e.runtimeCfg.Compaction = config.CompactionSettings{Enabled: true, ThresholdTokens: 20, KeepRecentTokens: 20}
	var compactionCalls atomic.Int32
	if _, err := e.RegisterHook(HookSessionBeforeCompact, "compaction-gate", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		switch compactionCalls.Add(1) {
		case 1:
			return HookResponse{Block: true}, nil // First request must use the unmodified history.
		case 2:
			close(compacting)
			select {
			case <-releaseCompaction:
			case <-ctx.Done():
				return HookResponse{}, ctx.Err()
			}
			return HookResponse{Payload: map[string]any{"summary": "retained decisions"}}, nil
		default:
			return HookResponse{Block: true}, nil
		}
	}); err != nil {
		t.Fatal(err)
	}
	turn, err := e.SubmitPrompt(ctx, RunInput{SessionID: sessionID, Prompt: "first request", Model: "mock-tool"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-firstStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("first provider request did not start")
	}
	first, err := e.SubmitPrompt(ctx, RunInput{SessionID: sessionID, Prompt: "steer one", Model: "mock-tool"})
	if err != nil || first.Queued || first.TurnID != turn.TurnID {
		t.Fatalf("first steering admission: %#v %v", first, err)
	}
	close(releaseFirst)
	select {
	case <-compacting:
	case <-time.After(5 * time.Second):
		t.Fatal("automatic compaction hook did not start")
	}
	second, err := e.SubmitPrompt(ctx, RunInput{SessionID: sessionID, Prompt: "steer two", Model: "mock-tool"})
	if err != nil || second.Queued || second.TurnID != turn.TurnID {
		t.Fatalf("second steering admission: %#v %v", second, err)
	}
	close(releaseCompaction)
	for i := 0; i < 3; i++ {
		select {
		case users := <-requests:
			if (i == 0 && (len(users) == 0 || users[len(users)-1] != "first request")) ||
				(i == 1 && (len(users) == 0 || users[len(users)-1] != "steer one" || containsUser(users, "steer two"))) ||
				(i == 2 && (len(users) < 2 || users[len(users)-2] != "steer one" || users[len(users)-1] != "steer two")) {
				t.Fatalf("request %d users=%q", i+1, users)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("request %d missing", i+1)
		}
	}
	waitForCondition(t, 5*time.Second, func() bool {
		record, err := s.GetTurn(ctx, turn.TurnID)
		if err != nil || record.Status != "completed" {
			return false
		}
		_, _, err = s.GetSessionActiveTurn(ctx, sessionID)
		return err == sql.ErrNoRows
	}, "compacted steering turn completion and release")
	if count.Load() != 3 || compactionCalls.Load() < 2 {
		t.Fatalf("provider requests=%d compaction hooks=%d, want 3 requests and real second-iteration compaction", count.Load(), compactionCalls.Load())
	}
	// Provider replay now preserves the earlier live response exactly (#84).
	// This context is therefore a valid durable prefix. Neither steering
	// admission may disappear behind that prefix.
	snapshot, err := s.ContextSnapshot(ctx, sessionID)
	if err != nil || snapshot.Summary != "retained decisions" {
		t.Fatalf("unexpected durable summary=%q: %v", snapshot.Summary, err)
	}
	for _, text := range []string{"steer one", "steer two"} {
		found := false
		for _, message := range snapshot.Messages {
			if message.Role == "user" && message.Content == text {
				found = true
			}
		}
		if !found {
			t.Fatalf("durable compaction swallowed steering %q: %+v", text, snapshot)
		}
	}
	events, err := s.ListTurnEvents(ctx, turn.TurnID)
	if err != nil {
		t.Fatal(err)
	}
	completed := 0
	for _, event := range events {
		if event.Type == "compaction.completed" {
			completed++
			if event.Payload["durable_context"] != true {
				t.Fatalf("equivalent replay context did not advance durable coverage: %#v", event.Payload)
			}
		}
	}
	if completed != 1 {
		t.Fatalf("completed run-local compaction checkpoints=%d want 1", completed)
	}
	if queued, err := s.SteeringQueueLength(ctx, sessionID); err != nil || queued != 0 {
		t.Fatalf("pending steering after completion=%d: %v", queued, err)
	}
	messages, err := s.ListMessages(ctx, sessionID)
	if err != nil {
		t.Fatal(err)
	}
	var steering []string
	for _, message := range messages {
		if message.Role == "user" && (message.Content == "steer one" || message.Content == "steer two") {
			steering = append(steering, message.Content)
			if message.Payload["turn_id"] != turn.TurnID {
				t.Fatalf("steering %q persisted against turn %v, want %s", message.Content, message.Payload["turn_id"], turn.TurnID)
			}
		}
	}
	if !reflect.DeepEqual(steering, []string{"steer one", "steer two"}) {
		t.Fatalf("persisted steering=%q", steering)
	}
}

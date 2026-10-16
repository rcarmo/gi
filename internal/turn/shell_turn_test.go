package turn

import (
	"context"
	"github.com/rcarmo/gi/internal/topics"
	"strings"
	"testing"
	"time"
)

func TestShellTurnPublishesTerminalRuntimeTopics(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_shell_topics", "Shell", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	turnTopicCh, unsubTurn := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 16, SessionID: "session_shell_topics"})
	defer unsubTurn()
	sessionTopicCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 16, SessionID: "session_shell_topics"})
	defer unsubSession()
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_shell_topics", Prompt: "hello shell", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit shell prompt: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "completed" && turnRec.FinishedAt != ""
	}, "shell turn completion")
	foundTurnCompleted := false
	foundSessionIdle := false
	deadline := time.After(2 * time.Second)
	for !(foundTurnCompleted && foundSessionIdle) {
		select {
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_completed" && env.Payload["turn_id"] == result.TurnID && env.Payload["status"] == "completed" && env.Payload["reason"] == "completed" && env.Payload["completion_kind"] == "response" {
				foundTurnCompleted = true
			}
		case env := <-sessionTopicCh:
			if env.Payload["type"] == "session_idle" && env.Payload["turn_id"] == result.TurnID && env.Payload["turn_status"] == "completed" && env.Payload["failure_kind"] == "" && env.Payload["model"] == "bootstrap" && env.Payload["completion_kind"] == "response" {
				foundSessionIdle = true
			}
		case <-deadline:
			t.Fatalf("expected shell turn to publish runtime turn/session terminal topics, got turn_completed=%v session_idle=%v", foundTurnCompleted, foundSessionIdle)
		}
	}
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundFinishedAudit := false
	foundToolFinishedAudit := false
	for _, event := range events {
		if event.Type == "turn.finished" && event.Payload["status"] == "completed" && event.Payload["completion_kind"] == "response" {
			foundFinishedAudit = true
		}
		if event.Type == "tool.finished" && event.Payload["tool"] == "shell" && event.Payload["output_length"] != nil {
			foundToolFinishedAudit = true
		}
	}
	if !foundFinishedAudit {
		t.Fatalf("expected shell completion to persist completion_kind on turn.finished, got %#v", events)
	}
	if !foundToolFinishedAudit {
		t.Fatalf("expected shell completion to persist normalized tool.finished audit row, got %#v", events)
	}
}

func TestShellCancelBroadcastsSystemMessageToTurnResponseTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_shell_cancel_topics", "Shell", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	turnResponseCh, unsub := engine.Topics().Subscribe(ctx, "turn.response", topics.SubscribeOptions{Buffer: 16, SessionID: "session_shell_cancel_topics"})
	defer unsub()
	setupEntered := make(chan struct{})
	engine.beforeSetupHook = func(ctx context.Context, sessionID, turnID string) {
		select {
		case <-setupEntered:
		default:
			close(setupEntered)
		}
		<-ctx.Done()
	}
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_shell_cancel_topics", Prompt: "cancel shell", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit shell prompt: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		select {
		case <-setupEntered:
			return true
		default:
			return false
		}
	}, "shell setup phase entry")
	if err := engine.CancelTurn(ctx, "session_shell_cancel_topics", result.TurnID); err != nil {
		t.Fatalf("cancel shell turn during setup: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "cancelled" && turnRec.FinishedAt != ""
	}, "shell turn cancellation")
	deadline := time.After(2 * time.Second)
	foundSystemMessage := false
	for !foundSystemMessage {
		select {
		case env := <-turnResponseCh:
			if env.Payload["sender"] == "system" {
				data, _ := env.Payload["data"].(map[string]any)
				if data["type"] == "system_message" && data["content"] == "Turn cancelled" {
					foundSystemMessage = true
				}
			}
		case <-deadline:
			t.Fatal("expected shell cancel to publish turn.response system message")
		}
	}
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundFinishedAudit := false
	for _, event := range events {
		if event.Type == "turn.finished" && event.Payload["status"] == "cancelled" && event.Payload["reason"] == "cancelled" && event.Payload["failure_kind"] == "" {
			foundFinishedAudit = true
		}
	}
	if !foundFinishedAudit {
		t.Fatalf("expected shell cancel to persist turn.finished audit row, got %#v", events)
	}
}

func TestShellFailureBroadcastsSystemMessageToTurnResponseTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_shell_failure_topics", "Shell", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	turnResponseCh, unsub := engine.Topics().Subscribe(ctx, "turn.response", topics.SubscribeOptions{Buffer: 16, SessionID: "session_shell_failure_topics"})
	defer unsub()
	t.Setenv("PATH", "")
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_shell_failure_topics", Prompt: "fail shell", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit shell prompt: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "failed" && turnRec.FinishedAt != ""
	}, "shell turn failure")
	deadline := time.After(2 * time.Second)
	foundSystemMessage := false
	for !foundSystemMessage {
		select {
		case env := <-turnResponseCh:
			if env.Payload["sender"] == "system" {
				data, _ := env.Payload["data"].(map[string]any)
				if data["type"] == "system_message" {
					content, _ := data["content"].(string)
					if strings.Contains(content, "Shell tool failed:") {
						foundSystemMessage = true
					}
				}
			}
		case <-deadline:
			t.Fatal("expected shell failure to publish turn.response system message")
		}
	}
	// Worker status is fenced before it publishes the terminal system post;
	// observing that post does not imply its turn.finished audit is written.
	waitForCondition(t, 3*time.Second, func() bool {
		events, err := s.ListTurnEvents(ctx, result.TurnID)
		if err != nil {
			return false
		}
		for _, event := range events {
			if event.Type == "turn.finished" {
				return true
			}
		}
		return false
	}, "shell failure finish audit")
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundFinishedAudit := false
	foundToolFailedAudit := false
	for _, event := range events {
		if event.Type == "turn.finished" && event.Payload["status"] == "failed" && event.Payload["reason"] == "shell_error" && event.Payload["failure_kind"] == "shell_error" {
			foundFinishedAudit = true
		}
		if event.Type == "tool.failed" && event.Payload["tool"] == "shell" && event.Payload["failure_kind"] == "shell_error" {
			foundToolFailedAudit = true
		}
	}
	if !foundFinishedAudit {
		t.Fatalf("expected shell failure to persist turn.finished audit row, got %#v", events)
	}
	if !foundToolFailedAudit {
		t.Fatalf("expected shell failure to persist normalized tool.failed audit row, got %#v", events)
	}
}

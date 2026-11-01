package turn

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestBootstrapTerminalJoinsChildBeforeParentAndQueue(t *testing.T) {
	for _, outcome := range []string{"completed", "failed", "cancelled"} {
		t.Run(outcome, func(t *testing.T) {
			s := openChildJoinStore(t)
			e := New(s)
			ctx := context.Background()
			for _, name := range []string{"parent", "child"} {
				if _, err := s.CreateSession(ctx, name, name, nil); err != nil {
					t.Fatal(err)
				}
			}
			parentStarted, parentRelease := make(chan struct{}), make(chan struct{})
			childStarted, childCancelled, childRelease := make(chan struct{}), make(chan struct{}), make(chan struct{})
			var parentOnce, childOnce sync.Once
			defer func() {
				parentOnce.Do(func() { close(parentRelease) })
				childOnce.Do(func() { close(childRelease) })
				e.Close()
			}()
			var shells atomic.Int32
			e.shellPromptRunner = func(runCtx context.Context, _ string, delta func(string)) (string, error, bool) {
				if shells.Add(1) > 1 {
					return "queued done", nil, false
				}
				delta("observed output")
				close(parentStarted)
				select {
				case <-parentRelease:
				case <-runCtx.Done():
					return "observed output", nil, true
				}
				if outcome == "failed" {
					return "", errors.New("shell failed"), false
				}
				return "observed output", nil, false
			}
			withStreamWithToolsStub(t, func(runCtx context.Context, _ string, _ *goai.Context, _ func(map[string]any)) (*inference.StreamResult, error) {
				close(childStarted)
				<-runCtx.Done()
				close(childCancelled)
				<-childRelease
				return nil, runCtx.Err()
			})
			parent, err := e.SubmitPrompt(ctx, RunInput{SessionID: "parent", Prompt: "parent", Model: "bootstrap"})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-parentStarted:
			case <-time.After(2 * time.Second):
				t.Fatal("parent did not start")
			}
			child, err := e.SubmitPrompt(ctx, RunInput{SessionID: "child", Prompt: "child", Model: "mock-stream", ParentTurnID: parent.TurnID})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-childStarted:
			case <-time.After(2 * time.Second):
				t.Fatal("child did not start")
			}
			queued, err := e.SubmitPrompt(ctx, RunInput{SessionID: "parent", Prompt: "queued", Model: "bootstrap", Intent: "queue"})
			if err != nil || !queued.Queued {
				t.Fatal(queued, err)
			}
			if outcome == "cancelled" {
				if err = e.CancelTurn(ctx, "parent", parent.TurnID); err != nil {
					t.Fatal(err)
				}
			} else {
				parentOnce.Do(func() { close(parentRelease) })
			}
			select {
			case <-childCancelled:
			case <-time.After(2 * time.Second):
				t.Fatal("child not cancelled")
			}
			waitForCondition(t, 2*time.Second, func() bool { row, err := s.GetTurn(ctx, parent.TurnID); return err == nil && row.Phase == "settling" }, "bootstrap staged decision")
			row, err := s.GetTurn(ctx, parent.TurnID)
			if err != nil || isTerminalSubTurnStatus(row.Status) || row.FinishedAt != "" {
				t.Fatal("parent completed before child", row, err)
			}
			next, err := s.GetTurn(ctx, queued.TurnID)
			if err != nil || next.Status != "queued" || shells.Load() != 1 {
				t.Fatal("queue passed child join", next, err, shells.Load())
			}
			active, _, err := s.GetSessionActiveTurn(ctx, "parent")
			if err != nil || active != parent.TurnID {
				t.Fatal("parent claim released early", active, err)
			}
			var receipts int
			if err = s.DB().QueryRow(`select count(*) from turn_receipts where turn_id=? and kind='tool'`, parent.TurnID).Scan(&receipts); err != nil || receipts != 1 {
				t.Fatal("shell result not settled", receipts, err)
			}
			messages, err := s.ListMessages(ctx, "parent")
			if err != nil {
				t.Fatal(err)
			}
			retained := false
			for _, m := range messages {
				if m.Role == "assistant" && m.Content == "observed output" {
					retained = true
					if outcome != "completed" && (m.Payload["partial"] != true || m.Payload["context_contribution"] != false) {
						t.Fatal("unsafe partial contribution", m)
					}
				}
			}
			if !retained {
				t.Fatal("known output lost before terminal settlement")
			}
			contributes := false
			for _, m := range e.runner("parent").projectContextSnapshot(ctx, "parent", store.ContextSnapshot{Messages: messages}) {
				for _, block := range m.Content {
					if m.Role == goai.RoleAssistant && block.Text == "observed output" {
						contributes = true
					}
				}
			}
			if contributes != (outcome == "completed") {
				t.Fatal("partial/context projection mismatch", outcome, contributes)
			}
			childOnce.Do(func() { close(childRelease) })
			waitForCondition(t, 2*time.Second, func() bool { row, err := s.GetTurn(ctx, parent.TurnID); return err == nil && row.Status == outcome }, "bootstrap terminal")
			if terminalEventSequence(t, s, child.TurnID) >= terminalEventSequence(t, s, parent.TurnID) {
				t.Fatal("parent terminal receipt preceded child")
			}
			waitForCondition(t, 2*time.Second, func() bool { row, err := s.GetTurn(ctx, queued.TurnID); return err == nil && row.Status == "completed" }, "queued bootstrap completion")
			var late int
			if err = s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='tool.output' and id>(select id from turn_events where turn_id=? and event_type='turn.finished')`, parent.TurnID, parent.TurnID).Scan(&late); err != nil || late != 0 {
				t.Fatal("progress writer outlived terminal", late, err)
			}
		})
	}
}

// This workload deliberately checks the same visible result on both the legacy
// shell path and the common settlement path for equivalent profiling runs.
func TestBootstrapResponseCompletesOnce(t *testing.T) {
	s := openChildJoinStore(t)
	e := NewWithRuntimeConfig(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir(), DefaultModel: "bootstrap"}, "")
	defer e.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "bootstrap profile", Model: "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	waitForCondition(t, 2*time.Second, func() bool { row, err := s.GetTurn(ctx, run.TurnID); return err == nil && row.Status == "completed" }, "bootstrap completion")
	if err = e.Close(); err != nil {
		t.Fatal(err)
	}
	messages, err := s.ListMessages(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	responses := 0
	for _, m := range messages {
		if m.Role == "assistant" {
			responses++
			if m.Content != "Gi received: bootstrap profile" {
				t.Fatal("shell response changed", m.Content)
			}
		}
	}
	if responses != 1 {
		t.Fatal("duplicate or missing response", responses)
	}
	for _, event := range []string{"tool.started", "tool.finished", "turn.finished"} {
		var count int
		if err = s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type=?`, run.TurnID, event).Scan(&count); err != nil || count != 1 {
			t.Fatal(event, count, err)
		}
	}
}

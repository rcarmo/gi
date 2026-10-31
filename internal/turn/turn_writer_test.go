package turn

import (
	"context"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestTurnWriterReplacedClaimRejectsLateNativeEffects(t *testing.T) {
	for _, kind := range []string{"provider", "tool"} {
		t.Run(kind, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "writer.db")
			s, err := store.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			other, err := store.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer other.Close()
			e := New(s)
			started, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			defer func() { once.Do(func() { close(release) }); _ = e.Close() }()
			var requests, effects atomic.Int32
			if err := e.RegisterTool(tools.RegisteredTool{Name: "lease-fixture", Executor: func(_ context.Context, rt tools.ToolRuntime, _ goai.ToolCall) (string, error) {
				effects.Add(1)
				close(started)
				<-release
				rt.OnOutput("must not publish after takeover")
				return "must not adopt after takeover", nil
			}}); err != nil {
				t.Fatal(err)
			}
			withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
				requests.Add(1)
				if kind == "provider" {
					close(started)
					<-release
					return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "must not adopt after takeover"}}}}, nil
				}
				return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Content: []goai.ContentBlock{{Type: "toolCall", ID: "call", Name: "lease-fixture", Arguments: map[string]any{}}}}}, nil
			})
			ctx := context.Background()
			if _, err := s.CreateSession(ctx, "writer", "writer", nil); err != nil {
				t.Fatal(err)
			}
			result, err := e.SubmitPrompt(ctx, RunInput{SessionID: "writer", Prompt: "one request", Model: "mock-stream"})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-started:
			case <-time.After(2 * time.Second):
				t.Fatal("host did not start")
			}
			r := e.runner("writer")
			r.mu.Lock()
			done := r.current.done
			r.mu.Unlock()
			_, old, err := other.GetSessionActiveTurn(ctx, "writer")
			if err != nil {
				t.Fatal(err)
			}
			if err := other.ReleaseSessionActiveTurn(ctx, "writer", old); err != nil {
				t.Fatal(err)
			}
			if ok, err := other.ClaimSessionActiveTurn(ctx, "writer", result.TurnID, "replacement", "replacement"); err != nil || !ok {
				t.Fatal(ok, err)
			}
			once.Do(func() { close(release) })
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("old host did not join")
			}
			if requests.Load() != 1 {
				t.Fatal("old host dispatched another request", requests.Load())
			}
			if kind == "tool" && effects.Load() != 1 {
				t.Fatal("tool effects repeated", effects.Load())
			}
			row, err := s.GetTurn(ctx, result.TurnID)
			if err != nil || row.Status != "running" || row.FinishedAt != "" {
				t.Fatal("old host finalised replacement", row, err)
			}
			_, token, err := other.GetSessionActiveTurn(ctx, "writer")
			if err != nil || token != "replacement" {
				t.Fatal("old cleanup released replacement", token, err)
			}
			var n int
			if err := s.DB().QueryRow(`select count(*) from messages where content like 'must not%'`).Scan(&n); err != nil || n != 0 {
				t.Fatal("late native result adopted", n, err)
			}
			if err := s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='tool.output'`, result.TurnID).Scan(&n); err != nil || n != 0 {
				t.Fatal("late native progress committed", n, err)
			}
		})
	}
}

func TestShellFinalizationFailureKeepsClaimAndQueuedWork(t *testing.T) {
	s := openChildJoinStore(t)
	e := New(s)
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	hook := func(context.Context) { close(started); <-release }
	previous := shellTurnStartHook.Swap(&hook)
	defer func() { once.Do(func() { close(release) }); _ = e.Close(); shellTurnStartHook.Store(previous) }()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "shell-pending", "shell", nil); err != nil {
		t.Fatal(err)
	}
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "shell-pending", Prompt: "first", Model: "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("shell did not start")
	}
	r := e.runner("shell-pending")
	r.mu.Lock()
	active := r.current
	r.mu.Unlock()
	queued, err := e.SubmitPrompt(ctx, RunInput{SessionID: "shell-pending", Prompt: "next", Model: "bootstrap", Intent: "queue"})
	if err != nil || !queued.Queued {
		t.Fatal(queued, err)
	}
	if _, err := s.DB().Exec(`create trigger reject_shell_terminal before update of status on turns when new.status in ('completed','cancelled','failed') begin select raise(abort,'terminal write rejected'); end`); err != nil {
		t.Fatal(err)
	}
	once.Do(func() { close(release) })
	select {
	case <-active.done:
	case <-time.After(2 * time.Second):
		t.Fatal("shell host did not join")
	}
	row, err := s.GetTurn(ctx, run.TurnID)
	if err != nil || row.Status != "running" || row.FinishedAt != "" {
		t.Fatal("failed finalization lost checkpoint", row, err)
	}
	id, token, err := s.GetSessionActiveTurn(ctx, "shell-pending")
	if err != nil || id != run.TurnID || token != active.claimToken {
		t.Fatal("failed finalization released ownership", id, token, err)
	}
	next, err := s.GetTurn(ctx, queued.TurnID)
	if err != nil || next.Status != "queued" {
		t.Fatal("queue passed unresolved settlement", next, err)
	}
	if _, err := s.DB().Exec(`drop trigger reject_shell_terminal`); err != nil {
		t.Fatal(err)
	}
}

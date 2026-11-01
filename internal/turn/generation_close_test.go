package turn

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestGenerationCloseJoinsThenResumesWithoutLifecycleReplay(t *testing.T) {
	for _, mode := range []string{"ready", "silent", "partial", "commit-fault", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			s := openTestStore(t)
			defer s.Close()
			ctx := context.Background()
			if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			entered, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			e := NewWithRuntimeConfig(s, config.RuntimeConfig{DefaultModel: "test/before-close"}, "")
			defer e.Close()
			defer unblock()
			var requests, starts, ends atomic.Int32
			hook := func(e *Engine, name string, count *atomic.Int32) {
				t.Helper()
				if _, err := e.RegisterHook(name, string(name), func(context.Context, HookRequest) (HookResponse, error) { count.Add(1); return HookResponse{}, nil }); err != nil {
					t.Fatal(err)
				}
			}
			hook(e, HookAgentStart, &starts)
			hook(e, HookAgentEnd, &ends)
			if mode == "ready" {
				if _, err := e.RegisterHook(HookBeforeProviderRequest, "blocking-context", func(_ context.Context, req HookRequest) (HookResponse, error) {
					if req.Payload["stage"] == "context" {
						close(entered)
						<-release
					}
					return HookResponse{}, nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			withStreamWithToolsHookStub(t, func(_ context.Context, model string, conv *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
				n := requests.Add(1)
				if mode != "ready" && n == 1 {
					if mode == "partial" {
						partial := &goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: "committed interrupted prefix"}, {Type: "thinking", Thinking: "private thought", ThinkingSignature: "preserved signature"}}, Usage: &goai.Usage{Input: 17}}
						if err := hooks.OnProgress(partial); err != nil {
							return nil, err
						}
					}
					close(entered)
					<-release // Deliberately ignore cancellation.
					return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: "late response must not commit"}}, StopReason: goai.StopReasonStop, Usage: &goai.Usage{Input: 19}}}, nil
				}
				raw, err := json.Marshal(conv)
				if err != nil {
					return nil, err
				}
				if strings.Contains(string(raw), "committed interrupted prefix") || strings.Contains(string(raw), "late response must not commit") {
					return nil, errors.New("interrupted output entered fresh context")
				}
				if model != "test/after-close" {
					return nil, errors.New("stale selected model: " + model)
				}
				return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: "continued after close"}}, StopReason: goai.StopReasonStop, Usage: &goai.Usage{Input: 11}}, Text: "continued after close"}, nil
			})
			run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "one original prompt", Model: "test/before-close"})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-entered:
			case <-time.After(3 * time.Second):
				t.Fatal("checkpoint boundary not reached")
			}
			_, oldToken, err := s.GetSessionActiveTurn(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			if mode == "commit-fault" {
				if _, err = s.DB().Exec(`create trigger reject_close before update on turns when new.phase='checkpoint_generation_ready' begin select raise(abort,'close rejected');end`); err != nil {
					t.Fatal(err)
				}
			}
			closed, alsoClosed := make(chan error, 1), make(chan error, 1)
			go func() { closed <- e.Close() }()
			go func() { alsoClosed <- e.Close() }()
			deadline := time.Now().Add(3 * time.Second)
			for {
				var worker string
				if err = s.DB().QueryRow(`select worker_id from session_active_turns where session_id='A'`).Scan(&worker); err != nil {
					t.Fatal(err)
				}
				if worker == "generation-closing" {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("close did not reserve", worker)
				}
				time.Sleep(time.Millisecond)
			}
			for _, done := range []chan error{closed, alsoClosed} {
				select {
				case err := <-done:
					t.Fatal("Close escaped live callback", err)
				default:
				}
			}
			row, err := s.GetTurn(ctx, run.TurnID)
			if err != nil || row.Status != "running" || row.FinishedAt != "" {
				t.Fatal("reservation published before join", row, err)
			}
			stale := store.WithTurnWriter(ctx, "A", run.TurnID, oldToken)
			if err = s.AdmitTurnEffect(stale, "A", run.TurnID, "running", "inference.started", map[string]any{"request_id": "stale"}); err == nil {
				t.Fatal("cancelled host retained writer")
			}
			observer := New(s)
			if err = observer.ResumeInterruptedWork(ctx); err != nil && !errors.Is(err, store.ErrQueueConflict) {
				t.Fatal(err)
			}
			if err = observer.Close(); err != nil {
				t.Fatal(err)
			}
			unblock()
			for _, done := range []chan error{closed, alsoClosed} {
				select {
				case err = <-done:
					if (err != nil) != (mode == "commit-fault") {
						t.Fatal("unexpected shared close outcome", err)
					}
				case <-time.After(3 * time.Second):
					t.Fatal("Close did not finish joining")
				}
			}
			row, err = s.GetTurn(ctx, run.TurnID)
			if mode == "commit-fault" {
				if err != nil || row.Status != "running" || row.FinishedAt != "" {
					t.Fatal("failed publication fabricated terminal/queue", row, err)
				}
				if err = e.Close(); err == nil {
					t.Fatal("repeat Close lost shared error")
				}
				if _, err = s.DB().Exec(`drop trigger reject_close`); err != nil {
					t.Fatal(err)
				}
				// The joined reservation remains recoverable after a failed
				// publication, but takes the stale-owner rather than graceful path.
				if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z' where session_id='A'`); err != nil {
					t.Fatal(err)
				}
			} else if err != nil || row.Status != "queued" || row.Phase != store.ClosedGenerationReadyPhase || row.FinishedAt != "" {
				t.Fatal(row, err)
			}
			wantRequests := int32(1)
			if mode == "ready" {
				wantRequests = 0
			}
			if requests.Load() != wantRequests || starts.Load() != 1 || ends.Load() != 0 {
				t.Fatal("close dispatched or replayed lifecycle", requests.Load(), starts.Load(), ends.Load())
			}
			if err = s.TouchSessionState(ctx, "A", map[string]any{"selected_provider": "test", "selected_model": "after-close"}); err != nil {
				t.Fatal(err)
			}
			if mode == "invalid" {
				if _, err = s.DB().Exec(`update turn_events set payload_json=json_set(payload_json,'$.continuation.version',9) where turn_id=? and event_type='generation.ready'`, run.TurnID); err != nil {
					t.Fatal(err)
				}
			}
			resumed := NewWithRuntimeConfig(s, config.RuntimeConfig{DefaultModel: "test/after-close"}, "")
			defer resumed.Close()
			if requests.Load() != wantRequests {
				t.Fatal("constructor dispatched checkpoint")
			}
			hook(resumed, HookAgentStart, &starts)
			hook(resumed, HookAgentEnd, &ends)
			if err = resumed.ResumeInterruptedWork(ctx); err != nil {
				t.Fatal(err)
			}
			if mode == "invalid" {
				row, err = s.GetTurn(ctx, run.TurnID)
				if err != nil || row.Status != "failed" || row.Phase != "held_for_retry_or_skip" {
					t.Fatal("invalid checkpoint became prompt replay", row, err)
				}
				if requests.Load() != wantRequests || starts.Load() != 1 || ends.Load() != 0 {
					t.Fatal("invalid checkpoint dispatched/replayed lifecycle", requests.Load(), starts.Load(), ends.Load())
				}
				return
			}
			waitRetryDone(t, s, run.TurnID)
			resumed.runs.Wait()
			row, err = s.GetTurn(ctx, run.TurnID)
			if err != nil || row.Status != "completed" {
				t.Fatal(row, err)
			}
			if requests.Load() != wantRequests+1 || starts.Load() != 1 || ends.Load() != 1 {
				t.Fatal("replayed lifecycle/provider", requests.Load(), starts.Load(), ends.Load())
			}
			var prompts, late, failures, crashes int
			for query, dest := range map[string]*int{`select count(*) from messages where session_id='A' and role='user'`: &prompts, `select count(*) from messages where content='late response must not commit'`: &late, `select count(*) from turn_failures where turn_id=?`: &failures, `select count(*) from turn_events where turn_id=? and event_type='generation.recovered'`: &crashes} {
				var args []any
				if strings.Contains(query, "?") {
					args = []any{run.TurnID}
				}
				if err = s.DB().QueryRow(query, args...).Scan(dest); err != nil {
					t.Fatal(err)
				}
			}
			wantCrashes := 0
			if mode == "commit-fault" {
				wantCrashes = 1
			}
			if prompts != 1 || late != 0 || failures != 0 || crashes != wantCrashes {
				t.Fatal(prompts, late, failures, crashes)
			}
			accounting, err := s.TurnAccounting(ctx, run.TurnID)
			unknown, input := 0, 11
			if mode != "ready" {
				unknown = 1
			}
			if mode == "partial" {
				input += 17
			}
			if err != nil || accounting.Unknown != unknown || accounting.Usage.Input != input {
				t.Fatal(accounting, err)
			}
		})
	}
}

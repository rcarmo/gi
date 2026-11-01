package turn

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestInterruptedCloseJoinsThenHoldsWithoutAbortOrReplay(t *testing.T) {
	for _, mode := range []string{"preparation", "unproven-tool", "unproven-provider", "unproven-provider-empty", "commit-fault"} {
		t.Run(mode, func(t *testing.T) {
			providerMode := mode == "unproven-provider" || mode == "unproven-provider-empty"
			s := openTestStore(t)
			defer s.Close()
			ctx := context.Background()
			if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			entered, release := make(chan struct{}), make(chan struct{})
			var models, ends atomic.Int32
			withStreamWithToolsHookStub(t, func(_ context.Context, _ string, _ *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
				models.Add(1)
				if providerMode {
					if err := hooks.OnProgress(&goai.Message{Role: goai.RoleAssistant, Usage: &goai.Usage{Input: 5}, Content: []goai.ContentBlock{{Type: "text", Text: "committed provider prefix"}}}); err != nil {
						return nil, err
					}
					close(entered)
					<-release
					var blocks []goai.ContentBlock
					if mode != "unproven-provider-empty" {
						blocks = []goai.ContentBlock{{Type: "text", Text: "differing received terminal"}}
					}
					return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Usage: &goai.Usage{Input: 23}, Content: blocks}}, nil
				}
				return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Content: []goai.ContentBlock{{Type: "toolCall", ID: "wire", Name: "held", Arguments: map[string]any{}}}}}, nil
			})
			e := NewWithRuntimeConfig(s, config.RuntimeConfig{DefaultModel: "test/model"}, "")
			defer e.Close()
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			defer unblock()
			var escaped tools.ToolRuntime
			if mode == "unproven-tool" {
				if err := e.RegisterTool(tools.RegisteredTool{Name: "held", Parameters: json.RawMessage(`{"type":"object"}`), ReplaySafe: false, Executor: func(ctx context.Context, rt tools.ToolRuntime, _ goai.ToolCall) (string, error) {
					escaped = rt
					rt.AddUsage(goai.Usage{Input: 7})
					if err := rt.OnOutput("committed prefix"); err != nil {
						return "", err
					}
					close(entered)
					<-release
					rt.AddUsage(goai.Usage{Input: 10})
					if err := rt.OnOutput("committed prefix and joined tail"); err != nil {
						return "", err
					}
					return "", ctx.Err()
				}}); err != nil {
					t.Fatal(err)
				}
			} else if !providerMode {
				if _, err := e.RegisterHook(HookBeforeAgentStart, "unproven preparation", func(context.Context, HookRequest) (HookResponse, error) {
					close(entered)
					<-release
					return HookResponse{}, nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := e.RegisterHook(HookAgentEnd, "end", func(context.Context, HookRequest) (HookResponse, error) { ends.Add(1); return HookResponse{}, nil }); err != nil {
				t.Fatal(err)
			}
			run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "hold an unproven frontier", Model: "test/model"})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-entered:
			case <-time.After(3 * time.Second):
				t.Fatal("frontier not reached")
			}
			if mode == "unproven-tool" || providerMode {
				// An unsupported journal extension must not become replay authority.
				if err = s.AppendTurnEvent(ctx, run.TurnID, "A", "legacy.unverified_boundary", map[string]any{"checkpoint": true}); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "commit-fault" {
				if _, err = s.DB().Exec(`create trigger reject_hold before update on turns when new.phase='held_for_retry_or_skip' begin select raise(abort,'hold rejected');end`); err != nil {
					t.Fatal(err)
				}
			}
			_, old, err := s.GetSessionActiveTurn(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			done, also := make(chan error, 1), make(chan error, 1)
			go func() { done <- e.Close() }()
			go func() { also <- e.Close() }()
			deadline := time.Now().Add(3 * time.Second)
			for {
				var worker string
				if err = s.DB().QueryRow(`select worker_id from session_active_turns where session_id='A'`).Scan(&worker); err != nil {
					t.Fatal(err)
				}
				if worker == store.InterruptedClosingWorker {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("hold not reserved", worker)
				}
				time.Sleep(time.Millisecond)
			}
			for _, ch := range []chan error{done, also} {
				select {
				case err := <-ch:
					t.Fatal("Close escaped live callback", err)
				default:
				}
			}
			row, err := s.GetTurn(ctx, run.TurnID)
			if err != nil || row.Status != "running" || row.FinishedAt != "" {
				t.Fatal("published before join", row, err)
			}
			if err = s.AdmitTurnEffect(store.WithTurnWriter(ctx, "A", run.TurnID, old), "A", run.TurnID, "inference", "inference.started", map[string]any{"request_id": "stale"}); err == nil {
				t.Fatal("old writer accepted")
			}
			unblock()
			for _, ch := range []chan error{done, also} {
				select {
				case err = <-ch:
					if (err != nil) != (mode == "commit-fault") {
						t.Fatal("shared close outcome", err)
					}
				case <-time.After(3 * time.Second):
					t.Fatal("Close did not join")
				}
			}
			if mode == "commit-fault" {
				row, err = s.GetTurn(ctx, run.TurnID)
				if err != nil || row.Status != "running" || row.FinishedAt != "" {
					t.Fatal("failed commit published outcome", row, err)
				}
				if err = e.Close(); err == nil {
					t.Fatal("repeat Close lost error")
				}
				if _, err = s.DB().Exec(`drop trigger reject_hold; update session_active_turns set updated_at='2000-01-01T00:00:00Z' where session_id='A'`); err != nil {
					t.Fatal(err)
				}
			}
			observer := New(s)
			defer observer.Close()
			if err = observer.ResumeInterruptedWork(ctx); err != nil {
				t.Fatal(err)
			}
			row, err = s.GetTurn(ctx, run.TurnID)
			if err != nil || row.Status != "failed" || row.Phase != "held_for_retry_or_skip" {
				t.Fatal(row, err)
			}
			failure, err := s.GetTurnFailure(ctx, run.TurnID)
			if err != nil || failure.HoldState != "review" || failure.FailureKind != InterruptedFailureKind {
				t.Fatal(failure, err)
			}
			if receipt, err := s.HasTurnTerminalReceipt(ctx, run.TurnID); err != nil || receipt {
				t.Fatal("invented terminal", receipt, err)
			}
			wantModels := int32(0)
			if mode == "unproven-tool" {
				wantModels = 1
				var input int
				if err = s.DB().QueryRow(`select json_extract(payload_json,'$.received_usage.input') from turn_events where turn_id=? and event_type='tool.usage_received' order by seq desc limit 1`, run.TurnID).Scan(&input); err != nil || input != 17 {
					t.Fatal("received usage lost", input, err)
				}
				var preview string
				if err = s.DB().QueryRow(`select json_extract(payload_json,'$.output_preview') from turn_events where turn_id=? and event_type='tool.output' order by seq desc limit 1`, run.TurnID).Scan(&preview); err != nil || preview != "committed prefix and joined tail" {
					t.Fatal("joined output lost", preview, err)
				}
				before, err := s.ListTurnEvents(ctx, run.TurnID)
				if err != nil {
					t.Fatal(err)
				}
				escaped.AddUsage(goai.Usage{Input: 99})
				_ = escaped.OnOutput("escaped")
				after, err := s.ListTurnEvents(ctx, run.TurnID)
				if err != nil || len(before) != len(after) {
					t.Fatal("escaped recorder mutated joined hold", err)
				}
			}
			if providerMode {
				wantModels = 1
				accounting, err := s.TurnAccounting(ctx, run.TurnID)
				if err != nil || accounting.Usage.Input != 23 || accounting.Unknown != 1 {
					t.Fatal("received provider usage lost", accounting, err)
				}
				var content string
				want := "committed provider prefix"
				if mode != "unproven-provider-empty" {
					want += "\n\nProvider response:\n\ndiffering received terminal"
				}
				if err = s.DB().QueryRow(`select content from messages where session_id='A' and role='assistant'`).Scan(&content); err != nil || content != want {
					t.Fatal("committed prefix/received terminal lost", content, err)
				}
			}
			if models.Load() != wantModels || ends.Load() != 0 {
				t.Fatal("replay or fabricated lifecycle", models.Load(), ends.Load())
			}
		})
	}
}

func TestInterruptedClosePreservesExplicitUserCancellation(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	e := NewWithRuntimeConfig(s, config.RuntimeConfig{DefaultModel: "test/model"}, "")
	defer e.Close()
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	if _, err := e.RegisterHook(HookBeforeAgentStart, "blocked preparation", func(context.Context, HookRequest) (HookResponse, error) {
		close(entered)
		<-release
		return HookResponse{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "cancel explicitly", Model: "test/model"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("hook not entered")
	}
	if err = e.CancelTurn(ctx, "A", run.TurnID); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- e.Close() }()
	deadline := time.Now().Add(3 * time.Second)
	for !e.closing.Load() {
		if time.Now().After(deadline) {
			t.Fatal("Close did not begin")
		}
		time.Sleep(time.Millisecond)
	}
	select {
	case err = <-done:
		t.Fatal("Close escaped cancelled callback", err)
	default:
	}
	unblock()
	select {
	case err = <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Close did not join")
	}
	row, err := s.GetTurn(ctx, run.TurnID)
	if err != nil || row.Status != "cancelled" {
		t.Fatal("Close replaced user intent", row, err)
	}
	if receipt, err := s.HasTurnTerminalReceipt(ctx, run.TurnID); err != nil || !receipt {
		t.Fatal(receipt, err)
	}
}

func TestInterruptedCloseRecoveryNeverRequeuesSetupOrCompaction(t *testing.T) {
	for _, phase := range []string{"setup", "compacting", "manual-compaction"} {
		t.Run(phase, func(t *testing.T) {
			s := openTestStore(t)
			defer s.Close()
			ctx := context.Background()
			if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			metadata := map[string]any{"model": "test/model", "intent": "prompt"}
			if phase == "manual-compaction" {
				metadata["operation"] = "manual_compaction"
			}
			if _, err := s.CreateTurnWithStatus(ctx, "t", "A", "running", "unproven preparation", metadata); err != nil {
				t.Fatal(err)
			}
			if err := s.UpdateTurnStatusAndPhase(ctx, "t", "running", phase); err != nil {
				t.Fatal(err)
			}
			if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "t", "runner", "old"); err != nil || !ok {
				t.Fatal(ok, err)
			}
			if ok, err := s.ReserveInterruptedClose(ctx, "A", "t", "old", "closing"); err != nil || !ok {
				t.Fatal(ok, err)
			}
			if _, err := s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z' where session_id='A'`); err != nil {
				t.Fatal(err)
			}
			e := NewWithRuntimeConfig(s, config.RuntimeConfig{DefaultModel: "test/model"}, "")
			defer e.Close()
			if err := e.ResumeInterruptedWork(ctx); err != nil {
				t.Fatal(err)
			}
			e.runs.Wait()
			row, err := s.GetTurn(ctx, "t")
			if err != nil || row.Status != "failed" || row.Phase != "held_for_retry_or_skip" {
				t.Fatal("unproven preparation replayed/aborted", row, err)
			}
			failure, err := s.GetTurnFailure(ctx, "t")
			if err != nil || failure.HoldState != "review" {
				t.Fatal(failure, err)
			}
			if receipt, err := s.HasTurnTerminalReceipt(ctx, "t"); err != nil || receipt {
				t.Fatal("invented root terminal", receipt, err)
			}
		})
	}
}

package turn

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func stageRecoveryTurn(t *testing.T, s *store.Store, beforeStage ...func()) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "recovery", "recovery", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "interrupted", "recovery", "running", "do not resend", map[string]any{"model": "mock-stream"}); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "recovery", "interrupted", "dead", "old"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if err := s.AddMessage(ctx, "user-original", "recovery", "user", "do not resend", map[string]any{"turn_id": "interrupted"}); err != nil {
		t.Fatal(err)
	}
	message := goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "already committed"}}}
	if _, err := s.SettleProviderAttempt(ctx, "interrupted", "recovery", "request-original", store.ProviderReceipt{MessageID: "reply-original", Content: "already committed", Payload: map[string]any{"provider_message": message}, Usage: &goai.Usage{Input: 7, Output: 3, TotalTokens: 10}, Decision: "reply_boundary", Iteration: 1}); err != nil {
		t.Fatal(err)
	}
	for _, prepare := range beforeStage {
		prepare()
	}
	if _, _, err := s.StageTurnTerminal(ctx, "interrupted", "recovery", store.TurnSettlement{Status: "failed", Phase: "failed", MessageID: "outcome-original", Message: "chosen outcome", MessagePayload: map[string]any{"turn_id": "interrupted", "kind": "chat", "source": "system"}, FailureKind: "provider_error", Payload: map[string]any{"iterations": 1}, ClaimToken: "old"}); err != nil {
		t.Fatal(err)
	}
}
func staleRecoveryClaim(t *testing.T, s *store.Store) store.ActiveTurnClaim {
	t.Helper()
	if _, err := s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z' where session_id='recovery'`); err != nil {
		t.Fatal(err)
	}
	claims, err := s.ListStaleActiveTurnClaims(context.Background(), time.Now().Add(-interruptedTurnStaleAfter), "recovery")
	if err != nil || len(claims) != 1 {
		t.Fatal(claims, err)
	}
	return claims[0]
}
func assertRecoveredTerminal(t *testing.T, s *store.Store, calls *atomic.Int32, want string) {
	t.Helper()
	ctx := context.Background()
	waitForCondition(t, 2*time.Second, func() bool { turn, err := s.GetTurn(ctx, "interrupted"); return err == nil && turn.Status == want }, "staged terminal recovery")
	if calls.Load() != 0 {
		t.Fatal("recovery dispatched inference", calls.Load())
	}
	messages, err := s.ListMessages(ctx, "recovery")
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 3 {
		t.Fatal("recovery duplicated history", messages)
	}
	identities := map[string]bool{}
	for _, message := range messages {
		identities[message.ID] = true
	}
	if !identities["user-original"] || !identities["reply-original"] || !identities["outcome-original"] {
		t.Fatal("recovery changed committed identities", messages)
	}
	accounting, err := s.TurnAccounting(ctx, "interrupted")
	if err != nil || accounting.Usage.TotalTokens != 10 || accounting.Iterations != 1 || accounting.Unknown != 0 {
		t.Fatal("recovery changed accounting", accounting, err)
	}
	var n int
	if err := s.DB().QueryRow(`select count(*) from turn_events where turn_id='interrupted' and event_type='turn.finished'`).Scan(&n); err != nil || n != 1 {
		t.Fatal("terminal decision duplicated", n, err)
	}
}
func TestStagedTerminalRecoverySkipsSetupAndDispatch(t *testing.T) {
	for _, abort := range []bool{false, true} {
		t.Run(fmt.Sprint(abort), func(t *testing.T) {
			s := openChildJoinStore(t)
			e := New(s)
			defer e.Close()
			stageRecoveryTurn(t, s)
			claim := staleRecoveryClaim(t, s)
			var calls atomic.Int32
			withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
				calls.Add(1)
				return nil, fmt.Errorf("unexpected replay")
			})
			if abort {
				claim.Phase = "cancelling"
			}
			if err := e.recoverInterruptedTurn(context.Background(), claim); err != nil {
				t.Fatal(err)
			}
			want := "failed"
			if abort {
				want = "cancelled"
			}
			assertRecoveredTerminal(t, s, &calls, want)
		})
	}
}
func TestStagedTerminalRecoveryRefusesLiveOrRefreshedOwner(t *testing.T) {
	for _, live := range []bool{false, true} {
		t.Run(fmt.Sprint(live), func(t *testing.T) {
			s := openChildJoinStore(t)
			e := New(s)
			defer e.Close()
			stageRecoveryTurn(t, s)
			claim := staleRecoveryClaim(t, s)
			r := e.runner("recovery")
			if live {
				r.mu.Lock()
				r.current = &runningTurn{turnID: "interrupted", claimToken: "old", cancel: func() {}}
				r.mu.Unlock()
				defer func() { r.mu.Lock(); r.current = nil; r.mu.Unlock() }()
			} else if err := s.TouchSessionActiveTurn(context.Background(), "recovery", "old"); err != nil {
				t.Fatal(err)
			}
			if err := e.recoverInterruptedTurn(context.Background(), claim); !errors.Is(err, store.ErrQueueConflict) {
				t.Fatal("live owner replaced", err)
			}
			_, token, err := s.GetSessionActiveTurn(context.Background(), "recovery")
			if err != nil || token != "old" {
				t.Fatal(token, err)
			}
			if terminal, err := s.HasTurnTerminalReceipt(context.Background(), "interrupted"); err != nil || terminal {
				t.Fatal("live turn finalised", terminal, err)
			}
		})
	}
}

// The helper is killed after both receipt and staged decision commits. Reopening
// the WAL must consume those exact records without entering setup or inference.
func TestStagedTerminalRecoveryProcessHelper(t *testing.T) {
	path := os.Getenv("GI_STAGED_RECOVERY_DB")
	if path == "" {
		return
	}
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	stageRecoveryTurn(t, s)
	if err := os.WriteFile(path+".ready", []byte("ready"), 0600); err != nil {
		t.Fatal(err)
	}
	for {
		time.Sleep(time.Hour)
	}
}
func TestStagedTerminalRecoveryAfterKilledProcess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "killed.db")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(executable, "-test.run=^TestStagedTerminalRecoveryProcessHelper$")
	cmd.Env = append(os.Environ(), "GI_STAGED_RECOVERY_DB="+path)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	joined := false
	defer func() {
		if !joined {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	waitForCondition(t, 5*time.Second, func() bool { _, err := os.Stat(path + ".ready"); return err == nil }, "child process checkpoint")
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	err = cmd.Wait()
	joined = true
	if err == nil {
		t.Fatal("helper was not killed", output.String())
	}
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	staleRecoveryClaim(t, s)
	var calls atomic.Int32
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		calls.Add(1)
		return nil, fmt.Errorf("unexpected replay")
	})
	e := New(s)
	defer e.Close()
	assertRecoveredTerminal(t, s, &calls, "failed")
}

func TestStagedTerminalRecoveryRetainsUnjoinedForeignChild(t *testing.T) {
	s := openChildJoinStore(t)
	e := New(s)
	defer e.Close()
	ctx := context.Background()
	stageRecoveryTurn(t, s, func() {
		if _, err := s.CreateSession(ctx, "foreign", "foreign", nil); err != nil {
			t.Fatal(err)
		}
		if _, err := s.CreateTurnWithStatus(ctx, "child", "foreign", "running", "child", nil); err != nil {
			t.Fatal(err)
		}
		if _, err := s.CreateSubTurn(ctx, "interrupted", "recovery", "child", "foreign", "sync", 1, nil); err != nil {
			t.Fatal(err)
		}
		if ok, err := s.ClaimSessionActiveTurn(ctx, "foreign", "child", "foreign-process", "foreign-live"); err != nil || !ok {
			t.Fatal(ok, err)
		}
		if _, err := s.SettleRunningTurn(ctx, "child", "foreign", store.TurnSettlement{Status: "completed", Phase: "completed", ClaimToken: "foreign-live"}); err != nil {
			t.Fatal(err)
		}
	})
	claim := staleRecoveryClaim(t, s)
	if err := e.recoverInterruptedTurn(ctx, claim); err != nil {
		t.Fatal(err)
	}
	e.runs.Wait()
	turn, err := s.GetTurn(ctx, "interrupted")
	if err != nil || turn.Status != "running" || turn.Phase != "settling" {
		t.Fatal("foreign owner did not retain checkpoint", turn, err)
	}
	if terminal, err := s.HasTurnTerminalReceipt(ctx, "interrupted"); err != nil || terminal {
		t.Fatal("unjoined foreign child finalised parent", terminal, err)
	}
	_, token, err := s.GetSessionActiveTurn(ctx, "recovery")
	if err != nil || token == "old" || token == "" {
		t.Fatal("pending recovery lost claim", token, err)
	}
	if err := s.ReleaseSessionActiveTurn(ctx, "foreign", "foreign-live"); err != nil {
		t.Fatal(err)
	}
	claim = staleRecoveryClaim(t, s)
	var calls atomic.Int32
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		calls.Add(1)
		return nil, fmt.Errorf("unexpected replay")
	})
	if err := e.recoverInterruptedTurn(ctx, claim); err != nil {
		t.Fatal(err)
	}
	assertRecoveredTerminal(t, s, &calls, "failed")
	if terminalEventSequence(t, s, "child") >= terminalEventSequence(t, s, "interrupted") {
		t.Fatal("parent receipt preceded child")
	}
}

func BenchmarkTerminalRecoveryClassification(b *testing.B) {
	s, err := store.Open(filepath.Join(b.TempDir(), "recovery.db"))
	if err != nil {
		b.Fatal(err)
	}
	e := New(s)
	b.Cleanup(func() { _ = e.Close(); _ = s.Close() })
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "probe", "probe", nil); err != nil {
		b.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "probe", "probe", "running", "probe", nil); err != nil {
		b.Fatal(err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, "probe", "cancelling", "cancelling"); err != nil {
		b.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "probe", "probe", "dead", "old"); err != nil || !ok {
		b.Fatal(ok, err)
	}
	payload := map[string]any{"text": string(bytes.Repeat([]byte("x"), 1024))}
	for range 1000 {
		if err := s.AppendTurnEvent(ctx, "probe", "probe", "tool.output", payload); err != nil {
			b.Fatal(err)
		}
	}
	claim := store.ActiveTurnClaim{TurnID: "probe", SessionID: "probe", ClaimToken: "old", Status: "cancelling", Phase: "cancelling"}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if handled, err := e.recoverStagedTerminal(ctx, claim); err != nil || handled {
			b.Fatal(handled, err)
		}
	}
}

package turn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// SIGKILL intentionally bypasses Close: the next host must resume the committed
// retry successor, not the failed request or its once-only lifecycle hooks.
func TestProviderRetryRecoveryProcessHelper(t *testing.T) {
	path := os.Getenv("GI_RETRY_RECOVERY_DB")
	if path == "" {
		return
	}
	root := filepath.Dir(path)
	t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(root, "agent"))
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	e := NewWithRuntimeConfig(s, config.Load(root), "")
	if _, err = s.CreateSession(context.Background(), "recovery", "recovery", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = e.RegisterHook(HookBeforeAgentStart, "once", func(context.Context, HookRequest) (HookResponse, error) {
		return HookResponse{Message: "captured once before crash"}, nil
	}); err != nil {
		t.Fatal(err)
	}
	woken := os.Getenv("GI_RETRY_RECOVERY_WAKE") == "1"
	if woken {
		var preparations atomic.Int32
		if _, err = e.RegisterHook(HookBeforeProviderRequest, "stop after wake", func(ctx context.Context, req HookRequest) (HookResponse, error) {
			if req.Payload["stage"] == "context" && preparations.Add(1) == 2 {
				if err := publishRetryProcessReady(path, req.TurnID); err != nil {
					return HookResponse{}, err
				}
				<-ctx.Done()
			}
			return HookResponse{}, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Usage: &goai.Usage{Input: 3}}, errors.New("HTTP 503 unavailable")
	})
	result, err := e.SubmitPrompt(context.Background(), RunInput{SessionID: "recovery", Prompt: "original killed-process prompt", Model: "mock-stream"})
	if err != nil {
		t.Fatal(err)
	}
	if woken {
		waitForCondition(t, 5*time.Second, func() bool { _, err := os.Stat(path + ".ready"); return err == nil }, "wake before provider admission")
	} else {
		waitForCondition(t, 5*time.Second, func() bool {
			row, err := s.GetTurn(context.Background(), result.TurnID)
			return err == nil && row.Phase == "retry_wait"
		}, "committed retry successor")
		if err = publishRetryProcessReady(path, result.TurnID); err != nil {
			t.Fatal(err)
		}
	}
	for {
		time.Sleep(time.Hour)
	}
}

// Publish the complete identity before the parent can observe readiness and
// kill us. WriteFile alone exposes an empty file between open and write.
func publishRetryProcessReady(path, turnID string) error {
	if err := os.WriteFile(path+".ready.tmp", []byte(turnID), 0600); err != nil {
		return err
	}
	return os.Rename(path+".ready.tmp", path+".ready")
}

func TestProviderRetryRecoveryAfterKilledProcess(t *testing.T) {
	for _, mode := range []string{"waiting", "woken"} {
		t.Run(mode, func(t *testing.T) { testProviderRetryRecoveryAfterKilledProcess(t, mode == "woken") })
	}
}

func testProviderRetryRecoveryAfterKilledProcess(t *testing.T, woken bool) {
	root := t.TempDir()
	path := filepath.Join(root, "killed.db")
	t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(root, "agent"))
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"defaultModel":"mock-stream","retry":{"maxRetries":2,"baseDelayMs":1500,"maxAgentDelayMs":1500}}`), 0600); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(executable, "-test.run=^TestProviderRetryRecoveryProcessHelper$")
	child.Env = append(os.Environ(), "GI_RETRY_RECOVERY_DB="+path)
	if woken {
		child.Env = append(child.Env, "GI_RETRY_RECOVERY_WAKE=1")
	}
	var output bytes.Buffer
	child.Stdout = &output
	child.Stderr = &output
	if err = child.Start(); err != nil {
		t.Fatal(err)
	}
	joined := false
	defer func() {
		if !joined {
			_ = child.Process.Kill()
			_ = child.Wait()
		}
	}()
	waitForCondition(t, 5*time.Second, func() bool { _, err := os.Stat(path + ".ready"); return err == nil }, "retry process checkpoint")
	if err = child.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	err = child.Wait()
	joined = true
	if err == nil {
		t.Fatal("helper was not killed", output.String())
	}
	rawID, err := os.ReadFile(path + ".ready")
	if err != nil {
		t.Fatal(err)
	}
	turnID := string(rawID)
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if woken {
		row, err := s.GetTurn(context.Background(), turnID)
		if err != nil || row.Phase != "running" {
			t.Fatal("helper was not killed after wake", row, err)
		}
		var intents int
		if err = s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='inference.started'`, turnID).Scan(&intents); err != nil || intents != 1 {
			t.Fatal("second request was admitted before crash", intents, err)
		}
	}
	if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z' where session_id='recovery'`); err != nil {
		t.Fatal(err)
	}
	var raw string
	if err = s.DB().QueryRow(`select decision_json from turn_receipts where turn_id=? and kind='provider'`, turnID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var saved struct {
		Wait *store.ProviderRetryWait `json:"retry_wait"`
	}
	if err = json.Unmarshal([]byte(raw), &saved); err != nil || saved.Wait == nil || saved.Wait.Lifecycle == nil {
		t.Fatal("child did not commit lifecycle", raw, err)
	}
	var calls, starts atomic.Int32
	withStreamWithToolsStub(t, func(_ context.Context, _ string, request *goai.Context, _ func(map[string]any)) (*inference.StreamResult, error) {
		calls.Add(1)
		if time.Now().Before(saved.Wait.RetryAt) {
			t.Error("restart reset absolute deadline")
		}
		prefixes := 0
		for _, message := range request.Messages {
			for _, block := range message.Content {
				if block.Text == "captured once before crash" {
					prefixes++
				}
			}
		}
		if prefixes != 1 {
			t.Error("lost or replayed lifecycle prefix", prefixes)
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "continued after restart"}}, Usage: &goai.Usage{Output: 2}}}, nil
	})
	e := NewWithRuntimeConfig(s, config.Load(root), "")
	defer e.Close()
	if calls.Load() != 0 {
		t.Fatal("dispatch before runtime initialization")
	}
	if _, err = e.RegisterHook(HookBeforeAgentStart, "no replay", func(context.Context, HookRequest) (HookResponse, error) { starts.Add(1); return HookResponse{}, nil }); err != nil {
		t.Fatal(err)
	}
	if err = e.ResumeInterruptedWork(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitForCondition(t, 4*time.Second, func() bool {
		row, err := s.GetTurn(context.Background(), turnID)
		return err == nil && row.Status == "completed"
	}, "killed process continuation")
	if err = e.Close(); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 || starts.Load() != 0 {
		t.Fatal("replayed process lifecycle", calls.Load(), starts.Load())
	}
	var users, receipts int
	if err = s.DB().QueryRow(`select count(*) from messages where session_id='recovery' and role='user' and content='original killed-process prompt'`).Scan(&users); err != nil || users != 1 {
		t.Fatal("original prompt replayed", users, err)
	}
	if err = s.DB().QueryRow(`select count(*) from turn_receipts where turn_id=? and kind='provider'`, turnID).Scan(&receipts); err != nil || receipts != 2 {
		t.Fatal("provider receipts replayed", receipts, err)
	}
	var wakes int
	if err = s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='inference.retry_resumed'`, turnID).Scan(&wakes); err != nil || wakes != 1 {
		t.Fatal("wake was repeated", wakes, err)
	}
	accounting, err := s.TurnAccounting(context.Background(), turnID)
	if err != nil || accounting.Usage.Input != 3 || accounting.Usage.Output != 2 || accounting.Iterations != 1 {
		t.Fatal("restart changed accounting", accounting, err)
	}
}

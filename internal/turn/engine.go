package turn

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sort"
	"sync"
	"syscall"

	"github.com/rcarmo/gi/internal/store"
)

type Engine struct {
	store    *store.Store
	sessions sync.Map // sessionID -> *sessionRunner
}

type sessionRunner struct {
	mu      sync.Mutex
	store   *store.Store
	busy    bool
	current *runningTurn
}

type runningTurn struct {
	turnID string
	cancel context.CancelFunc
	cmd    *exec.Cmd
	cmdMu  sync.Mutex
}

type RunInput struct {
	SessionID string
	Prompt    string
	Intent    string
	Model     string
}

type SubmitResult struct {
	TurnID    string `json:"turn_id"`
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Queued    bool   `json:"queued"`
}

type Summary struct {
	TurnID    string            `json:"turn_id"`
	SessionID string            `json:"session_id"`
	Status    string            `json:"status"`
	Assistant string            `json:"assistant"`
	Events    []store.TurnEvent `json:"events"`
}

func New(s *store.Store) *Engine { return &Engine{store: s} }

func (e *Engine) SubmitPrompt(ctx context.Context, in RunInput) (*SubmitResult, error) {
	if in.Intent == "" {
		in.Intent = "prompt"
	}
	turnID := store.NowID("turn")
	runner := e.runner(in.SessionID)
	runner.mu.Lock()
	queued := runner.busy
	status := "running"
	if queued {
		status = "queued"
	}
	metadata := map[string]any{"intent": in.Intent, "model": in.Model, "queued": queued}
	if _, err := e.store.CreateTurnWithStatus(ctx, turnID, in.SessionID, status, in.Prompt, metadata); err != nil {
		runner.mu.Unlock()
		return nil, err
	}
	if err := e.store.AppendTurnEvent(ctx, turnID, in.SessionID, "turn.submitted", map[string]any{"phase": "queue", "intent": in.Intent, "queued": queued, "checkpoint": true}); err != nil {
		runner.mu.Unlock()
		return nil, err
	}
	if queued {
		if err := e.store.AddMessage(ctx, store.NowID("msg"), in.SessionID, "system", fmt.Sprintf("Queued prompt: %s", in.Prompt), map[string]any{"kind": "queue", "turn_id": turnID, "intent": in.Intent}); err != nil {
			runner.mu.Unlock()
			return nil, err
		}
	} else {
		runner.busy = true
		go runner.runTurn(e.store, turnID)
	}
	runner.mu.Unlock()
	_ = e.store.TouchSessionState(ctx, in.SessionID, map[string]any{"model": in.Model})
	return &SubmitResult{TurnID: turnID, SessionID: in.SessionID, Status: status, Queued: queued}, nil
}

func (e *Engine) CancelTurn(ctx context.Context, sessionID, turnID string) error {
	runner := e.runner(sessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.current != nil && runner.current.turnID == turnID {
		_ = e.store.AppendTurnEvent(ctx, turnID, sessionID, "turn.cancelling", map[string]any{"phase": "cancel", "checkpoint": true})
		_ = e.store.UpdateTurnStatus(ctx, turnID, "cancelling")
		runner.current.cancel()
		runner.current.cmdMu.Lock()
		if runner.current.cmd != nil && runner.current.cmd.Process != nil {
			_ = syscall.Kill(-runner.current.cmd.Process.Pid, syscall.SIGKILL)
		}
		runner.current.cmdMu.Unlock()
		return nil
	}
	turn, err := e.store.GetTurn(ctx, turnID)
	if err != nil {
		return err
	}
	if turn.Status == "queued" {
		if err := e.store.UpdateTurnStatus(ctx, turnID, "cancelled"); err != nil {
			return err
		}
		return e.store.AppendTurnEvent(ctx, turnID, sessionID, "turn.cancelled", map[string]any{"phase": "cancel", "checkpoint": true, "queued": true})
	}
	return fmt.Errorf("turn not cancellable")
}

func (e *Engine) runner(sessionID string) *sessionRunner {
	v, _ := e.sessions.LoadOrStore(sessionID, &sessionRunner{store: e.store})
	return v.(*sessionRunner)
}

func (r *sessionRunner) runTurn(s *store.Store, turnID string) {
	ctx, cancel := context.WithCancel(context.Background())
	r.mu.Lock()
	r.current = &runningTurn{turnID: turnID, cancel: cancel}
	r.mu.Unlock()
	defer func() {
		cancel()
		r.mu.Lock()
		r.current = nil
		next, _ := s.GetNextQueuedTurn(context.Background(), turnIDSession(s, turnID))
		if next != nil {
			_ = s.UpdateTurnStatus(context.Background(), next.ID, "running")
			go r.runTurn(s, next.ID)
		} else {
			r.busy = false
		}
		r.mu.Unlock()
	}()

	turnRec, err := s.GetTurn(ctx, turnID)
	if err != nil {
		return
	}
	sessionID := turnRec.SessionID
	prompt := turnRec.Prompt
	intent := stringValue(turnRec.Metadata["intent"], "prompt")
	model := stringValue(turnRec.Metadata["model"], "bootstrap")
	_ = s.TouchSessionState(ctx, sessionID, map[string]any{"active_turn_id": turnID, "model": model, "status": "running"})
	_ = s.AddMessage(ctx, store.NowID("msg"), sessionID, "user", prompt, map[string]any{"kind": "chat", "intent": intent, "turn_id": turnID})
	_ = s.AppendTurnEvent(ctx, turnID, sessionID, "turn.started", map[string]any{"phase": "turn", "prompt": prompt, "intent": intent, "checkpoint": true})
	_ = s.AppendTurnEvent(ctx, turnID, sessionID, "tool.started", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "command": []string{"sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\""}})

	out, runErr, cancelled := runShell(ctx, prompt, func(cmd *exec.Cmd) {
		r.mu.Lock()
		if r.current != nil && r.current.turnID == turnID {
			r.current.cmdMu.Lock()
			r.current.cmd = cmd
			r.current.cmdMu.Unlock()
		}
		r.mu.Unlock()
	})
	if cancelled {
		_ = s.AppendTurnEvent(ctx, turnID, sessionID, "turn.cancelled", map[string]any{"phase": "cancel", "checkpoint": true})
		_ = s.UpdateTurnStatus(context.Background(), turnID, "cancelled")
		_ = s.AddMessage(context.Background(), store.NowID("msg"), sessionID, "system", "Turn cancelled", map[string]any{"kind": "status", "turn_id": turnID, "clipped": true})
		_ = s.TouchSessionState(context.Background(), sessionID, map[string]any{"status": "idle", "active_turn_id": nil})
		return
	}
	if runErr != nil {
		_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "tool.failed", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "error": runErr.Error()})
		_ = s.UpdateTurnStatus(context.Background(), turnID, "failed")
		_ = s.TouchSessionState(context.Background(), sessionID, map[string]any{"status": "idle", "active_turn_id": nil})
		return
	}
	_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "tool.finished", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "output": out})
	_ = s.AddMessage(context.Background(), store.NowID("msg"), sessionID, "assistant", out, map[string]any{"kind": "chat", "source": "shell", "turn_id": turnID})
	_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "completed"})
	_ = s.UpdateTurnStatus(context.Background(), turnID, "completed")
	_ = s.TouchSessionState(context.Background(), sessionID, map[string]any{"status": "idle", "active_turn_id": nil})
}

func (e *Engine) Summary(ctx context.Context, turnID string) (*Summary, error) {
	turnRec, err := e.store.GetTurn(ctx, turnID)
	if err != nil {
		return nil, err
	}
	events, err := e.store.ListTurnEvents(ctx, turnID)
	if err != nil {
		return nil, err
	}
	msgs, _ := e.store.ListMessages(ctx, turnRec.SessionID)
	assistant := ""
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "assistant" {
			assistant = msgs[i].Content
			break
		}
	}
	return &Summary{TurnID: turnID, SessionID: turnRec.SessionID, Status: turnRec.Status, Assistant: assistant, Events: events}, nil
}

func runShell(ctx context.Context, prompt string, onStart func(*exec.Cmd)) (string, error, bool) {
	cmd := exec.Command("sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\"; sleep 1")
	cmd.Env = append(cmd.Environ(), "GI_PROMPT="+prompt)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return "", err, false
	}
	if onStart != nil {
		onStart(cmd)
	}
	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()
	select {
	case <-ctx.Done():
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		<-waitCh
		return stdout.String(), nil, true
	case err := <-waitCh:
		if err != nil {
			if stderr.Len() > 0 {
				return "", fmt.Errorf("%w: %s", err, stderr.String()), false
			}
			return "", err, false
		}
		if stderr.Len() > 0 {
			return stdout.String(), fmt.Errorf("stderr: %s", stderr.String()), false
		}
		return stdout.String(), nil, false
	}
}

func stringValue(v any, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}

func turnIDSession(s *store.Store, turnID string) string {
	turnRec, err := s.GetTurn(context.Background(), turnID)
	if err != nil {
		return ""
	}
	return turnRec.SessionID
}

func SortQueuedTurns(turns []store.Turn) {
	sort.SliceStable(turns, func(i, j int) bool { return turns[i].CreatedAt < turns[j].CreatedAt })
}

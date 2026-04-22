package turn

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"syscall"

	"github.com/rcarmo/gi/internal/store"
)

type Engine struct {
	store *store.Store
	mu    sync.Map
}

type RunInput struct {
	SessionID string
	Prompt    string
	Intent    string
	Model     string
}

type Summary struct {
	TurnID      string         `json:"turn_id"`
	SessionID   string         `json:"session_id"`
	Status      string         `json:"status"`
	Assistant   string         `json:"assistant"`
	Events      []store.TurnEvent `json:"events"`
}

func New(s *store.Store) *Engine {
	return &Engine{store: s}
}

func (e *Engine) RunPrompt(ctx context.Context, in RunInput) (*Summary, error) {
	unlocker, _ := e.mu.LoadOrStore(in.SessionID, &sync.Mutex{})
	mu := unlocker.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	turnID := store.NowID("turn")
	if in.Intent == "" {
		in.Intent = "prompt"
	}
	metadata := map[string]any{"intent": in.Intent, "model": in.Model}
	if _, err := e.store.CreateTurn(ctx, turnID, in.SessionID, in.Prompt, metadata); err != nil {
		return nil, err
	}
	if err := e.store.SetSessionState(ctx, in.SessionID, map[string]any{"active_turn_id": turnID, "model": in.Model}); err != nil {
		return nil, err
	}
	if err := e.store.AddMessage(ctx, store.NowID("msg"), in.SessionID, "user", in.Prompt, map[string]any{"kind": "chat", "intent": in.Intent}); err != nil {
		return nil, err
	}
	if err := e.store.AppendTurnEvent(ctx, turnID, in.SessionID, "turn.started", map[string]any{"phase": "turn", "prompt": in.Prompt, "intent": in.Intent, "checkpoint": true}); err != nil {
		return nil, err
	}
	if err := e.store.AppendTurnEvent(ctx, turnID, in.SessionID, "tool.started", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "command": []string{"sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\""}}); err != nil {
		return nil, err
	}

	out, err := runShell(ctx, in.Prompt)
	if err != nil {
		_ = e.store.AppendTurnEvent(ctx, turnID, in.SessionID, "tool.failed", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "error": err.Error()})
		_ = e.store.UpdateTurnStatus(ctx, turnID, "failed")
		return e.summary(ctx, turnID, in.SessionID, "failed", "", err)
	}

	if err := e.store.AppendTurnEvent(ctx, turnID, in.SessionID, "tool.finished", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "output": out}); err != nil {
		return nil, err
	}
	if err := e.store.AddMessage(ctx, store.NowID("msg"), in.SessionID, "assistant", out, map[string]any{"kind": "chat", "source": "shell"}); err != nil {
		return nil, err
	}
	if err := e.store.AppendTurnEvent(ctx, turnID, in.SessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "completed"}); err != nil {
		return nil, err
	}
	if err := e.store.UpdateTurnStatus(ctx, turnID, "completed"); err != nil {
		return nil, err
	}
	return e.summary(ctx, turnID, in.SessionID, "completed", out, nil)
}

func (e *Engine) summary(ctx context.Context, turnID, sessionID, status, assistant string, cause error) (*Summary, error) {
	events, err := e.store.ListTurnEvents(ctx, turnID)
	if err != nil {
		return nil, err
	}
	s := &Summary{TurnID: turnID, SessionID: sessionID, Status: status, Assistant: assistant, Events: events}
	if cause != nil {
		return s, cause
	}
	return s, nil
}

func runShell(ctx context.Context, prompt string) (string, error) {
	cmd := exec.CommandContext(ctx, "sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\"")
	cmd.Env = append(cmd.Environ(), "GI_PROMPT="+prompt)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("%w: %s", err, stderr.String())
		}
		return "", err
	}
	if stderr.Len() > 0 {
		return stdout.String(), fmt.Errorf("stderr: %s", stderr.String())
	}
	return stdout.String(), nil
}


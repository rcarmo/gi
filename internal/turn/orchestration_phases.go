package turn

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/routing"
	"github.com/rcarmo/gi/internal/store"
)

type preparedTurnRun struct {
	turn            *store.Turn
	sessionID       string
	turnID          string
	prompt          string
	intent          string
	model           string
	agentID         string
	initialSteering []store.SteeringMessage
}

func (r *sessionRunner) cleanupTurnRun(sessionID, claimToken string, active *runningTurn) {
	ctx := r.engine.backgroundContext()
	r.mu.Lock()
	defer r.mu.Unlock()
	warnStore("release session active turn", r.store.ReleaseSessionActiveTurn(ctx, sessionID, claimToken))
	warnStore("sync session queue count", r.store.SyncSessionQueueCount(ctx, sessionID))
	if r.current == active {
		r.current = nil
	}
	if hook := r.engine.beforeCleanupNextWorkHook; hook != nil {
		hook(ctx, sessionID)
	}
	launched, err := r.engine.startNextQueuedTurnLocked(ctx, r, sessionID)
	if err != nil {
		log.Printf("turn coordination: launch queued turn failed: %v", err)
		return
	}
	if launched {
		return
	}
	if _, _, err := r.store.GetSessionActiveTurn(ctx, sessionID); err == sql.ErrNoRows {
		if _, err := r.engine.continueQueuedSteeringLocked(ctx, r, sessionID); err != nil {
			log.Printf("steering continuation: %v", err)
		}
	}
}

func (r *sessionRunner) setupTurnRun(ctx context.Context, s *store.Store, sessionID, turnID string) (*preparedTurnRun, error) {
	turnRec, err := s.GetTurn(ctx, turnID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(sessionID) == "" {
		sessionID = turnRec.SessionID
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	initialSteering := steeringMessagesFromMetadata(turnRec.Metadata)
	prompt := turnRec.Prompt
	if strings.TrimSpace(prompt) == "" && len(initialSteering) > 0 {
		prompt = steeringPromptForShell(initialSteering)
	}
	intent := stringValue(turnRec.Metadata["intent"], "prompt")
	agentID, model := r.resolveTurnAgentAndModel(ctx, s, turnRec, sessionID, prompt)
	if hook := r.engine.beforeSetupHook; hook != nil {
		hook(ctx, sessionID, turnID)
	}
	if hook := r.engine.beforeSetupErrorHook; hook != nil {
		if err := hook(ctx, sessionID, turnID); err != nil {
			return nil, err
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	warnStore("touch session state running", s.TouchSessionState(ctx, sessionID, map[string]any{"active_turn_id": turnID, "model": model, "status": "running"}))
	r.engine.PublishRuntimeSessionEvent("session_running", sessionID, agentID, "running", map[string]any{"reason": "setup", "active_turn_id": turnID, "model": model})
	r.emitSessionStateHookOnly(ctx, sessionID, agentID, model, "running", map[string]any{"reason": "setup", "active_turn_id": turnID})
	r.engine.PublishRuntimeTurnEvent("turn_started", sessionID, turnID, agentID, "running", "setup", map[string]any{"reason": "setup", "model": model})
	r.emitTurnStateHookOnly(ctx, sessionID, turnID, agentID, model, "running", "setup", map[string]any{"reason": "setup"})
	userPayload := map[string]any{"kind": "chat", "intent": intent, "turn_id": turnID}
	for _, key := range []string{"source_session_id", "source_agent_id", "target_agent_id", "routed_from_prompt", "ingress_kind", "ingress_source_kind", "ingress_source_id", "ingress_role", "ingress_label"} {
		if value, ok := turnRec.Metadata[key]; ok {
			userPayload[key] = value
		}
	}
	if strings.TrimSpace(prompt) != "" && len(initialSteering) == 0 {
		warnStore("add user prompt message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "user", prompt, userPayload))
	}
	startedPayload := map[string]any{"phase": "turn", "prompt": prompt, "intent": intent, "model": model, "checkpoint": true}
	for _, key := range []string{"source_session_id", "source_agent_id", "target_agent_id", "routed_from_prompt", "parent_turn_id", "route_mode", "route_matched_by", "ingress_kind", "ingress_source_kind", "ingress_source_id", "ingress_role", "ingress_label"} {
		if value, ok := turnRec.Metadata[key]; ok {
			startedPayload[key] = value
		}
	}
	warnStore("append turn.started event", s.AppendTurnEvent(ctx, turnID, sessionID, "turn.started", startedPayload))
	return &preparedTurnRun{
		turn:            turnRec,
		sessionID:       sessionID,
		turnID:          turnID,
		prompt:          prompt,
		intent:          intent,
		model:           model,
		agentID:         agentID,
		initialSteering: initialSteering,
	}, nil
}

func (r *sessionRunner) resolveTurnAgentAndModel(ctx context.Context, s *store.Store, turnRec *store.Turn, sessionID, prompt string) (string, string) {
	model := stringValue(turnRec.Metadata["model"], "bootstrap")
	agentID := "agent"
	if sess, err := s.GetSession(ctx, sessionID); err == nil {
		agentID = sessionAgentIDWithStore(ctx, s, sess)
	}
	agentModel := r.engine.modelForAgent(agentID)
	if strings.TrimSpace(model) == "" {
		model = agentModel
	}
	if model == agentModel {
		history, _ := s.ListMessages(ctx, sessionID)
		routingHistory := make([]routing.HistoryMessage, 0, len(history))
		for _, msg := range history {
			routingHistory = append(routingHistory, routing.HistoryMessage{Payload: msg.Payload})
		}
		selected, usedLight, score := r.engine.modelRouter.SelectModel(prompt, routingHistory, agentModel)
		if usedLight && strings.TrimSpace(selected) != "" {
			model = selected
			turnRec.Metadata["route_model_score"] = score
			turnRec.Metadata["route_used_light_model"] = usedLight
		}
	}
	return agentID, model
}

func (r *sessionRunner) runPreparedTurn(ctx context.Context, s *store.Store, run *preparedTurnRun) {
	if run.model != "bootstrap" && run.model != "test-model" && run.model != "" {
		r.runAgentLoop(ctx, s, run.turnID, run.sessionID, run.model, run.agentID, run.initialSteering)
		return
	}
	r.runShellTurn(ctx, s, run)
}

func (r *sessionRunner) runShellTurn(ctx context.Context, s *store.Store, run *preparedTurnRun) {
	if len(run.initialSteering) > 0 {
		r.persistSteeringMessages(ctx, run.sessionID, run.turnID, run.initialSteering)
	}
	r.engine.PublishRuntimeToolEvent("tool_started", run.sessionID, run.turnID, run.agentID, "shell", "", 0, nil, map[string]any{"phase": "tool", "command": []string{"sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\""}})
	warnStore("append shell tool.started event", s.AppendTurnEvent(ctx, run.turnID, run.sessionID, "tool.started", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "command": []string{"sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\""}}))

	out, runErr, cancelled := runShell(ctx, run.prompt, func(cmd *exec.Cmd) {
		r.mu.Lock()
		if r.current != nil && r.current.turnID == run.turnID {
			r.current.cmdMu.Lock()
			r.current.cmd = cmd
			r.current.cmdMu.Unlock()
		}
		r.mu.Unlock()
	}, func(delta string) {
		if strings.TrimSpace(delta) == "" {
			return
		}
		r.engine.broadcast(run.sessionID, map[string]any{
			"type":     "agent_draft_delta",
			"chat_jid": "gi:" + run.sessionID,
			"delta":    delta,
			"turn_id":  run.turnID,
		})
	})
	if cancelled {
		bgCtx := r.engine.backgroundContext()
		r.appendFinalSteeringCheckpoint(s, run.turnID, run.sessionID)
		warnStore("append turn.cancelled event", s.AppendTurnEvent(ctx, run.turnID, run.sessionID, "turn.cancelled", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "cancelled", "status": "cancelled", "turn_phase": "aborted", "failure_kind": ""}))
		warnStore("append turn.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "cancelled", "reason": "cancelled", "failure_kind": ""}))
		warnStore("update turn status cancelled", s.UpdateTurnStatus(bgCtx, run.turnID, "cancelled"))
		r.engine.PublishRuntimeTurnEvent("turn_terminal", run.sessionID, run.turnID, run.agentID, "cancelled", "aborted", map[string]any{"reason": "cancelled", "failure_kind": ""})
		r.emitTurnStateHookOnly(bgCtx, run.sessionID, run.turnID, run.agentID, run.model, "cancelled", "aborted", map[string]any{"reason": "cancelled", "failure_kind": ""})
		r.propagateChildSubTurnCancellation(bgCtx, run.turnID, "cancelled", "")
		r.publishSubTurnLifecycle(bgCtx, run.turnID, "cancelled")
		msgID := store.NowID("msg")
		warnStore("add turn cancelled system message", s.AddMessage(bgCtx, msgID, run.sessionID, "system", "Turn cancelled", map[string]any{"kind": "status", "turn_id": run.turnID, "clipped": true}))
		r.broadcastSystemPost(run.sessionID, run.turnID, msgID, "Turn cancelled")
		warnStore("touch session idle after cancel", s.TouchSessionState(bgCtx, run.sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
		r.engine.PublishRuntimeSessionEvent("session_idle", run.sessionID, run.agentID, "idle", map[string]any{"reason": "turn_terminal", "turn_id": run.turnID, "turn_status": "cancelled", "failure_kind": "", "model": run.model})
		r.emitSessionStateHookOnly(bgCtx, run.sessionID, run.agentID, run.model, "idle", map[string]any{"reason": "turn_terminal", "turn_status": "cancelled", "failure_kind": ""})
		return
	}
	if runErr != nil {
		bgCtx := r.engine.backgroundContext()
		r.appendFinalSteeringCheckpoint(s, run.turnID, run.sessionID)
		markTurnFailure(bgCtx, s, run.turnID, run.sessionID, "shell_error", runErr.Error())
		r.engine.PublishRuntimeToolEvent("tool_failed", run.sessionID, run.turnID, run.agentID, "shell", "", 0, runErr, map[string]any{"phase": "tool"})
		warnStore("append shell tool.failed event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "tool.failed", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "error": runErr.Error()}))
		msgID := store.NowID("msg")
		warnStore("add shell failure system message", s.AddMessage(bgCtx, msgID, run.sessionID, "system", fmt.Sprintf("Shell tool failed: %v", runErr), map[string]any{"kind": "status", "turn_id": run.turnID, "source": "system", "failure_kind": "shell_error"}))
		r.broadcastSystemPost(run.sessionID, run.turnID, msgID, fmt.Sprintf("Shell tool failed: %v", runErr))
		warnStore("append turn.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "failed", "reason": "shell_error", "failure_kind": "shell_error"}))
		warnStore("update turn status failed", s.UpdateTurnStatus(bgCtx, run.turnID, "failed"))
		r.engine.PublishRuntimeTurnEvent("turn_terminal", run.sessionID, run.turnID, run.agentID, "failed", "failed", map[string]any{"reason": "shell_error", "failure_kind": "shell_error"})
		r.emitTurnStateHookOnly(bgCtx, run.sessionID, run.turnID, run.agentID, run.model, "failed", "failed", map[string]any{"reason": "shell_error", "failure_kind": "shell_error"})
		r.propagateChildSubTurnCancellation(bgCtx, run.turnID, "failed", "shell_error")
		r.publishSubTurnLifecycle(bgCtx, run.turnID, "failed")
		warnStore("touch session idle after failure", s.TouchSessionState(bgCtx, run.sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
		r.engine.PublishRuntimeSessionEvent("session_idle", run.sessionID, run.agentID, "idle", map[string]any{"reason": "turn_terminal", "turn_id": run.turnID, "turn_status": "failed", "failure_kind": "shell_error", "model": run.model})
		r.emitSessionStateHookOnly(bgCtx, run.sessionID, run.agentID, run.model, "idle", map[string]any{"reason": "turn_terminal", "turn_status": "failed", "failure_kind": "shell_error"})
		return
	}
	bgCtx := r.engine.backgroundContext()
	r.appendFinalSteeringCheckpoint(s, run.turnID, run.sessionID)
	r.engine.PublishRuntimeToolEvent("tool_finished", run.sessionID, run.turnID, run.agentID, "shell", "", 0, nil, map[string]any{"phase": "tool", "output_length": len(out)})
	warnStore("append shell tool.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "tool.finished", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "output": out}))
	msgID := store.NowID("msg")
	warnStore("add shell assistant message", s.AddMessage(bgCtx, msgID, run.sessionID, "assistant", out, map[string]any{"kind": "chat", "source": "shell", "turn_id": run.turnID, "agent_id": run.agentID}))
	r.engine.broadcast(run.sessionID, map[string]any{
		"type": "new_post", "id": msgID, "chat_jid": "gi:" + run.sessionID,
		"content": out, "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"sender": "agent", "is_bot_message": true,
		"data": map[string]any{"type": "agent_response", "content": out, "agent_id": run.agentID},
	})
	completionPayload := map[string]any{"reason": "completed", "completion_kind": "response"}
	warnStore("append turn.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "completed", "reason": "completed", "failure_kind": "", "completion_kind": "response"}))
	warnStore("update turn status completed", s.UpdateTurnStatus(bgCtx, run.turnID, "completed"))
	r.engine.PublishRuntimeTurnEvent("turn_completed", run.sessionID, run.turnID, run.agentID, "completed", "completed", completionPayload)
	r.emitTurnStateHookOnly(bgCtx, run.sessionID, run.turnID, run.agentID, run.model, "completed", "completed", completionPayload)
	r.propagateChildSubTurnCancellation(bgCtx, run.turnID, "completed", "")
	r.publishSubTurnLifecycle(bgCtx, run.turnID, "completed")
	warnStore("touch session idle after completion", s.TouchSessionState(bgCtx, run.sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
	sessionCompletionPayload := map[string]any{"reason": "turn_completed", "turn_id": run.turnID, "turn_status": "completed", "failure_kind": "", "model": run.model, "completion_kind": "response"}
	r.engine.PublishRuntimeSessionEvent("session_idle", run.sessionID, run.agentID, "idle", sessionCompletionPayload)
	r.emitSessionStateHookOnly(bgCtx, run.sessionID, run.agentID, run.model, "idle", map[string]any{"reason": "turn_completed", "turn_status": "completed", "failure_kind": "", "completion_kind": "response"})
}

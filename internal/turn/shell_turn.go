package turn

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rcarmo/gi/internal/logutil"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
	"strings"
	"sync/atomic"
)

// shellTurnStartHook is a test seam: tests that assert on a still-active
// bootstrap turn hold it here instead of racing the shell's exit.
var shellTurnStartHook atomic.Pointer[func(context.Context)]

func (r *sessionRunner) runShellTurn(ctx context.Context, s *store.Store, run *preparedTurnRun) {
	if hook := shellTurnStartHook.Load(); hook != nil {
		(*hook)(ctx)
	}
	toolOccurrenceID := store.NowID("tool")
	if len(run.initialSteering) > 0 {
		r.persistSteeringMessages(ctx, run.sessionID, run.turnID, run.initialSteering)
	}
	intent := map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "occurrence_id": toolOccurrenceID, "command": []string{"sh", "-c", "printf 'Gi received: %s' \"$GI_PROMPT\""}, "prompt": run.prompt, "replay_safe": false}
	if err := s.AdmitTurnEffect(ctx, run.sessionID, run.turnID, "waiting_on_tools", "tool.started", intent); err != nil {
		r.finishTurn(s, run.turnID, run.sessionID, run.agentID, run.model, "failed", "Admit shell intent: "+err.Error(), "persistence_error")
		return
	}
	r.engine.PublishRuntimeToolEvent("tool_started", run.sessionID, run.turnID, run.agentID, "shell", "", 0, nil, intent)
	r.engine.broadcast(run.sessionID, map[string]any{"type": "tool_activity_changed", "chat_jid": "gi:" + run.sessionID, "turn_id": run.turnID})

	reportOutput, closeOutput := r.toolOutputReporter(run.turnID, run.sessionID, "", toolOccurrenceID)
	defer closeOutput()
	var streamedOutput strings.Builder
	executionStart := time.Now()
	out, runErr, cancelled := tools.RunShellPrompt(ctx, run.prompt, nil, func(delta string) {
		streamedOutput.WriteString(delta)
		_ = reportOutput(streamedOutput.String(), false)
	})
	durationMS := time.Since(executionStart).Milliseconds()
	if outputErr := reportOutput(out, true); outputErr != nil && !cancelled {
		runErr = outputErr
	}
	bgCtx := r.engine.backgroundContext()
	status, phase := "completed", "completed"
	if cancelled || ctx.Err() != nil {
		status, phase = "cancelled", "aborted"
	} else if runErr != nil {
		status, phase = "failed", "failed"
	}
	finalStatus, err := s.FinalizeRunningTurn(bgCtx, run.turnID, status, phase)
	if err != nil {
		log.Printf("shell turn finalization conflict for %s: %v", run.turnID, err)
		return
	}
	logutil.WarnIfErr("mark shell turn finished", s.MarkTurnFinished(bgCtx, run.turnID))
	if finalStatus == "cancelled" {
		r.persistStoppedTool(s, run.sessionID, run.turnID, goai.ToolCall{Name: "shell"}, toolOccurrenceID, "cancelled", durationMS)
		r.appendFinalSteeringCheckpoint(s, run.turnID, run.sessionID)
		logutil.WarnIfErr("append turn.cancelled event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.cancelled", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "cancelled", "status": "cancelled", "turn_phase": "aborted", "failure_kind": ""}))
		logutil.WarnIfErr("append turn.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "cancelled", "reason": "cancelled", "failure_kind": ""}))
		r.engine.PublishRuntimeTurnEvent("turn_terminal", run.sessionID, run.turnID, run.agentID, "cancelled", "aborted", map[string]any{"reason": "cancelled", "failure_kind": ""})
		r.emitTurnStateHookOnly(bgCtx, run.sessionID, run.turnID, run.agentID, run.model, "cancelled", "aborted", map[string]any{"reason": "cancelled", "failure_kind": ""})
		r.propagateChildSubTurnCancellation(bgCtx, run.turnID, "cancelled", "")
		r.publishSubTurnLifecycle(bgCtx, run.turnID, "cancelled")
		msgID := store.NowID("msg")
		logutil.WarnIfErr("add turn cancelled system message", s.AddMessage(bgCtx, msgID, run.sessionID, "system", "Turn cancelled", map[string]any{"kind": "status", "turn_id": run.turnID, "clipped": true}))
		r.broadcastSystemPost(run.sessionID, run.turnID, msgID, "Turn cancelled")
		logutil.WarnIfErr("touch session idle after cancel", s.TouchSessionState(bgCtx, run.sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
		r.engine.PublishRuntimeSessionEvent("session_idle", run.sessionID, run.agentID, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "cancelled", "turn_phase": "aborted", "failure_kind": "", "model": run.model})
		r.emitSessionStateHookOnly(bgCtx, run.sessionID, run.agentID, run.model, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "cancelled", "turn_phase": "aborted", "failure_kind": ""})
		return
	}
	if finalStatus == "failed" {
		r.appendFinalSteeringCheckpoint(s, run.turnID, run.sessionID)
		logutil.WarnIfErr("turn failure mark", s.MarkTurnFailureWithFallbackErr(bgCtx, nil, run.turnID, run.sessionID, "shell_error", "none", runErr.Error()))
		r.engine.PublishRuntimeToolEvent("tool_failed", run.sessionID, run.turnID, run.agentID, "shell", "", 0, runErr, map[string]any{"occurrence_id": toolOccurrenceID, "phase": "tool", "duration_ms": durationMS})
		logutil.WarnIfErr("append shell tool.failed event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "tool.failed", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "occurrence_id": toolOccurrenceID, "duration_ms": durationMS, "error": runErr.Error(), "failure_kind": "shell_error"}))
		r.engine.broadcast(run.sessionID, map[string]any{"type": "tool_activity_changed", "chat_jid": "gi:" + run.sessionID, "turn_id": run.turnID})
		msgID := store.NowID("msg")
		logutil.WarnIfErr("add shell failure system message", s.AddMessage(bgCtx, msgID, run.sessionID, "system", fmt.Sprintf("Shell tool failed: %v", runErr), map[string]any{"kind": "status", "turn_id": run.turnID, "source": "system", "failure_kind": "shell_error"}))
		r.broadcastSystemPost(run.sessionID, run.turnID, msgID, fmt.Sprintf("Shell tool failed: %v", runErr))
		logutil.WarnIfErr("append turn.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "failed", "reason": "shell_error", "failure_kind": "shell_error"}))
		r.engine.PublishRuntimeTurnEvent("turn_terminal", run.sessionID, run.turnID, run.agentID, "failed", "failed", map[string]any{"reason": "shell_error", "failure_kind": "shell_error"})
		r.emitTurnStateHookOnly(bgCtx, run.sessionID, run.turnID, run.agentID, run.model, "failed", "failed", map[string]any{"reason": "shell_error", "failure_kind": "shell_error"})
		r.propagateChildSubTurnCancellation(bgCtx, run.turnID, "failed", "shell_error")
		r.publishSubTurnLifecycle(bgCtx, run.turnID, "failed")
		logutil.WarnIfErr("touch session idle after failure", s.TouchSessionState(bgCtx, run.sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
		r.engine.PublishRuntimeSessionEvent("session_idle", run.sessionID, run.agentID, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "failed", "turn_phase": "failed", "failure_kind": "shell_error", "model": run.model})
		r.emitSessionStateHookOnly(bgCtx, run.sessionID, run.agentID, run.model, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "failed", "turn_phase": "failed", "failure_kind": "shell_error"})
		return
	}
	r.appendFinalSteeringCheckpoint(s, run.turnID, run.sessionID)
	r.engine.PublishRuntimeToolEvent("tool_finished", run.sessionID, run.turnID, run.agentID, "shell", "", 0, nil, map[string]any{"occurrence_id": toolOccurrenceID, "phase": "tool", "output_length": len(out), "duration_ms": durationMS})
	logutil.WarnIfErr("append shell tool.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "tool.finished", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "occurrence_id": toolOccurrenceID, "duration_ms": durationMS, "output_length": len(out)}))
	r.engine.broadcast(run.sessionID, map[string]any{"type": "tool_activity_changed", "chat_jid": "gi:" + run.sessionID, "turn_id": run.turnID})
	msgID := store.NowID("msg")
	messagePayload := map[string]any{"kind": "chat", "source": "shell", "turn_id": run.turnID, "agent_id": run.agentID}
	r.addRecoveryMarker(bgCtx, run.sessionID, run.turnID, messagePayload)
	logutil.WarnIfErr("add shell assistant message", s.AddMessage(bgCtx, msgID, run.sessionID, "assistant", out, messagePayload))
	r.engine.broadcast(run.sessionID, map[string]any{
		"type": "new_post", "turn_id": run.turnID, "id": msgID, "chat_jid": "gi:" + run.sessionID,
		"content": out, "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"sender": "agent", "is_bot_message": true,
		"data": map[string]any{"type": "agent_response", "content": out, "agent_id": run.agentID, "content_blocks": messagePayload["content_blocks"]},
	})
	completionPayload := map[string]any{"reason": "completed", "completion_kind": "response"}
	logutil.WarnIfErr("append turn.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "completed", "reason": "completed", "failure_kind": "", "completion_kind": "response"}))
	r.engine.PublishRuntimeTurnEvent("turn_completed", run.sessionID, run.turnID, run.agentID, "completed", "completed", completionPayload)
	r.emitTurnStateHookOnly(bgCtx, run.sessionID, run.turnID, run.agentID, run.model, "completed", "completed", completionPayload)
	r.propagateChildSubTurnCancellation(bgCtx, run.turnID, "completed", "")
	r.publishSubTurnLifecycle(bgCtx, run.turnID, "completed")
	logutil.WarnIfErr("touch session idle after completion", s.TouchSessionState(bgCtx, run.sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
	sessionCompletionPayload := map[string]any{"reason": "turn_completed", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "completed", "turn_phase": "completed", "failure_kind": "", "model": run.model, "completion_kind": "response"}
	r.engine.PublishRuntimeSessionEvent("session_idle", run.sessionID, run.agentID, "idle", sessionCompletionPayload)
	r.emitSessionStateHookOnly(bgCtx, run.sessionID, run.agentID, run.model, "idle", map[string]any{"reason": "turn_completed", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "completed", "turn_phase": "completed", "failure_kind": "", "completion_kind": "response"})
}

// admissionThinking is the thinking level captured for a turn: the session's
// validated selection, else Pi's default (the configured default thinking
// level, else "medium") clamped to what the model supports. Non-reasoning and
// unknown models get none.

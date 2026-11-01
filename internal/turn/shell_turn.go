package turn

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
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

	reportOutput, closeOutput := r.toolOutputReporter(ctx, run.turnID, run.sessionID, "", toolOccurrenceID)
	defer closeOutput()
	var streamedOutput strings.Builder
	executionStart := time.Now()
	runPrompt := r.engine.shellPromptRunner
	if runPrompt == nil {
		runPrompt = func(ctx context.Context, prompt string, delta func(string)) (string, error, bool) {
			return tools.RunShellPrompt(ctx, prompt, nil, delta)
		}
	}
	out, runErr, cancelled := runPrompt(ctx, run.prompt, func(delta string) {
		streamedOutput.WriteString(delta)
		_ = reportOutput(streamedOutput.String(), false)
	})
	durationMS := time.Since(executionStart).Milliseconds()
	if (cancelled || ctx.Err() != nil || runErr != nil) && streamedOutput.Len() > len(out) {
		out = streamedOutput.String()
	}
	if outputErr := reportOutput(out, true); outputErr != nil && !cancelled {
		runErr = outputErr
	}
	// Join every coalesced progress write before recording the result or
	// entering the common child-before-parent terminal barrier.
	closeOutput()
	bgCtx := context.WithoutCancel(ctx)
	status, eventType, systemMessage, failureKind := "completed", "tool.finished", "", ""
	if cancelled || ctx.Err() != nil {
		status, eventType, systemMessage = "cancelled", "tool.cancelled", "Turn cancelled"
	} else if runErr != nil {
		status, eventType, failureKind = "failed", "tool.failed", "shell_error"
		systemMessage = fmt.Sprintf("Shell tool failed: %v", runErr)
	}
	out, diagnostics := boundToolOutput(out)
	messagePayload := withOutputDiagnostics(map[string]any{"kind": "chat", "source": "shell", "turn_id": run.turnID, "agent_id": run.agentID}, diagnostics)
	if status != "completed" {
		messagePayload["partial"] = true
		messagePayload["context_contribution"] = false
		messagePayload["adopted"] = false
		stopReason := goai.StopReasonError
		if status == "cancelled" {
			stopReason = goai.StopReasonAborted
		}
		messagePayload["stop_reason"] = stopReason
	} else {
		r.addRecoveryMarker(bgCtx, run.sessionID, run.turnID, messagePayload)
	}
	messageID := ""
	if out != "" {
		messageID = store.NowID("msg")
	}
	event := map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "occurrence_id": toolOccurrenceID, "duration_ms": durationMS, "output_length": len(out), "output_policy": diagnostics}
	if runErr != nil {
		event["error"] = runErr.Error()
	}
	if failureKind != "" {
		event["failure_kind"] = failureKind
	}
	r.setTerminalPending(run.turnID, true)
	if err := s.SettleToolResponse(bgCtx, run.turnID, run.sessionID, eventType, event, messageID, out, messagePayload); err != nil {
		log.Printf("shell result settlement conflict for %s: %v", run.turnID, err)
		return
	}
	r.engine.PublishRuntimeToolEvent(strings.ReplaceAll(eventType, ".", "_"), run.sessionID, run.turnID, run.agentID, "shell", "", 0, runErr, event)
	r.engine.broadcast(run.sessionID, map[string]any{"type": "tool_activity_changed", "chat_jid": "gi:" + run.sessionID, "turn_id": run.turnID})
	if messageID != "" {
		r.broadcastPost(run.sessionID, run.turnID, messageID, out, run.agentID, messagePayload["content_blocks"])
	}
	var completion map[string]any
	if status == "completed" {
		completion = map[string]any{"completion_kind": "response"}
	}
	r.finishTurnWithPayload(s, run.turnID, run.sessionID, run.agentID, run.model, status, systemMessage, failureKind, completion)
}

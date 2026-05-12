package turn

import (
	"context"
	"log"
)

func (r *sessionRunner) emitTurnStateHook(ctx context.Context, sessionID, turnID, agentID, model, status, phase string, payload map[string]any) {
	r.emitTurnState(ctx, sessionID, turnID, agentID, model, status, phase, payload, true)
}

func (r *sessionRunner) emitTurnStateHookOnly(ctx context.Context, sessionID, turnID, agentID, model, status, phase string, payload map[string]any) {
	r.emitTurnState(ctx, sessionID, turnID, agentID, model, status, phase, payload, false)
}

func (r *sessionRunner) emitTurnState(ctx context.Context, sessionID, turnID, agentID, model, status, phase string, payload map[string]any, publishTopic bool) {
	if stringsMapEmpty(payload) {
		payload = map[string]any{}
	}
	if publishTopic {
		r.engine.PublishRuntimeTurnEvent("turn_state", sessionID, turnID, agentID, status, phase, payload)
	}
	if _, err := r.engine.emitHook(ctx, HookRequest{
		Name:       HookTurnState,
		SessionID:  sessionID,
		TurnID:     turnID,
		AgentID:    agentID,
		Model:      model,
		TurnStatus: status,
		TurnPhase:  phase,
		Payload:    payload,
	}); err != nil {
		log.Printf("hook turn_state error: %v", err)
	}
}

func (r *sessionRunner) emitSessionStateHook(ctx context.Context, sessionID, agentID, model, status string, payload map[string]any) {
	r.emitSessionState(ctx, sessionID, agentID, model, status, payload, true)
}

func (r *sessionRunner) emitSessionStateHookOnly(ctx context.Context, sessionID, agentID, model, status string, payload map[string]any) {
	r.emitSessionState(ctx, sessionID, agentID, model, status, payload, false)
}

func (r *sessionRunner) emitSessionState(ctx context.Context, sessionID, agentID, model, status string, payload map[string]any, publishTopic bool) {
	if stringsMapEmpty(payload) {
		payload = map[string]any{}
	}
	if publishTopic {
		r.engine.PublishRuntimeSessionEvent("session_state", sessionID, agentID, status, payload)
	}
	if _, err := r.engine.emitHook(ctx, HookRequest{
		Name:          HookSessionState,
		SessionID:     sessionID,
		AgentID:       agentID,
		Model:         model,
		SessionStatus: status,
		Payload:       payload,
	}); err != nil {
		log.Printf("hook session_state error: %v", err)
	}
}

func stringsMapEmpty(v map[string]any) bool {
	return len(v) == 0
}

package turn

import (
	"context"
	"log"
)

func (r *sessionRunner) emitTurnStateHook(ctx context.Context, sessionID, turnID, agentID, model, status, phase string, payload map[string]any) {
	if stringsMapEmpty(payload) {
		payload = map[string]any{}
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
	if stringsMapEmpty(payload) {
		payload = map[string]any{}
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

package turn

import (
	"context"
	"fmt"
	"strings"
)

const (
	DirectKindPrompt      = "prompt"
	DirectKindPeerMessage = "peer-message"
	DirectKindContinue    = "continue"

	DirectSourceKindDirect   = "direct"
	DirectSourceKindIPC      = "ipc"
	DirectSourceKindSystem   = "system"
	DirectSourceKindInternal = "internal"
)

func normalizeDirectKind(kind string) string {
	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "", DirectKindPrompt:
		return DirectKindPrompt
	case DirectKindPeerMessage:
		return DirectKindPeerMessage
	case DirectKindContinue:
		return DirectKindContinue
	default:
		return strings.TrimSpace(strings.ToLower(kind))
	}
}

func normalizeDirectSourceKind(kind string) string {
	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "", DirectSourceKindDirect:
		return DirectSourceKindDirect
	case DirectSourceKindIPC:
		return DirectSourceKindIPC
	case DirectSourceKindSystem:
		return DirectSourceKindSystem
	case DirectSourceKindInternal:
		return DirectSourceKindInternal
	default:
		return strings.TrimSpace(strings.ToLower(kind))
	}
}

func (e *Engine) ProcessSystemDirect(ctx context.Context, in DirectInput) (*SubmitResult, error) {
	in.Origin.SourceKind = DirectSourceKindSystem
	if strings.TrimSpace(in.Origin.Role) == "" {
		in.Origin.Role = "system"
	}
	return e.ProcessDirect(ctx, in)
}

func (e *Engine) ProcessInternalDirect(ctx context.Context, in DirectInput) (*SubmitResult, error) {
	in.Origin.SourceKind = DirectSourceKindInternal
	if strings.TrimSpace(in.Origin.Role) == "" {
		in.Origin.Role = "system"
	}
	return e.ProcessDirect(ctx, in)
}

func (e *Engine) resolveDirectSessionID(ctx context.Context, in DirectInput) (string, error) {
	opCtx := ctx
	if opCtx == nil || opCtx.Err() != nil {
		opCtx = e.backgroundContext()
	}
	sessionID := strings.TrimSpace(in.SessionID)
	sessionKey := strings.TrimSpace(in.SessionKey)
	if sessionKey != "" {
		if e.store == nil {
			return "", fmt.Errorf("direct processing requires store-backed session resolution")
		}
		sess, err := e.store.ResolveSessionByKeyOrAlias(opCtx, sessionKey)
		if err != nil {
			return "", err
		}
		if sessionID != "" && sessionID != sess.ID {
			return "", fmt.Errorf("direct processing session id %q does not match session key %q", sessionID, sessionKey)
		}
		return sess.ID, nil
	}
	if sessionID != "" {
		return sessionID, nil
	}
	return "", fmt.Errorf("direct processing requires session id or session key")
}

func (e *Engine) ProcessDirect(ctx context.Context, in DirectInput) (*SubmitResult, error) {
	opCtx := ctx
	if opCtx == nil || opCtx.Err() != nil {
		opCtx = e.backgroundContext()
	}
	kind := normalizeDirectKind(in.Kind)
	metadata := map[string]any{}
	for k, v := range in.Metadata {
		metadata[k] = v
	}
	metadata["ingress_kind"] = "direct"
	metadata["ingress_source_kind"] = normalizeDirectSourceKind(in.Origin.SourceKind)
	if value := strings.TrimSpace(in.Origin.SourceID); value != "" {
		metadata["ingress_source_id"] = value
	}
	if value := strings.TrimSpace(in.Origin.Role); value != "" {
		metadata["ingress_role"] = value
	}
	if value := strings.TrimSpace(in.Origin.Label); value != "" {
		metadata["ingress_label"] = value
	}
	if value := strings.TrimSpace(in.SessionKey); value != "" {
		metadata["ingress_session_key"] = value
	}
	sessionID, err := e.resolveDirectSessionID(opCtx, in)
	if err != nil {
		return nil, err
	}
	switch kind {
	case DirectKindPrompt:
		return e.SubmitPromptRouted(opCtx, RunInput{SessionID: sessionID, Prompt: in.Prompt, Intent: in.Intent, Model: in.Model, ParentTurnID: in.ParentTurnID, Metadata: metadata})
	case DirectKindPeerMessage:
		if strings.TrimSpace(in.TargetAgentID) == "" {
			return nil, fmt.Errorf("direct peer-message requires target agent id")
		}
		return e.submitPeerMessageWithMetadata(opCtx, sessionID, in.TargetAgentID, in.Prompt, in.Intent, in.Model, in.ParentTurnID, metadata)
	case DirectKindContinue:
		continued, err := e.ContinueSession(opCtx, sessionID)
		if err != nil {
			return nil, err
		}
		return &SubmitResult{SessionID: sessionID, Status: map[bool]string{true: "continued", false: "idle"}[continued], Queued: false}, nil
	default:
		return nil, fmt.Errorf("direct input kind not supported: %s", in.Kind)
	}
}

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

func (e *Engine) ProcessDirect(ctx context.Context, in DirectInput) (*SubmitResult, error) {
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
	switch kind {
	case DirectKindPrompt:
		if strings.TrimSpace(in.SessionID) == "" {
			return nil, fmt.Errorf("direct prompt requires session id")
		}
		return e.SubmitPromptRouted(ctx, RunInput{SessionID: in.SessionID, Prompt: in.Prompt, Intent: in.Intent, Model: in.Model, ParentTurnID: in.ParentTurnID, Metadata: metadata})
	case DirectKindPeerMessage:
		if strings.TrimSpace(in.SessionID) == "" {
			return nil, fmt.Errorf("direct peer-message requires session id")
		}
		if strings.TrimSpace(in.TargetAgentID) == "" {
			return nil, fmt.Errorf("direct peer-message requires target agent id")
		}
		return e.submitPeerMessageWithMetadata(ctx, in.SessionID, in.TargetAgentID, in.Prompt, in.Intent, in.Model, in.ParentTurnID, metadata)
	case DirectKindContinue:
		if strings.TrimSpace(in.SessionID) == "" {
			return nil, fmt.Errorf("direct continue requires session id")
		}
		continued, err := e.ContinueSession(ctx, in.SessionID)
		if err != nil {
			return nil, err
		}
		return &SubmitResult{SessionID: in.SessionID, Status: map[bool]string{true: "continued", false: "idle"}[continued], Queued: false}, nil
	default:
		return nil, fmt.Errorf("direct input kind not supported: %s", in.Kind)
	}
}

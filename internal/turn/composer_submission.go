package turn

import (
	"context"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/routing"
	"github.com/rcarmo/gi/internal/routing/routedsession"
	"github.com/rcarmo/gi/internal/store"
)

// SubmitTUIMediaPrompt bridges existing media-only TUI sends. Public metadata
// cannot act as a receipt: the token and refs must match the current stored
// non-paired claim. Paired claims must use SubmitTUIComposer.
func (e *Engine) SubmitTUIMediaPrompt(ctx context.Context, in RunInput, token string) (*SubmitResult, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if opCtx == nil {
		return nil, context.Canceled
	}
	state, err := e.store.LoadTUIMediaDraft(opCtx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if token == "" || state.Claim == nil || state.Claim.Token != token || state.Claim.Text {
		return nil, store.ErrTUIDraftConflict
	}
	in.Metadata = cloneMap(in.Metadata)
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	in.Metadata["media"] = state.Claim.Refs
	return e.submitPrompt(opCtx, in, nil, "", token)
}

// SubmitTUIComposer accepts no caller prompt/media/metadata. Only the persisted
// claim is sent. This session-local adapter is not wired to the TUI yet.
func (e *Engine) SubmitTUIComposer(ctx context.Context, sessionID, token string, expected int64, model string) (*SubmitResult, store.TUIComposerDraft, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if opCtx == nil {
		return nil, store.TUIComposerDraft{}, context.Canceled
	}
	state, err := e.store.LoadTUIComposerDraft(opCtx, sessionID)
	if err != nil {
		return nil, state, err
	}
	if state.Text.Revision != expected || state.Text.Claim == nil || state.Text.Claim.Token != token {
		return nil, state, store.ErrTUIDraftConflict
	}
	if state.Text.Claim.Dispatched || state.Text.Claim.Rejected {
		return nil, state, store.ErrTUIDraftHeld
	}
	prompt := strings.TrimSpace(state.Text.Claim.Text)
	if model = strings.TrimSpace(model); model == "" {
		return nil, state, fmt.Errorf("composer: choose a model before submission")
	}
	if prompt == "" || strings.HasPrefix(prompt, "/") || strings.HasPrefix(prompt, "!!") {
		return nil, state, fmt.Errorf("composer: native/local commands are not prompt claims")
	}
	// Preview routing without ResolveOrCreate: an unsupported route must not
	// create sessions, messages or receipts in either source or target.
	identity, err := e.store.RequireSessionIdentityRuntime(opCtx, sessionID)
	if err != nil {
		return nil, state, err
	}
	inbound, err := routedsession.RequireInboundContextFromSession(opCtx, e.store, sessionID)
	if err != nil {
		return nil, state, err
	}
	inbound.SenderID = "user"
	route, body, _, err := routing.PreparePromptRoutedInput(prompt, inbound, e.routeResolver)
	if err != nil {
		return nil, state, err
	}
	if routing.NormalizeAgentID(route.AgentID) != routing.NormalizeAgentID(identity.AgentID) {
		return nil, state, fmt.Errorf("composer: cross-session routing not supported for durable draft; original claim retained")
	}
	if strings.HasPrefix(prompt, "!") {
		body = "Run this shell command and summarize the result: " + strings.TrimSpace(strings.TrimPrefix(prompt, "!"))
	}
	// The transaction prevents concurrent engines or restarted callers from
	// submitting the same token, even if validation raced a newer edit.
	state, err = e.store.BeginTUIComposerSubmission(opCtx, sessionID, token, expected)
	if err != nil {
		return nil, state, err
	}
	metadata := routing.ApplyPromptRouteMetadata(nil, sessionID, sessionID, identity.AgentID, route, false)
	mediaToken := ""
	if state.Text.Claim.Media {
		metadata["media"] = state.Media.Claim.Refs
		mediaToken = token
	}
	result, submitErr := e.submitPrompt(opCtx, RunInput{SessionID: sessionID, Prompt: body, Intent: "prompt", Model: model, Metadata: metadata}, nil, token, mediaToken)
	// Only this dispatch owner may prove rejection. A settlement failure must
	// retain the durable claim; returning an error never invites implicit resend.
	settled, settleErr := e.store.FinishTUIComposerDraft(opCtx, sessionID, token, submitErr != nil)
	if settleErr != nil {
		return result, state, fmt.Errorf("composer admission requires recovery: %w", settleErr)
	}
	return result, settled, submitErr
}

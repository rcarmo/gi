package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

// Durable refs are session-local. A live admission claim is owned by its
// frontend until native SubmitPrompt returns; restarted claims stay held.
const maxPendingMedia = store.MaxTUIPendingMedia

type mediaClaim struct {
	refs          []store.MediaRef
	token         string
	uncertain     bool
	restoreAbsent bool // only this frontend observed native rejection return
	recovered     bool
}

func (c *chatTUI) applyMediaDraft(id string, state store.TUIMediaDraft) {
	if c.pendingMedia == nil {
		c.pendingMedia = map[string][]store.MediaRef{}
	}
	if c.mediaClaims == nil {
		c.mediaClaims = map[string]*mediaClaim{}
	}
	c.pendingMedia[id] = state.Pending
	if state.Claim == nil {
		delete(c.mediaClaims, id)
		return
	}
	if claim := c.mediaClaims[id]; claim != nil && claim.token == state.Claim.Token {
		return
	}
	c.mediaClaims[id] = &mediaClaim{refs: state.Claim.Refs, token: state.Claim.Token, uncertain: true, recovered: true}
}
func (c *chatTUI) refreshPendingMedia() error {
	if c.store == nil || c.sessionID == "" {
		return nil
	}
	state, err := c.store.LoadTUIMediaDraft(context.Background(), c.sessionID)
	if err == nil {
		c.applyMediaDraft(c.sessionID, state)
	}
	return err
}
func (c *chatTUI) mediaSlotsUsed() int {
	n := len(c.pendingMedia[c.sessionID])
	if claim := c.mediaClaims[c.sessionID]; claim != nil {
		n += len(claim.refs)
	}
	return n
}
func (c *chatTUI) stageMedia(ref store.MediaRef) error {
	state, err := c.store.StageTUIMedia(context.Background(), c.sessionID, ref)
	if err == nil {
		c.applyMediaDraft(c.sessionID, state)
	}
	return err
}
func (c *chatTUI) pendingMediaLines(fields []string) []string {
	if len(fields) > 1 {
		return []string{"attachments: usage: /attachments"}
	}
	// A live failed call can retry proof-of-absence; recovered claims cannot.
	if claim := c.mediaClaims[c.sessionID]; claim != nil && claim.uncertain && !claim.recovered {
		c.settleMediaClaim(c.selectionScope(), claim, true)
	}
	if err := c.refreshPendingMedia(); err != nil {
		return []string{"attachments: state unavailable; references held: " + err.Error()}
	}
	refs := c.pendingMedia[c.sessionID]
	lines := []string{fmt.Sprintf("attachments: %d pending (session-local; survives restart)", len(refs))}
	for _, ref := range refs {
		lines = append(lines, fmt.Sprintf("  %s %s (%d bytes)", ref.ID, ref.Filename, ref.Size))
	}
	if claim := c.mediaClaims[c.sessionID]; claim != nil {
		if claim.recovered {
			lines = append(lines, fmt.Sprintf("attachments: %d unresolved admission references; held without resend. /attachments rechecks; /detach unresolved discards references only", len(claim.refs)))
		} else {
			lines = append(lines, fmt.Sprintf("attachments: %d awaiting admission; cannot detach yet", len(claim.refs)))
		}
	}
	return lines
}
func (c *chatTUI) detachMediaLines(fields []string) []string {
	if len(fields) != 2 {
		return []string{"detach: usage: /detach <media:id|all|unresolved> (references only)"}
	}
	if err := c.refreshPendingMedia(); err != nil {
		return []string{"detach: state unavailable; references held: " + err.Error()}
	}
	if fields[1] == "unresolved" {
		claim := c.mediaClaims[c.sessionID]
		if claim != nil && !claim.recovered {
			return []string{"detach: live admission pending; cannot detach yet"}
		}
	}
	expected := ""
	if claim := c.mediaClaims[c.sessionID]; claim != nil {
		expected = claim.token
	}
	state, n, err := c.store.DetachTUIMedia(context.Background(), c.sessionID, fields[1], expected)
	if err != nil {
		return []string{"detach: references unchanged: " + err.Error()}
	}
	c.applyMediaDraft(c.sessionID, state)
	if fields[1] == "all" {
		return []string{fmt.Sprintf("detach: removed %d pending references; stored files kept", n)}
	}
	if fields[1] == "unresolved" {
		return []string{fmt.Sprintf("detach: discarded %d unresolved references; stored files and any admitted work kept", n)}
	}
	if n > 0 {
		return []string{"detach: removed " + fields[1] + "; stored file kept"}
	}
	return []string{"detach: no pending reference " + fields[1]}
}
func (c *chatTUI) submitMediaInput(scope sessionScope, input turn.RunInput, claim *mediaClaim) (*turn.SubmitResult, error) {
	if claim == nil {
		return c.engine.SubmitPromptRouted(context.Background(), input)
	}
	result, err := c.engine.SubmitTUIMediaPrompt(context.Background(), input, claim.token)
	c.settleMediaClaim(scope, claim, err != nil)
	return result, err
}
func ordinaryMediaPrompt(text string) bool {
	text = strings.TrimSpace(text)
	return text != "" && !strings.HasPrefix(text, "/") && !strings.HasPrefix(text, "!!")
}
func (c *chatTUI) claimPendingMedia(allow bool) (*mediaClaim, error) {
	token := store.NowID("tui-media")
	state, err := c.store.ClaimTUIMedia(context.Background(), c.sessionID, token, allow)
	if err != nil {
		return nil, err
	}
	c.applyMediaDraft(c.sessionID, state)
	if state.Claim == nil {
		return nil, nil
	}
	claim := &mediaClaim{refs: state.Claim.Refs, token: token}
	c.mediaClaims[c.sessionID] = claim
	return claim, nil
}

// Settle against the exact durable token, merging newer staged refs inside the
// store transaction. Failure leaves the journal intact for explicit recovery.
func (c *chatTUI) settleMediaClaim(scope sessionScope, claim *mediaClaim, rejected bool) {
	if claim == nil {
		return
	}
	apply := func() {
		if c.mediaClaims[scope.id] != claim {
			return
		}
		if rejected && !claim.recovered {
			claim.restoreAbsent = true
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		state, restored, err := c.store.SettleTUIMedia(ctx, scope.id, claim.token, rejected, claim.restoreAbsent)
		cancel()
		if err != nil {
			claim.uncertain = true
			if c.ownsScope(scope) {
				c.appendTranscript("attachments: admission unknown; held to prevent duplicate send. /attachments retries the check")
			}
			return
		}
		c.applyMediaDraft(scope.id, state)
		if restored && c.ownsScope(scope) {
			c.appendTranscript("attachments: admission rejected; references retained. /attachments to review before retry")
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	}
	if c.app != nil {
		c.app.QueueUpdate(apply)
	} else {
		apply()
	}
}

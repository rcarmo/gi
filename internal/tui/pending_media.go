package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

// Pending refs live for this TUI process, indexed by source session rather than
// selection generation. An admission claim owns its refs until native acceptance
// or rejection; changing the selected session cannot lose or reuse that claim.
const maxPendingMedia = 6

type mediaClaim struct {
	refs      []store.MediaRef
	token     string
	uncertain bool
}

func (c *chatTUI) mediaSlotsUsed() int {
	n := len(c.pendingMedia[c.sessionID])
	if claim := c.mediaClaims[c.sessionID]; claim != nil {
		n += len(claim.refs)
	}
	return n
}

func (c *chatTUI) stageMedia(ref store.MediaRef) {
	if c.pendingMedia == nil {
		c.pendingMedia = map[string][]store.MediaRef{}
	}
	c.pendingMedia[c.sessionID] = append(c.pendingMedia[c.sessionID], ref)
}

func (c *chatTUI) pendingMediaLines(fields []string) []string {
	if len(fields) > 1 {
		return []string{"attachments: usage: /attachments"}
	}
	if claim := c.mediaClaims[c.sessionID]; claim != nil && claim.uncertain {
		c.settleMediaClaim(c.selectionScope(), claim, true)
	}
	refs := c.pendingMedia[c.sessionID]
	lines := []string{fmt.Sprintf("attachments: %d pending (session-local; this process)", len(refs))}
	for _, ref := range refs {
		lines = append(lines, fmt.Sprintf("  %s %s (%d bytes)", ref.ID, ref.Filename, ref.Size))
	}
	if claim := c.mediaClaims[c.sessionID]; claim != nil {
		lines = append(lines, fmt.Sprintf("attachments: %d awaiting admission; cannot detach yet", len(claim.refs)))
	}
	return lines
}

func (c *chatTUI) detachMediaLines(fields []string) []string {
	if len(fields) != 2 {
		return []string{"detach: usage: /detach <media:id|all> (pending refs only)"}
	}
	refs := c.pendingMedia[c.sessionID]
	if fields[1] == "all" {
		delete(c.pendingMedia, c.sessionID)
		return []string{fmt.Sprintf("detach: removed %d pending references; stored files kept", len(refs))}
	}
	for i, ref := range refs {
		if ref.ID == fields[1] {
			c.pendingMedia[c.sessionID] = append(refs[:i:i], refs[i+1:]...)
			return []string{"detach: removed " + ref.ID + "; stored file kept"}
		}
	}
	return []string{"detach: no pending reference " + fields[1]}
}

func (c *chatTUI) submitMediaInput(scope sessionScope, input turn.RunInput, claim *mediaClaim) (*turn.SubmitResult, error) {
	if claim == nil {
		return c.engine.SubmitPromptRouted(context.Background(), input)
	}
	// Attachments are explicitly bound to the selected session. Directed text
	// is rejected before this point; implicit agent routing must not move files
	// or create a peer session before rejecting ownership.
	result, err := c.engine.SubmitPrompt(context.Background(), input)
	c.settleMediaClaim(scope, claim, err != nil)
	return result, err
}

func ordinaryMediaPrompt(text string) bool {
	text = strings.TrimSpace(text)
	return text != "" && !strings.HasPrefix(text, "/") && !strings.HasPrefix(text, "!!")
}

func (c *chatTUI) claimPendingMedia() *mediaClaim {
	refs := c.pendingMedia[c.sessionID]
	if len(refs) == 0 {
		return nil
	}
	if c.mediaClaims == nil {
		c.mediaClaims = map[string]*mediaClaim{}
	}
	claim := &mediaClaim{refs: append([]store.MediaRef(nil), refs...), token: store.NowID("tui-media")}
	c.mediaClaims[c.sessionID] = claim
	delete(c.pendingMedia, c.sessionID)
	return claim
}

// Queue all claim settlement on the UI loop, even for a background session.
// Do not use applySessionCompletion: dropping a stale selection's error would
// silently lose its attachments. Only this exact claim can settle its refs.
func (c *chatTUI) settleMediaClaim(scope sessionScope, claim *mediaClaim, rejected bool) {
	if claim == nil {
		return
	}
	apply := func() {
		if c.mediaClaims[scope.id] != claim {
			return
		}
		if rejected {
			// Submit may return an error after storing a turn/steering row. Never
			// restore refs unless durable absence is confirmed; DB errors hold the claim.
			var admitted bool
			// Native SubmitPrompt returns only after its SQLite write/rollback;
			// no HTTP/eventual-consistency inference is involved in this read.
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := c.store.DB().QueryRowContext(ctx, `SELECT EXISTS (
    SELECT 1 FROM turns WHERE session_id=? AND json_extract(metadata_json,'$.tui_media_claim')=?
    UNION ALL SELECT 1 FROM steering_queue WHERE session_id=? AND json_extract(payload_json,'$.tui_media_claim')=?
    UNION ALL SELECT 1 FROM messages WHERE session_id=? AND json_extract(payload_json,'$.tui_media_claim')=?
   )`, scope.id, claim.token, scope.id, claim.token, scope.id, claim.token).Scan(&admitted)
			cancel()
			if err != nil {
				claim.uncertain = true
				if c.ownsScope(scope) {
					c.appendTranscript("attachments: admission unknown; held to prevent duplicate send. /attachments retries the check")
				}
				return
			}
			rejected = !admitted
			if admitted && c.ownsScope(scope) {
				c.appendTranscript("attachments: stored admission found; files will not be reattached on retry")
			}
		}
		delete(c.mediaClaims, scope.id)
		if rejected {
			if c.pendingMedia == nil {
				c.pendingMedia = map[string][]store.MediaRef{}
			}
			c.pendingMedia[scope.id] = append(claim.refs, c.pendingMedia[scope.id]...)
			if c.ownsScope(scope) {
				c.appendTranscript("attachments: admission rejected; references retained. /attachments to review before retry")
			}
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

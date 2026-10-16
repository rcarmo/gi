package tui

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
)

func durableTestChat(t *testing.T) *chatTUI {
	t.Helper()
	c := sessionTestChat(t)
	c.durableDrafts = true
	c.loadDurableDraft()
	return c
}
func draftClaimFixture(t *testing.T, c *chatTUI) (*terminalDraftState, store.TUIComposerDraft) {
	t.Helper()
	c.input.SetText("original 中文🙂")
	d := c.textDrafts[c.sessionID]
	claim, err := c.store.ClaimTUIComposerDraft(context.Background(), c.sessionID, d.pair.Text.Revision)
	if err != nil {
		t.Fatal(err)
	}
	d.pair = claim
	d.local = claim.Text.TUITextSnapshot
	d.frozen = true
	c.applyDraftSnapshot(d.local)
	return d, claim
}
func settleFixture(t *testing.T, c *chatTUI, claim store.TUIComposerDraft) store.TUIComposerDraft {
	t.Helper()
	ctx := context.Background()
	token := claim.Text.Claim.Token
	if _, err := c.store.BeginTUIComposerSubmission(ctx, "A", token, claim.Text.Revision); err != nil {
		t.Fatal(err)
	}
	if _, err := c.store.CreateTurnWithStatus(ctx, "accepted", "A", "queued", "original", map[string]any{"tui_text_claim": token}); err != nil {
		t.Fatal(err)
	}
	settled, err := c.store.FinishTUIComposerDraft(ctx, "A", token, false)
	if err != nil {
		t.Fatal(err)
	}
	return settled
}
func TestDurableDraftSaveSwitchConflictAndReload(t *testing.T) {
	c := durableTestChat(t)
	ctx := context.Background()
	c.input.SetText(" A中文🙂\nline ")
	c.input.cursorPos = 3
	c.saveDurableDraft()
	a, err := c.store.LoadTUITextDraft(ctx, "A")
	if err != nil || a.Text != c.input.Text() || a.Cursor != 3 {
		t.Fatal(a, err)
	}
	c.switchSession("B")
	c.input.SetText("B draft")
	c.switchSession("A")
	if c.input.Text() != a.Text || c.input.cursorPos != 3 {
		t.Fatal("session draft lost")
	}
	if _, err = c.store.SaveTUITextDraft(ctx, "A", a.Revision, store.TUITextSnapshot{Text: "other terminal", Cursor: 2}); err != nil {
		t.Fatal(err)
	}
	c.input.SetText("local unsaved")
	if !errors.Is(c.textDrafts["A"].err, store.ErrTUIDraftConflict) {
		t.Fatal("conflict not held")
	}
	c.switchSession("B")
	c.switchSession("A")
	if c.input.Text() != "local unsaved" {
		t.Fatal("conflict erased on switch")
	}
	c.submitDurableDraft("local unsaved")
	turns, _ := c.store.ListTurns(ctx, "A")
	if len(turns) != 0 {
		t.Fatal("conflict submitted")
	}
	out := c.draftCommand([]string{"/draft", "reload"})
	if !strings.Contains(strings.Join(out, ""), "loaded") || c.input.Text() != "other terminal" || c.input.cursorPos != 2 {
		t.Fatal(out, c.input.Text())
	}
}
func TestDurableDraftCompletionOwnsOriginAndKeepsNewerText(t *testing.T) {
	for _, visit := range []string{"A", "B", "ABA"} {
		t.Run(visit, func(t *testing.T) {
			c := durableTestChat(t)
			scope := c.selectionScope()
			d, claim := draftClaimFixture(t, c)
			c.input.SetText("newer A")
			c.input.cursorPos = 2
			c.saveDurableDraft()
			if visit != "A" {
				c.switchSession("B")
				c.input.SetText("B draft")
				if visit == "ABA" {
					c.switchSession("A")
				}
			}
			settled := settleFixture(t, c, claim)
			c.finishDurableDraft(scope, d, claim, settled, nil)
			stored, err := c.store.LoadTUITextDraft(context.Background(), "A")
			if err != nil || stored.Text != "newer A" || stored.Cursor != 2 || stored.Claim != nil {
				t.Fatal(stored, err)
			}
			if visit == "B" {
				if c.input.Text() != "B draft" {
					t.Fatal("other session overwritten")
				}
				c.switchSession("A")
			}
			if c.input.Text() != "newer A" || c.input.cursorPos != 2 {
				t.Fatal("newer draft overwritten")
			}
		})
	}
}
func TestDurableDraftLateExternalWriteNeverOverwritten(t *testing.T) {
	c := durableTestChat(t)
	scope := c.selectionScope()
	d, claim := draftClaimFixture(t, c)
	c.input.SetText("local after claim")
	settled := settleFixture(t, c, claim)
	if _, err := c.store.SaveTUITextDraft(context.Background(), "A", settled.Text.Revision, store.TUITextSnapshot{Text: "external", Cursor: 1}); err != nil {
		t.Fatal(err)
	}
	c.finishDurableDraft(scope, d, claim, settled, nil)
	if d.err == nil || c.input.Text() != "local after claim" {
		t.Fatal("external conflict lost local")
	}
	stored, _ := c.store.LoadTUITextDraft(context.Background(), "A")
	if stored.Text != "external" {
		t.Fatal("external overwritten")
	}
}
func TestDurableDraftQuestionsAndCommandsDoNotOverwriteDraft(t *testing.T) {
	c := durableTestChat(t)
	ctx := context.Background()
	c.input.SetText("ordinary draft")
	before, _ := c.store.LoadTUITextDraft(ctx, "A")
	c.setEditorAsk("secret", "Question", "private answer")
	c.input.SetText("secret changed")
	c.saveDurableDraft()
	after, _ := c.store.LoadTUITextDraft(ctx, "A")
	if !reflect.DeepEqual(before, after) {
		t.Fatal("question saved")
	}
	c.cancelEditorAsk()
	if c.input.Text() != "ordinary draft" {
		t.Fatal("question erased draft")
	}
	c.input.SetText("/where")
	c.onSubmit("/where")
	if c.input.Text() != "ordinary draft" {
		t.Fatal("command erased draft")
	}
	after, _ = c.store.LoadTUITextDraft(ctx, "A")
	if after.Text != "ordinary draft" {
		t.Fatal("command persisted")
	}
}
func TestDurableDraftUnknownRecoveryDoesNotReplayOrDiscard(t *testing.T) {
	c := durableTestChat(t)
	d, claim := draftClaimFixture(t, c)
	d.frozen = false
	c.loadDurableDraft()
	out := strings.Join(c.draftCommand([]string{"/draft", "check"}), " ")
	if !strings.Contains(out, "unresolved") {
		t.Fatal(out)
	}
	out = strings.Join(c.draftCommand([]string{"/draft", "discard", claim.Text.Claim.Token}), " ")
	if !strings.Contains(out, "refused") {
		t.Fatal(out)
	}
	c.input.SetText("newer")
	c.submitDurableDraft("newer")
	turns, _ := c.store.ListTurns(context.Background(), "A")
	if len(turns) != 0 {
		t.Fatal("unknown replayed")
	}
	stored, _ := c.store.LoadTUITextDraft(context.Background(), "A")
	if stored.Claim == nil || stored.Text != "newer" {
		t.Fatal(stored)
	}
}
func TestDurableDraftClaimWriteFailureKeepsEditor(t *testing.T) {
	c := durableTestChat(t)
	c.cfg.DefaultModel = "bootstrap"
	c.input.SetText("retain bytes")
	c.input.cursorPos = 3
	c.saveDurableDraft()
	before := c.editorSnapshot()
	if _, err := c.store.DB().Exec(`create trigger deny_draft before update on kv_store when OLD.namespace='tui_text_draft_v1' begin select raise(abort,'write failed'); end`); err != nil {
		t.Fatal(err)
	}
	c.submitDurableDraft(strings.TrimSpace(before.Text))
	if c.editorSnapshot() != before || c.textDrafts["A"].frozen {
		t.Fatal("failed claim cleared editor")
	}
}

func TestDurableDraftCheckNeverAdoptsExternalRevision(t *testing.T) {
	c := durableTestChat(t)
	d, claim := draftClaimFixture(t, c)
	d.frozen = false
	ctx := context.Background()
	external, err := c.store.SaveTUITextDraft(ctx, "A", claim.Text.Revision, store.TUITextSnapshot{Text: "external", Cursor: 3})
	if err != nil {
		t.Fatal(err)
	}
	out := c.draftCommand([]string{"/draft", "check"})
	if !strings.Contains(strings.Join(out, " "), "failed") {
		t.Fatal(out)
	}
	if d.pair.Text.Revision == external.Revision {
		t.Fatal("adopted foreign edit revision")
	}
	c.input.SetText("local")
	stored, err := c.store.LoadTUITextDraft(ctx, "A")
	if err != nil || stored.Text != "external" || d.err == nil {
		t.Fatal("external overwritten", stored, err)
	}
}
func TestDurableDraftSynchronousAdmissionClearsOnlyAcceptedText(t *testing.T) {
	c := durableTestChat(t)
	ctx := context.Background()
	c.cfg.DefaultModel = "bootstrap"
	// sessionTestChat uses agent gi; an explicit self-route keeps the adapter local.
	c.input.SetText("@gi sent once 中文🙂")
	c.submitDurableDraft(c.input.Text())
	pair, err := c.store.LoadTUIComposerDraft(ctx, "A")
	if err != nil || pair.Text.Claim != nil || pair.Text.Text != "" || c.input.Text() != "" {
		t.Fatal("accepted restored as unsent", pair, c.input.Text(), err)
	}
	c.input.SetText("successor")
	pair, err = c.store.LoadTUIComposerDraft(ctx, "A")
	if err != nil || pair.Text.Text != "successor" {
		t.Fatal(pair, err)
	}
	turns, err := c.store.ListTurns(ctx, "A")
	if err != nil || len(turns) != 1 {
		t.Fatal(turns, err)
	}
}
func TestDurableDraftConfirmedCheckPreservesEditableSuccessor(t *testing.T) {
	c := durableTestChat(t)
	d, claim := draftClaimFixture(t, c)
	d.frozen = false
	c.input.SetText("successor")
	ctx := context.Background()
	if _, err := c.store.CreateTurnWithStatus(ctx, "receipt", "A", "queued", "original", map[string]any{"tui_text_claim": claim.Text.Claim.Token}); err != nil {
		t.Fatal(err)
	}
	if err := c.store.AppendTurnEvent(ctx, "receipt", "A", "turn.submitted", nil); err != nil {
		t.Fatal(err)
	}
	out := c.draftCommand([]string{"/draft", "check"})
	if !strings.Contains(strings.Join(out, " "), "confirmed") {
		t.Fatal(out)
	}
	if c.input.Text() != "successor" || d.local.Text != "successor" || d.pair.Text.Claim != nil {
		t.Fatal("confirmed check lost successor")
	}
	c.input.SetText("successor edited")
	pair, _ := c.store.LoadTUIComposerDraft(ctx, "A")
	if pair.Text.Text != "successor edited" || pair.Text.Claim != nil {
		t.Fatal(pair)
	}
}

func TestDurableModelCommandPreservesDraftAcrossSessionSwitch(t *testing.T) {
	c := durableTestChat(t)
	origin := c.sessionID
	c.input.SetText("ordinary draft")
	c.input.SetText("/model test-model")
	c.onSubmit(c.input.Text())
	if c.input.Text() != "ordinary draft" {
		t.Fatalf("after command: %q", c.input.Text())
	}
	other, err := c.store.CreateSession(context.Background(), "draft-other", "other", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !c.switchSession(other.ID) || !c.switchSession(origin) {
		t.Fatal("switch failed")
	}
	if c.input.Text() != "ordinary draft" {
		t.Fatalf("restored: %q", c.input.Text())
	}
}

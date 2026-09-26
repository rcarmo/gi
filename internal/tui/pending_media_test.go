package tui

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func pendingMediaChat(t *testing.T) *chatTUI {
	t.Helper()
	root := t.TempDir()
	s, err := store.Open(filepath.Join(root, "gi.db"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.RuntimeConfig{WorkspaceRoot: root, DefaultModel: "bootstrap", AssistantName: "Test", UserName: "User"}
	e := turn.NewWithRuntimeConfig(s, cfg, "")
	c := &chatTUI{store: s, engine: e, cfg: cfg, stickToBottom: true, draftLineIndex: -1}
	c.ensureInput()
	if _, _, err := s.ResolveOrCreateMainSessionFromAllocation(context.Background(), store.ResolveOrCreateSessionFromAllocationInput{ID: "A", Title: "@A", State: map[string]any{"model": "bootstrap", "status": "idle"}, Allocation: gisession.AllocateDefaultSession("agent", "agent", "gi", "default")}); err != nil {
		t.Fatal(err)
	}
	c.bindSession("A")
	t.Cleanup(func() { c.stopSessionSubscription(); e.Close(); s.Close() })
	return c
}
func pendingRef(t *testing.T, c *chatTUI, name string) store.MediaRef {
	t.Helper()
	m, err := c.store.CreateMedia(context.Background(), c.sessionID, name, "text/plain", []byte("native αβ bytes"), map[string]any{"source": "tui"})
	if err != nil {
		t.Fatal(err)
	}
	return store.MediaRef{ID: store.MediaRefID(m.ID), MediaID: m.ID, SessionID: m.SessionID, Filename: m.Filename, Size: m.OriginalSize}
}
func claimedInput(c *chatTUI, claim *mediaClaim) turn.RunInput {
	return turn.RunInput{SessionID: c.sessionID, Prompt: "echo pending-media-native", Model: "bootstrap", Intent: "prompt", Metadata: map[string]any{"media": claim.refs, "tui_media_claim": claim.token}}
}
func TestPendingMediaAdmissionAndRejectedRecovery(t *testing.T) {
	c := pendingMediaChat(t)
	ref := pendingRef(t, c, "a.txt")
	stageMediaForTest(t, c, ref)
	claim := claimMediaForTest(t, c)
	input := claimedInput(c, claim)
	result, err := c.submitMediaInput(c.selectionScope(), input, claim)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.pendingMedia[c.sessionID]) != 0 || c.mediaClaims[c.sessionID] != nil {
		t.Fatal("accepted refs retained")
	}
	rec, err := c.store.GetTurn(context.Background(), result.TurnID)
	if err != nil {
		t.Fatal(err)
	}
	refs, err := store.NormalizeMediaReferences(rec.Metadata["media"])
	if err != nil || len(refs) != 1 || refs[0].MediaID != ref.MediaID || rec.Metadata["tui_media_claim"] != claim.token {
		t.Fatalf("native metadata: %#v %v", rec.Metadata, err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		r, _ := c.store.GetTurn(context.Background(), result.TurnID)
		if r.FinishedAt != "" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	// Missing identity rejects before any durable admission, not a provider mock.
	scope := sessionScope{id: "missing", generation: 90}
	bad := &mediaClaim{refs: []store.MediaRef{ref}, token: store.NowID("rejected")}
	c.mediaClaims[scope.id] = bad
	installClaimJournalForTest(t, c, scope.id, bad)
	_, err = c.submitMediaInput(scope, turn.RunInput{SessionID: scope.id, Prompt: "rejected", Model: "bootstrap"}, bad)
	if err == nil {
		t.Fatal("expected native rejection")
	}
	if !reflect.DeepEqual(c.pendingMedia[scope.id], bad.refs) {
		t.Fatal("background rejection lost refs")
	}
	var count int
	if err = c.store.DB().QueryRow(`SELECT COUNT(*) FROM turns WHERE session_id='missing'`).Scan(&count); err != nil || count != 0 {
		t.Fatal(count, err)
	}
}
func TestPendingMediaClaimsSurviveSwitchNewerStageAndUncertainDB(t *testing.T) {
	c := pendingMediaChat(t)
	id := c.sessionID
	scope := c.selectionScope()
	ref := pendingRef(t, c, "old")
	stageMediaForTest(t, c, ref)
	claim := claimMediaForTest(t, c)
	newer := pendingRef(t, c, "new")
	stageMediaForTest(t, c, newer)
	// An old callback can't settle a replacement claim, or affect editor/session.
	c.mediaClaims[id] = &mediaClaim{token: "replacement"}
	c.settleMediaClaim(scope, claim, true)
	if len(c.pendingMedia[id]) != 1 {
		t.Fatal("stale claim restored")
	}
	c.mediaClaims[id] = claim
	c.sessionID = "B"
	c.sessionGeneration++
	c.input.SetText("newer 中文\ndraft")
	c.input.cursorPos = 3
	c.settleMediaClaim(scope, claim, true)
	if !reflect.DeepEqual(c.pendingMedia[id], []store.MediaRef{ref, newer}) || c.input.cursorPos != 3 || c.input.Text() != "newer 中文\ndraft" {
		t.Fatal("origin/newer editor lost")
	}
	c.sessionID = id
	c.sessionGeneration++
	claim = claimMediaForTest(t, c)
	scope = c.selectionScope()
	// Durable post-admission errors must consume, not restore: turn and steering.
	for _, table := range []string{"turns", "steering_queue", "messages"} {
		claim = &mediaClaim{refs: []store.MediaRef{ref}, token: store.NowID("admitted")}
		c.mediaClaims[id] = claim
		installClaimJournalForTest(t, c, id, claim)
		switch table {
		case "turns":
			_, err := c.store.CreateTurn(context.Background(), store.NowID("t"), id, "accepted", map[string]any{"tui_media_claim": claim.token})
			if err != nil {
				t.Fatal(err)
			}
		case "steering_queue":
			_, err := c.store.DB().Exec(`INSERT INTO steering_queue(session_id,content,payload_json,created_at,updated_at) VALUES(?, '',json_object('tui_media_claim',?),datetime('now'),datetime('now'))`, id, claim.token)
			if err != nil {
				t.Fatal(err)
			}
		case "messages":
			err := c.store.AddMessage(context.Background(), store.NowID("m"), id, "user", "accepted", map[string]any{"tui_media_claim": claim.token})
			if err != nil {
				t.Fatal(err)
			}
		}
		c.settleMediaClaim(scope, claim, true)
		if c.mediaClaims[id] != nil || len(c.pendingMedia[id]) != 0 {
			t.Fatal("admitted refs retried", table)
		}
	}
	// A transient schema-read failure is unknown; retry only after authority works.
	claim = &mediaClaim{refs: []store.MediaRef{ref}, token: store.NowID("unknown")}
	c.mediaClaims[id] = claim
	installClaimJournalForTest(t, c, id, claim)
	if _, err := c.store.DB().Exec(`ALTER TABLE steering_queue RENAME TO held_steering`); err != nil {
		t.Fatal(err)
	}
	c.settleMediaClaim(scope, claim, true)
	if !claim.uncertain || len(c.pendingMedia[id]) != 0 {
		t.Fatal("uncertain refs restored")
	}
	if _, err := c.store.DB().Exec(`ALTER TABLE held_steering RENAME TO steering_queue`); err != nil {
		t.Fatal(err)
	}
	c.pendingMediaLines([]string{"/attachments"})
	if c.mediaClaims[id] != nil || len(c.pendingMedia[id]) != 1 {
		t.Fatal("reconciliation retry failed")
	}
}
func TestPendingMediaCommandsLimitsAndNoModelDraft(t *testing.T) {
	c := pendingMediaChat(t)
	path := filepath.Join(c.cfg.WorkspaceRoot, "file.txt")
	os.WriteFile(path, []byte("bytes"), 0600)
	for i := 0; i < maxPendingMedia; i++ {
		c.attachCommand("/attach file.txt", []string{"/attach", "file.txt"})
	}
	if c.mediaSlotsUsed() != 6 {
		t.Fatal(c.pendingMediaLines([]string{"/attachments"}))
	}
	if out := c.attachCommand("/attach absent", []string{"/attach", "absent"}); !strings.Contains(strings.Join(out, " "), "limit 6") {
		t.Fatal(out)
	}
	var count int
	c.store.DB().QueryRow(`SELECT COUNT(*) FROM media`).Scan(&count)
	if count != 6 {
		t.Fatal("overflow stored bytes")
	}
	ref := c.pendingMedia[c.sessionID][0]
	c.detachMediaLines([]string{"/detach", ref.ID})
	if c.mediaSlotsUsed() != 5 {
		t.Fatal("detach")
	}
	c.cfg.DefaultModel = ""
	c.input.SetText("no model 中文")
	c.input.cursorPos = 4
	c.onSubmit(c.input.Text())
	if c.input.Text() != "no model 中文" || c.input.cursorPos != 4 || c.mediaSlotsUsed() != 5 {
		t.Fatal("no model lost draft")
	}
	c.cfg.DefaultModel = "bootstrap"
	c.input.SetText("@peer hello")
	c.onSubmit(c.input.Text())
	if c.input.Text() != "@peer hello" || c.mediaSlotsUsed() != 5 {
		t.Fatal("directed send lost refs")
	}
	c.handleCommand("/where")
	if c.mediaSlotsUsed() != 5 {
		t.Fatal("command consumed refs")
	}
	c.detachMediaLines([]string{"/detach", "all"})
	c.store.DB().QueryRow(`SELECT COUNT(*) FROM media`).Scan(&count)
	if c.mediaSlotsUsed() != 0 || count != 6 {
		t.Fatal("detach deleted files")
	}
}

func TestPendingMediaSlotReservationClipboardAndSafeFileLimit(t *testing.T) {
	c := pendingMediaChat(t)
	for i := 0; i < 5; i++ {
		stageMediaForTest(t, c, pendingRef(t, c, "file"))
	}
	claim := claimMediaForTest(t, c)
	c.clipboardImageReader = func() ([]byte, string, error) { return []byte("PNG native"), "image/png", nil }
	c.pasteImageCommand("/paste-image", []string{"/paste-image"})
	if c.mediaSlotsUsed() != 6 {
		t.Fatal("claimed slots not reserved")
	}
	if got := strings.Join(c.pasteImageCommand("/paste-image", []string{"/paste-image"}), " "); !strings.Contains(got, "limit 6") {
		t.Fatal(got)
	}
	c.detachMediaLines([]string{"/detach", "all"})
	if c.mediaSlotsUsed() != 5 || c.mediaClaims[c.sessionID] != claim {
		t.Fatal("detach touched in-flight refs")
	}
	c.settleMediaClaim(c.selectionScope(), claim, true)
	if len(c.pendingMedia[c.sessionID]) != 5 {
		t.Fatal("rejection did not restore reserved slots")
	}
	c.detachMediaLines([]string{"/detach", "all"})
	big := filepath.Join(c.cfg.WorkspaceRoot, "large")
	f, err := os.Create(big)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate((10 << 20) + 1); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if got := strings.Join(c.attachCommand("/attach large", []string{"/attach", "large"}), " "); !strings.Contains(got, "10 MiB") {
		t.Fatal(got)
	}
	if got := strings.Join(c.attachCommand("/attach .", []string{"/attach", "."}), " "); !strings.Contains(got, "regular file required") {
		t.Fatal(got)
	}
	if c.mediaSlotsUsed() != 0 {
		t.Fatal("invalid file staged")
	}
}

func TestPendingMediaNativePostInsertErrorDoesNotRequeue(t *testing.T) {
	c := pendingMediaChat(t)
	c.stageMedia(pendingRef(t, c, "accepted-before-error"))
	claim := claimMediaForTest(t, c)
	// Real SQL write error after turn INSERT, during queue-count synchronization.
	_, err := c.store.DB().Exec(`CREATE TRIGGER fail_media_sync BEFORE UPDATE ON sessions
 WHEN EXISTS(SELECT 1 FROM turns WHERE session_id=NEW.id)
 BEGIN SELECT RAISE(ABORT,'post-insert media sync failure'); END`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.submitMediaInput(c.selectionScope(), claimedInput(c, claim), claim)
	if err == nil {
		t.Fatal("expected post-admission native error")
	}
	var n int
	if err = c.store.DB().QueryRow(`SELECT COUNT(*) FROM turns WHERE session_id=? AND json_extract(metadata_json,'$.tui_media_claim')=?`, c.sessionID, claim.token).Scan(&n); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	if len(c.pendingMedia[c.sessionID]) != 0 || c.mediaClaims[c.sessionID] != nil {
		t.Fatal("accepted file would be resent")
	}
	if _, err = c.store.DB().Exec(`DROP TRIGGER fail_media_sync`); err != nil {
		t.Fatal(err)
	}
}

func stageMediaForTest(t *testing.T, c *chatTUI, ref store.MediaRef) {
	t.Helper()
	if err := c.stageMedia(ref); err != nil {
		t.Fatal(err)
	}
}
func claimMediaForTest(t *testing.T, c *chatTUI) *mediaClaim {
	t.Helper()
	claim, err := c.claimPendingMedia(true)
	if err != nil {
		t.Fatal(err)
	}
	return claim
}

// Inject crash/error boundary state, including deliberately missing session
// identity, without weakening the public staging ownership check.
func installClaimJournalForTest(t *testing.T, c *chatTUI, id string, claim *mediaClaim) {
	t.Helper()
	raw, err := json.Marshal(store.TUIMediaDraft{Claim: &store.TUIMediaClaim{Token: claim.token, Refs: claim.refs}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.store.DB().Exec(`INSERT INTO kv_store(namespace,key,value,created_at,updated_at) VALUES('tui_pending_media_v1',?,?,datetime('now'),datetime('now')) ON CONFLICT(namespace,key) DO UPDATE SET value=excluded.value`, id, raw)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPendingMediaRecoveredClaimNeverRestoresOnAbsence(t *testing.T) {
	c := pendingMediaChat(t)
	ref := pendingRef(t, c, "restart")
	stageMediaForTest(t, c, ref)
	old := claimMediaForTest(t, c)
	// A new frontend has no in-memory evidence that SubmitPrompt returned.
	reopened := &chatTUI{store: c.store, engine: c.engine, cfg: c.cfg, sessionID: c.sessionID, sessionGeneration: 8}
	reopened.ensureInput()
	out := strings.Join(reopened.pendingMediaLines([]string{"/attachments"}), " ")
	if !strings.Contains(out, "unresolved admission") || !reopened.mediaClaims[c.sessionID].recovered {
		t.Fatal(out)
	}
	reopened.input.SetText("do not resend")
	reopened.input.cursorPos = 3
	reopened.onSubmit(reopened.input.Text())
	if reopened.input.Text() != "do not resend" || reopened.input.cursorPos != 3 {
		t.Fatal("held draft cleared")
	}
	var count int
	c.store.DB().QueryRow(`SELECT COUNT(*) FROM turns`).Scan(&count)
	if count != 0 {
		t.Fatal("recovered claim sent")
	}
	// Another client stages a new reference while the old admission is held.
	newer := pendingRef(t, c, "newer")
	stageMediaForTest(t, c, newer)
	reopened.pendingMediaLines([]string{"/attachments"})
	if len(reopened.pendingMedia[c.sessionID]) != 1 {
		t.Fatal("new reference lost")
	}
	reopened.detachMediaLines([]string{"/detach", "unresolved"})
	if reopened.mediaClaims[c.sessionID] != nil || len(reopened.pendingMedia[c.sessionID]) != 1 {
		t.Fatal("discard touched pending")
	}
	// A late live callback must not restore an explicitly discarded token.
	c.settleMediaClaim(c.selectionScope(), old, true)
	if len(c.pendingMedia[c.sessionID]) != 1 || c.pendingMedia[c.sessionID][0].ID != newer.ID {
		t.Fatal("late callback restored discarded refs")
	}
}

func TestPendingMediaJournalWriteFailurePreservesInputAndStagedFiles(t *testing.T) {
	c := pendingMediaChat(t)
	ref := pendingRef(t, c, "draft")
	stageMediaForTest(t, c, ref)
	if _, err := c.store.DB().Exec(`CREATE TRIGGER fail_tui_claim BEFORE UPDATE ON kv_store BEGIN SELECT RAISE(ABORT,'journal failure'); END`); err != nil {
		t.Fatal(err)
	}
	c.input.SetText("unsent after claim failure")
	c.input.cursorPos = 5
	history := len(c.history)
	c.onSubmit(c.input.Text())
	if c.input.Text() != "unsent after claim failure" || c.input.cursorPos != 5 || len(c.history) != history {
		t.Fatal("failed claim mutated editor")
	}
	var count int
	c.store.DB().QueryRow(`SELECT COUNT(*) FROM turns`).Scan(&count)
	if count != 0 {
		t.Fatal("failed claim admitted")
	}
	c.store.DB().Exec(`DROP TRIGGER fail_tui_claim`)
	if err := c.refreshPendingMedia(); err != nil || len(c.pendingMedia[c.sessionID]) != 1 {
		t.Fatal("refs lost", err)
	}
}

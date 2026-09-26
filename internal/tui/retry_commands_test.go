package tui

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
)

func seedRetryFailure(t *testing.T, c *chatTUI, id, session string) {
	t.Helper()
	ctx := context.Background()
	if _, err := c.store.CreateTurnWithStatus(ctx, id, session, "failed", "original "+id, map[string]any{"model": "bootstrap", "custom": "kept"}); err != nil {
		t.Fatal(err)
	}
	if err := c.store.HoldTurnFailure(ctx, id, "review", "held\n\x1b[31m summary"); err != nil {
		t.Fatal(err)
	}
}
func retryOutput(c *chatTUI, text string) string {
	return strings.Join(c.retryCommand(strings.Fields(text)), "\n")
}

func TestTUIRetryCommandsBoundedReadOnlyAndSessionGuards(t *testing.T) {
	c := pendingMediaChat(t)
	s := c.store
	ctx := context.Background()
	c.input.SetText("untouched draft 中文🙂")
	for i := 0; i < 8; i++ {
		seedRetryFailure(t, c, fmt.Sprintf("held-%d", i), "A")
	}
	if _, err := s.CreateSession(ctx, "B", "B", nil); err != nil {
		t.Fatal(err)
	}
	seedRetryFailure(t, c, "foreign", "B")
	before, err := s.ListTurns(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	out := retryOutput(c, "/retry")
	if !strings.Contains(out, "page 1") || strings.Count(out, "  held-") != 6 || strings.Contains(out, "foreign") || strings.Contains(out, "\x1b") {
		t.Fatal(out)
	}
	out = retryOutput(c, "/retry 2")
	if strings.Count(out, "  held-") != 2 {
		t.Fatal(out)
	}
	for _, cmd := range []string{"/retry 0", "/retry -1", "/retry 9999999999999999999999999999999", "/retry run", "/retry release held-0", "/retry check held-0 extra"} {
		if out := retryOutput(c, cmd); !strings.Contains(out, retryUsage) {
			t.Fatal(cmd, out)
		}
	}
	if out := retryOutput(c, "/retry 3"); !strings.Contains(out, "out of range") {
		t.Fatal(out)
	}
	for _, cmd := range []string{"/retry check foreign", "/retry run foreign", "/retry release foreign token", "/retry run missing", "/retry run held-"} {
		if out := retryOutput(c, cmd); !strings.Contains(out, "not found in this session") {
			t.Fatal(cmd, out)
		}
	}
	if out := retryOutput(c, "/retry check held-0"); !strings.Contains(out, "held-0 · held") {
		t.Fatal(out)
	}
	after, err := s.ListTurns(ctx, "A")
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("read/foreign commands mutated", err)
	}
	for i := 0; i < 8; i++ {
		f, err := s.GetTurnFailure(ctx, fmt.Sprintf("held-%d", i))
		if err != nil || f.ResolutionState != "" || f.RetryAdmissionToken != "" {
			t.Fatal(f, err)
		}
	}
	if c.input.text != "untouched draft 中文🙂" {
		t.Fatal("draft changed")
	}
}

func TestTUIRetryCommandsNativeAdmissionReleaseAndLegacy(t *testing.T) {
	c := pendingMediaChat(t)
	s := c.store
	ctx := context.Background()
	seedRetryFailure(t, c, "old", "A")
	seedRetryFailure(t, c, "pending", "A")
	seedRetryFailure(t, c, "legacy", "A")
	// Active fixture keeps admitted retries queued and prevents external work.
	if _, err := s.CreateTurnWithStatus(ctx, "active", "A", "running", "other work", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "active", "live", "claim"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	c.input.SetText("retained draft")
	c.pendingMedia = map[string][]store.MediaRef{"A": {{ID: "pending"}}}
	beforeMedia := append([]store.MediaRef(nil), c.pendingMedia["A"]...)
	c.queueSnapshot = []string{"stale"}
	out := retryOutput(c, "/retry run old")
	if !strings.Contains(out, "retry: admitted") || !strings.Contains(out, "queued") {
		t.Fatal(out)
	}
	f, err := s.GetTurnFailure(ctx, "old")
	if err != nil || f.ResolutionState != "retried" {
		t.Fatal(f, err)
	}
	admitted, err := s.GetTurn(ctx, f.ResolvedTurnID)
	if err != nil || admitted.Prompt != "original old" || admitted.Metadata["custom"] != "kept" {
		t.Fatal(admitted, err)
	}
	if out := retryOutput(c, "/retry run old"); !strings.Contains(out, f.ResolvedTurnID) {
		t.Fatal("duplicate did not return same ID", out)
	}
	if out := retryOutput(c, "/retry check old"); !strings.Contains(out, "admitted: "+f.ResolvedTurnID) {
		t.Fatal(out)
	}
	if len(c.queueSnapshot) != 0 {
		t.Fatal("queue snapshot retained after mutation")
	}
	for _, legacy := range []bool{false, true} {
		id := "pending"
		var ok bool
		var err error
		if legacy {
			id = "legacy"
			ok, err = s.ReserveHeldRetry(ctx, id, "token")
		} else {
			ok, err = s.ReserveAtomicHeldRetry(ctx, id, "token")
		}
		if err != nil || !ok {
			t.Fatal(ok, err)
		}
		before, err := s.GetTurnFailure(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		out = retryOutput(c, "/retry check "+id)
		if !strings.Contains(out, "pending") {
			t.Fatal(out)
		}
		after, err := s.GetTurnFailure(ctx, id)
		if err != nil || *before != *after {
			t.Fatal("check mutated pending state")
		}
		out = retryOutput(c, "/retry run "+id)
		if !strings.Contains(out, "no resend") {
			t.Fatal(out)
		}
		if out := retryOutput(c, "/retry release "+id+" wrong"); !strings.Contains(out, "release failed") {
			t.Fatal(out)
		}
		out = retryOutput(c, "/retry release "+id+" token")
		if legacy {
			if !strings.Contains(out, "release failed") {
				t.Fatal("released legacy", out)
			}
		} else {
			if !strings.Contains(out, "no work submitted") {
				t.Fatal(out)
			}
		}
	}
	turns, err := s.ListTurns(ctx, "A")
	if err != nil || len(turns) != 5 {
		t.Fatal("unexpected turns", turns, err)
	}
	if c.input.text != "retained draft" || !reflect.DeepEqual(beforeMedia, c.pendingMedia["A"]) {
		t.Fatal("draft/media mutated")
	}
	// The live reservation token cannot submit after the explicit release.
	err = s.AdmitHeldRetry(ctx, store.HeldRetryAdmission{OriginalTurnID: "pending", Token: "token"}, "late", "A", "old", map[string]any{"retry_of_turn_id": "pending", "retry_admission_token": "token"}, nil)
	if err == nil {
		t.Fatal("released owner admitted")
	}
}

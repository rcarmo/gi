package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestSessionDisplayCapabilitiesMatchMutationRules(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "caps.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.CreateSession(ctx, "root", "Root", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CloneSession(ctx, "root", "child", "Child", "child"); err != nil {
		t.Fatal(err)
	}
	caps, err := s.SessionDisplayCapabilities(ctx, "root")
	if err != nil || !caps.CanPin || !caps.CanRename || caps.CanArchive || caps.CanRestore {
		t.Fatal(caps, err)
	}
	check := func(archive bool) {
		t.Helper()
		caps, err := s.SessionDisplayCapabilities(ctx, "child")
		if err != nil || caps.CanArchive != archive {
			t.Fatal(caps, err)
		}
	}
	check(true)
	for _, status := range []string{"queued", "running"} {
		if _, err := s.CreateTurnWithStatus(ctx, "busy", "child", status, "busy", nil); err != nil {
			t.Fatal(err)
		}
		check(false)
		if err := s.DeleteTurn(ctx, "busy"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreateTurnWithStatus(ctx, "ended", "child", "completed", "done", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "child", "ended", "owner", "claim"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	check(false)
	if err := s.ReleaseSessionActiveTurn(ctx, "child", "claim"); err != nil {
		t.Fatal(err)
	}
	check(true)
	pin := true
	if err := s.MutateSession(ctx, "child", SessionMutation{Action: "pin", Pinned: &pin}); err != nil {
		t.Fatal(err)
	}
	if err := s.MutateSession(ctx, "child", SessionMutation{Action: "archive"}); err != nil {
		t.Fatal(err)
	}
	caps, err = s.SessionDisplayCapabilities(ctx, "child")
	if err != nil || caps.CanPin || caps.CanRename || caps.CanArchive || !caps.CanRestore || !caps.Pinned {
		t.Fatal(caps, err)
	}
	if _, err = s.SessionDisplayCapabilities(ctx, "missing"); err == nil {
		t.Fatal("missing treated as empty capabilities")
	}
}

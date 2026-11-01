package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestProviderRetryRunningGuard(t *testing.T) {
	s, ctx, _ := retryReceiptFixture(t)
	if err := s.CheckProviderRetryRunning(ctx, "s", "t"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckProviderRetryRunning(ctx, "other", "t"); !errors.Is(err, context.Canceled) {
		t.Fatal("cross-session guard", err)
	}
	if err := s.CheckProviderRetryRunning(ctx, "s", "missing"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("missing turn", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := s.CheckProviderRetryRunning(cancelled, "s", "t"); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled read", err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, "t", "cancelling", "cancelling"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckProviderRetryRunning(ctx, "s", "t"); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled turn continued", err)
	}
}

// Compare equivalent session/status checks on one database. The metadata read
// remains in the downstream context builder in both preparation paths.
func BenchmarkProviderRetryRunningGuard(b *testing.B) {
	s, err := Open(filepath.Join(b.TempDir(), "guard.db"))
	if err != nil {
		b.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err = s.CreateSession(ctx, "s", "s", nil); err != nil {
		b.Fatal(err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "t", "s", "running", "prompt", map[string]any{"model": "mock-stream", "selected_thinking_model": "mock-stream", "selected_thinking_level": "high", "allowed_tools": []string{"read", "edit", "write", "shell"}}); err != nil {
		b.Fatal(err)
	}
	for _, full := range []bool{true, false} {
		name := "full_record"
		if !full {
			name = "status_only"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				if full {
					record, err := s.GetTurn(ctx, "t")
					if err != nil {
						b.Fatal(err)
					}
					if record.SessionID != "s" || record.Status != "running" {
						b.Fatal("invalid guard")
					}
				} else if err := s.CheckProviderRetryRunning(ctx, "s", "t"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

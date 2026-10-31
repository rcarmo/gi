package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	goai "github.com/rcarmo/go-ai"
)

func TestTurnWriterRevocationFencesEveryEffectBoundary(t *testing.T) {
	s, _, claim := stagedRecoveryFixture(t)
	base := context.Background()
	old := WithTurnWriter(base, "s", "t", "old")
	if _, err := s.AcquireStagedTerminalRecovery(base, claim, "new", time.Now().Add(-time.Second), false); err != nil {
		t.Fatal(err)
	}
	for _, released := range []bool{false, true} {
		if released {
			if err := s.ReleaseSessionActiveTurn(base, "s", "new"); err != nil {
				t.Fatal(err)
			}
		}
		for name, write := range map[string]func(context.Context) error{
			"intent": func(ctx context.Context) error {
				return s.AdmitTurnEffect(ctx, "s", "t", "running", "inference.started", map[string]any{"request_id": "late"})
			},
			"phase":  func(ctx context.Context) error { return s.SetClaimedRunningPhase(ctx, "s", "t", "retry_wait") },
			"status": func(ctx context.Context) error { return s.UpdateTurnStatusAndPhase(ctx, "t", "failed", "failed") },
			"finalize": func(ctx context.Context) error {
				_, err := s.FinalizeRunningTurn(ctx, "t", "failed", "failed")
				return err
			},
			"provider": func(ctx context.Context) error {
				_, err := s.SettleProviderAttempt(ctx, "t", "s", "late", sampleProviderReceipt())
				return err
			},
			"partial": func(ctx context.Context) error { return s.SaveProviderPartial(ctx, "t", "s", "late", "late", nil) },
			"tool": func(ctx context.Context) error {
				return s.SettleToolResult(ctx, "t", "s", "tool.finished", map[string]any{"occurrence_id": "late"}, "late-result", "late", nil)
			},
			"progress": func(ctx context.Context) error {
				return s.AppendTurnEffectEvent(ctx, "t", "s", "tool.output", map[string]any{"occurrence_id": "late"})
			},
			"prompt_sections": func(ctx context.Context) error {
				return s.RecordTurnPromptSections(ctx, "s", "t", map[string]any{"late": true})
			},
			"begin_compaction": func(ctx context.Context) error { _, err := s.BeginCompaction(ctx, "s", "t", nil); return err },
			"finish_compaction": func(ctx context.Context) error {
				_, err := s.FinishCompaction(ctx, "s", "t", 1, "completed", "late", nil)
				return err
			},
		} {
			t.Run(name+map[bool]string{false: "_active", true: "_released"}[released], func(t *testing.T) {
				for _, ctx := range []context.Context{old, context.WithoutCancel(old)} {
					if err := write(ctx); err == nil {
						t.Fatal("revoked writer admitted")
					}
				}
			})
		}
	}
	var count int
	if err := s.db.QueryRow(`select count(*) from turn_receipts where turn_id='t'`).Scan(&count); err != nil || count != 0 {
		t.Fatal("revoked receipt stored", count, err)
	}
	if err := s.db.QueryRow(`select count(*) from messages where id in ('late','late-result')`).Scan(&count); err != nil || count != 0 {
		t.Fatal("revoked message stored", count, err)
	}
	rec, err := s.GetTurn(base, "t")
	if err != nil || rec.Status != "running" || rec.Phase != "settling" {
		t.Fatal("revoked phase changed", rec, err)
	}
}
func TestTurnWriterScopeAndCancellationAreIndependent(t *testing.T) {
	s, _, _ := stagedRecoveryFixture(t)
	base := context.Background()
	// A staged owner cannot reopen effect admission, even before takeover.
	current := WithTurnWriter(base, "s", "t", "old")
	if err := s.AdmitTurnEffect(current, "s", "t", "running", "inference.started", nil); err == nil {
		t.Fatal("staged owner reopened inference")
	}
	if err := s.SetClaimedRunningPhase(current, "s", "t", "running"); err == nil {
		t.Fatal("staged owner reopened phase")
	}
	if _, err := s.BeginCompaction(current, "s", "t", nil); err == nil {
		t.Fatal("staged owner reopened compaction")
	}
	if _, err := s.db.Exec(`update turns set phase='running' where id='t'`); err != nil {
		t.Fatal(err)
	}
	if err := s.AdmitTurnEffect(current, "s", "t", "running", "inference.started", map[string]any{"request_id": "known"}); err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(current)
	cancel()
	if err := s.SaveProviderPartial(cancelled, "t", "s", "known", "partial", nil); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled direct write admitted", err)
	}
	// Only the still-owned host can deliberately finish its receipt after cancel.
	if err := s.SaveProviderPartial(context.WithoutCancel(cancelled), "t", "s", "known", "partial", map[string]any{"provider_message": goai.Message{Role: goai.RoleAssistant}}); err != nil {
		t.Fatal(err)
	}
	for _, ctx := range []context.Context{WithTurnWriter(base, "foreign", "t", "old"), WithTurnWriter(base, "s", "foreign", "old"), WithTurnWriter(base, "s", "t", "")} {
		if err := s.AppendTurnEffectEvent(ctx, "t", "s", "tool.output", nil); !errors.Is(err, ErrQueueConflict) {
			t.Fatal("foreign writer scope accepted", err)
		}
		if _, err := s.FinalizeRunningTurn(ctx, "t", "completed", "completed"); !errors.Is(err, ErrQueueConflict) {
			t.Fatal("foreign finalizer scope accepted", err)
		}
	}
	if _, err := s.SettleProviderAttempt(context.WithoutCancel(cancelled), "t", "s", "known", sampleProviderReceipt()); err != nil {
		t.Fatal(err)
	}
	accounting, err := s.TurnAccounting(base, "t")
	if err != nil || accounting.Usage.TotalTokens == 0 {
		t.Fatal("owned cancellation lost known accounting", accounting, err)
	}
}

func BenchmarkTurnWriterEvent(b *testing.B) {
	for _, scoped := range []bool{false, true} {
		b.Run(map[bool]string{false: "unbound", true: "scoped"}[scoped], func(b *testing.B) {
			s, err := Open(filepath.Join(b.TempDir(), "writer.db"))
			if err != nil {
				b.Fatal(err)
			}
			defer s.Close()
			ctx := context.Background()
			if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
				b.Fatal(err)
			}
			if _, err := s.CreateTurnWithStatus(ctx, "t", "s", "running", "prompt", nil); err != nil {
				b.Fatal(err)
			}
			if ok, err := s.ClaimSessionActiveTurn(ctx, "s", "t", "worker", "epoch"); err != nil || !ok {
				b.Fatal(ok, err)
			}
			if scoped {
				ctx = WithTurnWriter(ctx, "s", "t", "epoch")
			}
			payload := map[string]any{"occurrence_id": "tool", "output_preview": "bounded progress"}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if err := s.AppendTurnEffectEvent(ctx, "t", "s", "tool.output", payload); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestTurnWriterPromptPreparationOwnedRecord(t *testing.T) {
	s, ctx, _ := retryReceiptFixture(t)
	if err := s.RecordTurnPromptSections(ctx, "s", "t", map[string]any{"owned": true}); err != nil {
		t.Fatal(err)
	}
	row, err := s.GetSession(context.Background(), "s")
	if err != nil {
		t.Fatal(err)
	}
	if record, ok := row.State["system_prompt_sections"].(map[string]any); !ok || record["owned"] != true {
		t.Fatal(row.State)
	}
	if err := s.RecordTurnPromptSections(WithTurnWriter(context.Background(), "other", "t", "epoch"), "s", "t", nil); err == nil {
		t.Fatal("cross-session preparation admitted")
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, "t", "cancelling", "cancelling"); err != nil {
		t.Fatal(err)
	}
	if err := s.RecordTurnPromptSections(ctx, "s", "t", map[string]any{"late": true}); err == nil {
		t.Fatal("cancelled preparation adopted")
	}
}

package store

import (
	"context"
	"database/sql/driver"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"modernc.org/sqlite"
)

var settlementBarrierSequence atomic.Uint64

func TestToolSettlementCancellationDuringSQLiteWrite(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	name := fmt.Sprintf("settlement_barrier_%d", settlementBarrierSequence.Add(1))
	if err := sqlite.RegisterScalarFunction(name, 0, func(*sqlite.FunctionContext, []driver.Value) (driver.Value, error) {
		close(entered)
		<-release
		return int64(1), nil
	}); err != nil {
		t.Fatal(err)
	}
	s, err := Open(filepath.Join(t.TempDir(), "cancel-write.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "s", "session", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`create trigger settlement_pause before insert on messages begin select ` + name + `(); end`); err != nil {
		t.Fatal(err)
	}
	parent, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- s.SettleToolResult(parent, "t", "s", "tool.finished", nil, "m", "result", map[string]any{"kind": "tool_result"})
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("write never reached barrier")
	}
	cancel()
	close(release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("admitted write cancelled: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("settlement did not finish")
	}
	for _, table := range []string{"turn_events", "messages"} {
		var count int
		if err := s.db.QueryRow(`select count(*) from ` + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("partial admitted settlement: %s=%d", table, count)
		}
	}
}

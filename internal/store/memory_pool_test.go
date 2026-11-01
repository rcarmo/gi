package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestMemoryStoreSerializesTransactionsAndPreservesCancellation(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if s.db.Stats().MaxOpenConnections != 1 {
		t.Fatal("shared-cache memory permits conflicting transaction upgrades")
	}
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err = s.db.ExecContext(ctx, `delete from sessions`); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("waiting writer ignored cancellation", err)
	}
	if err = tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateSession(context.Background(), "after", "after", nil); err != nil {
		t.Fatal(err)
	}
	disk, err := Open(filepath.Join(t.TempDir(), "wal.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer disk.Close()
	if disk.db.Stats().MaxOpenConnections == 1 {
		t.Fatal("file-backed WAL pool was serialised")
	}
}

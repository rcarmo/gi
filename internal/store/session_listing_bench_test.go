package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
)

func BenchmarkListSessions(b *testing.B) {
	s, err := Open(filepath.Join(b.TempDir(), "bench.db"))
	if err != nil {
		b.Fatal(err)
	}
	defer s.Close()
	for i := 0; i < 100; i++ {
		if _, err := s.CreateSession(context.Background(), fmt.Sprintf("session-%03d", i), "title", map[string]any{"model": "fixture"}); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v, err := s.ListSessions(context.Background())
		if err != nil || len(v) != 100 {
			b.Fatal(len(v), err)
		}
	}
}

func TestListSessionsMatchesIndividualIdentityAndFailsClosed(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "list.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if _, err := s.CreateSession(ctx, fmt.Sprintf("session-%d", i), "title", map[string]any{"model": "fixture"}); err != nil {
			t.Fatal(err)
		}
	}
	list, err := s.ListSessions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range list {
		single, err := s.GetSession(ctx, v.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(v, *single) {
			t.Fatal("batched identity diverged", v, single)
		}
	}
	if _, err := s.db.ExecContext(ctx, `delete from session_identities where session_id='session-1'`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ListSessions(ctx); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("missing identity accepted", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := s.ListSessions(cancelled); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled listing", err)
	}
}

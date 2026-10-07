package store

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// Keep the database bounded while measuring a complete 100-message copy.
// Cleanup is included equally in baseline and tuned runs.
func BenchmarkCloneSession(b *testing.B) {
	s, err := Open(filepath.Join(b.TempDir(), "copies.db"))
	if err != nil {
		b.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "source", "Source", nil); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		if err := s.AddMessage(ctx, fmt.Sprintf("msg-%03d", i), "source", "user", strings.Repeat("fixture ", 128), map[string]any{"kind": "chat"}); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy, err := s.CloneSession(ctx, "source", "copy", "", "")
		if err != nil {
			b.Fatal(err)
		}
		if _, err := s.DB().ExecContext(ctx, "delete from sessions where id = ?", copy.ID); err != nil {
			b.Fatal(err)
		}
	}
}

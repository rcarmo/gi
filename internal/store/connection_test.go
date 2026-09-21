package store

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestSQLitePoolConnectionsHaveRuntimePragmas(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "pool.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var held []*sql.Conn
	defer func() {
		for _, c := range held {
			c.Close()
		}
	}()
	for i := 0; i < 4; i++ {
		conn, err := s.DB().Conn(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		held = append(held, conn)
		for pragma, want := range map[string]int{"busy_timeout": 5000, "foreign_keys": 1, "synchronous": 1, "temp_store": 2} {
			var got int
			if err := conn.QueryRowContext(t.Context(), "PRAGMA "+pragma).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("connection %d: %s=%d, want %d", i, pragma, got, want)
			}
		}
	}
}

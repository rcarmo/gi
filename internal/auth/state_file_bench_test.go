package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Keep the number and shape of reads fixed: the concurrent login regression's
// reader count varies with writer latency, so its totals are not a benchmark.
func BenchmarkAuthStateRead(b *testing.B) {
	for _, count := range []int{1, 32, 128} {
		b.Run(fmt.Sprintf("sessions-%d", count), func(b *testing.B) {
			state := State{Username: "fixture-admin", CreatedAt: time.Unix(1, 0).UTC(), UpdatedAt: time.Unix(2, 0).UTC(), extra: map[string]json.RawMessage{"future": json.RawMessage(`{"data":"` + strings.Repeat("x", 2048) + `"}`)}}
			for i := 0; i < count; i++ {
				state.Sessions = append(state.Sessions, Session{TokenHash: fmt.Sprintf("fixture-token-%064d", i), CreatedAt: time.Unix(1, 0).UTC(), ExpiresAt: time.Unix(100, 0).UTC(), Purpose: "web", AuthFactor: "totp"})
			}
			data, err := encodeState(state)
			if err != nil {
				b.Fatal(err)
			}
			dir := b.TempDir()
			if err = os.WriteFile(filepath.Join(dir, "auth.json"), data, 0600); err != nil {
				b.Fatal(err)
			}
			root, err := os.OpenRoot(dir)
			if err != nil {
				b.Fatal(err)
			}
			defer root.Close()
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				read, raw, err := readState(root)
				if err != nil || len(raw) != len(data) || len(read.Sessions) != count || read.Username != state.Username || read.extra["future"] == nil {
					b.Fatal("state contract lost", err)
				}
			}
		})
	}
}

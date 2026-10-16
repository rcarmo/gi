package store

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkConversationReplayPayload(b *testing.B) {
	s, err := Open(filepath.Join(b.TempDir(), "payload.db"))
	if err != nil {
		b.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "s", "session", nil); err != nil {
		b.Fatal(err)
	}
	if err := s.AddMessage(ctx, "m", "s", "assistant", "visible", map[string]any{"kind": "chat", "provider_message": map[string]any{"content": strings.Repeat("x", 1<<20)}}); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		page, err := s.PageConversationMessages(ctx, "s", "", "", 50)
		if err != nil {
			b.Fatal(err)
		}
		if len(page.Messages) != 1 || page.Messages[0].Content != "visible" || page.Messages[0].Payload["provider_message"] != nil {
			b.Fatal("invalid display projection")
		}
	}
}

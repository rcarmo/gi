package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestNormalizeMediaReferencesAcceptsStringsAndObjects(t *testing.T) {
	refs, err := NormalizeMediaReferences([]any{
		"media:12",
		map[string]any{"media_id": float64(34), "filename": "shot.png", "content_type": "image/png"},
		map[string]any{"id": "56", "source": "tui"},
	})
	if err != nil {
		t.Fatalf("normalize media refs: %v", err)
	}
	if len(refs) != 3 {
		t.Fatalf("unexpected refs: %#v", refs)
	}
	if refs[0].ID != "media:12" || refs[0].MediaID != 12 {
		t.Fatalf("unexpected string ref: %#v", refs[0])
	}
	if refs[1].ID != "media:34" || refs[1].MediaID != 34 || refs[1].Filename != "shot.png" || refs[1].ContentType != "image/png" {
		t.Fatalf("unexpected object ref: %#v", refs[1])
	}
	if refs[2].ID != "56" || refs[2].MediaID != 56 || refs[2].Source != "tui" {
		t.Fatalf("unexpected numeric-id ref: %#v", refs[2])
	}
}

func TestNormalizeMediaReferencesRejectsInvalidReferences(t *testing.T) {
	if _, err := NormalizeMediaReferences([]any{"not-media"}); err == nil {
		t.Fatal("expected invalid media ref error")
	}
	if _, err := NormalizeMediaReferences(map[string]any{"id": "media:1"}); err == nil {
		t.Fatal("expected invalid media list type error")
	}
}

func TestCreateMediaStoresHashAndDetectedContentType(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_media_contract", "Media", map[string]any{"status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	raw := []byte("hello media")
	item, err := s.CreateMedia(ctx, "session_media_contract", "hello.txt", "text/plain", raw, map[string]any{"source": "api"})
	if err != nil {
		t.Fatalf("create media: %v", err)
	}
	got, content, err := s.GetMediaContent(ctx, item.ID)
	if err != nil {
		t.Fatalf("get media content: %v", err)
	}
	if string(content) != string(raw) {
		t.Fatalf("unexpected content: %q", string(content))
	}
	sum := sha256.Sum256(raw)
	wantHash := hex.EncodeToString(sum[:])
	if got.Metadata["sha256"] != wantHash {
		t.Fatalf("unexpected hash metadata: %#v", got.Metadata)
	}
	if got.Metadata["detected_content_type"] == "" {
		t.Fatalf("expected detected content type metadata: %#v", got.Metadata)
	}
	if got.Metadata["source"] != "api" {
		t.Fatalf("expected source metadata: %#v", got.Metadata)
	}
}

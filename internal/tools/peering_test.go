package tools_test

import (
	"context"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestPeeringToolStatus(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := turn.New(s)
	out, err := e.ExecuteToolByName(context.Background(), "peering", "", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"\"backend\": \"tsnet\"", "\"state\": \"disabled\""} {
		if !strings.Contains(out, want) {
			t.Fatalf("peering output missing %q: %s", want, out)
		}
	}
}

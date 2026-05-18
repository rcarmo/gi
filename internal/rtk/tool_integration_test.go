package rtk

import (
	"strings"
	"testing"
)

func TestRTKToolFilterOnly(t *testing.T) {
	got := Filter("go test ./...", "ok a\n--- FAIL: TestX\nFAIL\n")
	if got.Mode != "go-test" || !strings.Contains(got.Output, "FAIL") || strings.Contains(got.Output, "ok a") {
		t.Fatalf("unexpected rtk output: %#v", got)
	}
}

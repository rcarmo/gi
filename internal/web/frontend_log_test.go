package web

import (
	"errors"
	"strings"
	"testing"
)

type badFrontendLogDetail struct{}

func (badFrontendLogDetail) MarshalJSON() ([]byte, error) {
	return nil, errors.New("boom")
}

func TestMarshalFrontendLogDetailFallsBackOnMarshalError(t *testing.T) {
	out := marshalFrontendLogDetail(badFrontendLogDetail{})
	if !strings.Contains(out, `"marshal_error"`) || !strings.Contains(out, `"detail_type":"web.badFrontendLogDetail"`) {
		t.Fatalf("unexpected fallback payload: %q", out)
	}
}

func TestMarshalFrontendLogDetailKeepsJSONForSerializableValues(t *testing.T) {
	out := marshalFrontendLogDetail(map[string]any{"ok": true})
	if out != `{"ok":true}` {
		t.Fatalf("unexpected serialized payload: %q", out)
	}
}

package web

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
)

func TestSessionThinkingHTTPRejectsUnsupportedUnknownAndStaleFields(t *testing.T) {
	s, st := newTestWebServer(t, t.TempDir())
	ctx := context.Background()
	a, err := st.CreateSession(ctx, "A", "A", map[string]any{"selected_model": "bootstrap", "selected_provider": "test"})
	if err != nil {
		t.Fatal(err)
	}
	token := store.SessionThinkingToken("A", a.State)
	for _, body := range []string{`{"model":"test/bootstrap","thinking_level":"high","thinking_token":"` + token + `"}`, `{"model":"test/bootstrap","thinking_level":"bogus"}`, `{"model":"test/bootstrap","thinking_token":"x"}`, `{"model":"test/bootstrap","thinking_level":3}`, `{"model":"test/bootstrap","thinking_level":null}`, `{"model":"test/bootstrap","extra":1}`, `{"model":"test/bootstrap"}{}`} {
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, httptest.NewRequest("PATCH", "/api/sessions/A/model", strings.NewReader(body)))
		if rr.Code != 400 {
			t.Fatal(rr.Code, rr.Body.String())
		}
	}
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/api/sessions/A/model", nil))
	var payload map[string]any
	json.Unmarshal(rr.Body.Bytes(), &payload)
	if payload["thinking_configurable"] != false {
		t.Fatal(payload)
	}
	after, _ := st.GetSession(ctx, "A")
	if store.SessionThinkingToken("A", after.State) != token {
		t.Fatal("rejection mutated state")
	}
	turns, _ := st.ListTurns(ctx, "A")
	if len(turns) != 0 {
		t.Fatal(turns)
	}
}

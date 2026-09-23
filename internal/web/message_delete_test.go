package web

import (
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
)

func TestMessageDeleteAuthSessionMethodsAndCascade(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "delete.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, id := range []string{"a", "b"} {
		if _, err := db.CreateSession(t.Context(), id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.AddMessage(t.Context(), "message", "a", "assistant", "keep until deleted", nil); err != nil {
		t.Fatal(err)
	}
	srv := New(db, nil, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	call := func(method, path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, httptest.NewRequest(method, path, nil))
		return w
	}
	for _, tc := range []struct {
		method, path string
		code         int
	}{{"GET", "/api/sessions/a/messages/message", 405}, {"DELETE", "/api/sessions/a/messages/message?cascade=true", 400}, {"DELETE", "/api/sessions/b/messages/message", 404}, {"DELETE", "/api/sessions/a/messages/message/extra", 404}, {"DELETE", "/api/sessions/a/messages/message", 200}, {"DELETE", "/api/sessions/a/messages/message", 404}} {
		if r := call(tc.method, tc.path); r.Code != tc.code {
			t.Fatal(tc, r.Code, r.Body.String())
		}
	}
	pending, err := srv.auth.StartEnrollment("rui")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = srv.auth.VerifyEnrollment("rui", webTestTOTPCode(pending.Secret)); err != nil {
		t.Fatal(err)
	}
	if r := call("DELETE", "/api/sessions/a/messages/message"); r.Code != 401 {
		t.Fatal(r.Code)
	}
}

package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkWriteJSONLargeSessionState(b *testing.B) {
	payload := map[string]any{"sessions": []any{map[string]any{"id": "session", "state": map[string]any{"system_prompt_sections": strings.Repeat("prompt context ", 75000)}}}}
	b.ReportAllocs()
	for b.Loop() {
		writeJSON(discardJSONWriter{}, http.StatusOK, payload)
	}
}

type discardJSONWriter struct{}

func (discardJSONWriter) Header() http.Header         { return make(http.Header) }
func (discardJSONWriter) WriteHeader(int)             {}
func (discardJSONWriter) Write(p []byte) (int, error) { return len(p), nil }

type failingJSONWriter struct {
	header http.Header
	calls  int
	failAt int
}

func (w *failingJSONWriter) Header() http.Header { return w.header }
func (w *failingJSONWriter) WriteHeader(int)     {}
func (w *failingJSONWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls == w.failAt {
		return 0, errors.New("closed")
	}
	return len(p), nil
}

func TestWriteJSONResponseAndEncodingFailure(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSON(rr, http.StatusCreated, map[string]any{"text": "<中文>\n"})
	if rr.Code != http.StatusCreated || rr.Header().Get("Content-Type") != "application/json" {
		t.Fatal(rr.Code, rr.Header())
	}
	if !bytes.HasSuffix(rr.Body.Bytes(), []byte("\n")) || !json.Valid(rr.Body.Bytes()) {
		t.Fatal(rr.Body.String())
	}
	rr = httptest.NewRecorder()
	writeJSON(rr, http.StatusCreated, map[string]any{"invalid": make(chan int)})
	if rr.Code != http.StatusInternalServerError || strings.Contains(rr.Body.String(), "\"invalid\"") {
		t.Fatal(rr.Code, rr.Body.String())
	}
	for _, failAt := range []int{1, 2} {
		w := &failingJSONWriter{header: make(http.Header), failAt: failAt}
		writeJSON(w, http.StatusOK, map[string]int{"ok": 1})
	}
}

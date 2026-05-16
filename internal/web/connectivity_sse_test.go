package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestConnectivitySSEStopsAfterMarshalFailure(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.New(s)
	defer engine.Close()
	srv := New(s, engine, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/connect/sse/runtime.test", nil).WithContext(ctx)
	res := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		srv.handleConnectivitySSE(res, req)
	}()

	time.Sleep(50 * time.Millisecond)
	_ = engine.Connectivity().Emit(context.Background(), "runtime.test", map[string]any{"bad": func() {}})
	time.Sleep(50 * time.Millisecond)
	_ = engine.Connectivity().Emit(context.Background(), "runtime.test", map[string]any{"ok": true})

	select {
	case <-done:
	case <-time.After(time.Second):
		cancel()
		t.Fatal("timed out waiting for connectivity sse handler to stop after marshal failure")
	}

	body := res.Body.String()
	if !strings.Contains(body, ": connected ") {
		t.Fatalf("expected initial connected prelude, got %q", body)
	}
	if strings.Contains(body, `"ok":true`) {
		t.Fatalf("expected stream to stop before later good event after marshal failure, got %q", body)
	}
}

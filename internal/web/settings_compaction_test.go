package web

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestSettingsCompactionReportsEnginePolicy(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	session, err := st.CreateSession(t.Context(), "settings", "settings", nil)
	if err != nil {
		t.Fatal(err)
	}
	policy := config.CompactionSettings{Enabled: true, ContextWindow: 8000, ReserveTokens: 1000, KeepRecentTokens: 1200, ThresholdTokens: 6500, Strategy: "local"}
	engine := turn.NewWithRuntimeConfig(st, config.RuntimeConfig{DefaultModel: "test-model", Compaction: policy}, "test-model")
	// Intentionally different server config: the engine is authoritative.
	srv := New(st, engine, config.RuntimeConfig{})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/api/sessions/"+session.ID+"/compaction", nil))
	if rec.Code != 200 || rec.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatalf("response %d %+v", rec.Code, rec.Header())
	}
	var result struct {
		Policy    config.CompactionSettings `json:"policy"`
		Scope     string                    `json:"policy_scope"`
		Available bool                      `json:"available"`
		Reason    string                    `json:"reason"`
	}
	if err = json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Policy != policy || result.Scope != "startup" || result.Available || result.Reason == "" {
		t.Fatalf("payload %+v", result)
	}
}

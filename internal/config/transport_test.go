package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPiTransportSettingsPrecedenceAndLegacy(t *testing.T) {
	for _, tc := range []struct{ name, user, project, want string }{
		{"default", `{}`, `{}`, "auto"},
		{"managed SSE", `{"transport":"sse"}`, `{}`, "sse"},
		{"project override", `{"transport":"sse"}`, `{"transport":"websocket"}`, "websocket"},
		{"explicit auto", `{"transport":"sse"}`, `{"transport":"auto"}`, "auto"},
		{"cached websocket", `{}`, `{"transport":"websocket-cached"}`, "websocket-cached"},
		{"legacy false", `{"websockets":false}`, `{}`, "sse"},
		{"legacy project true", `{"transport":"sse"}`, `{"websockets":true}`, "websocket"},
		{"enum wins", `{"transport":"sse","websockets":true}`, `{}`, "sse"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			agent, workspace := t.TempDir(), t.TempDir()
			t.Setenv("PI_CODING_AGENT_DIR", agent)
			t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(agent, "absent-gi"))
			if err := os.MkdirAll(filepath.Join(workspace, ".pi"), 0700); err != nil {
				t.Fatal(err)
			}
			files := map[string]string{filepath.Join(agent, "settings.json"): tc.user, filepath.Join(workspace, ".pi", "settings.json"): tc.project}
			for path, text := range files {
				if err := os.WriteFile(path, []byte(text), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if got := Load(workspace).Transport; got != tc.want {
				t.Fatalf("transport=%q want %q", got, tc.want)
			}
			for path, text := range files {
				got, err := os.ReadFile(path)
				if err != nil || string(got) != text {
					t.Fatalf("managed configuration was modified: %s", path)
				}
			}
		})
	}
}

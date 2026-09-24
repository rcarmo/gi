package web

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/skills"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestLoadedWebSkillCatalogueAndCapturedInvocation(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".gi", "skills", "proof", "SKILL.md")
	os.MkdirAll(filepath.Dir(path), 0700)
	content := "---\nname: proof\ndescription: cerulean evidence\n---\nNATIVE_SKILL_MARKER\n"
	os.WriteFile(path, []byte(content), 0600)
	duplicate := filepath.Join(root, ".pi", "skills", "duplicate", "SKILL.md")
	os.MkdirAll(filepath.Dir(duplicate), 0700)
	os.WriteFile(duplicate, []byte("---\nname: proof\n---\nBAD_DUPLICATE"), 0600)
	cfg := config.Load(root)
	cfg.DefaultModel = "bootstrap"
	s, err := store.Open(filepath.Join(root, "gi.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := turn.NewWithRuntimeConfig(s, cfg, "")
	defer e.Close()
	srv := New(s, e, cfg)
	for _, id := range []string{"A", "B"} {
		s.CreateSession(context.Background(), id, id, map[string]any{"model": "bootstrap"})
	}
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/quick-actions", nil))
	var catalogue struct {
		Commands []map[string]string `json:"commands"`
	}
	json.Unmarshal(w.Body.Bytes(), &catalogue)
	count := 0
	for _, cmd := range catalogue.Commands {
		if cmd["name"] == "/skill:proof" {
			count++
			if cmd["description"] != "cerulean evidence" {
				t.Fatal(cmd)
			}
		}
	}
	if count != 1 {
		t.Fatal(w.Body.String())
	}
	send := func(prompt, target string) *httptest.ResponseRecorder {
		t.Helper()
		raw, _ := json.Marshal(map[string]any{"prompt": prompt, "target_agent_id": target})
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, httptest.NewRequest("POST", "/api/sessions/A/prompt", strings.NewReader(string(raw))))
		return w
	}
	accepted := send("/skill:proof request β", "")
	if accepted.Code != 202 {
		t.Fatal(accepted.Code, accepted.Body.String())
	}
	var result turn.SubmitResult
	json.Unmarshal(accepted.Body.Bytes(), &result)
	rec, err := s.GetTurn(context.Background(), result.TurnID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.SessionID != "A" || !strings.Contains(rec.Prompt, "NATIVE_SKILL_MARKER") || !strings.Contains(rec.Prompt, "request β") || strings.Contains(rec.Prompt, "BAD_DUPLICATE") || rec.Metadata["skill_name"] != "proof" {
		t.Fatal(rec)
	}
	for _, command := range []string{"/skill:unknown nope", "/skill:../proof nope"} {
		if r := send(command, ""); r.Code != 400 {
			t.Fatal(r.Code, r.Body.String())
		}
	}
	if r := send("/skill:proof cross", "B"); r.Code != 400 {
		t.Fatal(r.Code)
	}
	os.WriteFile(path, []byte(content+"changed"), 0600)
	if r := send("/skill:proof stale", ""); r.Code != 400 || !strings.Contains(r.Body.String(), "changed") {
		t.Fatal(r.Code, r.Body.String())
	}
	os.Remove(path)
	if r := send("/skill:proof missing", ""); r.Code != 400 {
		t.Fatal(r.Code)
	}
	other, _ := s.ListTurns(context.Background(), "B")
	if len(other) != 0 {
		t.Fatal("cross-session execution")
	}
}
func TestLoadedWebSkillsFailClosedOnUnsafeUnboundedOrNewFiles(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	os.WriteFile(outside, []byte("external"), 0600)
	os.Symlink(outside, filepath.Join(root, "escape"))
	os.WriteFile(filepath.Join(root, "big"), []byte(strings.Repeat("a", maxWebSkillBytes+1)), 0600)
	os.WriteFile(filepath.Join(root, "binary"), []byte{0xff}, 0600)
	os.WriteFile(filepath.Join(root, "valid"), []byte("bounded UTF-8 β"), 0600)
	cfg := config.RuntimeConfig{WorkspaceRoot: root, Discovery: skills.Discovery{Skills: []skills.Skill{{Name: "escape", Path: filepath.Join(root, "escape")}, {Name: "big", Path: filepath.Join(root, "big")}, {Name: "binary", Path: filepath.Join(root, "binary")}, {Name: "../bad", Path: filepath.Join(root, "valid")}, {Name: "valid", Path: filepath.Join(root, "valid")}}}}
	loaded := loadWebSkills(cfg)
	if len(loaded) != 1 || loaded["valid"].name != "valid" {
		t.Fatal(loaded)
	}
	srv := &Server{cfg: cfg, webSkills: loaded}
	os.WriteFile(filepath.Join(root, "new"), []byte("new"), 0600)
	if _, _, err := srv.expandWebSkill("/skill:new request"); err == nil {
		t.Fatal("unloaded skill executed")
	}
	os.Remove(filepath.Join(root, "valid"))
	os.Symlink(outside, filepath.Join(root, "valid"))
	if _, _, err := srv.expandWebSkill("/skill:valid request"); err == nil {
		t.Fatal("symlink escape executed")
	}
}

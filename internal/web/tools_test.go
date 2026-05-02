package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func newTestWebServer(t *testing.T, root string) (*Server, *store.Store) {
	t.Helper()
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	cfg := config.RuntimeConfig{WorkspaceRoot: root}
	srv := New(s, turn.New(s), cfg)
	return srv, s
}

type toolReq struct {
	Tool  string `json:"tool"`
	Input any    `json:"input"`
}

type toolResp struct {
	Tools  []map[string]any `json:"tools"`
	Result string           `json:"result"`
	Error  string           `json:"error"`
}

func callToolExecute(t *testing.T, srv *Server, req toolReq) toolResp {
	t.Helper()
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	w := httptest.NewRecorder()
	h := httptest.NewRequest(http.MethodPost, "/api/tools/execute", bytes.NewBuffer(body))
	h.Header.Set("Content-Type", "application/json")
	srv.Handler().ServeHTTP(w, h)
	if w.Code != http.StatusOK {
		t.Fatalf("tool execute status: %d body=%s", w.Code, w.Body.String())
	}
	var resp toolResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode tool response: %v", err)
	}
	return resp
}

func TestToolsDiscoveryIncludesReadWriteShell(t *testing.T) {
	root := t.TempDir()
	srv, _ := newTestWebServer(t, root)
	w := httptest.NewRecorder()
	h := httptest.NewRequest(http.MethodGet, "/api/tools", nil)
	srv.Handler().ServeHTTP(w, h)
	if w.Code != http.StatusOK {
		t.Fatalf("tools status: %d body=%s", w.Code, w.Body.String())
	}
	var payload toolResp
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode tools: %v", err)
	}
	if len(payload.Tools) < 4 {
		t.Fatalf("expected at least 4 tools, got %+v", len(payload.Tools))
	}
	names := map[string]bool{}
	for _, tool := range payload.Tools {
		name, ok := tool["name"].(string)
		if !ok {
			continue
		}
		names[name] = true
	}
	for _, want := range []string{"script", "read", "write", "shell"} {
		if !names[want] {
			t.Fatalf("missing tool %q in /api/tools", want)
		}
	}
}

func TestToolExecuteReadWriteWorkspaceFiles(t *testing.T) {
	root := t.TempDir()
	rawPath := filepath.Join(root, "seed.txt")
	if err := os.WriteFile(rawPath, []byte("seed-content\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	srv, _ := newTestWebServer(t, root)

	readResp := callToolExecute(t, srv, toolReq{
		Tool:  "read",
		Input: map[string]any{"path": "seed.txt"},
	})
	if readResp.Error != "" {
		t.Fatalf("read should succeed: %v", readResp.Error)
	}
	if strings.TrimSpace(readResp.Result) != "seed-content" {
		t.Fatalf("unexpected read result: %q", readResp.Result)
	}

	writeResp := callToolExecute(t, srv, toolReq{
		Tool:  "write",
		Input: map[string]any{"path": "out.txt", "content": "from-tool"},
	})
	if writeResp.Error != "" || writeResp.Result != "written" {
		t.Fatalf("write failed: %#v", writeResp)
	}
	if b, err := os.ReadFile(filepath.Join(root, "out.txt")); err != nil {
		t.Fatalf("read written file: %v", err)
	} else if string(b) != "from-tool" {
		t.Fatalf("unexpected written content: %q", string(b))
	}
}

func TestToolExecuteReadWriteVFSFiles(t *testing.T) {
	root := t.TempDir()
	srv, storeDB := newTestWebServer(t, root)

	ctx := context.Background()
	if _, err := storeDB.SaveVFSFile(ctx, "scripts", "nested/seed.txt", "text/plain", []byte("vfs-seed"), map[string]any{"kind": "seed"}); err != nil {
		t.Fatalf("seed vfs: %v", err)
	}

	readResp := callToolExecute(t, srv, toolReq{
		Tool:  "read",
		Input: map[string]any{"path": "vfs://scripts/nested/seed.txt"},
	})
	if readResp.Error != "" {
		t.Fatalf("vfs read should succeed: %v", readResp.Error)
	}
	if strings.TrimSpace(readResp.Result) != "vfs-seed" {
		t.Fatalf("unexpected vfs read result: %q", readResp.Result)
	}

	writeResp := callToolExecute(t, srv, toolReq{
		Tool:  "write",
		Input: map[string]any{"path": "vfs://scripts/nested/write.txt", "content": "vfs-write"},
	})
	if writeResp.Error != "" || writeResp.Result != "written" {
		t.Fatalf("vfs write failed: %#v", writeResp)
	}
	if _, _, err := storeDB.GetVFSFileContent(ctx, "scripts", "nested/write.txt"); err != nil {
		t.Fatalf("read-back vfs write: %v", err)
	}
}

func TestToolExecuteReadPathErrors(t *testing.T) {
	root := t.TempDir()
	srv, _ := newTestWebServer(t, root)

	respEmpty := callToolExecute(t, srv, toolReq{
		Tool:  "read",
		Input: map[string]any{"path": ""},
	})
	if respEmpty.Error == "" {
		t.Fatal("expected error for empty read path")
	}
	if !strings.Contains(respEmpty.Error, "empty path") {
		t.Fatalf("unexpected empty-path error: %v", respEmpty.Error)
	}

	respTraversal := callToolExecute(t, srv, toolReq{
		Tool:  "read",
		Input: map[string]any{"path": "../evil.txt"},
	})
	if !strings.Contains(respTraversal.Error, "path escapes workspace") {
		t.Fatalf("unexpected traversal error: %v", respTraversal.Error)
	}
}

func TestToolExecutePreventsReferenceWrites(t *testing.T) {
	root := t.TempDir()
	srv, _ := newTestWebServer(t, root)

	resp := callToolExecute(t, srv, toolReq{
		Tool:  "write",
		Input: map[string]any{"path": "vfs://reference/system/immutable.txt", "content": "blocked"},
	})
	if resp.Error == "" {
		t.Fatal("expected error on reference namespace write")
	}
	if !strings.Contains(resp.Error, "read-only") {
		t.Fatalf("unexpected error for reference write: %v", resp.Error)
	}
}

func TestToolExecuteRejectsInvalidVFSWritePath(t *testing.T) {
	root := t.TempDir()
	srv, _ := newTestWebServer(t, root)

	resp := callToolExecute(t, srv, toolReq{
		Tool:  "write",
		Input: map[string]any{"path": "vfs://scripts/../evil.txt", "content": "blocked"},
	})
	if resp.Error == "" {
		t.Fatal("expected error for invalid vfs write path")
	}
	if !strings.Contains(resp.Error, "traversal outside namespace") {
		t.Fatalf("unexpected invalid vfs write error: %v", resp.Error)
	}
}

func TestToolExecuteShellTool(t *testing.T) {
	root := t.TempDir()
	srv, _ := newTestWebServer(t, root)

	resp := callToolExecute(t, srv, toolReq{
		Tool:  "shell",
		Input: map[string]any{"command": "printf hello"},
	})
	if resp.Error != "" {
		t.Fatalf("shell error: %v", resp.Error)
	}
	if strings.TrimSpace(resp.Result) != "hello" {
		t.Fatalf("unexpected shell output: %q", resp.Result)
	}
}

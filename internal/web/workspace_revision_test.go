package web

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/rcarmo/gi/internal/config"
)

func TestWorkspaceRevisionConditionalWritesAndCreateOnlyCopy(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a.md")
	os.WriteFile(path, []byte("loaded"), 0600)
	s := New(nil, nil, config.RuntimeConfig{WorkspaceRoot: root})
	call := func(method, url string, body any) (int, map[string]any) {
		raw, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, httptest.NewRequest(method, url, strings.NewReader(string(raw))))
		var out map[string]any
		json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	status, loaded := call("GET", "/api/workspace/file?mode=edit&path=a.md", nil)
	if status != 200 || loaded["revision"] == "" || loaded["truncated"] != false {
		t.Fatal(status, loaded)
	}
	revision := loaded["revision"]
	if status, out := call("GET", "/api/workspace/file?path=a.md&max_bytes=1", nil); status != 200 || out["revision"] != nil {
		t.Fatal("preview granted edit authority", status, out)
	}
	if status, out := call("PUT", "/api/workspace/file", map[string]any{"path": "a.md", "content": "no revision"}); status != 428 || out["code"] != "revision_required" {
		t.Fatal(status, out)
	}
	os.WriteFile(path, []byte("remote"), 0600)
	status, out := call("PUT", "/api/workspace/file", map[string]any{"path": "a.md", "content": "stale", "expected_revision": revision})
	if status != 409 || out["code"] != "revision_conflict" || out["revision"] == revision {
		t.Fatal(status, out)
	}
	disk, _ := os.ReadFile(path)
	if string(disk) != "remote" {
		t.Fatal("stale overwrite", string(disk))
	}
	status, reviewed := call("GET", "/api/workspace/file?mode=edit&path=a.md", nil)
	revision = reviewed["revision"]
	statuses := make(chan int, 2)
	var wg sync.WaitGroup
	for _, content := range []string{"one", "two"} {
		wg.Add(1)
		go func(content string) {
			defer wg.Done()
			status, _ := call("PUT", "/api/workspace/file", map[string]any{"path": "a.md", "content": content, "expected_revision": revision})
			statuses <- status
		}(content)
	}
	wg.Wait()
	close(statuses)
	counts := map[int]int{}
	for status := range statuses {
		counts[status]++
	}
	if counts[200] != 1 || counts[409] != 1 {
		t.Fatal(counts)
	}
	status, current := call("GET", "/workspace/file?mode=edit&path=a.md", nil)
	if status != 200 || current["revision"] == revision {
		t.Fatal(current)
	}
	info, _ := os.Stat(path)
	status, saved := call("PUT", "/workspace/file", map[string]any{"path": "a.md", "content": current["text"], "expected_revision": current["revision"]})
	after, _ := os.Stat(path)
	if status != 200 || saved["revision"] != current["revision"] || !info.ModTime().Equal(after.ModTime()) {
		t.Fatal("no-op changed revision", status, saved)
	}
	status, _ = call("POST", "/api/workspace/file", map[string]any{"path": ".", "name": "copy.md", "content": "copy"})
	if status != 200 {
		t.Fatal(status)
	}
	status, out = call("POST", "/api/workspace/file", map[string]any{"path": ".", "name": "copy.md", "content": "overwrite"})
	if status != 409 || out["code"] != "file_exists" {
		t.Fatal(status, out)
	}
	disk, _ = os.ReadFile(filepath.Join(root, "copy.md"))
	if string(disk) != "copy" {
		t.Fatal("copy overwritten")
	}
	status, _ = call("PUT", "/api/workspace/file", map[string]any{"path": "missing.md", "content": "copy", "expected_revision": current["revision"]})
	if status != 404 {
		t.Fatal("PUT became upsert", status)
	}
}

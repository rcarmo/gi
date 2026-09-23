package web

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/rcarmo/gi/internal/search/chunking"
	"github.com/rcarmo/gi/internal/search/indexer"
	searchstore "github.com/rcarmo/gi/internal/search/store"
)

type workspaceIndexStatus struct {
	Scope            string   `json:"scope"`
	State            string   `json:"state"`
	Roots            []string `json:"roots"`
	IndexedFileCount int      `json:"indexed_file_count"`
	LastIndexedAt    string   `json:"last_indexed_at,omitempty"`
	UpdatedAt        string   `json:"updated_at,omitempty"`
	LastError        string   `json:"last_error,omitempty"`
	Generation       int64    `json:"generation"`
	ConfigHash       string   `json:"config_hash"`
	RequiredRoots    bool     `json:"required_roots"`
}

func (s *Server) workspaceScope(r *http.Request) (searchstore.ScopeConfig, error) {
	root := s.cfg.WorkspaceRoot
	if root == "" {
		root = "/workspace"
	}
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "all"
	}
	return searchstore.DefaultScopeConfig(root, scope, nil, nil, chunking.LineVersion)
}
func (s *Server) indexStatus(r *http.Request, c searchstore.ScopeConfig) (workspaceIndexStatus, error) {
	stored, err := searchstore.NewRefreshStore(s.store.DB()).Status(r.Context(), c)
	result := workspaceIndexStatus{Scope: c.Scope(), State: stored.State, Roots: stored.Roots, IndexedFileCount: stored.IndexedFileCount, LastError: stored.LastError, Generation: stored.Generation, ConfigHash: stored.ConfigHash, RequiredRoots: true}
	if stored.LastIndexedAtMS.Valid {
		result.LastIndexedAt = time.UnixMilli(stored.LastIndexedAtMS.Int64).UTC().Format(time.RFC3339Nano)
	}
	if stored.UpdatedAtMS != 0 {
		result.UpdatedAt = time.UnixMilli(stored.UpdatedAtMS).UTC().Format(time.RFC3339Nano)
	}
	return result, err
}
func (s *Server) handleWorkspaceIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	config, err := s.workspaceScope(r)
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	if r.Method == http.MethodPost {
		worker := indexer.NewWorker(searchstore.NewRefreshStore(s.store.DB()))
		if err = worker.Run(r.Context(), config); err != nil {
			code := 500
			if errors.Is(err, searchstore.ErrRefreshBusy) || errors.Is(err, searchstore.ErrRefreshLost) {
				code = 409
			}
			writeJSON(w, code, map[string]any{"error": err.Error()})
			return
		}
	}
	status, err := s.indexStatus(r, config)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "Unable to read workspace index status"})
		return
	}
	writeJSON(w, 200, status)
}
func (s *Server) handleWorkspaceSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	config, err := s.workspaceScope(r)
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	limit, offset := 10, 0
	if r.URL.Query().Has("limit") {
		limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit < 1 || limit > 50 {
			writeJSON(w, 400, map[string]any{"error": "Invalid search limit"})
			return
		}
	}
	if r.URL.Query().Has("offset") {
		offset, err = strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil || offset < 0 || offset > 10000 {
			writeJSON(w, 400, map[string]any{"error": "Invalid search offset"})
			return
		}
	}
	// GET is strictly read-only, never a disguised refresh. Explicit POST /index
	// must complete before a caller requests refreshed query results.
	if r.URL.Query().Has("refresh") {
		writeJSON(w, 400, map[string]any{"error": "Use POST /api/workspace/index for explicit refresh"})
		return
	}
	q := r.URL.Query().Get("q")
	if err := searchstore.ValidateLexicalQuery(q, limit, offset); err != nil {
		writeJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	result, err := searchstore.NewRefreshStore(s.store.DB()).Query(r.Context(), config, q, limit, offset)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "Unable to search workspace index"})
		return
	}
	writeJSON(w, 200, map[string]any{"scope": config.Scope(), "query": q, "mode": result.Mode, "hits": result.Hits})
}

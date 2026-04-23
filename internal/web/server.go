package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	"github.com/rcarmo/gi/internal/turn"
)

//go:embed all:static
var staticFS embed.FS

type Server struct {
	store      *store.Store
	turns      *turn.Engine
	cfg        config.RuntimeConfig
	mux        *http.ServeMux
	version    string
	scriptTool *tools.ScriptTool
}

func New(s *store.Store, t *turn.Engine, cfg config.RuntimeConfig) *Server {
	srv := &Server{
		store:      s,
		turns:      t,
		cfg:        cfg,
		mux:        http.NewServeMux(),
		version:    fmt.Sprintf("%x", time.Now().UnixNano()),
		scriptTool: tools.NewScriptTool(s, cfg),
	}
	srv.routes()
	return srv
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	staticRoot, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(staticRoot))
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			s.serveIndex(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
	s.mux.HandleFunc("/api/runtime/config", s.handleRuntimeConfig)
	s.mux.HandleFunc("/api/frontend/log", s.handleFrontendLog)
	s.mux.HandleFunc("/api/workspace/tree", s.handleWorkspaceTree)
	s.mux.HandleFunc("/api/workspace/file", s.handleWorkspaceFile)
	s.mux.HandleFunc("/sse/stream", s.handleSSEStream)
	s.mux.HandleFunc("/api/system-metrics", s.handleSystemMetrics)
	s.mux.HandleFunc("/agent/system-metrics", s.handleSystemMetrics)
	s.mux.HandleFunc("/api/tools", s.handleTools)
	s.mux.HandleFunc("/api/tools/execute", s.handleToolExecute)
	s.mux.HandleFunc("/api/sessions", s.handleSessions)
	s.mux.HandleFunc("/api/sessions/", s.handleSessionSubroutes)
	s.mux.HandleFunc("/api/turns/", s.handleTurnSubroutes)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		sessions, err := s.store.ListSessions(ctx)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
	case http.MethodPost:
		var req struct {
			Title string `json:"title"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		id := store.NowID("session")
		session, err := s.store.CreateSession(ctx, id, req.Title, map[string]any{"status": "idle", "queue_count": 0, "model": s.cfg.DefaultModel, "provider": s.cfg.DefaultProvider, "thinking_level": s.cfg.DefaultThinkingLevel})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, session)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSessionSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	sessionID := parts[0]
	if len(parts) == 1 {
		s.handleSession(w, r, sessionID)
		return
	}
	switch parts[1] {
	case "messages":
		s.handleMessages(w, r, sessionID)
	case "prompt":
		s.handlePrompt(w, r, sessionID)
	case "turns":
		s.handleTurns(w, r, sessionID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleTurnSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/turns/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	turnID := parts[0]
	switch parts[1] {
	case "cancel":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		turnRec, err := s.store.GetTurn(r.Context(), turnID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		if err := s.turns.CancelTurn(r.Context(), turnRec.SessionID, turnID); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	case "events":
		events, err := s.store.ListTurnEvents(r.Context(), turnID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"events": events})
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, err := s.store.GetSession(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	msgs, err := s.store.ListMessages(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messages": msgs})
}

func (s *Server) handleTurns(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	turns, err := s.store.ListTurns(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"turns": turns})
}

func (s *Server) handlePrompt(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Prompt string `json:"prompt"`
		Intent string `json:"intent"`
		Model  string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	model := req.Model
	if model == "" {
		model = s.cfg.DefaultModel
	}
	result, err := s.turns.SubmitPrompt(r.Context(), turn.RunInput{SessionID: sessionID, Prompt: req.Prompt, Intent: req.Intent, Model: model})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, result)
}

func (s *Server) handleRuntimeConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"workspace_root":         s.cfg.WorkspaceRoot,
		"assistant_name":         s.cfg.AssistantName,
		"assistant_avatar":       s.cfg.AssistantAvatar,
		"user_name":              s.cfg.UserName,
		"user_avatar":            s.cfg.UserAvatar,
		"user_avatar_background": s.cfg.UserAvatarBackground,
		"default_provider":       s.cfg.DefaultProvider,
		"default_model":          s.cfg.DefaultModel,
		"default_thinking_level": s.cfg.DefaultThinkingLevel,
		"enabled_models":         s.cfg.EnabledModels,
		"version":                s.version,
	})
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}
	html := strings.ReplaceAll(string(data), ".js\"", ".js?v="+s.version+"\"")
	html = strings.ReplaceAll(html, ".css\"", ".css?v="+s.version+"\"")
	html = strings.ReplaceAll(html, ".ico\"", ".ico?v="+s.version+"\"")
	html = strings.ReplaceAll(html, ".png\"", ".png?v="+s.version+"\"")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

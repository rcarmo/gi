package web

import (
	"context"
	"database/sql"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	giauth "github.com/rcarmo/gi/internal/auth"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/search/indexer"
	searchstore "github.com/rcarmo/gi/internal/search/store"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	storeaudit "github.com/rcarmo/gi/internal/store/audit"
	"github.com/rcarmo/gi/internal/store/queue"
	"github.com/rcarmo/gi/internal/tools"
	"github.com/rcarmo/gi/internal/turn"
)

//go:embed all:static
var staticFS embed.FS

type Server struct {
	store                 *store.Store
	turns                 *turn.Engine
	cfg                   config.RuntimeConfig
	mux                   *http.ServeMux
	version               string
	scriptTool            *tools.ScriptTool
	auth                  *giauth.Manager
	inboundDispatcherOnce sync.Once
	indexMu               sync.Mutex
	indexScheduler        *indexer.Scheduler
	indexConfigs          map[string]searchstore.ScopeConfig
	indexClosed           bool
	webSkills             map[string]loadedWebSkill
}

func New(s *store.Store, t *turn.Engine, cfg config.RuntimeConfig) *Server {
	srv := &Server{
		store:      s,
		turns:      t,
		cfg:        cfg,
		mux:        http.NewServeMux(),
		version:    fmt.Sprintf("%x", time.Now().UnixNano()),
		scriptTool: tools.NewScriptTool(s, cfg),
		auth:       giauth.NewManagerWithPasskeys(cfg.WorkspaceRoot, giauth.PasskeyConfig{RPID: cfg.Passkeys.RPID, Origins: cfg.Passkeys.Origins}),
	}
	srv.webSkills = loadWebSkills(cfg)
	srv.configureScriptConnectivity()
	srv.routes()
	return srv
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) StartInboundWorkDispatcher(ctx context.Context) {
	if !s.cfg.InboundWork.Enabled {
		return
	}
	s.inboundDispatcherOnce.Do(func() {
		if ctx == nil {
			ctx = context.Background()
		}
		releaseCtx := context.WithoutCancel(ctx)
		interval := time.Duration(s.cfg.InboundWork.IntervalMS) * time.Millisecond
		if interval <= 0 {
			interval = 500 * time.Millisecond
		}
		batchSize := s.cfg.InboundWork.BatchSize
		if batchSize <= 0 {
			batchSize = 8
		}
		workerID := strings.TrimSpace(s.cfg.InboundWork.WorkerID)
		if workerID == "" {
			workerID = "web-runtime"
		}
		leaseOwner := workerID + ":" + s.version
		leaseTTL := time.Duration(s.cfg.InboundWork.LeaseTTLMS) * time.Millisecond
		if leaseTTL <= 0 {
			leaseTTL = 2 * time.Second
		}
		drain := func() {
			acquired, err := queue.AcquireInboundDispatcherLease(ctx, s.store.DB(), leaseOwner, leaseTTL)
			if err != nil {
				log.Printf("runtime inbound dispatcher lease: %v", err)
				return
			}
			if !acquired {
				s.turns.PublishRuntimeDispatcherEvent("dispatcher_lease_skipped", map[string]any{"worker_id": workerID, "lease_owner": leaseOwner})
				return
			}
			s.turns.PublishRuntimeDispatcherEvent("dispatcher_lease_acquired", map[string]any{"worker_id": workerID, "lease_owner": leaseOwner, "lease_ttl_ms": s.cfg.InboundWork.LeaseTTLMS})
			processed := 0
			for i := 0; i < batchSize; i++ {
				item, _, ok, err := s.turns.ProcessNextInboundWorkIfQueued(ctx, workerID)
				if !ok {
					break
				}
				processed++
				if err != nil {
					if item != nil {
						log.Printf("runtime inbound dispatcher item %d -> %s: %v", item.ID, item.Status, err)
					} else {
						log.Printf("runtime inbound dispatcher drain: %v", err)
					}
					continue
				}
			}
			if processed > 0 {
				log.Printf("runtime inbound dispatcher processed %d queued item(s)", processed)
				s.turns.PublishRuntimeDispatcherEvent("dispatcher_drain_processed", map[string]any{"worker_id": workerID, "lease_owner": leaseOwner, "processed": processed})
			}
		}
		go func() {
			defer func() {
				if err := queue.ReleaseInboundDispatcherLease(releaseCtx, s.store.DB(), leaseOwner); err != nil {
					log.Printf("runtime inbound dispatcher release lease: %v", err)
					return
				}
				s.turns.PublishRuntimeDispatcherEvent("dispatcher_lease_released", map[string]any{"worker_id": workerID, "lease_owner": leaseOwner})
			}()
			drain()
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					drain()
				}
			}
		}()
	})
}

func (s *Server) routes() {
	staticRoot, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(staticRoot))
	s.mux.HandleFunc("/manifest.json", s.serveManifest)
	s.mux.Handle("/static/icon-192.png", http.StripPrefix("/static", fileServer))
	s.mux.Handle("/static/icon-512.png", http.StripPrefix("/static", fileServer))
	// Supplied Adaptive Card renderer's lazy SDK URL (public static asset).
	s.mux.Handle("/static/js/vendor/adaptivecards.min.js", http.StripPrefix("/static", fileServer))
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			s.serveIndex(w, r)
			return
		}
		// The stable bootstrap points to this binary's hashed module graph. Even
		// a same-version rebuild must not reuse a bootstrap from an older binary.
		if r.URL.Path == "/dist/app.bundle.js" || r.URL.Path == "/dist/app.bundle.js.map" {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}
		fileServer.ServeHTTP(w, r)
	})
	s.mux.HandleFunc("/api/auth/status", s.handleAuthStatus)
	s.mux.HandleFunc("/api/auth/enroll/start", s.handleAuthEnrollStart)
	s.mux.HandleFunc("/api/auth/enroll/verify", s.handleAuthEnrollVerify)
	for _, operation := range []string{"start", "finish", "cancel"} {
		s.mux.HandleFunc("/api/auth/setup/"+operation, func(w http.ResponseWriter, r *http.Request) {
			s.handleBrowserSetup(w, r, operation)
		})
	}
	s.mux.HandleFunc("/api/auth/totp/verify", s.handleAuthTOTPVerify)
	s.mux.HandleFunc("/api/auth/session", s.handleAuthSession)
	s.mux.HandleFunc("/api/auth/session/logout", s.handleAuthLogout)
	s.mux.HandleFunc("/api/auth/session/proof", s.handleBrowserProof)
	s.mux.HandleFunc("/api/auth/policy", s.handleLoginPolicy)
	s.mux.HandleFunc("/api/auth/session/reauth/totp", s.handleBrowserReauthTOTP)
	s.mux.HandleFunc("/api/auth/passkeys", s.handlePasskeyList)
	s.mux.HandleFunc("/api/auth/passkeys/rename", s.handlePasskeyMutation)
	s.mux.HandleFunc("/api/auth/passkeys/remove", s.handlePasskeyMutation)
	for _, operation := range []string{"register", "login", "reauth"} {
		s.mux.HandleFunc("/api/auth/passkeys/"+operation+"/start", func(w http.ResponseWriter, r *http.Request) { s.handlePasskeyCeremony(w, r, operation, false) })
		s.mux.HandleFunc("/api/auth/passkeys/"+operation+"/finish", func(w http.ResponseWriter, r *http.Request) { s.handlePasskeyCeremony(w, r, operation, true) })
	}

	guard := s.withAuth
	s.mux.HandleFunc("/api/runtime/config", guard(s.handleRuntimeConfig))
	s.mux.HandleFunc("/api/settings/identity", guard(s.handleSettingsIdentity))
	s.mux.HandleFunc("/api/settings/compaction", guard(s.handleCompactionPolicy))
	s.mux.HandleFunc("/api/settings/providers", guard(s.handleProviderSettings))
	s.mux.HandleFunc("/api/runtime/inbound-work", guard(s.handleRuntimeInboundWork))
	s.mux.HandleFunc("/api/runtime/inbound-work/drain", guard(s.handleRuntimeInboundWorkDrain))
	s.mux.HandleFunc("/api/runtime/inbound-work/requeue", guard(s.handleRuntimeInboundWorkRequeue))
	s.mux.HandleFunc("/api/runtime/inbound-work/discard", guard(s.handleRuntimeInboundWorkDiscard))
	s.mux.HandleFunc("/api/frontend/log", guard(s.handleFrontendLog))
	s.mux.HandleFunc("/api/quick-actions", guard(s.handleQuickActions))
	s.mux.HandleFunc("/api/workspace/tree", guard(s.handleWorkspaceTree))
	s.mux.HandleFunc("/api/workspace/index", guard(s.handleWorkspaceIndex))
	s.mux.HandleFunc("/api/workspace/search", guard(s.handleWorkspaceSearch))
	s.mux.HandleFunc("/api/workspace/file", guard(s.handleWorkspaceFile))
	s.mux.HandleFunc("/api/workspace/raw", guard(s.handleWorkspaceRaw))
	s.mux.HandleFunc("/sse/stream", guard(s.handleSSEStream))
	s.mux.HandleFunc("/sse/topics", guard(s.handleTopicSSE))
	s.mux.HandleFunc("/api/system-metrics", guard(s.handleSystemMetrics))
	s.mux.HandleFunc("/agent/system-metrics", guard(s.handleSystemMetrics))
	s.mux.HandleFunc("/api/tools", guard(s.handleTools))
	s.mux.HandleFunc("/api/tools/execute", guard(s.handleToolExecute))
	s.mux.HandleFunc(connectivityRoutePrefix, guard(s.handleConnectivityRoutes))
	s.mux.HandleFunc(connectivitySSEPrefix, guard(s.handleConnectivitySSE))
	s.mux.HandleFunc("/api/sessions", guard(s.handleSessions))
	s.mux.HandleFunc("/api/sessions/", guard(s.handleSessionSubroutes))
	s.mux.HandleFunc("/api/media/", guard(s.handleMediaLookup))
	s.mux.HandleFunc("/api/turns/", guard(s.handleTurnSubroutes))
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
		if sessions == nil {
			sessions = []store.Session{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
	case http.MethodPost:
		var req struct {
			Title    string `json:"title"`
			AgentID  string `json:"agent_id"`
			ForkFrom string `json:"fork_from"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		id := store.NowID("session")
		if req.ForkFrom != "" {
			agentID := req.AgentID
			if agentID == "" {
				var err error
				agentID, err = s.nextForkAgentID(ctx, req.ForkFrom)
				if err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
					return
				}
			}
			title := req.Title
			if title == "" {
				title = "@" + agentID
			}
			session, err := s.store.CloneSession(ctx, req.ForkFrom, id, title, agentID)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusCreated, session)
			return
		}
		agentID := req.AgentID
		if agentID == "" {
			agentID = "agent"
		}
		title := req.Title
		if title == "" {
			title = "@" + agentID
		}
		alloc := gisession.AllocateDefaultSession(agentID, "gi", "default", id)
		session, _, err := s.store.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: id, Title: title, State: map[string]any{"status": "idle", "queue_count": 0, "model": s.cfg.DefaultModel, "provider": s.cfg.DefaultProvider, "thinking_level": s.cfg.DefaultThinkingLevel}, Allocation: alloc})
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
	case "search":
		s.handleSessionSearch(w, r, sessionID)
	case "messages":
		if len(parts) == 3 && parts[2] != "" {
			s.handleMessageDelete(w, r, sessionID, parts[2])
		} else if len(parts) == 2 {
			s.handleMessages(w, r, sessionID)
		} else {
			http.NotFound(w, r)
		}
	case "media":
		s.handleMedia(w, r, sessionID, parts[2:])
	case "prompt":
		s.handlePrompt(w, r, sessionID)
	case "turns":
		s.handleTurns(w, r, sessionID)
	case "send-receipt":
		s.handleSendReceipt(w, r, sessionID)
	case "compaction":
		s.handleSessionCompaction(w, r, sessionID)
	case "activity":
		s.handleSessionActivity(w, r, sessionID)
	case "resume-queue":
		s.handleSessionResume(w, r, sessionID)
	case "queue":
		s.handleSessionQueue(w, r, sessionID, parts[2:])
	case "model":
		s.handleSessionModel(w, r, sessionID)
	case "route-events":
		s.handleSessionRouteEvents(w, r, sessionID)
	case "introspect":
		s.handleSessionIntrospect(w, r, sessionID)
	case "fork":
		s.handleSessionFork(w, r, sessionID)
	case "peer-message":
		s.handleSessionPeerMessage(w, r, sessionID)
	case "continue":
		s.handleSessionContinue(w, r, sessionID)
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
	if r.Method != http.MethodGet && r.Method != http.MethodPatch {
		w.Header().Set("Allow", "GET, PATCH")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Method == http.MethodPatch {
		var mutation store.SessionMutation
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&mutation); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid session mutation"})
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Expected one session mutation"})
			return
		}
		if err := s.store.MutateSession(r.Context(), sessionID, mutation); err != nil {
			status := http.StatusInternalServerError
			switch {
			case errors.Is(err, sql.ErrNoRows):
				status = http.StatusNotFound
			case errors.Is(err, store.ErrSessionMutationInvalid):
				status = http.StatusBadRequest
			case errors.Is(err, store.ErrSessionMutationConflict):
				status = http.StatusConflict
			}
			writeJSON(w, status, map[string]any{"error": err.Error()})
			return
		}
	}
	session, err := s.store.GetSession(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) handleMedia(w http.ResponseWriter, r *http.Request, sessionID string, parts []string) {
	if _, err := s.store.GetSession(r.Context(), sessionID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}
	if len(parts) == 0 || parts[0] == "" {
		s.handleMediaCollection(w, r, sessionID)
		return
	}
	if len(parts) == 1 {
		s.handleMediaItem(w, r, sessionID, parts[0])
		return
	}
	http.NotFound(w, r)
}

func (s *Server) handleMediaCollection(w http.ResponseWriter, r *http.Request, sessionID string) {
	switch r.Method {
	case http.MethodGet:
		items, err := s.store.ListMedia(r.Context(), sessionID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		if items == nil {
			items = []store.Media{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"media": items})
	case http.MethodPost:
		media, err := s.createMediaFromRequest(r, sessionID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"media": media, "ref": store.MediaRef{ID: store.MediaRefID(media.ID), MediaID: media.ID, SessionID: media.SessionID, Filename: media.Filename, ContentType: media.ContentType, Size: media.OriginalSize, SHA256: internalString(media.Metadata["sha256"]), Source: internalString(media.Metadata["source"]), CreatedAt: media.CreatedAt}})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleMediaItem(w http.ResponseWriter, r *http.Request, sessionID, rawID string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	mediaID, ok := store.ParseMediaRefID(rawID)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid media id"})
		return
	}
	media, content, err := s.store.GetMediaContent(r.Context(), mediaID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}
	if media.SessionID != sessionID {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "media not found"})
		return
	}
	w.Header().Set("Content-Type", media.ContentType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+strings.ReplaceAll(media.Filename, "\"", "")+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (s *Server) createMediaFromRequest(r *http.Request, sessionID string) (*store.Media, error) {
	contentType := r.Header.Get("Content-Type")
	metadata := map[string]any{"source": "web"}
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			return nil, err
		}
		defer r.MultipartForm.RemoveAll()
		file, header, err := r.FormFile("file")
		if err != nil {
			return nil, err
		}
		defer file.Close()
		raw, err := io.ReadAll(io.LimitReader(file, (10<<20)+1))
		if err != nil {
			return nil, err
		}
		if len(raw) > 10<<20 {
			return nil, fmt.Errorf("media exceeds 10 MiB limit")
		}
		filename := header.Filename
		partContentType := header.Header.Get("Content-Type")
		return s.store.CreateOrReuseWebMedia(r.Context(), sessionID, filename, partContentType, raw)
	}
	var req struct {
		Filename    string         `json:"filename"`
		ContentType string         `json:"content_type"`
		Content     string         `json:"content_base64"`
		Source      string         `json:"source"`
		Metadata    map[string]any `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	if req.Source != "" {
		metadata["source"] = req.Source
	}
	for k, v := range req.Metadata {
		metadata[k] = v
	}
	raw, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		return nil, err
	}
	if len(raw) > 10<<20 {
		return nil, fmt.Errorf("media exceeds 10 MiB limit")
	}
	return s.store.CreateMedia(r.Context(), sessionID, req.Filename, req.ContentType, raw, metadata)
}

func internalString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Query().Has("limit") || r.URL.Query().Has("before") || r.URL.Query().Has("after") {
		if _, err := s.store.GetSession(r.Context(), sessionID); err != nil {
			writeJSON(w, 404, map[string]any{"error": "Session not found"})
			return
		}
		limit := 50
		if raw := r.URL.Query().Get("limit"); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil {
				writeJSON(w, 400, map[string]any{"error": "Invalid limit"})
				return
			}
			limit = value
		}
		page, err := s.store.PageMessages(r.Context(), sessionID, r.URL.Query().Get("before"), r.URL.Query().Get("after"), limit)
		if err != nil {
			code := 500
			if errors.Is(err, store.ErrMessageCursor) {
				code = 400
			}
			writeJSON(w, code, map[string]any{"error": "Unable to read message page"})
			return
		}
		writeJSON(w, 200, page)
		return
	}
	// Preserve the unpaged export API for existing consumers.
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

func (s *Server) handleSessionRouteEvents(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	events, err := storeaudit.ListRouteEvents(r.Context(), s.store.DB(), sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if events == nil {
		events = []storeaudit.RouteEvent{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"route_events": events})
}

func (s *Server) sessionInfo(ctx context.Context, sessionID string) (map[string]any, error) {
	session, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	messages, err := s.store.ListMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	turns, err := s.store.ListTurns(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	routeEvents, err := storeaudit.ListRouteEvents(ctx, s.store.DB(), sessionID)
	if err != nil {
		// Non-fatal introspection path; keep core payloads readable.
		routeEvents = nil
	}
	return map[string]any{
		"session":       session,
		"runtime":       map[string]any{"default_provider": s.cfg.DefaultProvider, "default_model": s.cfg.DefaultModel, "default_thinking_level": s.cfg.DefaultThinkingLevel, "workspace_root": s.cfg.WorkspaceRoot},
		"message_count": len(messages),
		"turn_count":    len(turns),
		"route_event_count": func() int {
			if routeEvents == nil {
				return 0
			}
			return len(routeEvents)
		}(),
		"messages":     messages,
		"turns":        turns,
		"route_events": routeEvents,
	}, nil
}

func (s *Server) handleSessionIntrospect(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	info, err := s.sessionInfo(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (s *Server) handlePrompt(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Prompt          string `json:"prompt"`
		Intent          string `json:"intent"`
		Model           string `json:"model"`
		TargetAgentID   string `json:"target_agent_id"`
		ParentTurnID    string `json:"parent_turn_id"`
		Media           []any  `json:"media"`
		ClientRequestID string `json:"client_request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if req.ClientRequestID != "" && !store.ValidWebSendToken(req.ClientRequestID) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid client_request_id"})
		return
	}
	if s.handleModelCommand(w, r, sessionID, req.Prompt) {
		return
	}
	expandedPrompt, skillMetadata, skillErr := s.expandWebSkill(req.Prompt)
	if skillErr != nil {
		writeJSON(w, 400, map[string]any{"error": skillErr.Error()})
		return
	}
	model := req.Model
	if model == "" {
		choice := inference.SessionModelChoice{Model: s.cfg.DefaultModel, Provider: s.cfg.DefaultProvider}
		if session, err := s.store.GetSession(r.Context(), sessionID); err == nil {
			choice = inference.SessionModel(session.State, choice)
		}
		model = choice.Label()
		// Native deterministic fixtures use bare sentinels, not provider calls.
		if choice.Provider == "test" && (choice.Model == "test-model" || choice.Model == "bootstrap") {
			model = choice.Model
		}
	}
	var (
		result *turn.SubmitResult
		err    error
	)
	// Prompt submission returns 202 while the turn continues asynchronously; do
	// not let client disconnect/navigation cancel accepted runtime work.
	submitCtx := context.WithoutCancel(r.Context())
	targetAgentID := strings.TrimPrefix(strings.TrimSpace(req.TargetAgentID), "@")
	currentAgentID := ""
	if identity, identityErr := s.store.RequireSessionIdentityRuntime(submitCtx, sessionID); identityErr == nil {
		currentAgentID = identity.AgentID
	}
	if skillMetadata != nil && targetAgentID != "" && targetAgentID != "default" && targetAgentID != currentAgentID {
		writeJSON(w, 400, map[string]any{"error": "skill invocation must use the selected session"})
		return
	}
	if targetAgentID != "" && targetAgentID != "default" && targetAgentID != currentAgentID {
		result, err = s.turns.SubmitPeerMessage(submitCtx, sessionID, targetAgentID, req.Prompt, req.Intent, model, req.ParentTurnID)
	} else {
		metadata := skillMetadata
		if metadata == nil {
			metadata = map[string]any{}
		}
		if req.ClientRequestID != "" {
			metadata["client_request_id"] = req.ClientRequestID
		}
		if len(req.Media) > 0 {
			metadata["media"] = req.Media
		}
		input := turn.RunInput{SessionID: sessionID, Prompt: expandedPrompt, Intent: req.Intent, Model: model, ParentTurnID: req.ParentTurnID, Metadata: metadata}
		if strings.HasPrefix(strings.TrimSpace(req.Prompt), "@") {
			result, err = s.turns.SubmitPromptRouted(submitCtx, input)
		} else {
			result, err = s.turns.SubmitPrompt(submitCtx, input)
		}
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if req.ClientRequestID != "" {
		// Native admission already happened: receipt failure is never a reason
		// to return rejection or retry the POST. Missing recovery stays unknown.
		if receiptErr := s.store.RecordWebSendReceipt(submitCtx, sessionID, req.ClientRequestID, result); receiptErr != nil {
			log.Printf("record web send receipt: %v", receiptErr)
		}
	}
	writeJSON(w, http.StatusAccepted, result)
}

func (s *Server) handleSessionContinue(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	continued, err := s.turns.ContinueSession(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"continued": continued})
}

func (s *Server) handleSessionPeerMessage(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		TargetAgentID string `json:"target_agent_id"`
		Content       string `json:"content"`
		Mode          string `json:"mode"`
		Model         string `json:"model"`
		ParentTurnID  string `json:"parent_turn_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	intent := req.Mode
	if intent == "" || intent == "auto" {
		intent = "prompt"
	}
	model := req.Model
	if model == "" {
		model = s.cfg.DefaultModel
	}
	result, err := s.turns.SubmitPeerMessage(r.Context(), sessionID, req.TargetAgentID, req.Content, intent, model, req.ParentTurnID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, result)
}

func (s *Server) handleSessionFork(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Title   string `json:"title"`
		AgentID string `json:"agent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	agentID := req.AgentID
	if agentID == "" {
		var err error
		agentID, err = s.nextForkAgentID(r.Context(), sessionID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
	}
	title := req.Title
	if title == "" {
		title = "@" + agentID
	}
	cloned, err := s.store.CloneSession(r.Context(), sessionID, store.NowID("session"), title, agentID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"branch": map[string]any{"chat_jid": "gi:" + cloned.ID, "label": cloned.Title, "agent_id": agentID, "source_chat_jid": "gi:" + sessionID}})
}

const defaultForkAgentID = "agent"
const minForkAgentIDSuffix = 1
const maxForkAgentIDSuffixExclusive = 1000

func normalizeForkSessionID(sessionID string) string {
	return strings.TrimSpace(sessionID)
}

func normalizeForkAgentID(agentID string) string {
	return strings.TrimSpace(agentID)
}

func mapForkAgentIDOrDefaultNormalized(normalizedSessionID string, agentBySession map[string]string) (agentID string, mapped bool) {
	if normalizedSessionID == "" {
		return defaultForkAgentID, false
	}
	if agentBySession == nil {
		return defaultForkAgentID, false
	}
	agentID = normalizeForkAgentID(agentBySession[normalizedSessionID])
	if agentID == "" {
		return defaultForkAgentID, false
	}
	return agentID, true
}

func (s *Server) sourceForkAgentID(ctx context.Context, sourceSessionID string, agentBySession map[string]string) string {
	if ctx == nil {
		ctx = context.Background()
	}
	normalizedSourceSessionID := normalizeForkSessionID(sourceSessionID)
	if normalizedSourceSessionID == "" {
		return defaultForkAgentID
	}
	agentID, mapped := mapForkAgentIDOrDefaultNormalized(normalizedSourceSessionID, agentBySession)
	if mapped {
		return agentID
	}
	identity, err := s.store.RequireSessionIdentityRuntime(ctx, normalizedSourceSessionID)
	if err != nil {
		return defaultForkAgentID
	}
	return identity.AgentID
}

func trimAgentNumericSuffix(agentID string) string {
	agentID = normalizeForkAgentID(agentID)
	trimmed := strings.TrimRightFunc(agentID, func(r rune) bool { return r >= '0' && r <= '9' })
	if trimmed == "" {
		return agentID
	}
	return trimmed
}

func buildUsedForkAgentIDs(agentBySession map[string]string) map[string]bool {
	if len(agentBySession) == 0 {
		return map[string]bool{}
	}
	used := make(map[string]bool, len(agentBySession))
	for sessionID := range agentBySession {
		normalizedSessionID := normalizeForkSessionID(sessionID)
		if normalizedSessionID == "" {
			continue
		}
		agentID, _ := mapForkAgentIDOrDefaultNormalized(normalizedSessionID, agentBySession)
		used[agentID] = true
	}
	return used
}

func normalizeForkAgentBase(base string) string {
	base = normalizeForkAgentID(base)
	if base == "" {
		return defaultForkAgentID
	}
	return base
}

func forkAgentIDCandidate(base string, suffix int) string {
	return base + strconv.Itoa(suffix)
}

func chooseNextForkAgentID(base string, used map[string]bool) (string, bool) {
	base = normalizeForkAgentBase(base)
	if len(used) == 0 {
		return forkAgentIDCandidate(base, minForkAgentIDSuffix), true
	}
	for i := minForkAgentIDSuffix; i < maxForkAgentIDSuffixExclusive; i++ {
		candidate := forkAgentIDCandidate(base, i)
		if !used[candidate] {
			return candidate, true
		}
	}
	return "", false
}

func (s *Server) nextForkAgentID(ctx context.Context, sourceSessionID string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	agentBySession, err := s.store.ListSessionAgentIDs(ctx)
	if err != nil {
		return "", err
	}
	sourceAgentID := s.sourceForkAgentID(ctx, sourceSessionID, agentBySession)
	base := trimAgentNumericSuffix(sourceAgentID)
	used := buildUsedForkAgentIDs(agentBySession)
	if candidate, ok := chooseNextForkAgentID(base, used); ok {
		return candidate, nil
	}
	return "", fmt.Errorf("could not allocate fork agent id from %s", normalizeForkAgentBase(base))
}

func (s *Server) handleRuntimeConfig(w http.ResponseWriter, r *http.Request) {
	providerOptions, modelOptions := inference.ListRuntimeOptions(s.cfg.DefaultProvider, s.cfg.DefaultModel, s.cfg.EnabledModels)
	currentModel := strings.TrimSpace(s.cfg.DefaultModel)
	if currentModel != "" && !strings.Contains(currentModel, "/") && strings.TrimSpace(s.cfg.DefaultProvider) != "" {
		currentModel = strings.TrimSpace(s.cfg.DefaultProvider) + "/" + currentModel
	}
	modelLabels := make([]string, 0, len(modelOptions))
	supportsThinking := false
	for _, option := range modelOptions {
		modelLabels = append(modelLabels, option.Label)
		if option.Label == currentModel || (option.ID == s.cfg.DefaultModel && option.Provider == s.cfg.DefaultProvider) {
			supportsThinking = option.Reasoning
			if currentModel == "" {
				currentModel = option.Label
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"workspace_root":         s.cfg.WorkspaceRoot,
		"workspace_index":        s.cfg.WorkspaceIndex,
		"assistant_name":         s.cfg.AssistantName,
		"assistant_avatar":       s.cfg.AssistantAvatar,
		"user_name":              s.cfg.UserName,
		"user_avatar":            s.cfg.UserAvatar,
		"user_avatar_background": s.cfg.UserAvatarBackground,
		"default_provider":       s.cfg.DefaultProvider,
		"default_model":          s.cfg.DefaultModel,
		"default_thinking_level": s.cfg.DefaultThinkingLevel,
		"enabled_models":         s.cfg.EnabledModels,
		"provider_options":       providerOptions,
		"model_options":          modelOptions,
		"models":                 modelLabels,
		"current":                currentModel,
		"thinking_level":         s.cfg.DefaultThinkingLevel,
		"supports_thinking":      supportsThinking,
		"version":                s.version,
	})
}

func (s *Server) handleRuntimeInboundWork(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		status := strings.TrimSpace(r.URL.Query().Get("status"))
		limit := 100
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid limit"})
				return
			}
			limit = parsed
		}
		var eligible *bool
		if raw := strings.TrimSpace(r.URL.Query().Get("eligible")); raw != "" {
			parsed, err := strconv.ParseBool(raw)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid eligible flag"})
				return
			}
			eligible = &parsed
		}
		items, err := queue.ListInboundWorkFiltered(r.Context(), s.store.DB(), status, limit, eligible)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		counts, err := queue.CountInboundWorkByStatus(r.Context(), s.store.DB())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		eligibleCount, err := queue.CountEligibleInboundWork(r.Context(), s.store.DB())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		if items == nil {
			items = []queue.InboundWorkItem{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"inbound_work": items, "counts": counts, "eligible_count": eligibleCount})
	case http.MethodPost:
		var req turn.DirectInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		if strings.TrimSpace(req.Model) == "" {
			req.Model = s.cfg.DefaultModel
		}
		item, err := s.turns.EnqueueDirectInbound(r.Context(), req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"item": item})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRuntimeInboundWorkDrain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ClaimedBy string `json:"claimed_by"`
		Limit     int    `json:"limit"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
	}
	items, results, err := s.turns.ProcessQueuedInboundWork(r.Context(), req.ClaimedBy, req.Limit)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if items == nil {
		items = []*queue.InboundWorkItem{}
	}
	if results == nil {
		results = []*turn.SubmitResult{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"processed": len(items), "items": items, "results": results})
}

func (s *Server) handleRuntimeInboundWorkRequeue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID            int64 `json:"id"`
		ResetAttempts bool  `json:"reset_attempts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if req.ID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id is required"})
		return
	}
	item, err := queue.RequeueInboundWork(r.Context(), s.store.DB(), req.ID, req.ResetAttempts)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	s.turns.PublishRuntimeInboundWorkEvent("inbound_work_requeued", item, map[string]any{"reset_attempts": req.ResetAttempts})
	writeJSON(w, http.StatusOK, map[string]any{"item": item})
}

func (s *Server) handleRuntimeInboundWorkDiscard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if req.ID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id is required"})
		return
	}
	item, err := queue.DiscardInboundWork(r.Context(), s.store.DB(), req.ID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	s.turns.PublishRuntimeInboundWorkEvent("inbound_work_discarded", item, nil)
	writeJSON(w, http.StatusOK, map[string]any{"item": item})
}

func (s *Server) serveManifest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := s.cfg.AssistantName
	if name == "" {
		name = "PiClaw"
	}
	icons := []map[string]string{}
	for _, size := range []string{"192", "512"} {
		for _, purpose := range []string{"any", "maskable"} {
			icons = append(icons, map[string]string{
				"src": "/static/icon-" + size + ".png", "sizes": size + "x" + size,
				"type": "image/png", "purpose": purpose,
			})
		}
	}
	body, err := json.Marshal(map[string]any{
		"name": name, "short_name": name,
		"description": "Slack-like interface for coding agents",
		"start_url":   "/", "display": "standalone",
		"display_override": []string{"window-controls-overlay"},
		"background_color": "#ffffff", "theme_color": "#ffffff", "color_scheme": "dark light",
		"icons": icons,
	})
	if err != nil {
		http.Error(w, "manifest encoding failed", http.StatusInternalServerError)
		return
	}
	body = append(body, '\n')
	w.Header().Set("Content-Type", "application/manifest+json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodGet {
		_, _ = w.Write(body)
	}
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
	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("serve index write: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	blob, err := json.Marshal(v)
	if err != nil {
		http.Error(w, fmt.Sprintf("json encode error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(blob, '\n')); err != nil {
		log.Printf("write json response: %v", err)
	}
}

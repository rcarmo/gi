package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	giauth "github.com/rcarmo/gi/internal/auth"
	"github.com/rcarmo/gi/internal/config"
	gisession "github.com/rcarmo/gi/internal/session"
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
	auth       *giauth.Manager
}

func New(s *store.Store, t *turn.Engine, cfg config.RuntimeConfig) *Server {
	srv := &Server{
		store:      s,
		turns:      t,
		cfg:        cfg,
		mux:        http.NewServeMux(),
		version:    fmt.Sprintf("%x", time.Now().UnixNano()),
		scriptTool: tools.NewScriptTool(s, cfg),
		auth:       giauth.NewManager(cfg.WorkspaceRoot),
	}
	srv.configureScriptConnectivity()
	srv.routes()
	return srv
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) StartInboundWorkDispatcher(ctx context.Context) {
	if !s.cfg.InboundWork.Enabled {
		return
	}
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
		acquired, err := s.store.AcquireInboundDispatcherLease(ctx, leaseOwner, leaseTTL)
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
			if err := s.store.ReleaseInboundDispatcherLease(context.Background(), leaseOwner); err != nil {
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
}

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
	s.mux.HandleFunc("/api/auth/status", s.handleAuthStatus)
	s.mux.HandleFunc("/api/auth/enroll/start", s.handleAuthEnrollStart)
	s.mux.HandleFunc("/api/auth/enroll/verify", s.handleAuthEnrollVerify)
	s.mux.HandleFunc("/api/auth/totp/verify", s.handleAuthTOTPVerify)

	guard := s.withAuth
	s.mux.HandleFunc("/api/runtime/config", guard(s.handleRuntimeConfig))
	s.mux.HandleFunc("/api/runtime/inbound-work", guard(s.handleRuntimeInboundWork))
	s.mux.HandleFunc("/api/runtime/inbound-work/drain", guard(s.handleRuntimeInboundWorkDrain))
	s.mux.HandleFunc("/api/runtime/inbound-work/requeue", guard(s.handleRuntimeInboundWorkRequeue))
	s.mux.HandleFunc("/api/runtime/inbound-work/discard", guard(s.handleRuntimeInboundWorkDiscard))
	s.mux.HandleFunc("/api/frontend/log", guard(s.handleFrontendLog))
	s.mux.HandleFunc("/api/workspace/tree", guard(s.handleWorkspaceTree))
	s.mux.HandleFunc("/api/workspace/file", guard(s.handleWorkspaceFile))
	s.mux.HandleFunc("/sse/stream", guard(s.handleSSEStream))
	s.mux.HandleFunc("/api/system-metrics", guard(s.handleSystemMetrics))
	s.mux.HandleFunc("/agent/system-metrics", guard(s.handleSystemMetrics))
	s.mux.HandleFunc("/api/tools", guard(s.handleTools))
	s.mux.HandleFunc("/api/tools/execute", guard(s.handleToolExecute))
	s.mux.HandleFunc(connectivityRoutePrefix, guard(s.handleConnectivityRoutes))
	s.mux.HandleFunc(connectivitySSEPrefix, guard(s.handleConnectivitySSE))
	s.mux.HandleFunc("/api/sessions", guard(s.handleSessions))
	s.mux.HandleFunc("/api/sessions/", guard(s.handleSessionSubroutes))
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
	case "messages":
		s.handleMessages(w, r, sessionID)
	case "prompt":
		s.handlePrompt(w, r, sessionID)
	case "turns":
		s.handleTurns(w, r, sessionID)
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

func (s *Server) handleSessionRouteEvents(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	events, err := s.store.ListRouteEvents(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if events == nil {
		events = []store.RouteEvent{}
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
	routeEvents, err := s.store.ListRouteEvents(ctx, sessionID)
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
		Prompt        string `json:"prompt"`
		Intent        string `json:"intent"`
		Model         string `json:"model"`
		TargetAgentID string `json:"target_agent_id"`
		ParentTurnID  string `json:"parent_turn_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	model := req.Model
	if model == "" {
		model = s.cfg.DefaultModel
	}
	var (
		result *turn.SubmitResult
		err    error
	)
	if req.TargetAgentID != "" && req.TargetAgentID != "default" {
		result, err = s.turns.SubmitPeerMessage(r.Context(), sessionID, req.TargetAgentID, req.Prompt, req.Intent, model, req.ParentTurnID)
	} else {
		result, err = s.turns.SubmitPromptRouted(r.Context(), turn.RunInput{SessionID: sessionID, Prompt: req.Prompt, Intent: req.Intent, Model: model, ParentTurnID: req.ParentTurnID})
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
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

func sessionAgentID(sess *store.Session, identities map[string]string) string {
	if sess != nil {
		if agentID := strings.TrimSpace(identities[sess.ID]); agentID != "" {
			return agentID
		}
		if sess.Scope != nil && sess.Scope.AgentID != "" {
			return sess.Scope.AgentID
		}
	}
	return "agent"
}

func (s *Server) nextForkAgentID(ctx context.Context, sourceSessionID string) (string, error) {
	source, err := s.store.GetSession(ctx, sourceSessionID)
	if err != nil {
		return "", err
	}
	identityRows, err := s.store.ListSessionIdentities(ctx)
	if err != nil {
		return "", err
	}
	identities := map[string]string{}
	for _, identity := range identityRows {
		if strings.TrimSpace(identity.Scope.AgentID) != "" {
			identities[identity.SessionID] = identity.Scope.AgentID
		}
	}
	base := sessionAgentID(source, identities)
	base = strings.TrimRightFunc(base, func(r rune) bool { return r >= '0' && r <= '9' })
	if base == "" {
		base = sessionAgentID(source, identities)
	}
	sessions, err := s.store.ListSessions(ctx)
	if err != nil {
		return "", err
	}
	used := map[string]bool{}
	for i := range sessions {
		used[sessionAgentID(&sessions[i], identities)] = true
	}
	for i := 1; i < 1000; i++ {
		candidate := base + strconv.Itoa(i)
		if !used[candidate] {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate fork agent id from %s", base)
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
		items, err := s.store.ListInboundWorkFiltered(r.Context(), status, limit, eligible)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		counts, err := s.store.CountInboundWorkByStatus(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		eligibleCount, err := s.store.CountEligibleInboundWork(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		if items == nil {
			items = []store.InboundWorkItem{}
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
		items = []*store.InboundWorkItem{}
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
	item, err := s.store.RequeueInboundWork(r.Context(), req.ID, req.ResetAttempts)
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
	item, err := s.store.DiscardInboundWork(r.Context(), req.ID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	s.turns.PublishRuntimeInboundWorkEvent("inbound_work_discarded", item, nil)
	writeJSON(w, http.StatusOK, map[string]any{"item": item})
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

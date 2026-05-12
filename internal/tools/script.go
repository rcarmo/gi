package tools

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	xwebsocket "golang.org/x/net/websocket"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/connectivity"
	"github.com/rcarmo/gi/internal/scripting"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
)

// ScriptTool is a tool that executes Joker/Clojure scripts with access
// to gi's bridge state. It can run inline scripts or script files from
// the workspace.
type ScriptTool struct {
	store *store.Store
	cfg   config.RuntimeConfig
	joker scripting.Runner
	js    scripting.Runner

	rawSocketMu sync.Mutex
	rawSockets  map[string]net.Conn
	webSocketMu sync.Mutex
	webSockets  map[string]*xwebsocket.Conn
	topicMu     sync.Mutex
	topicSubs   map[string]topicSubscription
	hookMu      sync.Mutex
	eventHooks  []scripting.EventHookSpec
	ioIDCounter atomic.Int64
	httpClient  *http.Client

	onRegisterEventHook func(ctx context.Context, sessionID string, hook scripting.EventHookSpec) error
	onRegisterTool      func(ctx context.Context, sessionID string, spec scripting.ToolSpec) error
	onSetActiveTools    func(ctx context.Context, sessionID string, names []string) error
	onGetActiveTools    func(ctx context.Context, sessionID string) ([]string, error)
	onRegisterRoute     func(ctx context.Context, sessionID string, route connectivity.RouteSpec) (connectivity.RouteInfo, error)
	onUnregisterRoute   func(ctx context.Context, sessionID string, id string) error
	onListRoutes        func(ctx context.Context, sessionID string, filter map[string]any) ([]connectivity.RouteInfo, error)
	onEmitConnectEvent  func(ctx context.Context, sessionID string, topic string, payload map[string]any) error
	onPublishTopic      func(ctx context.Context, sessionID string, envelope map[string]any) error
	onSubscribeTopic    func(ctx context.Context, sessionID string, pattern string, opts scripting.TopicSubscribeOptions) (<-chan topics.Envelope, func(), error)
}

type topicSubscription struct {
	ch          <-chan topics.Envelope
	unsubscribe func()
}

// ScriptInput is what the agent sends to invoke the script tool.
type ScriptInput struct {
	Script    string `json:"script,omitempty"`
	Path      string `json:"path,omitempty"`
	Engine    string `json:"engine,omitempty"` // "joker" or "js" (default: auto-detect)
	SessionID string `json:"session_id,omitempty"`
}

// ScriptOutput is what the script tool returns.
type ScriptOutput struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

func NewScriptTool(s *store.Store, cfg config.RuntimeConfig) *ScriptTool {
	return &ScriptTool{
		store:      s,
		cfg:        cfg,
		joker:      scripting.NewJokerRunner(),
		js:         scripting.NewGojaRunner(),
		rawSockets: make(map[string]net.Conn),
		webSockets: make(map[string]*xwebsocket.Conn),
		topicSubs:  make(map[string]topicSubscription),
		httpClient: &http.Client{},
	}
}

// SetAgenticCallbacks connects script-declared tools/hooks to the host engine.
func (t *ScriptTool) SetAgenticCallbacks(
	registerHook func(context.Context, string, scripting.EventHookSpec) error,
	registerTool func(context.Context, string, scripting.ToolSpec) error,
	setActiveTools func(context.Context, string, []string) error,
	getActiveTools func(context.Context, string) ([]string, error),
) {
	t.onRegisterEventHook = registerHook
	t.onRegisterTool = registerTool
	t.onSetActiveTools = setActiveTools
	t.onGetActiveTools = getActiveTools
}

// SetConnectivityCallbacks connects scripts to the host connectivity registry.
func (t *ScriptTool) SetConnectivityCallbacks(
	registerRoute func(context.Context, string, connectivity.RouteSpec) (connectivity.RouteInfo, error),
	unregisterRoute func(context.Context, string, string) error,
	listRoutes func(context.Context, string, map[string]any) ([]connectivity.RouteInfo, error),
	emitEvent func(context.Context, string, string, map[string]any) error,
	publishTopic func(context.Context, string, map[string]any) error,
	subscribeTopic func(context.Context, string, string, scripting.TopicSubscribeOptions) (<-chan topics.Envelope, func(), error),
) {
	t.onRegisterRoute = registerRoute
	t.onUnregisterRoute = unregisterRoute
	t.onListRoutes = listRoutes
	t.onEmitConnectEvent = emitEvent
	t.onPublishTopic = publishTopic
	t.onSubscribeTopic = subscribeTopic
}

// Definition returns the tool metadata for the agent.
func (t *ScriptTool) Definition() map[string]any {
	return map[string]any{
		"name":        "script",
		"description": "Execute a script with access to gi's live session state, session info, turns, config, and workspace files. Supports Joker/Clojure (.joke, .clj, baked into the gi binary) and JavaScript (.js). The gi bridge provides sessionId, getSessionInfo(), getSessionState(), setSessionState(), listTurns(), getRuntimeConfig(), listMessages(), readFile(), writeFile(), listDir(), and log().",
		"parameters": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"script": map[string]any{
					"type":        "string",
					"description": "Inline script to execute",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "Workspace-relative path or vfs://namespace/path to a script file",
				},
				"engine": map[string]any{
					"type":        "string",
					"description": "Script engine: 'js' (JavaScript/goja, compiled-in) or 'joker' (Clojure, baked into gi). Default: js for .js files, joker for .joke/.clj, js for inline.",
					"enum":        []string{"js", "joker"},
				},
			},
		},
	}
}

// Execute runs the script and returns the output.
func (t *ScriptTool) Execute(ctx context.Context, input ScriptInput) ScriptOutput {
	bridge := t.buildBridge(input.SessionID)
	runner := t.resolveRunner(input.Engine, input.Path)

	var result string
	var err error

	if input.Path != "" {
		resolved, err := ResolveToolPath(t.cfg.WorkspaceRoot, input.Path, false)
		if err != nil {
			return ScriptOutput{Error: err.Error()}
		}
		if resolved.IsVFS() {
			_, raw, err := t.store.GetVFSFileContent(ctx, resolved.VFSNamespace, resolved.VFSPath)
			if err != nil {
				return ScriptOutput{Error: err.Error()}
			}
			log.Printf("script[%s]: executing vfs file %s/%s", runner.Name(), resolved.VFSNamespace, resolved.VFSPath)
			result, err = runner.Execute(ctx, string(raw), bridge)
		} else {
			content, err := os.ReadFile(resolved.WorkspacePath)
			if err != nil {
				return ScriptOutput{Error: err.Error()}
			}
			log.Printf("script[%s]: executing file %s", runner.Name(), input.Path)
			result, err = runner.Execute(ctx, string(content), bridge)
		}
	} else if input.Script != "" {
		log.Printf("script[%s]: executing inline (%d chars)", runner.Name(), len(input.Script))
		result, err = runner.Execute(ctx, input.Script, bridge)
	} else {
		return ScriptOutput{Error: "either script or path is required"}
	}

	if err != nil {
		log.Printf("script[%s]: error: %v", runner.Name(), err)
		return ScriptOutput{Result: result, Error: err.Error()}
	}

	log.Printf("script[%s]: result (%d chars)", runner.Name(), len(result))
	return ScriptOutput{Result: result}
}

func (t *ScriptTool) resolveRunner(engine, path string) scripting.Runner {
	if engine == "joker" {
		return t.joker
	}
	if engine == "js" || engine == "javascript" {
		return t.js
	}
	// Auto-detect from file extension
	if strings.HasSuffix(path, ".joke") || strings.HasSuffix(path, ".clj") {
		return t.joker
	}
	if strings.HasSuffix(path, ".js") {
		return t.js
	}
	// Default: JS (compiled-in, no external dependency)
	return t.js
}

func (t *ScriptTool) getSessionInfo(ctx context.Context, sessionID string) (map[string]any, error) {
	session, err := t.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	messages, err := t.store.ListMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	turns, err := t.store.ListTurns(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	cfg, err := json.Marshal(t.cfg)
	if err != nil {
		return nil, err
	}
	cfgMap := map[string]any{}
	if err := json.Unmarshal(cfg, &cfgMap); err != nil {
		cfgMap = map[string]any{"raw": string(cfg), "decode_error": err.Error()}
	}

	info := map[string]any{
		"session":       session,
		"config":        cfgMap,
		"message_count": len(messages),
		"turn_count":    len(turns),
		"messages":      []map[string]any{},
		"turns":         []map[string]any{},
	}
	for _, m := range messages {
		info["messages"] = append(info["messages"].([]map[string]any), map[string]any{
			"id":         m.ID,
			"role":       m.Role,
			"content":    m.Content,
			"payload":    m.Payload,
			"created_at": m.CreatedAt,
		})
	}
	for _, tr := range turns {
		info["turns"] = append(info["turns"].([]map[string]any), map[string]any{
			"id":         tr.ID,
			"status":     tr.Status,
			"prompt":     tr.Prompt,
			"metadata":   tr.Metadata,
			"created_at": tr.CreatedAt,
			"updated_at": tr.UpdatedAt,
		})
	}
	return info, nil
}

func inferContentTypeFromFilename(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if extType := mime.TypeByExtension(ext); extType != "" {
		return extType
	}
	return "text/plain"
}

func (t *ScriptTool) buildBridge(sessionID string) *scripting.Bridge {
	return scripting.NewBridge(sessionID, scripting.BridgeFuncs{
		GetSessionState: func(ctx context.Context) (map[string]any, error) {
			session, err := t.store.GetSession(ctx, sessionID)
			if err != nil {
				return nil, err
			}
			return session.State, nil
		},
		GetSessionInfo: func(ctx context.Context) (map[string]any, error) {
			return t.getSessionInfo(ctx, sessionID)
		},
		SetSessionState: func(ctx context.Context, patch map[string]any) error {
			return t.store.TouchSessionState(ctx, sessionID, patch)
		},
		ListMessages: func(ctx context.Context, limit int) ([]map[string]any, error) {
			msgs, err := t.store.ListMessages(ctx, sessionID)
			if err != nil {
				return nil, err
			}
			var result []map[string]any
			start := 0
			if limit > 0 && len(msgs) > limit {
				start = len(msgs) - limit
			}
			for _, m := range msgs[start:] {
				result = append(result, map[string]any{
					"id":      m.ID,
					"role":    m.Role,
					"content": m.Content,
					"payload": m.Payload,
				})
			}
			return result, nil
		},
		AddMessage: func(ctx context.Context, role, content string) error {
			return t.store.AddMessage(ctx, store.NowID("msg"), sessionID, role, content, map[string]any{"kind": "script"})
		},
		ListTurns: func(ctx context.Context, limit int) ([]map[string]any, error) {
			turns, err := t.store.ListTurns(ctx, sessionID)
			if err != nil {
				return nil, err
			}
			start := 0
			if limit > 0 && len(turns) > limit {
				start = len(turns) - limit
			}
			result := make([]map[string]any, 0, len(turns[start:]))
			for _, tr := range turns[start:] {
				result = append(result, map[string]any{
					"id":         tr.ID,
					"status":     tr.Status,
					"prompt":     tr.Prompt,
					"metadata":   tr.Metadata,
					"created_at": tr.CreatedAt,
					"updated_at": tr.UpdatedAt,
				})
			}
			return result, nil
		},
		ListTurnEvents: func(ctx context.Context, turnID string) ([]map[string]any, error) {
			events, err := t.store.ListTurnEvents(ctx, turnID)
			if err != nil {
				return nil, err
			}
			var result []map[string]any
			for _, e := range events {
				result = append(result, map[string]any{
					"seq":     e.Seq,
					"type":    e.Type,
					"payload": e.Payload,
				})
			}
			return result, nil
		},
		GetConfig: func(ctx context.Context) (map[string]any, error) {
			b, _ := json.Marshal(t.cfg)
			var m map[string]any
			json.Unmarshal(b, &m)
			return m, nil
		},
		ReadFile: func(ctx context.Context, path string) (string, error) {
			resolved, err := ResolveToolPath(t.cfg.WorkspaceRoot, path, false)
			if err != nil {
				return "", err
			}
			if resolved.IsVFS() {
				if resolved.VFSNamespace == "fts" {
					return ReadFTSQuery(ctx, t.cfg.WorkspaceRoot, t.store, resolved.VFSPath)
				}
				_, raw, err := t.store.GetVFSFileContent(ctx, resolved.VFSNamespace, resolved.VFSPath)
				if err != nil {
					return "", err
				}
				return string(raw), nil
			}
			data, err := os.ReadFile(resolved.WorkspacePath)
			if err != nil {
				return "", err
			}
			return string(data), nil
		},
		WriteFile: func(ctx context.Context, path, content string) error {
			resolved, err := ResolveToolPath(t.cfg.WorkspaceRoot, path, true)
			if err != nil {
				return err
			}
			if resolved.IsVFS() {
				_, err := t.store.SaveVFSFile(ctx, resolved.VFSNamespace, resolved.VFSPath, inferContentTypeFromFilename(resolved.VFSPath), []byte(content), map[string]any{})
				return err
			}
			dir := filepath.Dir(resolved.WorkspacePath)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			return os.WriteFile(resolved.WorkspacePath, []byte(content), 0o644)
		},
		ListDir: func(ctx context.Context, path string) ([]map[string]any, error) {
			resolved, err := ResolveToolPath(t.cfg.WorkspaceRoot, path, false)
			if err != nil {
				return nil, err
			}
			if resolved.IsVFS() {
				entries, err := t.store.ListVFSChildren(ctx, resolved.VFSNamespace, resolved.VFSPath)
				if err != nil {
					return nil, err
				}
				result := make([]map[string]any, 0, len(entries))
				for _, e := range entries {
					result = append(result, map[string]any{
						"name":        e.Name,
						"isDir":       e.IsDir,
						"path":        e.Path,
						"contentType": e.ContentType,
						"size":        e.OriginalSize,
						"metadata":    e.Metadata,
					})
				}
				return result, nil
			}
			entries, err := os.ReadDir(resolved.WorkspacePath)
			if err != nil {
				return nil, err
			}
			var result []map[string]any
			for _, e := range entries {
				info, _ := e.Info()
				result = append(result, map[string]any{
					"name":  e.Name(),
					"isDir": e.IsDir(),
					"size":  info.Size(),
				})
			}
			return result, nil
		},
		Exec: func(ctx context.Context, command string) (string, error) {
			return "", fmt.Errorf("exec: not yet available in script bridge")
		},
		RegisterEventHook: func(ctx context.Context, hook scripting.EventHookSpec) error {
			if t.onRegisterEventHook != nil {
				return t.onRegisterEventHook(ctx, sessionID, hook)
			}
			return t.registerEventHook(ctx, sessionID, hook)
		},
		RegisterTool: func(ctx context.Context, spec scripting.ToolSpec) error {
			if t.onRegisterTool == nil {
				return fmt.Errorf("register tool: host engine callback is not available")
			}
			return t.onRegisterTool(ctx, sessionID, spec)
		},
		SetActiveTools: func(ctx context.Context, names []string) error {
			if t.onSetActiveTools == nil {
				return fmt.Errorf("set active tools: host engine callback is not available")
			}
			return t.onSetActiveTools(ctx, sessionID, names)
		},
		GetActiveTools: func(ctx context.Context) ([]string, error) {
			if t.onGetActiveTools == nil {
				return nil, fmt.Errorf("get active tools: host engine callback is not available")
			}
			return t.onGetActiveTools(ctx, sessionID)
		},
		SetModel: func(ctx context.Context, model string) error {
			return t.store.TouchSessionState(ctx, sessionID, map[string]any{"model": model})
		},
		AppendEntry: func(ctx context.Context, entryType string, data map[string]any) error {
			return t.store.AddMessage(ctx, store.NowID("msg"), sessionID, "system", "", map[string]any{"kind": "script_entry", "entry_type": entryType, "data": data, "clipped": true})
		},
		GetEntries: func(ctx context.Context, entryType string) ([]map[string]any, error) {
			msgs, err := t.store.ListMessages(ctx, sessionID)
			if err != nil {
				return nil, err
			}
			var entries []map[string]any
			for _, m := range msgs {
				if m.Payload["kind"] != "script_entry" {
					continue
				}
				if entryType != "" && m.Payload["entry_type"] != entryType {
					continue
				}
				entries = append(entries, map[string]any{"id": m.ID, "type": m.Payload["entry_type"], "data": m.Payload["data"], "created_at": m.CreatedAt})
			}
			return entries, nil
		},
		EmitEvent: func(ctx context.Context, name string, payload map[string]any) error {
			return t.emitEvent(ctx, sessionID, name, payload)
		},
		ClearEventHooks: func(ctx context.Context) error {
			return t.clearEventHooks(ctx, sessionID)
		},
		PublishTopic: func(ctx context.Context, envelope map[string]any) error {
			if t.onPublishTopic == nil {
				return fmt.Errorf("publish topic: host topic bus is not available")
			}
			return t.onPublishTopic(ctx, sessionID, envelope)
		},
		SubscribeTopic: func(ctx context.Context, pattern string, opts scripting.TopicSubscribeOptions) (string, error) {
			return t.subscribeTopic(ctx, sessionID, pattern, opts)
		},
		ReadTopicSubscription: func(ctx context.Context, id string, limit int) ([]map[string]any, error) {
			return t.readTopicSubscription(ctx, id, limit)
		},
		UnsubscribeTopic: func(ctx context.Context, id string) error {
			return t.unsubscribeTopic(ctx, id)
		},
		RegisterConnectivityRoute: func(ctx context.Context, route connectivity.RouteSpec) (connectivity.RouteInfo, error) {
			if t.onRegisterRoute == nil {
				return connectivity.RouteInfo{}, fmt.Errorf("register route: host connectivity registry is not available")
			}
			return t.onRegisterRoute(ctx, sessionID, route)
		},
		UnregisterConnectivityRoute: func(ctx context.Context, id string) error {
			if t.onUnregisterRoute == nil {
				return fmt.Errorf("unregister route: host connectivity registry is not available")
			}
			return t.onUnregisterRoute(ctx, sessionID, id)
		},
		ListConnectivityRoutes: func(ctx context.Context, filter map[string]any) ([]connectivity.RouteInfo, error) {
			if t.onListRoutes == nil {
				return nil, fmt.Errorf("list routes: host connectivity registry is not available")
			}
			return t.onListRoutes(ctx, sessionID, filter)
		},
		EmitConnectivityEvent: func(ctx context.Context, topic string, payload map[string]any) error {
			if t.onEmitConnectEvent == nil {
				return fmt.Errorf("emit connectivity event: host connectivity registry is not available")
			}
			return t.onEmitConnectEvent(ctx, sessionID, topic, payload)
		},
		OpenRawSocket: func(ctx context.Context, spec scripting.RawSocketSpec) (string, error) {
			return t.openRawSocket(ctx, spec)
		},
		WriteRawSocket: func(ctx context.Context, payload scripting.RawSocketPayload) (int, error) {
			return t.writeRawSocket(ctx, payload)
		},
		ReadRawSocket: func(ctx context.Context, payload scripting.RawSocketPayload) (string, error) {
			return t.readRawSocket(ctx, payload)
		},
		CloseRawSocket: func(ctx context.Context, socketID string) error {
			return t.closeRawSocket(ctx, socketID)
		},
		OpenWebSocket: func(ctx context.Context, spec scripting.WebSocketSpec) (string, error) {
			return t.openWebSocket(ctx, spec)
		},
		WriteWebSocket: func(ctx context.Context, socketID string, payload string) error {
			return t.writeWebSocket(ctx, socketID, payload)
		},
		ReadWebSocket: func(ctx context.Context, socketID string, timeoutMS int) (string, error) {
			return t.readWebSocket(ctx, socketID, timeoutMS)
		},
		CloseWebSocket: func(ctx context.Context, socketID string) error {
			return t.closeWebSocket(ctx, socketID)
		},
		DoHTTPRequest: func(ctx context.Context, req scripting.HTTPCallSpec) (scripting.HTTPResponse, error) {
			return t.doHTTPRequest(ctx, req)
		},
		Log: func(ctx context.Context, level, message string) {
			log.Printf("script[%s]: %s", level, message)
		},
	})
}

func (t *ScriptTool) registerEventHook(_ context.Context, sessionID string, hook scripting.EventHookSpec) error {
	t.hookMu.Lock()
	defer t.hookMu.Unlock()
	if hook.Name == "" {
		return fmt.Errorf("event hook name is required")
	}
	t.eventHooks = append(t.eventHooks, hook)
	log.Printf("script[%s]: registerEventHook name=%q source=%q", sessionID, hook.Name, hook.Source)
	return nil
}

func (t *ScriptTool) emitEvent(_ context.Context, sessionID, name string, payload map[string]any) error {
	t.hookMu.Lock()
	defer t.hookMu.Unlock()
	matched := 0
	for _, hook := range t.eventHooks {
		if hook.Name != name && hook.Name != "*" {
			continue
		}
		if hook.Source != "" {
			if payload == nil {
				continue
			}
			payloadValue, ok := payload["source"]
			if !ok {
				continue
			}
			src, ok := payloadValue.(string)
			if !ok || src != hook.Source {
				continue
			}
		}
		matched++
	}
	log.Printf("script[%s]: emitEvent name=%q matched=%d payload=%#v", sessionID, name, matched, payload)
	return nil
}

func (t *ScriptTool) clearEventHooks(_ context.Context, sessionID string) error {
	t.hookMu.Lock()
	defer t.hookMu.Unlock()
	t.eventHooks = nil
	log.Printf("script[%s]: clearEventHooks", sessionID)
	return nil
}

func (t *ScriptTool) subscribeTopic(ctx context.Context, sessionID, pattern string, opts scripting.TopicSubscribeOptions) (string, error) {
	if t.onSubscribeTopic == nil {
		return "", fmt.Errorf("subscribe topic: host topic bus is not available")
	}
	if strings.TrimSpace(pattern) == "" {
		pattern = "*"
	}
	if strings.TrimSpace(opts.SessionID) == "" {
		opts.SessionID = sessionID
	}
	ch, unsubscribe, err := t.onSubscribeTopic(ctx, sessionID, pattern, opts)
	if err != nil {
		return "", err
	}
	id := t.socketID("topic")
	t.topicMu.Lock()
	t.topicSubs[id] = topicSubscription{ch: ch, unsubscribe: unsubscribe}
	t.topicMu.Unlock()
	return id, nil
}

func (t *ScriptTool) readTopicSubscription(_ context.Context, id string, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 50
	}
	t.topicMu.Lock()
	sub, ok := t.topicSubs[id]
	t.topicMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("read topic subscription: unknown id %q", id)
	}
	out := make([]map[string]any, 0, limit)
	for len(out) < limit {
		select {
		case env, ok := <-sub.ch:
			if !ok {
				return out, nil
			}
			out = append(out, map[string]any{
				"topic":      env.Topic,
				"session_id": env.SessionID,
				"agent_id":   env.AgentID,
				"source":     env.Source,
				"type":       env.Type,
				"payload":    env.Payload,
				"timestamp":  env.Timestamp.Format(time.RFC3339Nano),
			})
		default:
			return out, nil
		}
	}
	return out, nil
}

func (t *ScriptTool) unsubscribeTopic(_ context.Context, id string) error {
	t.topicMu.Lock()
	sub, ok := t.topicSubs[id]
	if ok {
		delete(t.topicSubs, id)
	}
	t.topicMu.Unlock()
	if !ok {
		return fmt.Errorf("unsubscribe topic: unknown id %q", id)
	}
	if sub.unsubscribe != nil {
		sub.unsubscribe()
	}
	return nil
}

func (t *ScriptTool) socketID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, t.ioIDCounter.Add(1))
}

func (t *ScriptTool) openRawSocket(ctx context.Context, spec scripting.RawSocketSpec) (string, error) {
	proto := strings.ToLower(strings.TrimSpace(spec.Protocol))
	if proto == "" {
		proto = "tcp"
	}
	if proto != "tcp" && proto != "udp" {
		return "", fmt.Errorf("unsupported protocol: %s", spec.Protocol)
	}
	address := strings.TrimSpace(spec.Address)
	if address == "" {
		return "", fmt.Errorf("socket address is required")
	}
	localAddr := strings.TrimSpace(spec.LocalAddr)
	var dialer net.Dialer
	if localAddr != "" {
		var addr net.Addr
		var err error
		if proto == "tcp" {
			addr, err = net.ResolveTCPAddr(proto, localAddr)
			if err != nil {
				return "", fmt.Errorf("resolve local addr: %w", err)
			}
		} else {
			addr, err = net.ResolveUDPAddr(proto, localAddr)
			if err != nil {
				return "", fmt.Errorf("resolve local addr: %w", err)
			}
		}
		dialer.LocalAddr = addr
	}
	if spec.TimeoutMS > 0 {
		dialer.Timeout = time.Duration(spec.TimeoutMS) * time.Millisecond
	}
	deadlineCtx := ctx
	if spec.TimeoutMS > 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(spec.TimeoutMS)*time.Millisecond)
		defer cancel()
		deadlineCtx = ctxWithTimeout
	}
	conn, err := dialer.DialContext(deadlineCtx, proto, address)
	if err != nil {
		return "", fmt.Errorf("open socket: %w", err)
	}
	id := t.socketID("raw")
	t.rawSocketMu.Lock()
	t.rawSockets[id] = conn
	t.rawSocketMu.Unlock()
	return id, nil
}

func (t *ScriptTool) getRawSocket(socketID string) (net.Conn, error) {
	t.rawSocketMu.Lock()
	defer t.rawSocketMu.Unlock()
	conn, ok := t.rawSockets[socketID]
	if !ok {
		return nil, fmt.Errorf("socket not found")
	}
	return conn, nil
}

func (t *ScriptTool) writeRawSocket(_ context.Context, payload scripting.RawSocketPayload) (int, error) {
	if strings.TrimSpace(payload.SocketID) == "" {
		return 0, fmt.Errorf("socket_id is required")
	}
	conn, err := t.getRawSocket(payload.SocketID)
	if err != nil {
		return 0, err
	}
	if payload.TimeoutMS > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(time.Duration(payload.TimeoutMS) * time.Millisecond)); err != nil {
			return 0, fmt.Errorf("set write deadline: %w", err)
		}
		defer conn.SetWriteDeadline(time.Time{})
	}
	n, err := conn.Write([]byte(payload.Data))
	if err != nil {
		return n, fmt.Errorf("write socket: %w", err)
	}
	return n, nil
}

func (t *ScriptTool) readRawSocket(_ context.Context, payload scripting.RawSocketPayload) (string, error) {
	if strings.TrimSpace(payload.SocketID) == "" {
		return "", fmt.Errorf("socket_id is required")
	}
	conn, err := t.getRawSocket(payload.SocketID)
	if err != nil {
		return "", err
	}
	maxBytes := payload.MaxBytes
	if maxBytes <= 0 {
		maxBytes = 4096
	}
	buf := make([]byte, maxBytes)
	if payload.TimeoutMS > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(time.Duration(payload.TimeoutMS) * time.Millisecond)); err != nil {
			return "", fmt.Errorf("set read deadline: %w", err)
		}
		defer conn.SetReadDeadline(time.Time{})
	}
	n, err := conn.Read(buf)
	if n > 0 {
		return string(buf[:n]), nil
	}
	if err != nil {
		return "", fmt.Errorf("read socket: %w", err)
	}
	return "", nil
}

func (t *ScriptTool) closeRawSocket(_ context.Context, socketID string) error {
	if strings.TrimSpace(socketID) == "" {
		return fmt.Errorf("socket_id is required")
	}
	t.rawSocketMu.Lock()
	conn, ok := t.rawSockets[socketID]
	if ok {
		delete(t.rawSockets, socketID)
	}
	t.rawSocketMu.Unlock()
	if !ok {
		return fmt.Errorf("socket not found")
	}
	return conn.Close()
}

func (t *ScriptTool) openWebSocket(ctx context.Context, spec scripting.WebSocketSpec) (string, error) {
	wsURL := strings.TrimSpace(spec.URL)
	if wsURL == "" {
		return "", fmt.Errorf("url is required")
	}
	cfg, err := xwebsocket.NewConfig(wsURL, "http://localhost")
	if err != nil {
		return "", fmt.Errorf("websocket config: %w", err)
	}
	if spec.Subprotocol != "" {
		cfg.Protocol = []string{spec.Subprotocol}
	}
	if cfg.Header == nil {
		cfg.Header = make(http.Header)
	}
	for name, values := range spec.Headers {
		for _, value := range values {
			cfg.Header.Add(name, value)
		}
	}
	if spec.TimeoutMS > 0 {
		cfg.Dialer = &net.Dialer{Timeout: time.Duration(spec.TimeoutMS) * time.Millisecond}
	}
	if cfg.Dialer == nil {
		cfg.Dialer = &net.Dialer{}
	}
	deadlineCtx := ctx
	if spec.TimeoutMS > 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(spec.TimeoutMS)*time.Millisecond)
		defer cancel()
		deadlineCtx = ctxWithTimeout
	}
	conn, err := cfg.DialContext(deadlineCtx)
	if err != nil {
		return "", fmt.Errorf("open websocket: %w", err)
	}
	id := t.socketID("ws")
	t.webSocketMu.Lock()
	t.webSockets[id] = conn
	t.webSocketMu.Unlock()
	return id, nil
}

func (t *ScriptTool) getWebSocket(socketID string) (*xwebsocket.Conn, error) {
	t.webSocketMu.Lock()
	defer t.webSocketMu.Unlock()
	conn, ok := t.webSockets[socketID]
	if !ok {
		return nil, fmt.Errorf("websocket not found")
	}
	return conn, nil
}

func (t *ScriptTool) writeWebSocket(_ context.Context, socketID string, payload string) error {
	if strings.TrimSpace(socketID) == "" {
		return fmt.Errorf("socket_id is required")
	}
	conn, err := t.getWebSocket(socketID)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte(payload))
	if err != nil {
		return fmt.Errorf("write websocket: %w", err)
	}
	return nil
}

func (t *ScriptTool) readWebSocket(ctx context.Context, socketID string, timeoutMS int) (string, error) {
	if strings.TrimSpace(socketID) == "" {
		return "", fmt.Errorf("socket_id is required")
	}
	conn, err := t.getWebSocket(socketID)
	if err != nil {
		return "", err
	}
	if timeoutMS > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMS) * time.Millisecond)); err != nil {
			return "", fmt.Errorf("set websocket read deadline: %w", err)
		}
		defer conn.SetReadDeadline(time.Time{})
	}
	if deadline, ok := ctx.Deadline(); ok && timeoutMS <= 0 {
		conn.SetReadDeadline(deadline)
		defer conn.SetReadDeadline(time.Time{})
	}
	buf := make([]byte, 65536)
	n, err := conn.Read(buf)
	if n > 0 {
		return string(buf[:n]), nil
	}
	if err != nil {
		return "", fmt.Errorf("read websocket: %w", err)
	}
	return "", nil
}

func (t *ScriptTool) closeWebSocket(_ context.Context, socketID string) error {
	if strings.TrimSpace(socketID) == "" {
		return fmt.Errorf("socket_id is required")
	}
	t.webSocketMu.Lock()
	conn, ok := t.webSockets[socketID]
	if ok {
		delete(t.webSockets, socketID)
	}
	t.webSocketMu.Unlock()
	if !ok {
		return fmt.Errorf("websocket not found")
	}
	return conn.Close()
}

func (t *ScriptTool) doHTTPRequest(ctx context.Context, req scripting.HTTPCallSpec) (scripting.HTTPResponse, error) {
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = "GET"
	}
	url := strings.TrimSpace(req.URL)
	if url == "" {
		return scripting.HTTPResponse{}, fmt.Errorf("url is required")
	}
	body := strings.NewReader(req.Body)
	httpReq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return scripting.HTTPResponse{}, fmt.Errorf("build request: %w", err)
	}
	for name, values := range req.Headers {
		for _, value := range values {
			httpReq.Header.Add(name, value)
		}
	}
	client := t.httpClient
	if req.TimeoutMS > 0 {
		client = &http.Client{Timeout: time.Duration(req.TimeoutMS) * time.Millisecond}
	}
	if req.SkipTLS {
		client = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
		if client.Timeout == 0 && req.TimeoutMS > 0 {
			client.Timeout = time.Duration(req.TimeoutMS) * time.Millisecond
		}
	}
	if !req.AllowRedirects {
		client = &http.Client{
			Transport: client.Transport,
			Timeout:   client.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	if client == nil {
		client = &http.Client{}
	}
	attempts := 1
	if req.Retry > 0 {
		attempts += req.Retry
	}
	var resp *http.Response
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			body := strings.NewReader(req.Body)
			httpReq, err = http.NewRequestWithContext(ctx, method, url, body)
			if err != nil {
				return scripting.HTTPResponse{}, fmt.Errorf("build request: %w", err)
			}
			for name, values := range req.Headers {
				for _, value := range values {
					httpReq.Header.Add(name, value)
				}
			}
		}

		resp, err = client.Do(httpReq)
		if err == nil {
			break
		}
		if attempt == attempts-1 {
			return scripting.HTTPResponse{}, fmt.Errorf("request: %w", err)
		}
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return scripting.HTTPResponse{}, fmt.Errorf("read response: %w", err)
	}
	return scripting.HTTPResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Body:       string(raw),
		URL:        resp.Request.URL.String(),
	}, nil
}

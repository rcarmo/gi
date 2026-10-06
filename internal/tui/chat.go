package tui

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/shellenv"
	"github.com/rcarmo/gi/internal/skills"
	"github.com/rcarmo/gi/internal/store"
	gitools "github.com/rcarmo/gi/internal/tools"
	"github.com/rcarmo/gi/internal/topics"
	"github.com/rcarmo/gi/internal/turn"
)

const defaultTUIAgentID = "agent"

func initialSessionID(ctx context.Context, s *store.Store) (string, error) {
	if sessionID, err := s.ResolveMainSessionID(ctx, defaultTUIAgentID, "gi", "default"); err == nil {
		return sessionID, nil
	}
	return "", nil
}

func Run(dbPath, workspace, model string) error {
	return RunMode(dbPath, workspace, model, "")
}

func RunMode(dbPath, workspace, model, mode string) error {
	if mode != "" && mode != "regular" && mode != "fullscreen" {
		return fmt.Errorf("invalid tui mode %q: use regular or fullscreen", mode)
	}
	cfg := config.Load(workspace)
	mode = resolveTUIMode(mode, cfg.TUIMode)
	if model != "" {
		cfg.DefaultModel = model
	}
	// Pick Pi's theme for this terminal before the UI owns input (issue #12).
	themeName := initPiTheme(cfg.Theme)
	log.Printf("tui theme: %s (setting %q)", themeName, cfg.Theme)

	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	engine := turn.NewWithRuntimeConfig(s, cfg, cfg.SystemPrompt)
	engine.EnableMCP() // Pi mcp.json servers (.gi first, then .pi); connects in the background
	defer engine.Close()
	return runWithEngineMode(s, engine, cfg, mode == "regular")
}

// Fixtures can install native hooks while reusing the production terminal loop.
func runWithEngine(s *store.Store, engine *turn.Engine, cfg config.RuntimeConfig) error {
	return runWithEngineMode(s, engine, cfg, false)
}

func runWithEngineMode(s *store.Store, engine *turn.Engine, cfg config.RuntimeConfig, regular bool) error {
	sessionID, err := initialSessionID(context.Background(), s)
	if err != nil {
		return err
	}
	if sessionID == "" {
		id := store.NowID("session")
		alloc := gisession.AllocateDefaultSession(defaultTUIAgentID, "gi", "default", id)
		sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(context.Background(), store.ResolveOrCreateSessionFromAllocationInput{ID: id, Title: "@" + defaultTUIAgentID, State: map[string]any{"status": "idle", "queue_count": 0, "model": cfg.DefaultModel, "provider": cfg.DefaultProvider, "thinking_level": cfg.DefaultThinkingLevel}, Allocation: alloc})
		if err != nil {
			return fmt.Errorf("create session: %w", err)
		}
		sessionID = sess.ID
	}

	chat := &chatTUI{
		store:         s,
		engine:        engine,
		sessionID:     sessionID,
		cfg:           cfg,
		transcriptRef: gotui.NewRef(),
		stickToBottom: true,
		durableDrafts: true,
		regularMode:   regular,
		startupHeader: true,
	}

	options := []gotui.AppOption{gotui.WithLegacyKeyboard(), gotui.WithRowRedraw()}
	if regular {
		options = append(options, gotui.WithInlineHeight(5), gotui.WithPostRenderHook(chat.flushRegularTranscript))
	} else {
		options = append(options, gotui.WithMouse(), gotui.WithPreFlushHook(chat.compositeJumpToLatest))
	}
	enableGoTUIColorOutput()
	app, err := gotui.NewApp(options...)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}
	chat.app = app
	app.SetPasteHandler(chat.handlePaste)
	cleanup := chat.Init()
	defer cleanup()
	chat.watchActiveTheme()
	defer func() { chat.themeWatcher.stop() }()
	app.SetRootComponent(chat)
	return app.Run()
}

const transcriptBlockMarkerPrefix = "⟦gi:block:"

type transcriptBlockHitTarget struct {
	Key string
	Ref *gotui.Ref
}

type transcriptBlockMeta struct {
	Key            string  `json:"key"`
	Kind           string  `json:"kind"`
	Title          string  `json:"title"`
	Status         string  `json:"status,omitempty"`
	StartedAt      string  `json:"started_at,omitempty"`
	EndedAt        string  `json:"ended_at,omitempty"`
	Detail         string  `json:"detail,omitempty"`
	Footer         string  `json:"footer,omitempty"`
	MarkdownSource string  `json:"markdown_source,omitempty"`
	ToolPath       string  `json:"tool_path,omitempty"`
	ToolContent    *string `json:"tool_content,omitempty"`
	// Pi's file tool renderers: read's line range and truncation notice,
	// edit's diff or error preview.
	ToolRange  string  `json:"tool_range,omitempty"`
	ToolNotice string  `json:"tool_notice,omitempty"`
	EditDiff   *string `json:"edit_diff,omitempty"`
	EditError  string  `json:"edit_error,omitempty"`
	// Codemode: nested call rows and the saved full output.
	Calls          []codemodeCallRow `json:"calls,omitempty"`
	FullOutputPath string            `json:"full_output_path,omitempty"`
}

type transcriptBlockSpan struct {
	HeaderIndex int
	BodyCount   int
}

type transcriptRenderableBlock struct {
	MarkdownSource string
	Key            string
	Kind           string
	Header         string
	Subheader      string
	Body           []string
	Expanded       bool
	Expandable     bool
	PreviewLimit   int
	PreviewTail    bool
	Footer         string
	Status         string
	Selected       bool
	Border         gotui.BorderStyle
	BorderStyle    gotui.Style
	HeaderStyle    gotui.Style
	BodyStyle      gotui.Style
	HintStyle      gotui.Style
	SelectedHint   string
	// Static blocks are printed to terminal-owned scrollback in full; they
	// cannot be toggled, so no expand/collapse hints are shown.
	Static bool
	// Tool call rendering (Pi's renderCall): argument text and timing.
	ToolPath           string
	ToolContent        *string
	ToolRange          string
	ToolNotice         string
	EditDiff           *string
	EditError          string
	Calls              []codemodeCallRow
	FullOutputPath     string
	ToolArg            string
	StartedAt, EndedAt string
}

// bashPreviewLines mirrors Pi's BashExecutionComponent preview window: when
// collapsed, only the trailing N lines of `!` output are shown.
const bashPreviewLines = 20

// Pi's tool previews: shell tools show the last 5 lines, others the first 10.
const (
	toolShellPreviewLines = 5
	toolPreviewLines      = 10
)

type chatTUI struct {
	app                         *gotui.App
	store                       *store.Store
	engine                      *turn.Engine
	sessionID                   string
	cfg                         config.RuntimeConfig
	history                     []string
	histIdx                     int
	historySearchQuery          string
	historySearchIdx            int
	historyDraft                string
	historyDraftCursor          int
	historyApplying             bool
	running                     bool
	status                      string
	compaction                  terminalCompaction
	workspaceIndex              terminalIndex
	draft                       string
	inputActive                 bool
	eventCh                     chan sessionEvent
	topicEventCh                chan sessionTopicEvent
	sessionGeneration           uint64
	subscriptionCancel          context.CancelFunc
	sessionEditors              map[string]sessionEditorState
	durableDrafts               bool
	draftApplying               bool
	textDrafts                  map[string]*terminalDraftState
	sessionModelDefaults        *[3]string
	subscribedCh                chan map[string]any
	topicUnsubscribe            func()
	input                       *multilineInput
	search                      transcriptSearch
	textSelection               transcriptSelection
	selectionClicks             transcriptClickSequence
	selectionClickSnapshot      transcriptSelection
	nativeSelectionCopyPending  bool
	queuedDrafts                []string
	queueSnapshot               []string
	queueSnapshotScope          sessionScope
	pendingMedia                map[string][]store.MediaRef
	mediaClaims                 map[string]*mediaClaim
	inputRegion                 *gotui.Element
	transcriptRegion            *gotui.Element
	transcriptRef               *gotui.Ref
	transcript                  []string
	regularMode                 bool
	regularHeaderPrinted        bool      // startup header printed to scrollback (regular mode)
	startupHeader               bool      // show gi\'s startup header (set by the app, not by test fixtures)
	lastFooterSignature         string    // footer data at the last idle check
	lastFooterCheck             time.Time // when the idle check last read the session
	runningCheckLines           []string  // transcript at the last running-block check
	runningCheckResult          bool
	blockCache                  *transcriptBlockCache // block heights and recently rendered blocks (#34, #31)
	thinkingMemoKey             string                // effectiveThinking's last model+level and answer
	heldNoticeShown             map[string]bool       // sessions told about held interrupted turns
	thinkingMemoValue           string
	blocksMemo                  transcriptBlocksMemo      // block list and keys for unchanged transcripts (#34)
	projectionMemo              *transcriptProjectionMemo // bounded retained search/selection projection
	jumpToLatest                jumpToLatestRect          // where the cue was drawn (none: width 0)
	regularPrinted              int
	regularSessionPending       bool
	regularWidth, regularHeight int
	regularReflowGen            int
	regularResizeBase           int // height before the current resize burst
	regularResizeBaseWidth      int
	regularResizeWidthChanged   bool
	regularResizePeak           int    // tallest height during the burst
	modelMenuScope              string // Pi selector scope: "scoped" or "all"
	modelMenuDefault            string // saved default model (Pi's "default" badge)
	modelMenuSessionRows        map[string]sessionPickerRow
	slash                       slashMenu
	slashLastText               string
	footerUsage                 *footerUsageCache
	footerCached                *footerCache
	wheel                       wheelAccelerator
	footerAuth                  *footerAuthCache
	transcriptScroll            int
	stickToBottom               bool
	draftLineIndex              int
	draftLineCount              int
	outputWidth                 int
	outputHeight                int
	osc52Writer                 io.Writer
	clipboardLookPath           func(string) (string, error)
	clipboardImageReader        func() ([]byte, string, error)
	clipboardRun                func(context.Context, string, []string, string) error
	transcriptBlockRefs         []transcriptBlockHitTarget
	transcriptExpanded          map[string]bool
	transcriptBlockSpans        map[string]transcriptBlockSpan
	transcriptToolBlocks        map[string]string
	selectedTranscriptBlock     string
	thinkingText                string
	thinkingBlockKey            string
	thinkingStartedAt           string
	thinkingIndicatorKey        string
	thinkingIndicatorStart      string
	lastInputTokens             int
	lastOutputTokens            int
	lastContextTokens           int
	lastCacheRead               int
	lastCacheWrite              int
	lastCostTotal               float64
	modelMenuOpen               bool
	modelMenuKind               string
	modelMenuValues             map[string]string
	modelMenuMetadata           map[string]modelPickerMetadata
	sessionActions              sessionActions
	selectDialog                selectDialog       // Pi ctx.ui.select (modelMenuKind "select")
	mcpManager                  *mcpManagerState   // Pi's /mcp manager (modelMenuKind "mcp-manager")
	scopedModels                *scopedModelsState // Pi's /scoped-models (modelMenuKind "scoped-models")
	authSelector                *authSelectorState // Pi's /login and /logout provider selector ("auth-selector")
	loginDialog                 *loginDialogState  // Pi's login dialog ("login-dialog")
	settingsList                *settingsListState // Pi's /settings ("settings")
	treeSelector                *treeSelector      // Pi's /tree ("tree")
	editorDialog                *editorDialog      // Pi's ctx.ui.editor ("editor-dialog")
	branchSummaryCancel         context.CancelFunc // set while /tree summarizes a branch
	loader                      *borderedLoader    // Pi's BorderedLoader (modelMenuKind "loader")
	themeWatcher                *themeWatcher      // reloads the active custom theme's file
	lastCtrlC                   time.Time          // Pi's app.clear: a second press within 500ms exits
	modelMenuSession            sessionScope
	modelMenuAltScreen          bool
	modelMenuResized            bool
	modelMenuInlineHeight       int
	modelMenuRenderedHeight     int
	modelMenuChoices            []string
	modelMenuAll                []string
	modelMenuQuery              string
	modelMenuError              string
	modelMenuSelected           int
	modelMenuScroll             int
	extensionStatuses           map[string]string
	extensionWidgets            map[string][]string
	extensionToolModes          map[string]string
	editorAskActive             bool
	editorAskHandler            func(answer string, cancelled bool) // internal asks (e.g. /mcp login); nil: extension asks
	uiQueue                     chan func()                         // background UI updates when there is no app (tests)
	editorAskKey                string
	editorAskPrompt             string
	editorAskPrevPlaceholder    string
	editorAskPrevText           string
	editorAskPrevCursor         int
}

func (c *chatTUI) ensureInput() {
	if c.input != nil {
		if c.input.onChange == nil {
			c.input.onChange = c.onInputChanged
		}
		c.input.onEdit = c.scrollTranscriptToBottom
		c.input.onEscape = c.handleTranscriptEscape
		c.bindTranscriptNavigation()
		return
	}
	c.input = newMultilineInput(80, "Send a message…", c.onSubmit, c.onInputChanged)
	c.input.onEdit = c.scrollTranscriptToBottom
	c.input.onRestoreQueued = c.restoreQueuedDraft
	c.input.onFollowUp = c.onFollowUp
	c.input.onComplete = c.completeInputPath
	c.input.interceptKey = c.handleSlashKey
	c.input.onEscape = c.handleTranscriptEscape
	c.bindTranscriptNavigation()
}

func (c *chatTUI) onInputChanged(text string) {
	previous := c.slashLastText
	c.slashLastText = text
	c.updateSlashMenu(previous)
	if !c.historyApplying && !c.draftApplying {
		c.histIdx = -1
		c.historyDraft = ""
		c.historyDraftCursor = 0
		c.historySearchQuery = ""
		c.historySearchIdx = -1
	}
	if !c.historyApplying {
		c.saveDurableDraft()
	}
	// User edits resume following through onEdit. Programmatic restoration
	// preserves the reader position unless already following the newest edge.
	if c.stickToBottom {
		c.scrollTranscriptToBottom()
	}
}

func encodeTranscriptBlockMarker(meta transcriptBlockMeta) string {
	payload, err := json.Marshal(meta)
	if err != nil {
		return transcriptBlockMarkerPrefix + "error⟧"
	}
	return transcriptBlockMarkerPrefix + base64.RawURLEncoding.EncodeToString(payload) + "⟧"
}

// markerCache memoizes parsed block markers: marker lines are immutable (a
// status change writes a new line), and every render and idle check parses
// all of them.
var markerCache = struct {
	sync.Mutex
	m map[string]transcriptBlockMeta
}{m: map[string]transcriptBlockMeta{}}

const markerCacheMax = 8192

func parseTranscriptBlockMarker(line string) (transcriptBlockMeta, bool) {
	// Markers start their line; avoid scanning long message text.
	if !strings.HasPrefix(strings.TrimLeft(line, " \t"), transcriptBlockMarkerPrefix) {
		return transcriptBlockMeta{}, false
	}
	markerCache.Lock()
	meta, ok := markerCache.m[line]
	markerCache.Unlock()
	if ok {
		return meta, true
	}
	meta, ok = decodeTranscriptBlockMarker(line)
	if ok {
		markerCache.Lock()
		if len(markerCache.m) >= markerCacheMax {
			markerCache.m = map[string]transcriptBlockMeta{}
		}
		markerCache.m[line] = meta
		markerCache.Unlock()
	}
	return meta, ok
}

func decodeTranscriptBlockMarker(line string) (transcriptBlockMeta, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, transcriptBlockMarkerPrefix) || !strings.HasSuffix(line, "⟧") {
		return transcriptBlockMeta{}, false
	}
	encoded := strings.TrimSuffix(strings.TrimPrefix(line, transcriptBlockMarkerPrefix), "⟧")
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return transcriptBlockMeta{}, false
	}
	var meta transcriptBlockMeta
	if err := json.Unmarshal(payload, &meta); err != nil || strings.TrimSpace(meta.Key) == "" {
		return transcriptBlockMeta{}, false
	}
	return meta, true
}

func (c *chatTUI) ensureTranscriptBlockState() {
	if c.transcriptExpanded == nil {
		c.transcriptExpanded = map[string]bool{}
	}
	if c.transcriptBlockSpans == nil {
		c.transcriptBlockSpans = map[string]transcriptBlockSpan{}
	}
	if c.transcriptToolBlocks == nil {
		c.transcriptToolBlocks = map[string]string{}
	}
}

func (c *chatTUI) reindexTranscriptBlocks() {
	c.ensureTranscriptBlockState()
	spans := map[string]transcriptBlockSpan{}
	for i := 0; i < len(c.transcript); i++ {
		meta, ok := parseTranscriptBlockMarker(c.transcript[i])
		if !ok {
			continue
		}
		bodyCount := 0
		for j := i + 1; j < len(c.transcript); j++ {
			if _, isBlock := parseTranscriptBlockMarker(c.transcript[j]); isBlock {
				break
			}
			if strings.HasPrefix(c.transcript[j], "│ ") {
				bodyCount++
				continue
			}
			break
		}
		spans[meta.Key] = transcriptBlockSpan{HeaderIndex: i, BodyCount: bodyCount}
	}
	c.transcriptBlockSpans = spans
	for toolKey, blockKey := range c.transcriptToolBlocks {
		if _, ok := spans[blockKey]; !ok {
			delete(c.transcriptToolBlocks, toolKey)
		}
	}
	if c.selectedTranscriptBlock != "" {
		if _, ok := spans[c.selectedTranscriptBlock]; !ok {
			c.selectedTranscriptBlock = ""
		}
	}
}

func (c *chatTUI) appendTranscriptBlock(meta transcriptBlockMeta, body []string) {
	c.ensureTranscriptBlockState()
	lines := []string{encodeTranscriptBlockMarker(meta)}
	for _, line := range body {
		trimmed := strings.TrimRight(line, "\r")
		if trimmed == "" {
			trimmed = " "
		}
		lines = append(lines, "│ "+trimmed)
	}
	c.appendTranscript(lines...)
	c.reindexTranscriptBlocks()
	if meta.Key != "" {
		c.selectedTranscriptBlock = meta.Key
	}
}

func (c *chatTUI) replaceTranscriptBlock(meta transcriptBlockMeta, body []string) {
	c.ensureTranscriptBlockState()
	span, ok := c.transcriptBlockSpans[meta.Key]
	if !ok || span.HeaderIndex < 0 || span.HeaderIndex >= len(c.transcript) {
		c.appendTranscriptBlock(meta, body)
		return
	}
	prefix := append([]string(nil), c.transcript[:span.HeaderIndex]...)
	middle := []string{encodeTranscriptBlockMarker(meta)}
	for _, line := range body {
		trimmed := strings.TrimRight(line, "\r")
		if trimmed == "" {
			trimmed = " "
		}
		middle = append(middle, "│ "+trimmed)
	}
	end := span.HeaderIndex + 1 + span.BodyCount
	if end > len(c.transcript) {
		end = len(c.transcript)
	}
	suffix := append([]string(nil), c.transcript[end:]...)
	c.transcript = append(prefix, append(middle, suffix...)...)
	c.applyTranscriptLimit()
	c.reindexTranscriptBlocks()
	if meta.Key != "" {
		c.selectedTranscriptBlock = meta.Key
	}
}

func (c *chatTUI) deleteTranscriptBlock(key string) bool {
	c.ensureTranscriptBlockState()
	span, ok := c.transcriptBlockSpans[key]
	if !ok || span.HeaderIndex < 0 || span.HeaderIndex >= len(c.transcript) {
		delete(c.transcriptBlockSpans, key)
		return false
	}
	if meta, ok := parseTranscriptBlockMarker(c.transcript[span.HeaderIndex]); !ok || meta.Key != key {
		delete(c.transcriptBlockSpans, key)
		c.reindexTranscriptBlocks()
		return false
	}
	end := span.HeaderIndex + 1 + span.BodyCount
	if end > len(c.transcript) {
		end = len(c.transcript)
	}
	c.transcript = append(c.transcript[:span.HeaderIndex], c.transcript[end:]...)
	delete(c.transcriptBlockSpans, key)
	delete(c.transcriptExpanded, key)
	if c.selectedTranscriptBlock == key {
		c.selectedTranscriptBlock = ""
	}
	c.reindexTranscriptBlocks()
	return true
}

func (c *chatTUI) toggleTranscriptBlock(key string) bool {
	if strings.TrimSpace(key) == "" {
		return false
	}
	c.ensureTranscriptBlockState()
	c.selectedTranscriptBlock = key
	c.transcriptExpanded[key] = !c.transcriptExpanded[key]
	if c.app != nil {
		c.app.MarkDirty()
	}
	return true
}

func (c *chatTUI) transcriptBlockOrder() []string {
	order := make([]string, 0, len(c.transcriptBlockSpans))
	for _, line := range c.transcript {
		meta, ok := parseTranscriptBlockMarker(line)
		if ok {
			order = append(order, meta.Key)
		}
	}
	return order
}

func (c *chatTUI) selectTranscriptBlock(delta int) {
	c.reindexTranscriptBlocks()
	order := c.transcriptBlockOrder()
	if len(order) == 0 {
		return
	}
	idx := 0
	if c.selectedTranscriptBlock != "" {
		for i, key := range order {
			if key == c.selectedTranscriptBlock {
				idx = i
				break
			}
		}
	}
	idx = (idx + delta) % len(order)
	if idx < 0 {
		idx += len(order)
	}
	c.selectedTranscriptBlock = order[idx]
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) toggleSelectedTranscriptBlock() {
	c.reindexTranscriptBlocks()
	order := c.transcriptBlockOrder()
	if len(order) == 0 {
		return
	}
	key := c.selectedTranscriptBlock
	if key == "" {
		key = order[len(order)-1]
	}
	c.toggleTranscriptBlock(key)
}

func (c *chatTUI) Init() func() {
	c.initWorkspaceIndex()
	c.eventCh = make(chan sessionEvent, 64)
	c.bindSession(c.sessionID)
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.histIdx = -1
	c.historySearchIdx = -1
	c.history = c.loadCommandHistory()
	c.inputActive = true
	c.transcript = c.loadTranscript()
	c.reindexTranscriptBlocks()
	c.draftLineIndex = -1
	c.restoreActiveSessionWork()
	if strings.TrimSpace(c.cfg.DefaultModel) == "" {
		c.status = "Select a model with /model <name>"
		c.transcript = append(c.transcript, c.firstUseModelPromptLines()...)
	}
	c.ensureInput()
	c.loadDurableDraft()
	c.noticeHeldInterruptions()
	c.watchMCPNotices()
	c.scrollTranscriptToBottom()

	if c.app != nil {
		time.AfterFunc(200*time.Millisecond, func() {
			c.app.QueueUpdate(func() { c.focusInput() })
		})
	}

	return func() { c.stopSessionSubscription(); c.stopWorkspaceIndex(); c.closeModelPickerScreen() }
}

func (c *chatTUI) bindSession(sessionID string) {
	c.closeWorkspaceIndex()
	c.modelDefaults()
	if c.eventCh == nil {
		c.eventCh = make(chan sessionEvent, 64)
	}
	if c.topicEventCh == nil {
		c.topicEventCh = make(chan sessionTopicEvent, 64)
	}
	c.stopSessionSubscription()
	c.sessionGeneration++
	c.compaction = terminalCompaction{}
	c.sessionID = sessionID
	c.regularPrinted = 0
	c.regularSessionPending = true
	if c.store != nil {
		if session, err := c.store.GetSession(context.Background(), sessionID); err == nil {
			c.restoreSessionModel(session.State)
		}
	}
	if c.engine == nil {
		return
	}
	c.syncCompactionActivity()
	ctx, cancel := context.WithCancel(context.Background())
	c.subscriptionCancel = cancel
	scope := c.selectionScope()
	c.subscribedCh = c.engine.Subscribe(sessionID)
	if c.engine.Topics() != nil {
		ch, unsubscribe := c.engine.Topics().Subscribe(ctx, "*", topics.SubscribeOptions{Buffer: 64, SessionID: sessionID})
		c.topicUnsubscribe = unsubscribe
		go forwardSessionTopics(ctx, ch, c.topicEventCh, scope)
	}
	go forwardSessionEvents(ctx, c.subscribedCh, c.eventCh, scope)
}

func (c *chatTUI) Watchers() []gotui.Watcher {
	watchers := []gotui.Watcher{gotui.NewChannelWatcher(c.eventCh, c.handleSessionEvent)}
	if c.workspaceIndex.results != nil {
		watchers = append(watchers, gotui.NewChannelWatcher(c.workspaceIndex.results, c.handleWorkspaceIndexResult))
	}
	if c.topicEventCh != nil {
		watchers = append(watchers, gotui.NewChannelWatcher(c.topicEventCh, c.handleSessionTopicEvent))
	}
	// One 80 ms tick (Pi's spinner cadence) drives selection auto-scroll,
	// activity redraws and, every other tick, the durable draft save; three
	// separate timers woke the process ~33 times a second when idle.
	var ticks uint64
	watchers = append(watchers, gotui.OnTimer(80*time.Millisecond, func() {
		ticks++
		c.tickTranscriptSelection()
		if ticks%2 == 0 {
			c.saveDurableDraft()
		}
		if (c.running || c.compaction.active || c.branchSummaryCancel != nil || c.loader != nil || c.hasRunningTranscriptBlock()) && c.app != nil {
			c.app.MarkDirty()
		}
	}))
	watchers = append(watchers, gotui.OnTimer(time.Second, func() {
		c.refreshActiveSessionIndicator()
		if c.compaction.active {
			c.syncCompactionActivity()
		}
		// Re-render only when something can have changed: live activity, or
		// footer data updated elsewhere (another frontend, a finished turn).
		// An unconditional re-render rebuilt the whole transcript every second.
		live := c.running || c.compaction.active || c.thinkingIndicatorKey != ""
		// Footer data changes arrive as topic events (invalidateFooter); poll
		// the session only then, or every few seconds as a fallback.
		if c.footerCached == nil || time.Since(c.lastFooterCheck) >= footerIdlePoll {
			c.lastFooterCheck = time.Now()
			if footer := fmt.Sprintf("%+v", c.footerData()); footer != c.lastFooterSignature {
				c.lastFooterSignature = footer
				live = true
			}
		}
		if live && c.app != nil {
			c.app.MarkDirty()
		}
	}))
	return watchers
}

// footerIdlePoll is how often an idle TUI re-reads the footer's session
// data without an invalidating event.
const footerIdlePoll = 5 * time.Second

// hasRunningTranscriptBlock reports a running block; the answer is reused
// while the transcript lines are unchanged (lines are immutable strings).
func (c *chatTUI) hasRunningTranscriptBlock() bool {
	if c.runningCheckLines != nil && sameLines(c.runningCheckLines, c.transcript) {
		return c.runningCheckResult
	}
	result := false
	for _, line := range c.transcript {
		meta, ok := parseTranscriptBlockMarker(line)
		if ok && meta.Status == "running" {
			result = true
			break
		}
	}
	c.runningCheckLines = append(c.runningCheckLines[:0], c.transcript...)
	c.runningCheckResult = result
	return result
}

func (c *chatTUI) handleTopicEvent(env topics.Envelope) {
	c.invalidateFooter()
	if env.SessionID != "" && env.SessionID != c.sessionID {
		return
	}
	payload := env.Payload
	switch env.Topic {
	case "turn.status":
		if payload["phase"] == "retry_wait" {
			c.handleEvent(payload)
			return
		}
		title, _ := payload["title"].(string)
		status, _ := payload["status"].(string)
		if status == "running" {
			c.markRunning()
			c.showThinkingIndicator(env.Timestamp)
		}
		_ = title
		if status == "idle" {
			c.resetRunningDraftState()
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
		}
	case "turn.response":
		if topicPayloadType(payload, "new_post", "agent_response") {
			return
		}
		text := ""
		responseType := ""
		sender, _ := payload["sender"].(string)
		if data, _ := payload["data"].(map[string]any); data != nil {
			text, _ = data["content"].(string)
			responseType, _ = data["type"].(string)
		}
		if text == "" {
			text, _ = payload["content"].(string)
		}
		if text != "" {
			if sender == "system" || responseType == "system_message" {
				c.clearDraftTranscriptLine()
				c.appendTranscript("sys: " + truncate(text, 160))
			} else {
				c.finalizeDraftTranscript(text)
			}
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
			c.resetRunningDraftState()
		}
	case "turn.draft":
		if topicPayloadType(payload, "agent_draft_delta") {
			return
		}
		delta, _ := payload["delta"].(string)
		if delta != "" {
			c.markRunning()
			c.clearThinkingIndicator()
			c.draft += delta
			c.updateDraftTranscriptLine()
		}
	case "turn.thought":
		if topicPayloadType(payload, "agent_thought_delta") {
			return
		}
		c.markRunning()
		c.clearThinkingIndicator()
		delta, _ := payload["delta"].(string)
		c.updateThinkingTranscript(delta, env.Timestamp)
	case "runtime.tool":
		c.renderToolEvent(payload, env.Timestamp)
	case "runtime.hook":
		if hook, _ := payload["hook"].(string); hook == turn.HookSessionBeforeCompact || hook == turn.HookSessionCompact {
			c.syncCompactionActivity()
		} else {
			c.renderHookEvent(payload, env.Timestamp)
		}
	case "runtime.turn":
		typ, _ := payload["type"].(string)
		status, _ := payload["status"].(string)
		switch typ {
		case "turn_usage":
			c.updateUsageFromPayload(payload)
		case "turn_completed":
			c.resetRunningDraftState()
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
		case "turn_terminal":
			if status == "failed" || status == "aborted" || status == "cancelled" {
				c.resetRunningDraftState()
				c.status = fmt.Sprintf("Turn %s", status)
			}
		case "turn_state":
			phase, _ := payload["phase"].(string)
			if status == "running" && phase == "waiting_on_tools" {
				c.clearThinkingIndicator()
				c.markRunning()
			}
		}
	case "runtime.session":
		typ, _ := payload["type"].(string)
		status, _ := payload["status"].(string)
		switch typ {
		case "session_running", "session_state":
			if status == "running" {
				c.markRunning()
				c.showThinkingIndicator(env.Timestamp)
			}
			if status == "queued" {
				c.resetRunningDraftState()
				c.status = "Queued"
			}
			if status == "idle" {
				c.resetRunningDraftState()
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
			}
		case "session_idle":
			c.resetRunningDraftState()
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
		}
	case "runtime.routing":
		c.renderRoutingEvent(payload, env.Timestamp)
	case "runtime.inbound_work":
		c.renderInboundWorkEvent(payload, env.Timestamp)
	case "runtime.dispatcher":
		c.renderDispatcherEvent(payload, env.Timestamp)
	case "session.compaction":
		if c.store != nil {
			c.syncCompactionActivity()
		} else {
			c.renderCompactionEvent(payload, env.Timestamp)
		}
	case "session.routing":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		c.renderRoutingEvent(payload, env.Timestamp)
	case "session.steering":
		c.renderSteeringEvent(payload["type"])
	case "turn.subturn":
		c.renderSubturnEvent(payload, env.Timestamp)
	case "extension.status":
		key, _ := payload["key"].(string)
		text, _ := payload["text"].(string)
		c.setExtensionStatus(key, text)
	case "extension.widget":
		key, _ := payload["key"].(string)
		c.setExtensionWidget(key, widgetPayloadLines(payload))
	case "extension.tool_render":
		tool, _ := payload["tool"].(string)
		mode, _ := payload["mode"].(string)
		c.setExtensionToolRender(tool, mode)
	case "extension.editor":
		key, _ := payload["key"].(string)
		prompt, _ := payload["prompt"].(string)
		prefill, _ := payload["prefill"].(string)
		c.setEditorAsk(key, prompt, prefill)
	}
	if c.stickToBottom {
		c.scrollTranscriptToBottom()
	}
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func topicPayloadType(payload map[string]any, types ...string) bool {
	got, _ := payload["type"].(string)
	got = strings.TrimSpace(got)
	for _, typ := range types {
		if got == typ {
			return true
		}
	}
	return false
}

func (c *chatTUI) updateUsageFromPayload(payload map[string]any) {
	usage, _ := payload["usage"].(map[string]any)
	if usage == nil {
		return
	}
	c.lastInputTokens = intFromAny(usage["input"])
	c.lastOutputTokens = intFromAny(usage["output"])
	// Billing totals are cumulative over iterations, not context occupancy.
	// The explicit context.measured record supplies the context footer below.
	c.lastCacheRead = intFromAny(usage["cache_read"])
	c.lastCacheWrite = intFromAny(usage["cache_write"])
	if cost := floatFromAny(usage["cost_total"]); cost > 0 {
		c.lastCostTotal = cost
	}
}

func (c *chatTUI) useTopicNativeRuntimeStatus() bool {
	return c.topicUnsubscribe != nil
}

func (c *chatTUI) handleEvent(ev map[string]any) {
	if !sessionEventMatchesID(ev, c.sessionID) {
		return
	}
	evType, _ := ev["type"].(string)
	switch evType {
	case "agent_draft_delta":
		delta, _ := ev["delta"].(string)
		c.markRunning()
		c.clearThinkingIndicator()
		c.draft += delta
		c.updateDraftTranscriptLine()
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "new_post":
		data, _ := ev["data"].(map[string]any)
		if data != nil {
			text, _ := data["content"].(string)
			if text != "" {
				if data["type"] == "system_message" || ev["sender"] == "system" {
					// A terminal system notice is not an assistant response. Use
					// the same role projection as reload, so an earlier streamed
					// error can deduplicate instead of gaining a second Gi label.
					c.clearDraftTranscriptLine()
					c.appendTranscript("sys: " + text)
				} else {
					c.finalizeDraftTranscript(text)
				}
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
				c.resetRunningDraftState()
				if c.stickToBottom {
					c.scrollTranscriptToBottom()
				}
				if c.app != nil {
					c.app.MarkDirty()
				}
			}
		}
	case "agent_thought_delta":
		c.markRunning()
		c.clearThinkingIndicator()
		delta, _ := ev["delta"].(string)
		c.updateThinkingTranscript(delta, time.Time{})
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "tool_started", "tool_finished", "tool_failed", "tool_skipped":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		payload := cloneAnyMap(ev)
		payload["type"] = evType
		c.renderToolEvent(payload, time.Time{})
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "compaction":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		payload := cloneAnyMap(ev)
		payload["type"] = evType
		c.renderCompactionEvent(payload, time.Time{})
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "routing_decision", "routing_incoming":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		payload := cloneAnyMap(ev)
		payload["type"] = evType
		c.renderRoutingEvent(payload, time.Time{})
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "agent_status":
		if ev["phase"] == "retry_wait" {
			title, _ := ev["title"].(string)
			c.markRunning()
			c.showThinkingIndicator(time.Time{})
			if span, ok := c.transcriptBlockSpans[c.thinkingIndicatorKey]; ok && span.HeaderIndex >= 0 && span.HeaderIndex < len(c.transcript) {
				if meta, valid := parseTranscriptBlockMarker(c.transcript[span.HeaderIndex]); valid {
					meta.Title = title
					c.replaceTranscriptBlock(meta, nil)
				}
			}
			c.status = title
			if c.app != nil {
				c.app.MarkDirty()
			}
			return
		}
		title := ""
		if v, ok := ev["title"].(string); ok {
			title = v
		}
		if title != "" {
			c.markRunning()
			c.showThinkingIndicator(time.Time{})
		} else {
			c.resetRunningDraftState()
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "error":
		msg := ""
		if v, ok := ev["error"].(string); ok {
			msg = v
		}
		if msg == "" {
			if v, ok := ev["message"].(string); ok {
				msg = v
			}
		}
		if msg != "" {
			c.clearDraftTranscriptLine()
			c.running = false
			c.finishThinkingTranscript(time.Now().UTC())
			c.appendTranscript("error: " + truncate(msg, 160))
			c.status = "Error"
			if c.app != nil {
				c.app.MarkDirty()
			}
		}
	}
}

func (c *chatTUI) updateDraftTranscriptLine() {
	lines := c.renderStreamingDraftLines()
	if len(lines) == 0 {
		return
	}
	if c.draftLineIndex >= 0 && c.draftLineIndex < len(c.transcript) {
		end := c.draftLineIndex + c.draftLineCount
		if c.draftLineCount <= 0 || end > len(c.transcript) {
			end = c.draftLineIndex + 1
		}
		prefix := append([]string(nil), c.transcript[:c.draftLineIndex]...)
		suffix := append([]string(nil), c.transcript[end:]...)
		c.transcript = append(prefix, append(lines, suffix...)...)
		c.draftLineCount = len(lines)
		c.applyTranscriptLimit()
		return
	}
	c.appendTranscript(lines...)
	c.draftLineIndex = len(c.transcript) - len(lines)
	c.draftLineCount = len(lines)
}

// Render local submissions through the same Markdown path as stored messages.
// Otherwise the user's prompt remains raw until a session reload.
func (c *chatTUI) appendUserPrompt(text string, queued bool) {
	prefix := "you: "
	if queued {
		prefix = "you [queued]: "
	}
	c.appendTranscript(renderChatMarkdown("user", prefix, text, c.transcriptRenderWidth())...)
}

// Reproject the entire accumulated draft on every delta. Markdown structures
// (notably tables, emphasis and fences) may become valid only after a later
// token; a one-time looksLikeMarkdown check leaves the live preview raw.
func (c *chatTUI) renderStreamingDraftLines() []string {
	return renderChatMarkdown("assistant", c.cfg.AssistantName+": ", c.draft, c.transcriptRenderWidth())
}

func (c *chatTUI) finalizeDraftTranscript(text string) {
	lines := renderChatMarkdown("assistant", c.cfg.AssistantName+": ", text, c.transcriptRenderWidth())
	if c.draftLineIndex >= 0 && c.draftLineIndex < len(c.transcript) {
		end := c.draftLineIndex + c.draftLineCount
		if c.draftLineCount <= 0 || end > len(c.transcript) {
			end = c.draftLineIndex + 1
		}
		prefix := append([]string(nil), c.transcript[:c.draftLineIndex]...)
		suffix := append([]string(nil), c.transcript[end:]...)
		c.transcript = append(prefix, append(lines, suffix...)...)
		c.applyTranscriptLimit()
		c.draftLineIndex = -1
		c.draftLineCount = 0
		return
	}
	c.appendTranscript(lines...)
}

func (c *chatTUI) clearDraftTranscriptLine() {
	if c.draftLineIndex >= 0 && c.draftLineIndex < len(c.transcript) {
		end := c.draftLineIndex + c.draftLineCount
		if c.draftLineCount <= 0 || end > len(c.transcript) {
			end = c.draftLineIndex + 1
		}
		c.transcript = append(c.transcript[:c.draftLineIndex], c.transcript[end:]...)
	}
	c.draft = ""
	c.draftLineIndex = -1
	c.draftLineCount = 0
}

func (c *chatTUI) promoteDraftToThinking(ts time.Time) {
	text := strings.TrimSpace(c.draft)
	if text == "" {
		return
	}
	c.clearDraftTranscriptLine()
	c.updateThinkingTranscript(text, ts)
}

func (c *chatTUI) updateThinkingTranscript(delta string, ts time.Time) {
	if strings.TrimSpace(delta) == "" {
		c.showThinkingIndicator(ts)
		return
	}
	c.clearThinkingIndicator()
	c.thinkingText += delta
	if strings.TrimSpace(c.thinkingBlockKey) == "" {
		c.thinkingBlockKey = fmt.Sprintf("thought:%d", time.Now().UnixNano())
	}
	if strings.TrimSpace(c.thinkingStartedAt) == "" {
		c.thinkingStartedAt = normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)
	}
	body := renderMarkdownTranscript("", c.thinkingText, c.transcriptBlockContentWidth("thought"))
	meta := transcriptBlockMeta{Key: c.thinkingBlockKey, Kind: "thought", Title: "", Status: "running", StartedAt: c.thinkingStartedAt, MarkdownSource: c.thinkingText}
	c.replaceTranscriptBlock(meta, body)
}

func (c *chatTUI) finishThinkingTranscript(ts time.Time) {
	c.clearThinkingIndicator()
	if strings.TrimSpace(c.thinkingBlockKey) != "" {
		span, ok := c.transcriptBlockSpans[c.thinkingBlockKey]
		if ok && span.HeaderIndex >= 0 && span.HeaderIndex < len(c.transcript) {
			if meta, ok := parseTranscriptBlockMarker(c.transcript[span.HeaderIndex]); ok {
				meta.Status = "done"
				meta.EndedAt = normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)
				body := c.readTranscriptBlockBody(c.thinkingBlockKey)
				c.replaceTranscriptBlock(meta, body)
			}
		}
	}
	c.thinkingText = ""
	c.thinkingBlockKey = ""
	c.thinkingStartedAt = ""
}

func (c *chatTUI) readTranscriptBlockBody(key string) []string {
	span, ok := c.transcriptBlockSpans[key]
	if !ok || span.HeaderIndex < 0 || span.HeaderIndex >= len(c.transcript) {
		return nil
	}
	body := make([]string, 0, span.BodyCount)
	for i := 0; i < span.BodyCount; i++ {
		idx := span.HeaderIndex + 1 + i
		if idx >= len(c.transcript) {
			break
		}
		body = append(body, strings.TrimPrefix(c.transcript[idx], "│ "))
	}
	return body
}

func (c *chatTUI) latestToolResultBody(turnID, toolCallID, toolName string) []string {
	if c.store == nil || strings.TrimSpace(c.sessionID) == "" {
		return nil
	}
	messages, err := c.store.ListMessages(context.Background(), c.sessionID)
	if err != nil {
		return nil
	}
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role != "tool_result" {
			continue
		}
		payload := msg.Payload
		if strings.TrimSpace(toolCallID) != "" {
			if got, _ := payload["tool_call_id"].(string); strings.TrimSpace(got) != strings.TrimSpace(toolCallID) {
				continue
			}
		}
		if strings.TrimSpace(turnID) != "" {
			if got, _ := payload["turn_id"].(string); strings.TrimSpace(got) != strings.TrimSpace(turnID) {
				continue
			}
		}
		if strings.TrimSpace(toolName) != "" {
			if got, _ := payload["tool_name"].(string); strings.TrimSpace(got) != strings.TrimSpace(toolName) {
				continue
			}
		}
		trimmed := strings.TrimRight(plainTerminalOutput(msg.Content), "\r\n")
		if strings.TrimSpace(trimmed) == "" {
			return []string{"(empty)"}
		}
		parts := strings.Split(trimmed, "\n")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			out = append(out, strings.TrimRight(part, "\r"))
		}
		return out
	}
	return nil
}

func normalizeThinkingStatus(title string) string {
	if strings.Contains(strings.ToLower(title), "thinking") {
		return "Thinking..."
	}
	return title
}

func (c *chatTUI) markRunning() {
	c.running = true
}

func (c *chatTUI) showThinkingIndicator(ts time.Time) {
	if strings.TrimSpace(c.draft) != "" || strings.TrimSpace(c.thinkingText) != "" || strings.TrimSpace(c.thinkingBlockKey) != "" || c.toolIsRunning() {
		return
	}
	c.ensureTranscriptBlockState()
	if strings.TrimSpace(c.thinkingIndicatorKey) == "" {
		c.thinkingIndicatorKey = fmt.Sprintf("thinking:%d", time.Now().UnixNano())
		c.thinkingIndicatorStart = normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)
	}
	meta := transcriptBlockMeta{Key: c.thinkingIndicatorKey, Kind: "thinking_indicator", Title: "Thinking...", Status: "running", StartedAt: c.thinkingIndicatorStart}
	if _, ok := c.transcriptBlockSpans[c.thinkingIndicatorKey]; ok {
		c.replaceTranscriptBlock(meta, nil)
	} else {
		c.appendTranscriptBlock(meta, nil)
	}
}

// A running tool has its own spinner; never add a second Thinking spinner
// while waiting for that tool to finish (including on repeated status events).
func (c *chatTUI) toolIsRunning() bool {
	for _, key := range c.transcriptToolBlocks {
		span, ok := c.transcriptBlockSpans[key]
		if !ok || span.HeaderIndex < 0 || span.HeaderIndex >= len(c.transcript) {
			continue
		}
		if meta, ok := parseTranscriptBlockMarker(c.transcript[span.HeaderIndex]); ok && meta.Status == "running" {
			return true
		}
	}
	return false
}

func (c *chatTUI) clearThinkingIndicator() {
	key := strings.TrimSpace(c.thinkingIndicatorKey)
	if key == "" {
		return
	}
	c.deleteTranscriptBlock(key)
	c.thinkingIndicatorKey = ""
	c.thinkingIndicatorStart = ""
}

func (c *chatTUI) renderSteeringEvent(eventTypeValue any) {}

func normalizeBlockTimestamp(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now().UTC()
	}
	return ts.UTC()
}

func formatBlockClock(ts string) string {
	if parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(ts)); err == nil {
		return parsed.Local().Format("15:04:05")
	}
	return ""
}

func formatBlockElapsed(startedAt, endedAt string) string {
	start, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(startedAt))
	if err != nil {
		return ""
	}
	end := time.Now().UTC()
	if strings.TrimSpace(endedAt) != "" {
		if parsed, parseErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(endedAt)); parseErr == nil {
			end = parsed
		}
	}
	d := end.Sub(start)
	if d < 0 {
		d = 0
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d / time.Minute)
	seconds := int((d % time.Minute) / time.Second)
	return fmt.Sprintf("%dm%02ds", minutes, seconds)
}

func (c *chatTUI) renderSubturnEvent(payload map[string]any, ts time.Time) {
	typ, _ := payload["type"].(string)
	childTurn, _ := payload["child_turn_id"].(string)
	status, _ := payload["status"].(string)
	meta := transcriptBlockMeta{Key: fmt.Sprintf("subturn:%d:%s", time.Now().UnixNano(), childTurn), Kind: "subturn", Status: strings.TrimSpace(status), StartedAt: normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)}
	switch typ {
	case "subturn_created":
		meta.Title = "Sub-turn started"
	case "subturn_status":
		meta.Title = "Sub-turn update"
	case "subturn_result_ready", "subturn_result_delivered", "subturn_orphaned":
		meta.Title = "Sub-turn result"
	default:
		return
	}
	body := []string{}
	if childTurn != "" {
		body = append(body, "turn="+childTurn)
	}
	if status != "" {
		body = append(body, "status="+status)
	}
	c.appendTranscriptBlock(meta, body)
}

func (c *chatTUI) renderHookEvent(payload map[string]any, ts time.Time) {
	typ, _ := payload["type"].(string)
	hookName, _ := payload["hook"].(string)
	reason, _ := payload["reason"].(string)
	toolName, _ := payload["tool"].(string)
	errText, _ := payload["error"].(string)
	durationMS := intFromAny(payload["duration_ms"])
	title := "Hook event"
	status := "info"
	switch typ {
	case "hook_deny", "hook_abort":
		title, status = "Hook denied", "error"
	case "hook_modify":
		title, status = "Hook modified", "ok"
	case "hook_respond":
		title, status = "Hook responded directly", "ok"
	case "hook_invocation":
		if errText != "" {
			title, status = "Hook invocation error", "error"
		} else {
			title, status = "Hook invoked", "running"
		}
	}
	body := []string{}
	if hookName != "" {
		body = append(body, "hook="+hookName)
	}
	if toolName != "" {
		body = append(body, "tool="+toolName)
	}
	if durationMS > 0 {
		body = append(body, fmt.Sprintf("duration=%dms", durationMS))
	}
	if reason != "" {
		body = append(body, "reason="+truncate(reason, 160))
	}
	if errText != "" {
		body = append(body, "error="+truncate(errText, 160))
		if durationMS > 0 && reason == "" {
			body = append(body, fmt.Sprintf("timeout=%dms", durationMS))
		}
	}
	c.appendTranscriptBlock(transcriptBlockMeta{Key: fmt.Sprintf("hook:%d:%s", time.Now().UnixNano(), hookName), Kind: "hook", Title: title, Status: status, StartedAt: normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)}, body)
	statusLine := strings.ToLower(title)
	if hookName != "" {
		statusLine += " via " + hookName
	}
	if toolName != "" {
		statusLine += " for " + toolName
	}
	if reason != "" {
		statusLine += ": " + truncate(reason, 120)
	} else if errText != "" {
		statusLine += ": " + truncate(errText, 120)
	}
	c.status = statusLine
}

func (c *chatTUI) renderInboundWorkEvent(payload map[string]any, ts time.Time) {
	typ, _ := payload["type"].(string)
	sourceKind, _ := payload["source_kind"].(string)
	status, _ := payload["status"].(string)
	attemptCount := intFromAny(payload["attempt_count"])
	errText, _ := payload["error"].(string)
	title := "Inbound work"
	blockStatus := "info"
	switch typ {
	case "inbound_work_enqueued":
		title = "Inbound work queued"
	case "inbound_work_retry_scheduled":
		title = "Inbound work retry scheduled"
	case "inbound_work_failed":
		title, blockStatus = "Inbound work failed", "error"
	case "inbound_work_completed":
		title, blockStatus = "Inbound work completed", "ok"
	case "inbound_work_requeued":
		title = "Inbound work requeued"
	case "inbound_work_discarded":
		title, blockStatus = "Inbound work discarded", "error"
	}
	body := []string{}
	if sourceKind != "" {
		body = append(body, "source="+sourceKind)
	}
	if status != "" {
		body = append(body, "status="+status)
	}
	if attemptCount > 0 {
		body = append(body, fmt.Sprintf("attempt=%d", attemptCount))
	}
	if errText != "" {
		body = append(body, "error="+truncate(errText, 160))
	}
	c.appendTranscriptBlock(transcriptBlockMeta{Key: fmt.Sprintf("inbound:%d:%s", time.Now().UnixNano(), typ), Kind: "dispatcher", Title: title, Status: blockStatus, StartedAt: normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)}, body)
	c.status = strings.ToLower(title)
	if sourceKind != "" {
		c.status += fmt.Sprintf(" (%s)", sourceKind)
	}
	if attemptCount > 0 && typ == "inbound_work_retry_scheduled" {
		c.status += fmt.Sprintf(" attempt %d", attemptCount)
	}
	if status != "" {
		c.status += fmt.Sprintf(" [%s]", status)
	}
	if errText != "" {
		c.status += ": " + truncate(errText, 120)
	}
}

func (c *chatTUI) renderDispatcherEvent(payload map[string]any, ts time.Time) {
	typ, _ := payload["type"].(string)
	workerID, _ := payload["worker_id"].(string)
	processedCount := intFromAny(payload["processed_count"])
	errText, _ := payload["error"].(string)
	title := "Dispatcher event"
	status := "info"
	switch typ {
	case "dispatcher_lease_acquired":
		title = "Inbound dispatcher lease acquired"
	case "dispatcher_lease_released":
		title = "Inbound dispatcher lease released"
	case "dispatcher_drain_completed":
		title, status = "Inbound dispatcher drain completed", "ok"
	case "dispatcher_error":
		title, status = "Inbound dispatcher error", "error"
	}
	body := []string{}
	if workerID != "" {
		body = append(body, "worker="+workerID)
	}
	if processedCount > 0 {
		body = append(body, fmt.Sprintf("processed=%d", processedCount))
	}
	if errText != "" {
		body = append(body, "error="+truncate(errText, 160))
	}
	c.appendTranscriptBlock(transcriptBlockMeta{Key: fmt.Sprintf("dispatcher:%d:%s", time.Now().UnixNano(), typ), Kind: "dispatcher", Title: title, Status: status, StartedAt: normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)}, body)
	c.status = strings.ToLower(title)
	if processedCount > 0 && typ == "dispatcher_drain_completed" {
		c.status += fmt.Sprintf(" (%d processed)", processedCount)
	}
	if workerID != "" {
		c.status += fmt.Sprintf(" [%s]", workerID)
	}
	if errText != "" {
		c.status += ": " + truncate(errText, 120)
	}
}

func (c *chatTUI) renderCompactionEvent(payload map[string]any, ts time.Time) {
	if kind, _ := payload["type"].(string); kind != "" && kind != "compaction" {
		return
	}
	before := intFromAny(payload["messages_before"])
	after := intFromAny(payload["messages_after"])
	tokens := intFromAny(payload["tokens_before"])
	c.appendTranscriptBlock(transcriptBlockMeta{Key: fmt.Sprintf("compact:%d", time.Now().UnixNano()), Kind: "compact", Title: "Context compacted", Status: "ok", StartedAt: normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)}, []string{fmt.Sprintf("messages=%d→%d", before, after), fmt.Sprintf("tokens_before=%d", tokens)})
	c.status = "Compacted context"
}

func (c *chatTUI) toolRuntimeBlockKey(payload map[string]any, toolName string) string {
	toolCallID, _ := payload["tool_call_id"].(string)
	turnID, _ := payload["turn_id"].(string)
	iteration := intFromAny(payload["iteration"])
	if strings.TrimSpace(toolCallID) != "" {
		// Providers can reuse a call ID in a later turn. The tuple is local to
		// this session's transcript; length encoding avoids delimiter collisions.
		turnID, _ := payload["turn_id"].(string)
		return fmt.Sprintf("toolcall:%d:%s:%s", len(turnID), turnID, toolCallID)
	}
	return fmt.Sprintf("tool:%s:%s:%d", strings.TrimSpace(turnID), strings.TrimSpace(toolName), iteration)
}

func (c *chatTUI) toolInvocationBody(toolName string, payload map[string]any) []string {
	if payload == nil {
		return nil
	}
	if command := toolInvocationText(toolName, payload["arguments"]); command != "" {
		return []string{command}
	}
	return nil
}

func toolInvocationText(toolName string, args any) string {
	m, ok := args.(map[string]any)
	if !ok || len(m) == 0 {
		if raw, ok := args.(map[string]interface{}); ok && len(raw) > 0 {
			m = map[string]any{}
			for k, v := range raw {
				m[k] = v
			}
		} else {
			return ""
		}
	}
	for _, key := range []string{"command", "cmd", "script", "path", "query"} {
		if v, ok := m[key]; ok {
			if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
				return truncate(strings.Join(strings.Fields(s), " "), 500)
			}
		}
	}
	payload, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return truncate(string(payload), 500)
}

func (c *chatTUI) toolResultBody(payload map[string]any, turnID, toolCallID, toolName string) []string {
	if output, _ := payload["output"].(string); strings.TrimSpace(output) != "" {
		if toolName == "codemode" {
			output = stripScriptHeader(output)
		}
		return toolOutputBodyLines(output)
	}
	return c.latestToolResultBody(turnID, toolCallID, toolName)
}

func toolOutputBodyLines(text string) []string {
	text = strings.TrimRight(plainTerminalOutput(text), "\r\n")
	if strings.TrimSpace(text) == "" {
		return []string{"(empty)"}
	}
	parts := strings.Split(text, "\n")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, strings.TrimRight(part, "\r"))
	}
	return out
}

func (c *chatTUI) renderToolEvent(payload map[string]any, ts time.Time) {
	c.ensureTranscriptBlockState()
	typ, _ := payload["type"].(string)
	toolName, _ := payload["tool"].(string)
	errText, _ := payload["error"].(string)
	reason, _ := payload["reason"].(string)
	turnID, _ := payload["turn_id"].(string)
	toolCallID, _ := payload["tool_call_id"].(string)
	if strings.TrimSpace(toolName) == "" {
		toolName = "tool"
	}
	// Codemode's nested calls are rows of its block, not tool blocks (Pi).
	if parent, _ := payload["parent_tool_call_id"].(string); parent != "" {
		c.updateCodemodeCall(parent, typ, payload, ts)
		return
	}
	startedAt := normalizeBlockTimestamp(ts)
	toolKey := c.toolRuntimeBlockKey(payload, toolName)
	blockKey := c.transcriptToolBlocks[toolKey]
	var previous transcriptBlockMeta
	if span, ok := c.transcriptBlockSpans[blockKey]; ok && span.HeaderIndex >= 0 && span.HeaderIndex < len(c.transcript) {
		previous, _ = parseTranscriptBlockMarker(c.transcript[span.HeaderIndex])
		// A terminal occurrence is immutable. Late starts and duplicate or
		// conflicting ends must not reopen it, append output, or move its time.
		if previous.EndedAt != "" {
			return
		}
	}
	// Pi renders the call (e.g. "$ ls -la", "read path") as the tool header;
	// the body holds only the result output.
	meta := transcriptBlockMeta{Key: blockKey, Kind: "tool", Title: toolName, Detail: previous.Detail, ToolPath: previous.ToolPath, ToolContent: previous.ToolContent, ToolRange: previous.ToolRange, ToolNotice: previous.ToolNotice, EditDiff: previous.EditDiff, EditError: previous.EditError, Calls: previous.Calls}
	if invocation := c.toolInvocationBody(toolName, payload); len(invocation) > 0 {
		meta.Detail = invocation[0]
	}
	setFileToolArguments(&meta, payload["arguments"])
	setCodemodeArguments(&meta, payload["arguments"])
	if calls, path := codemodeDetails(payload["details"]); calls != nil || path != "" {
		meta.Calls, meta.FullOutputPath = calls, path
	}
	var body []string
	switch typ {
	case "tool_started":
		c.promoteDraftToThinking(startedAt)
		c.finishThinkingTranscript(startedAt)
		c.markRunning()
		meta.Status = "running"
		meta.StartedAt = startedAt.Format(time.RFC3339Nano)
		c.setEditPreview(&meta, payload["arguments"])
		if meta.Key == "" {
			meta.Key = fmt.Sprintf("tool:%d:%s", time.Now().UnixNano(), toolName)
			c.transcriptToolBlocks[toolKey] = meta.Key
			c.appendTranscriptBlock(meta, body)
		} else {
			oldMeta, ok := parseTranscriptBlockMarker(c.transcript[c.transcriptBlockSpans[meta.Key].HeaderIndex])
			if ok && oldMeta.StartedAt != "" {
				meta.StartedAt = oldMeta.StartedAt
			}
			c.replaceTranscriptBlock(meta, body)
		}
	case "tool_finished", "tool_failed", "tool_skipped":
		if meta.Key == "" {
			meta.Key = fmt.Sprintf("tool:%d:%s", time.Now().UnixNano(), toolName)
		}
		existingBody := c.readTranscriptBlockBody(meta.Key)
		if span, ok := c.transcriptBlockSpans[meta.Key]; ok && span.HeaderIndex < len(c.transcript) {
			if oldMeta, ok := parseTranscriptBlockMarker(c.transcript[span.HeaderIndex]); ok && oldMeta.StartedAt != "" {
				meta.StartedAt = oldMeta.StartedAt
			}
		}
		// An end without its start has no measured duration. Keep it unknown
		// instead of inventing a zero-length call.
		meta.EndedAt = startedAt.Format(time.RFC3339Nano)
		if len(existingBody) > 0 {
			body = append([]string(nil), existingBody...)
		}
		resultBody := c.toolResultBody(payload, turnID, toolCallID, toolName)
		switch typ {
		case "tool_finished":
			meta.Status = "ok"
		case "tool_failed":
			meta.Status = "error"
			// The result text already carries the error (e.g. shell output and
			// "Command exited with code N"); show the bare error only without it.
			if errText != "" && len(resultBody) == 0 {
				body = append(body, "error="+truncate(errText, 160))
			}
		case "tool_skipped":
			meta.Status = "skipped"
			if reason != "" {
				body = append(body, "reason="+truncate(reason, 160))
			}
		}
		body = append(body, resultBody...)
		setFileToolDetails(&meta, payload["details"], strings.TrimSpace(strings.Join(resultBody, "\n")))
		if meta.Key == "" {
			c.appendTranscriptBlock(meta, body)
		} else {
			c.replaceTranscriptBlock(meta, body)
		}
		c.transcriptToolBlocks[toolKey] = meta.Key
	}
}

func (c *chatTUI) renderRoutingEvent(payload map[string]any, ts time.Time) {
	typ, _ := payload["type"].(string)
	targetAgent, _ := payload["target_agent_id"].(string)
	sourceAgent, _ := payload["source_agent_id"].(string)
	switch typ {
	case "routing_decision":
		if targetAgent != "" {
			c.status = fmt.Sprintf("routed to @%s", targetAgent)
		}
	case "routing_incoming":
		if sourceAgent != "" {
			c.status = fmt.Sprintf("incoming route from @%s", sourceAgent)
		}
	}
}

func (c *chatTUI) resetRunningDraftState() {
	c.clearThinkingIndicator()
	c.finishThinkingTranscript(time.Now().UTC())
	// Preserve any streamed assistant draft when terminal/runtime status events
	// arrive before the final response event. This keeps partial/final output in
	// the timeline instead of erasing it on turn_completed/session_idle.
	c.running = false
	c.draft = ""
}

func intFromEvent(ev map[string]any, key string) int {
	return intFromAny(ev[key])
}

func cloneAnyMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case int32:
		return int(n)
	case float64:
		return int(n)
	case float32:
		return int(n)
	default:
		return 0
	}
}

func floatFromAny(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}

func (c *chatTUI) KeyMap() gotui.KeyMap {
	if c.workspaceIndex.active {
		return c.workspaceIndexKeys()
	}
	if c.search.active {
		return c.transcriptSearchKeys()
	}
	if c.modelMenuOpen && c.modelMenuKind == "session-rename" {
		return c.sessionRenameKeys()
	}
	if c.modelMenuOpen && c.modelMenuKind == "fork" {
		return c.forkSelectorKeys()
	}
	if c.modelMenuOpen && c.modelMenuKind == "select" {
		return c.selectDialogKeys()
	}
	if c.modelMenuOpen && c.modelMenuKind == "loader" && c.loader != nil {
		return c.loaderKeys()
	}
	if c.modelMenuOpen && c.modelMenuKind == "mcp-manager" && c.mcpManager != nil {
		return c.mcpManagerKeys()
	}
	if c.modelMenuOpen && c.modelMenuKind == "scoped-models" && c.scopedModels != nil {
		return c.scopedModelsKeys()
	}
	if c.modelMenuOpen && c.modelMenuKind == "auth-selector" && c.authSelector != nil {
		return c.authSelectorKeys()
	}
	if c.modelMenuOpen && c.modelMenuKind == "login-dialog" && c.loginDialog != nil {
		return c.loginDialogKeys()
	}
	if c.modelMenuOpen && c.modelMenuKind == "settings" && c.settingsList != nil {
		return c.settingsKeys()
	}
	if c.modelMenuOpen && c.modelMenuKind == "tree" && c.treeSelector != nil {
		return treeSelectorKeys(c.treeSelector, c.markDirty)
	}
	if c.modelMenuOpen && c.modelMenuKind == "editor-dialog" && c.editorDialog != nil {
		return c.editorDialogKeys()
	}
	if c.modelMenuOpen {
		if c.modelMenuKind == "thinking" {
			return gotui.KeyMap{
				gotui.OnPreemptStop(gotui.KeyCtrlC, func(ke gotui.KeyEvent) { c.closeModelMenu() }),
				gotui.OnPreemptStop(gotui.KeyEscape, func(ke gotui.KeyEvent) { c.closeModelMenu() }),
				gotui.OnPreemptStop(gotui.Rune('s').Ctrl(), func(ke gotui.KeyEvent) { c.acceptThinkingMenuAsDefault() }),
				gotui.OnPreemptStop(gotui.KeyUp, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(-1) }),
				gotui.OnPreemptStop(gotui.KeyDown, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(1) }),
				gotui.OnPreemptStop(gotui.KeyEnter, func(ke gotui.KeyEvent) { c.acceptModelMenuSelection() }),
				gotui.OnPreemptStop(gotui.KeyBackspace, func(ke gotui.KeyEvent) { c.modelMenuBackspace() }),
				gotui.OnFocused(gotui.AnyRune, func(ke gotui.KeyEvent) { c.modelMenuTypeRune(ke.Rune) }),
			}
		}
		if c.modelMenuKind == "model" {
			// Pi's selector: Escape/Ctrl+C cancel, Tab scope, Ctrl+S save default.
			return gotui.KeyMap{
				gotui.OnPreemptStop(gotui.KeyCtrlC, func(ke gotui.KeyEvent) { c.closeModelMenu() }),
				gotui.OnPreemptStop(gotui.KeyEscape, func(ke gotui.KeyEvent) { c.closeModelMenu() }),
				gotui.OnPreemptStop(gotui.KeyTab, func(ke gotui.KeyEvent) { c.toggleModelMenuScope() }),
				gotui.OnPreemptStop(gotui.Rune('s').Ctrl(), func(ke gotui.KeyEvent) { c.acceptModelMenuAsDefault() }),
				gotui.OnPreemptStop(gotui.KeyUp, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(-1) }),
				gotui.OnPreemptStop(gotui.KeyDown, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(1) }),
				gotui.OnPreemptStop(gotui.KeyPageUp, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(-5) }),
				gotui.OnPreemptStop(gotui.KeyPageDown, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(5) }),
				gotui.OnPreemptStop(gotui.KeyHome, func(ke gotui.KeyEvent) { c.setModelMenuSelection(0) }),
				gotui.OnPreemptStop(gotui.KeyEnd, func(ke gotui.KeyEvent) { c.setModelMenuSelection(len(c.modelMenuChoices) - 1) }),
				gotui.OnPreemptStop(gotui.KeyEnter, func(ke gotui.KeyEvent) { c.acceptModelMenuSelection() }),
				gotui.OnPreemptStop(gotui.KeyBackspace, func(ke gotui.KeyEvent) { c.modelMenuBackspace() }),
				gotui.OnFocused(gotui.AnyRune, func(ke gotui.KeyEvent) { c.modelMenuTypeRune(ke.Rune) }),
			}
		}
		return gotui.KeyMap{
			gotui.OnStop(gotui.KeyCtrlC, func(ke gotui.KeyEvent) { c.app.Stop() }),
			gotui.OnPreemptStop(gotui.KeyEscape, func(ke gotui.KeyEvent) { c.backFromSessionActions() }),
			gotui.OnPreemptStop(gotui.KeyRight, func(ke gotui.KeyEvent) { c.openSessionActions() }),
			gotui.OnPreemptStop(gotui.KeyLeft, func(ke gotui.KeyEvent) {
				if c.modelMenuKind == "session-actions" {
					c.backFromSessionActions()
				}
			}),
			gotui.OnPreemptStop(gotui.KeyUp, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(-1) }),
			gotui.OnPreemptStop(gotui.KeyDown, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(1) }),
			gotui.OnPreemptStop(gotui.KeyPageUp, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(-5) }),
			gotui.OnPreemptStop(gotui.KeyPageDown, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(5) }),
			gotui.OnPreemptStop(gotui.KeyHome, func(ke gotui.KeyEvent) { c.setModelMenuSelection(0) }),
			gotui.OnPreemptStop(gotui.KeyEnd, func(ke gotui.KeyEvent) { c.setModelMenuSelection(len(c.modelMenuChoices) - 1) }),
			gotui.OnPreemptStop(gotui.KeyEnter, func(ke gotui.KeyEvent) { c.acceptModelMenuSelection() }),
			gotui.OnPreemptStop(gotui.KeyBackspace, func(ke gotui.KeyEvent) { c.modelMenuBackspace() }),
			gotui.OnFocused(gotui.AnyRune, func(ke gotui.KeyEvent) { c.modelMenuTypeRune(ke.Rune) }),
		}
	}
	keys := defaultPiKeys
	pasteKey, _ := keys.pasteImage()
	cycleBackKey, _ := keys.cycleModelBackward()
	bindings := gotui.KeyMap{
		// Pi's app.* bindings (docs/internal/keybindings.md).
		gotui.OnPreemptStop(gotui.KeyCtrlC, func(gotui.KeyEvent) { c.handleCtrlC() }),
		gotui.OnPreemptStop(gotui.Rune('x').Ctrl(), func(gotui.KeyEvent) { c.copySelectionOrLastMessage() }),
		gotui.OnStop(gotui.KeyCtrlD, func(ke gotui.KeyEvent) {
			if c.input.Text() == "" {
				c.app.Stop()
			} else { // Pi's tui.editor.deleteCharForward
				c.input.jumpMode, c.input.lastAction = "", ""
				c.input.delete()
			}
		}),
		gotui.OnPreemptStop(gotui.KeyCtrlG, func(gotui.KeyEvent) { c.openExternalEditor() }),
		gotui.OnPreemptStop(pasteKey, func(gotui.KeyEvent) { c.pasteClipboard() }),
		gotui.OnPreemptStop(gotui.KeyTab.Shift(), func(gotui.KeyEvent) { c.cycleThinking(1) }),
		gotui.OnPreemptStop(gotui.KeyEscape, func(ke gotui.KeyEvent) {
			if c.textSelection.active {
				c.clearTranscriptSelection()
				return
			}
			if c.compaction.active {
				c.stopCompaction()
				return
			}
			if c.editorAskActive {
				c.cancelEditorAsk()
				return
			}
			c.inputActive = false
			if c.app != nil {
				c.app.BlurFocused()
			}
		}),
		gotui.OnStop(gotui.KeyTab, func(ke gotui.KeyEvent) { c.focusInput() }),
		gotui.OnPreemptStop(keys.search(), func(ke gotui.KeyEvent) { c.toggleTranscriptSearch() }),
		gotui.OnPreemptStop(gotui.KeyUp.Ctrl(), func(ke gotui.KeyEvent) { c.jumpTranscriptPrompt(-1) }),
		gotui.OnPreemptStop(gotui.KeyDown.Ctrl(), func(ke gotui.KeyEvent) { c.jumpTranscriptPrompt(1) }),
		gotui.OnPreemptStop(gotui.KeyPageUp, func(ke gotui.KeyEvent) { c.pageTranscript(-1) }),
		gotui.OnPreemptStop(gotui.KeyPageDown, func(ke gotui.KeyEvent) { c.pageTranscript(1) }),
		gotui.OnPreemptStop(gotui.KeyHome, func(ke gotui.KeyEvent) { c.scrollTranscriptToTop() }),
		gotui.OnPreemptStop(gotui.KeyEnd, func(ke gotui.KeyEvent) { c.scrollTranscriptToBottom() }),
		gotui.OnPreemptStop(gotui.Rune('o').Ctrl(), func(ke gotui.KeyEvent) { c.toggleToolOutput() }),
		gotui.OnPreemptStop(gotui.KeyF6, func(ke gotui.KeyEvent) { c.selectTranscriptBlock(-1) }),
		gotui.OnPreemptStop(gotui.KeyF7, func(ke gotui.KeyEvent) { c.selectTranscriptBlock(1) }),
		gotui.OnPreemptStop(gotui.KeyF8, func(ke gotui.KeyEvent) { c.toggleSelectedTranscriptBlock() }),
		gotui.OnPreemptStop(gotui.KeyF2, func(ke gotui.KeyEvent) { c.recallHistory(-1) }),
		gotui.OnPreemptStop(gotui.KeyF3, func(ke gotui.KeyEvent) { c.recallHistory(1) }),
		gotui.OnPreemptStop(gotui.Rune('p').Ctrl(), func(ke gotui.KeyEvent) { c.cycleModel(1) }),
		gotui.OnPreemptStop(cycleBackKey, func(ke gotui.KeyEvent) { c.cycleModel(-1) }),
		gotui.OnPreemptStop(gotui.KeyCtrlL, func(ke gotui.KeyEvent) { c.openModelMenu() }),
		gotui.OnPreemptStop(gotui.KeyCtrlT, func(ke gotui.KeyEvent) { c.toggleThinkingBlocks() }),
		// gi-only, on keys Pi leaves free:
		gotui.OnPreemptStop(gotui.Rune('l').Alt(), func(ke gotui.KeyEvent) { c.cycleModel(-1) }),
		gotui.OnPreemptStop(gotui.KeyCtrlR, func(ke gotui.KeyEvent) { c.searchHistoryBackward() }),
		gotui.OnPreemptStop(gotui.Rune('i').Alt(), func(gotui.KeyEvent) { c.openWorkspaceIndex() }),
		gotui.OnPreemptStop(gotui.Rune('s').Alt(), func(ke gotui.KeyEvent) { c.openSessionMenu() }),
		gotui.OnPreemptStop(gotui.Rune('c').Alt(), func(ke gotui.KeyEvent) { c.startCompaction() }),
		gotui.OnPreemptStop(gotui.Rune('m').Alt(), func(ke gotui.KeyEvent) { c.openModelMenu() }),
		gotui.OnPreemptStop(gotui.Rune('t').Alt(), func(ke gotui.KeyEvent) { c.cycleThinking(-1) }),
		gotui.OnPreemptStop(gotui.KeyUp, func(ke gotui.KeyEvent) {
			c.recallHistory(-1)
		}),
		gotui.OnPreemptStop(gotui.KeyDown, func(ke gotui.KeyEvent) {
			c.recallHistory(1)
		}),
	}
	if !keys.windows { // Pi's app.suspend (none on Windows, where Ctrl+Z undoes)
		bindings = append(bindings, gotui.OnPreemptStop(gotui.KeyCtrlZ, func(gotui.KeyEvent) { c.suspend() }))
	}
	if !keys.windowsKeys() { // Pi's tui.altScreen.previousPrompt/nextPrompt
		bindings = append(bindings,
			gotui.OnPreemptStop(gotui.KeyUp.Ctrl().Shift(), func(ke gotui.KeyEvent) { c.jumpTranscriptPrompt(-1) }),
			gotui.OnPreemptStop(gotui.KeyDown.Ctrl().Shift(), func(ke gotui.KeyEvent) { c.jumpTranscriptPrompt(1) }))
	}
	if slash := c.slashMenuKeys(); slash != nil {
		bindings = append(slash, bindings...)
	}
	if c.regularMode {
		return regularKeyMap(bindings)
	}
	return bindings
}

func (c *chatTUI) openModelMenu() {
	choices := c.availableModelChoices()
	if len(choices) == 0 {
		c.appendTranscript("sys: no available models; use /model <provider/model>")
		return
	}
	selected := 0
	current := canonicalModelRef(c.cfg.DefaultProvider, c.cfg.DefaultModel)
	for i, model := range choices {
		if canonicalModelRef(c.cfg.DefaultProvider, model) == current {
			selected = i
			break
		}
	}
	c.modelMenuError = ""
	c.modelMenuOpen = true
	c.modelMenuKind = "model"
	c.modelMenuDefault = c.modelMenuDefaultLabel()
	c.modelMenuScope = "all"
	if scoped, _ := c.modelMenuScopes(); len(scoped) > 0 {
		// Pi opens on the scoped (enabled) models when there are any.
		c.modelMenuScope = "scoped"
		choices = scoped
		selected = 0
		for i, model := range choices {
			if canonicalModelRef(c.cfg.DefaultProvider, model) == current {
				selected = i
			}
		}
	}
	c.captureModelPickerMetadata()
	c.openModelPickerScreen()
	c.modelMenuValues = nil
	c.modelMenuAll = choices
	c.modelMenuQuery = ""
	c.modelMenuChoices = choices
	c.modelMenuSelected = selected
	c.modelMenuScroll = 0
	c.ensureModelMenuSelectionVisible()
	c.inputActive = false
	if c.app != nil {
		c.app.BlurFocused()
		c.app.MarkDirty()
	}
}

func (c *chatTUI) openSessionMenu() {
	sessions, err := c.store.ListSessions(context.Background())
	if err != nil {
		c.appendTranscript(fmt.Sprintf("error: list sessions: %v", err))
		return
	}
	if len(sessions) == 0 {
		c.appendTranscript("sys: no sessions to resume")
		return
	}
	labels := make([]string, 0, len(sessions))
	values := map[string]string{}
	c.modelMenuSessionRows = map[string]sessionPickerRow{}
	selected := 0
	for i := range sessions {
		sess := sessions[i]
		label := c.sessionPickerLabel(&sess)
		labels = append(labels, label)
		values[label] = sess.ID
		c.modelMenuSessionRows[label] = c.sessionPickerRowFor(&sess)
		if sess.ID == c.sessionID {
			selected = i
		}
	}
	c.modelMenuError = ""
	c.modelMenuOpen = true
	c.modelMenuKind = "session"
	c.modelMenuSession = c.selectionScope()
	c.openModelPickerScreen()
	c.modelMenuValues = values
	c.modelMenuAll = labels
	c.modelMenuQuery = ""
	c.modelMenuChoices = labels
	c.modelMenuSelected = selected
	c.modelMenuScroll = 0
	c.ensureModelMenuSelectionVisible()
	c.inputActive = false
	if c.app != nil {
		c.app.BlurFocused()
		c.app.MarkDirty()
	}
}

func fuzzyMatch(query, candidate string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	candidate = strings.ToLower(candidate)
	// Each whitespace-separated token must appear as a substring; this keeps
	// filtering intuitive ("gpt" matches only gpt models) while still allowing
	// multi-term queries like "openai mini".
	for _, token := range strings.Fields(query) {
		if !strings.Contains(candidate, token) {
			return false
		}
	}
	return true
}

func filterModelMenuChoices(all []string, query string) []string {
	if strings.TrimSpace(query) == "" {
		return append([]string(nil), all...)
	}
	out := make([]string, 0, len(all))
	for _, model := range all {
		if fuzzyMatch(query, model) {
			out = append(out, model)
		}
	}
	return out
}

func (c *chatTUI) applyModelMenuFilter() {
	c.modelMenuChoices = filterModelMenuChoices(c.modelMenuAll, c.modelMenuQuery)
	if c.modelMenuKind == "model" {
		// Pi's model selector: token fuzzy filter over id/provider/name
		// (plus gi's context/reasoning metadata), best matches first.
		items := make([]slashItem, len(c.modelMenuAll))
		for i, label := range c.modelMenuAll {
			items[i] = slashItem{name: label, description: label + " " + c.modelMenuMetadata[label].search}
		}
		c.modelMenuChoices = nil
		for _, it := range piFuzzyFilter(items, c.modelMenuQuery, func(it slashItem) string { return it.description }) {
			c.modelMenuChoices = append(c.modelMenuChoices, it.name)
		}
	}
	c.modelMenuSelected = 0
	c.modelMenuScroll = 0
	c.setModelMenuSelection(0)
	c.ensureModelMenuSelectionVisible()
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) modelMenuTypeRune(r rune) {
	if c.modelMenuKind == "session-actions" {
		return
	}
	if r == 0 {
		return
	}
	c.modelMenuError = ""
	c.modelMenuQuery += string(r)
	c.applyModelMenuFilter()
}

func (c *chatTUI) modelMenuBackspace() {
	if c.modelMenuKind == "session-actions" {
		return
	}
	if c.modelMenuQuery == "" {
		return
	}
	c.modelMenuError = ""
	q := []rune(c.modelMenuQuery)
	c.modelMenuQuery = string(q[:len(q)-1])
	c.applyModelMenuFilter()
}

func (c *chatTUI) closeModelMenu() {
	// Hide before the screen-restore resize is dispatched.
	c.modelMenuOpen = false
	c.sessionActions = sessionActions{}
	c.scopedModels = nil
	c.authSelector = nil
	c.settingsList = nil
	if c.mcpManager != nil {
		c.mcpManager = nil
		if c.engine != nil {
			c.engine.SetMCPChangeListener(nil)
		}
	}
	if loader := c.loader; loader != nil {
		// Closed by something else (session switch, quit): the work stops.
		c.loader = nil
		if loader.onAbort != nil {
			defer loader.onAbort()
		}
	}
	if c.selectDialog.onCancel != nil || c.selectDialog.onSelect != nil {
		// Closed by something else (session switch, quit): Pi resolves undefined.
		cancel := c.selectDialog.onCancel
		c.selectDialog = selectDialog{}
		if cancel != nil {
			defer cancel()
		}
	}
	c.closeModelPickerScreen()
	c.resetModelMenuMetadata()
	c.modelMenuError = ""
	c.modelMenuKind = ""
	c.modelMenuScope, c.modelMenuDefault = "", ""
	c.modelMenuSessionRows = nil
	c.modelMenuValues = nil
	c.modelMenuChoices = nil
	c.modelMenuAll = nil
	c.modelMenuQuery = ""
	c.modelMenuSelected = 0
	c.modelMenuScroll = 0
	if c.input != nil {
		c.input.suspended = false
		c.input.Focus()
	}
	c.focusInput()
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) moveModelMenuSelection(delta int) {
	if c.modelMenuKind == "model" {
		enabled := c.enabledModelMenuIndices()
		if len(enabled) == 0 {
			return
		}
		position := 0
		for i, index := range enabled {
			if index == c.modelMenuSelected {
				position = i
				break
			}
		}
		position += delta
		if delta == 1 || delta == -1 {
			position = (position + len(enabled)) % len(enabled)
		}
		position = max(0, min(position, len(enabled)-1))
		c.setModelMenuSelection(enabled[position])
		return
	}
	index := c.modelMenuSelected + delta
	if len(c.modelMenuChoices) > 0 && (delta == 1 || delta == -1) {
		index = (index + len(c.modelMenuChoices)) % len(c.modelMenuChoices)
	}
	c.setModelMenuSelection(index)
}

func (c *chatTUI) setModelMenuSelection(idx int) {
	if c.modelMenuKind == "model" {
		enabled := c.enabledModelMenuIndices()
		if len(enabled) == 0 {
			c.modelMenuSelected = -1
			return
		}
		selected := enabled[len(enabled)-1]
		for _, index := range enabled {
			if index >= idx {
				selected = index
				break
			}
		}
		idx = selected
	}
	if len(c.modelMenuChoices) == 0 {
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(c.modelMenuChoices) {
		idx = len(c.modelMenuChoices) - 1
	}
	c.modelMenuSelected = idx
	c.ensureModelMenuSelectionVisible()
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) modelMenuVisibleRows() int {
	height, width := c.outputHeight, c.outputWidth
	if c.app != nil {
		width, height = c.app.Size()
	}
	if height == 0 {
		height = 24
	}
	if width == 0 {
		width = 80
	}
	// Pi's session selector and actions list replace the editor: their rows
	// share the screen with the spacer and footer only.
	if c.modelMenuKind == "session" || c.modelMenuKind == "session-actions" {
		chrome := 11
		if c.modelMenuKind == "session-actions" {
			chrome = 10
		}
		if !c.regularMode {
			chrome += 1 + len(c.footerLines(width))
		}
		return max(1, min(6, height-chrome))
	}
	padding := 0 // Pi draws transcript, editor and footer edge to edge.
	inputRows := 1
	if c.input != nil {
		input := *c.input
		input.width = max(1, width-padding)
		input.maxLines = editorViewportRows(height, height-padding-len(c.footerLines(width))-len(c.pendingDockLines(width-padding, height))-len(c.extensionWidgetLines())-1-2-4-3)
		inputRows = max(1, len(input.renderLines()))
	}
	// Leave transcript, editor/separators and the existing footer intact.
	available := height - padding - len(c.footerLines(width)) - len(c.pendingDockLines(width-padding, height)) - len(c.extensionWidgetLines()) - 1 - inputRows - 2 - 4 - 2
	return min(6, max(1, available))
}

func (c *chatTUI) ensureModelMenuSelectionVisible() {
	if c.modelMenuKind == "model" && c.modelMenuSelected >= 0 && c.modelMenuSelected < len(c.modelMenuChoices) && c.modelPickerUnavailable(c.modelMenuChoices[c.modelMenuSelected]) != "" {
		c.modelMenuSelected = -1
		for _, index := range c.enabledModelMenuIndices() {
			c.modelMenuSelected = index
			break
		}
	}
	rows := c.modelMenuVisibleRows()
	if c.modelMenuSelected < c.modelMenuScroll {
		c.modelMenuScroll = c.modelMenuSelected
	}
	if c.modelMenuSelected >= c.modelMenuScroll+rows {
		c.modelMenuScroll = c.modelMenuSelected - rows + 1
	}
	if c.modelMenuScroll < 0 {
		c.modelMenuScroll = 0
	}
	maxScroll := len(c.modelMenuChoices) - rows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if c.modelMenuScroll > maxScroll {
		c.modelMenuScroll = maxScroll
	}
}

func (c *chatTUI) acceptModelMenuSelection() {
	if c.modelMenuKind == "fork" {
		c.acceptForkSelection()
		return
	}
	if c.modelMenuKind == "session-actions" {
		c.applySessionAction()
		return
	}
	if c.modelMenuOpen && (c.modelMenuKind == "model" || c.modelMenuKind == "session") {
		if !c.ownsScope(c.modelMenuSession) {
			c.modelMenuError = "session changed; reopen picker"
			if c.app != nil {
				c.app.MarkDirty()
			}
			return
		}
		if c.modelMenuKind == "model" && len(c.modelMenuChoices) > 0 {
			label := c.modelMenuChoices[max(0, min(c.modelMenuSelected, len(c.modelMenuChoices)-1))]
			if c.modelPickerUnavailable(label) != "" {
				// Enter on a blocked-only result is an explicit metadata retry.
				// Recovery only highlights it; a second Enter is required to apply.
				c.captureModelPickerMetadata()
				if reason := c.modelPickerUnavailable(label); reason != "" {
					c.modelMenuError = "model unavailable: " + reason
				} else {
					c.modelMenuError = ""
					c.setModelMenuSelection(max(0, c.modelMenuSelected))
				}
				if c.app != nil {
					c.app.MarkDirty()
				}
				return
			}
		}
	}
	if !c.modelMenuOpen || len(c.modelMenuChoices) == 0 || c.modelMenuSelected < 0 || c.modelMenuSelected >= len(c.modelMenuChoices) {
		return
	}
	label := c.modelMenuChoices[c.modelMenuSelected]
	kind := c.modelMenuKind
	value := label
	if c.modelMenuValues != nil {
		if v, ok := c.modelMenuValues[label]; ok {
			value = v
		}
	}
	if kind == "model" {
		if err := c.chooseSessionModel(value); err != nil {
			c.modelMenuError = err.Error()
			if c.app != nil {
				c.app.MarkDirty()
			}
			return
		}
		c.closeModelMenu()
		return
	}
	if kind == "session" {
		// Switch performs the authoritative read before touching the origin.
		// Keep the selector/screen until that single validation succeeds.
		if !c.switchSession(value) {
			c.modelMenuError = "session unavailable; Enter to retry or Esc to close"
			if c.app != nil {
				c.app.MarkDirty()
			}
			return
		}
		c.closeModelMenu()
		return
	}
	c.closeModelMenu()
	switch kind {
	case "thinking":
		c.appendTranscript(c.thinkingCommand([]string{"/thinking", value})...)
	default:
		c.appendTranscript(c.modelCommand([]string{"/model", value})...)
	}
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) modelMenuHeight() int {
	if c.modelMenuOpen && c.modelMenuKind == "session-rename" {
		return 4
	}
	if !c.modelMenuOpen {
		return 0
	}
	if c.modelMenuKind == "model" || c.modelMenuKind == "thinking" || c.modelMenuKind == "session" || c.modelMenuKind == "session-actions" || c.modelMenuKind == "fork" || c.modelMenuKind == "select" || c.modelMenuKind == "mcp-manager" || c.modelMenuKind == "scoped-models" || c.modelMenuKind == "auth-selector" || c.modelMenuKind == "login-dialog" || c.modelMenuKind == "settings" || c.modelMenuKind == "tree" || c.modelMenuKind == "editor-dialog" {
		width := c.currentContentWidth()
		if c.app != nil {
			width, _ = c.app.Size()
		}
		if c.modelMenuKind == "model" {
			return len(c.piModelSelectorRows(width))
		}
		rows, _ := c.piMenuRows(width)
		return len(rows)
	}
	rows := c.modelMenuVisibleRows()
	if len(c.modelMenuChoices) < rows {
		rows = len(c.modelMenuChoices)
	}
	// Title + search + at least one result/empty-state row; no box border.
	return max(1, rows) + 2
}

func (c *chatTUI) renderModelMenu(width int) *gotui.Element {
	if c.modelMenuKind == "model" {
		return c.renderPiModelSelector(width)
	}
	if rows, ok := c.piMenuRows(width); ok {
		return renderSpanRows(rows)
	}
	if c.modelMenuKind == "session-rename" {
		return c.renderSessionRename(width)
	}
	c.ensureModelMenuSelectionVisible()
	rows := c.modelMenuVisibleRows()
	start := c.modelMenuScroll
	end := start + rows
	if end > len(c.modelMenuChoices) {
		end = len(c.modelMenuChoices)
	}
	menu := gotui.New(
		gotui.WithWidthPercent(100),
		gotui.WithHeight(c.modelMenuHeight()),
		gotui.WithDirection(gotui.Column),
		gotui.WithPaddingTRBL(0, 0, 0, 0),
	)
	current := strings.TrimSpace(c.cfg.DefaultModel)
	noun := "model"
	if c.modelMenuKind == "session" {
		noun = "session"
	} else if c.modelMenuKind == "thinking" {
		noun = "thinking level"
	}
	title := "Select " + noun + " · ↑/↓ navigate · Enter select · Esc cancel"
	if width < 72 {
		title = "Select " + noun + " · ↑/↓ Enter Esc"
	}
	if c.modelMenuKind == "session" {
		title = "Select session · ↑↓ Enter · → actions · Esc"
	}
	if c.modelMenuKind == "session-actions" {
		title = "Session actions · ↑↓ Enter · Esc back"
	}
	if c.modelMenuKind == "model" && current != "" {
		title += " · current " + compactMaybe(current, c.compactOutput(), 28)
	}
	menu.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(selectorText(title, width)), gotui.WithTextStyle(gotui.NewStyle().Bold())))
	search := "search: " + c.modelMenuQuery + "▌"
	if strings.TrimSpace(c.modelMenuQuery) == "" {
		search = "search: (type to filter)"
	} else {
		search += fmt.Sprintf("  (%d match)", len(c.modelMenuChoices))
	}
	if c.modelMenuKind == "session-actions" {
		search = "target: " + c.sessionActions.targetLabel
	}
	if c.modelMenuError != "" {
		search = "error: " + c.modelMenuError
	}
	menu.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(selectorText(search, width)), gotui.WithTextStyle(piFg(piMuted))))
	if len(c.modelMenuChoices) == 0 {
		menu.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(selectorText("  no matching "+noun+"s", width)), gotui.WithTextStyle(piFg(piMuted))))
		return menu
	}
	for i := start; i < end; i++ {
		model := c.modelMenuChoices[i]
		prefix := "  "
		style := gotui.NewStyle()
		if i == c.modelMenuSelected {
			prefix = "› "
			style = piFg(piAccent)
		} else if canonicalModelRef(c.cfg.DefaultProvider, model) == canonicalModelRef(c.cfg.DefaultProvider, c.cfg.DefaultModel) {
			prefix = "* "
			style = piFg(piSuccess)
		}
		prefix = fmt.Sprintf("%s%d. ", prefix, i+1)
		label := prefix + c.modelPickerRowLabel(model, width-gotui.StringWidth(prefix))
		if reason := c.modelPickerUnavailable(model); reason != "" {
			label = fmt.Sprintf("× %d. %s · %s", i+1, model, reason)
			style = piFg(piDim)
		}
		menu.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(selectorText(label, width)), gotui.WithTextStyle(style)))
	}
	return menu
}

// cycleModel is Pi's cycleModel: through the scoped models (Pi's
// enabledModels) that are available, else through every available model.
func (c *chatTUI) cycleModel(delta int) {
	models, _ := c.modelMenuScopes()
	only := "Only one model in scope"
	if len(models) == 0 {
		_, models = c.modelMenuScopes()
		only = "Only one model available"
	}
	if len(models) <= 1 {
		c.selectionNotice(only)
		return
	}
	current := canonicalModelRef(c.cfg.DefaultProvider, c.cfg.DefaultModel)
	idx := 0 // Pi: a current model outside the list counts as the first
	for i, model := range models {
		if model == current {
			idx = i
			break
		}
	}
	idx = (idx + delta + len(models)) % len(models)
	c.appendTranscript(c.modelCommand([]string{"/model", models[idx]})...)
}

func (c *chatTUI) cycleThinking(delta int) {
	// Pi cycles every level the model supports.
	levels := inference.ThinkingLevels(c.sessionModelLabel())
	if len(levels) == 0 {
		if known, reasoning := inference.ModelReasoning(c.sessionModelLabel()); known && !reasoning {
			c.appendTranscript("sys: current model does not support thinking")
			return
		}
		levels = []string{"low", "medium", "high"}
	}
	current := strings.ToLower(strings.TrimSpace(c.cfg.DefaultThinkingLevel))
	idx := 0
	for i, level := range levels {
		if level == current {
			idx = i
			break
		}
	}
	idx = (idx + delta) % len(levels)
	if idx < 0 {
		idx += len(levels)
	}
	c.appendTranscript(c.thinkingCommand([]string{"/thinking", levels[idx]})...)
}

func (c *chatTUI) loadCommandHistory() []string {
	if c.store == nil || strings.TrimSpace(c.sessionID) == "" {
		return nil
	}
	history, err := c.store.ListTUIInputHistory(context.Background(), c.sessionID, c.currentHistoryLimit())
	if err != nil {
		return nil
	}
	return history
}

// Only prompts submitted to the turn engine belong in cursor history. Unsent
// drafts, slash commands and local shell shortcuts are not prompts.
func (c *chatTUI) recordInputHistory(text string) {
	c.history = append(c.history, text)
	c.applyHistoryLimit()
	c.histIdx = -1
	c.historyDraft = ""
	c.historyDraftCursor = 0
	c.historySearchIdx = -1
	c.historySearchQuery = ""
	if c.store != nil && c.sessionID != "" {
		if err := c.store.RecordTUIInput(context.Background(), c.sessionID, text, c.currentHistoryLimit()); err != nil {
			c.appendTranscript("warn: input history not saved: " + err.Error())
		}
	}
}

func (c *chatTUI) currentHistoryLimit() int {
	if c.cfg.TUIHistoryLimit > 0 {
		return c.cfg.TUIHistoryLimit
	}
	return 10000
}

func (c *chatTUI) applyHistoryLimit() {
	limit := c.currentHistoryLimit()
	if limit <= 0 || len(c.history) <= limit {
		return
	}
	c.history = append([]string(nil), c.history[len(c.history)-limit:]...)
}

func (c *chatTUI) searchHistoryBackward() {
	if c.input == nil || len(c.history) == 0 {
		return
	}
	query := strings.TrimSpace(c.input.Text())
	start := len(c.history) - 1
	if c.historySearchIdx >= 0 && c.historySearchIdx < len(c.history) &&
		c.input.Text() == c.history[c.historySearchIdx] {
		query = c.historySearchQuery
		start = c.historySearchIdx - 1
	}
	if query == "" {
		c.recallHistory(-1)
		return
	}
	needle := strings.ToLower(query)
	for offset := 0; offset < len(c.history); offset++ {
		i := (start - offset + len(c.history)*2) % len(c.history)
		if strings.Contains(strings.ToLower(c.history[i]), needle) {
			if c.histIdx < 0 {
				c.historyDraft, c.historyDraftCursor = c.input.Text(), c.input.cursorPos
			}
			c.historySearchQuery, c.historySearchIdx = query, i
			c.histIdx = i
			c.setHistoryInput(c.history[i], utf8.RuneCountInString(c.history[i]))
			c.status = fmt.Sprintf("History match %d/%d", i+1, len(c.history))
			return
		}
	}
	c.status = fmt.Sprintf("No history match for %q", query)
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) setHistoryInput(text string, cursor int) {
	c.historyApplying = true
	c.focusInput()
	c.input.SetText(text)
	c.input.cursorPos = cursor
	c.historyApplying = false
}

func (c *chatTUI) recallHistory(delta int) {
	if c.input == nil || len(c.history) == 0 {
		return
	}
	if delta < 0 {
		if c.histIdx < 0 {
			c.historyDraft, c.historyDraftCursor = c.input.Text(), c.input.cursorPos
			c.histIdx = len(c.history) - 1
		} else if c.histIdx > 0 {
			c.histIdx--
		}
	} else {
		if c.histIdx < 0 {
			return
		}
		c.histIdx++
		if c.histIdx >= len(c.history) {
			c.histIdx = -1
			c.setHistoryInput(c.historyDraft, c.historyDraftCursor)
			c.historyDraft = ""
			c.historyDraftCursor = 0
			c.historySearchIdx = -1
			c.historySearchQuery = ""
			return
		}
	}
	c.historySearchIdx = -1
	c.historySearchQuery = ""
	c.setHistoryInput(c.history[c.histIdx], utf8.RuneCountInString(c.history[c.histIdx]))
}

func (c *chatTUI) HandleMouse(me gotui.MouseEvent) bool {
	if me.Button != gotui.MouseLeft {
		c.selectionClicks = transcriptClickSequence{}
		c.selectionClickSnapshot = transcriptSelection{}
	}
	if c.modelMenuOpen && c.modelMenuKind == "session-rename" {
		return true
	}
	if c.workspaceIndex.active {
		return true
	}
	if c.handleJumpToLatestClick(me) {
		return true
	}
	if c.handleTranscriptSelection(me) {
		return true
	}
	if c.handleTranscriptScrollEvent(me) {
		return true
	}
	if me.Button != gotui.MouseLeft || me.Action != gotui.MousePress {
		return false
	}
	if c.handleTranscriptBlockClick(me) {
		return true
	}
	if c.inputRegion != nil && c.inputRegion.ContainsPoint(me.X, me.Y) {
		c.focusInput()
		return true
	}
	if c.app != nil {
		_, h := c.app.Size()
		if me.Y >= h-4 {
			c.focusInput()
			return true
		}
	}
	return false
}

func (c *chatTUI) handleTranscriptBlockClick(me gotui.MouseEvent) bool {
	for _, target := range c.transcriptBlockRefs {
		if target.Ref == nil || target.Ref.El() == nil || !target.Ref.El().ContainsPoint(me.X, me.Y) {
			continue
		}
		return c.toggleTranscriptBlock(target.Key)
	}
	return false
}

func (c *chatTUI) handleTranscriptScrollEvent(me gotui.MouseEvent) bool {
	switch me.Button {
	case gotui.MouseWheelUp, gotui.MouseWheelDown:
	default:
		return false
	}
	// Pi fullscreen scrolls the transcript for wheel input over it and over
	// the editor/footer (selectors keep their own input), with Pi's
	// velocity-based line count (fullscreenWheelScrollLines "auto") rather
	// than go-tui's fixed single line per event.
	over := c.transcriptRegion != nil && c.transcriptRegion.ContainsPoint(me.X, me.Y)
	below := c.transcriptRegion == nil || (!c.modelMenuOpen && me.Y >= c.transcriptRegion.Rect().Y+c.transcriptRegion.Rect().Height)
	if !over && !below {
		return false
	}
	direction := 1
	if me.Button == gotui.MouseWheelUp {
		direction = -1
	}
	c.wheel.configure(c.cfg.TUIWheelScrollLines)
	c.scrollTranscript(direction * c.wheel.next(direction, time.Now()))
	return true
}

func (c *chatTUI) focusInput() {
	if c.workspaceIndex.active {
		return
	}
	c.inputActive = true
	if c.app != nil && c.app.Focused() == nil {
		c.app.FocusNext()
	}
}

func (c *chatTUI) restoreQueuedDraft() {
	c.restoreQueuedDraftForActive("")
}

// Escape uses this only for an observed active run. Unlike Alt+Up, it must
// commit the cancellation request and the editor restore together.
func (c *chatTUI) restoreQueuedDraftForActive(activeID string) {
	if c.input == nil || c.store == nil || c.sessionID == "" || !c.durableDrafts {
		return
	}
	d := c.textDrafts[c.sessionID]
	if d == nil || d.frozen || d.err != nil || c.editorAskActive {
		c.showQueueCommand([]string{"queue: draft not ready; /draft to inspect"})
		return
	}
	// Slash commands are ephemeral, not saved journal text. Never replace one
	// or dequeue its underlying work while the editor contains it.
	if strings.HasPrefix(strings.TrimSpace(c.input.Text()), "/") {
		c.showQueueCommand([]string{"queue: finish the command before restoring queued text"})
		return
	}
	if !c.saveDurableDraft() {
		return
	}
	if d.local != c.editorSnapshot() {
		c.showQueueCommand([]string{"queue: editor changed; /draft to inspect"})
		return
	}
	ctx, cancel := c.draftContext()
	var restored store.TUITextDraft
	var ids []string
	var err error
	if activeID == "" {
		restored, ids, err = c.store.RestoreQueuedTUITextDraft(ctx, c.sessionID, d.pair.Text)
	} else if c.engine != nil {
		restored, ids, err = c.engine.AbortActiveAndRestoreTUITextDraft(ctx, c.sessionID, activeID, d.pair.Text)
	} else {
		err = store.ErrQueueConflict
	}
	cancel()
	if err != nil {
		if err != sql.ErrNoRows {
			c.showQueueCommand([]string{"queue: restore failed; delivery unchanged: " + err.Error()})
		}
		return
	}
	d.pair.Text = restored
	d.local = restored.TUITextSnapshot
	c.applyDraftSnapshot(d.local)
	c.queuedDrafts = nil
	c.queueSnapshot = nil
	c.publishQueueCommandChange(c.sessionID)
	c.status = fmt.Sprintf("Restored %d queued messages", len(ids))
	if activeID != "" {
		c.status = fmt.Sprintf("Stopped active turn; restored %d queued messages", len(ids))
	}
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) onSubmit(text string) {
	if c.workspaceIndex.active {
		return
	}
	c.submitWithMetadata(text, nil)
}

func (c *chatTUI) onFollowUp(text string) {
	c.submitWithIntent(text, nil, "queue")
}

func (c *chatTUI) submitWithMetadata(text string, metadata map[string]any) {
	c.submitWithIntent(text, metadata, "prompt")
}

func (c *chatTUI) submitWithIntent(text string, metadata map[string]any, intent string) {
	if c.workspaceIndex.active {
		return
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if c.editorAskActive {
		c.completeEditorAsk(text)
		return
	}
	// Reject before changing the editor/history or claiming media. A plaintext
	// draft needs the same preservation as an attached draft, including while
	// another turn is active. Local/slash commands remain available.
	if ordinaryMediaPrompt(text) && strings.TrimSpace(c.cfg.DefaultModel) == "" {
		c.showQueueCommand(c.firstUseModelPromptLines())
		c.status = "Select a model with /model <name>"
		if c.app != nil {
			c.app.MarkDirty()
		}
		return
	}
	if c.durableDrafts && ordinaryMediaPrompt(text) {
		c.submitDurableDraftWithIntent(text, intent)
		return
	}
	if c.durableDrafts && strings.HasPrefix(text, "/") {
		if strings.Fields(text)[0] == "/draft" {
			c.showQueueCommand(c.draftCommand(strings.Fields(text)))
			if c.input.ExpandedText() == text {
				if d := c.textDrafts[c.sessionID]; d != nil {
					c.applyDraftSnapshot(d.local)
				}
			}
			return
		}
		origin := c.sessionID
		c.draftApplying = true
		c.input.Reset("")
		c.draftApplying = false
		c.handleCommand(text)
		if c.sessionID == origin && !c.editorAskActive && c.input.Text() == "" {
			if d := c.textDrafts[origin]; d != nil {
				c.applyDraftSnapshot(d.local)
			}
		}
		return
	}
	var claim *mediaClaim
	if ordinaryMediaPrompt(text) {
		if err := c.refreshPendingMedia(); err != nil {
			c.appendTranscript("attachments: pending state unavailable; draft retained: " + err.Error())
			return
		}
	}
	if ordinaryMediaPrompt(text) && (len(c.pendingMedia[c.sessionID]) > 0 || c.mediaClaims[c.sessionID] != nil) {
		// Preserve text/cursor until it is safe to claim refs. Files belong to
		// this session, not a directed peer or a second in-flight submission.
		if c.mediaClaims[c.sessionID] != nil || strings.HasPrefix(text, "@") || strings.TrimSpace(c.cfg.DefaultModel) == "" {
			c.appendTranscript("attachments: retained; wait for admission, choose a model, or /detach all before a directed send")
			return
		}
	}
	if ordinaryMediaPrompt(text) {
		var err error
		claim, err = c.claimPendingMedia(!strings.HasPrefix(text, "@") && strings.TrimSpace(c.cfg.DefaultModel) != "")
		if err != nil {
			c.appendTranscript("attachments: could not claim references; draft retained: " + err.Error())
			return
		}
	}
	historyPrompt := text
	c.input.Reset("")
	if strings.HasPrefix(text, "/skill:") {
		c.appendTranscript(c.skillCommandLines(text)...)
		return
	}
	if strings.HasPrefix(text, "/") {
		c.handleCommand(text)
		return
	}
	if strings.HasPrefix(text, "!!") {
		c.appendTranscript(c.localShellShortcutLines(strings.TrimSpace(strings.TrimPrefix(text, "!!")))...)
		return
	}
	if strings.HasPrefix(text, "!") {
		cmd := strings.TrimSpace(strings.TrimPrefix(text, "!"))
		if cmd == "" {
			c.appendTranscript("sys: usage: !command sends a shell request to the model; !!command runs it locally")
			return
		}
		text = fmt.Sprintf("Run this shell command and summarize the result: %s", cmd)
	}
	// Store the typed prompt (not its expansion) after validation. Accepted
	// turns and queued follow-ups remain reachable after TUI restart.
	c.recordInputHistory(historyPrompt)
	scope := c.selectionScope()
	if claim != nil {
		merged := make(map[string]any, len(metadata)+2)
		for key, value := range metadata {
			merged[key] = value
		}
		merged["media"], merged["tui_media_claim"] = claim.refs, claim.token
		metadata = merged
	}
	input := turn.RunInput{SessionID: scope.id, Prompt: text, Intent: intent, Model: c.cfg.DefaultModel, Metadata: metadata}
	if c.running {
		c.queuedDrafts = append(c.queuedDrafts, text)
		c.appendUserPrompt(text, true)
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
		go func() {
			_, err := c.submitMediaInput(scope, input, claim)
			if err != nil {
				c.applySessionCompletion(scope, func() { c.appendTranscript(fmt.Sprintf("error: queue follow-up: %v", err)) })
			}
		}()
		return
	}
	if strings.TrimSpace(c.cfg.DefaultModel) == "" {
		c.transcript = append(c.transcript, c.firstUseModelPromptLines()...)
		c.status = "Select a model with /model <name>"
		c.scrollTranscriptToBottom()
		if c.app != nil {
			c.app.MarkDirty()
		}
		return
	}
	c.running = true
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.draft = ""
	c.draftLineIndex = -1
	c.draftLineCount = 0
	c.appendUserPrompt(text, false)
	c.showThinkingIndicator(time.Now())
	if c.stickToBottom {
		c.scrollTranscriptToBottom()
	}
	if c.app != nil {
		c.app.MarkDirty()
	}

	go func() {
		result, err := c.submitMediaInput(scope, input, claim)
		if err != nil {
			c.applySessionCompletion(scope, func() {
				c.finishThinkingTranscript(time.Now())
				c.clearDraftTranscriptLine()
				c.appendTranscript(fmt.Sprintf("error: %v", err))
				c.running = false
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
			})
			return
		}
		if result != nil && strings.TrimSpace(result.SessionID) != "" && result.SessionID != scope.id {
			apply := func() {
				c.switchSession(result.SessionID)
				c.running = result.Status == "running" || result.Status == "queued"
				if c.running {
					c.showThinkingIndicator(time.Now())
				}
				if result.Routed {
					c.appendTranscript(fmt.Sprintf("sys: routed to @%s (%s)", result.TargetAgentID, result.SessionID))
				}
				c.stickToBottom = true
				c.scrollTranscriptToBottom()
			}
			c.applySessionCompletion(scope, apply)
		}
	}()
}

func (c *chatTUI) handleCommand(text string) {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return
	}
	switch fields[0] {
	case "/help":
		c.transcript = append(c.transcript, c.helpLines()...)
	case "/hotkeys":
		c.transcript = append(c.transcript, c.hotkeyLines()...)
	case "/commands", "/palette":
		query := ""
		if len(fields) > 1 {
			query = strings.Join(fields[1:], " ")
		}
		c.appendTranscript(c.commandPaletteLines(query)...)
	case "/session":
		c.appendTranscript(c.sessionLines()...)
	case "/new":
		c.appendTranscript(c.newSessionLines()...)
	case "/name":
		c.appendTranscript(c.nameSessionLines(text, fields)...)
	case "/resume":
		if len(fields) == 1 { // Pi: /resume opens the session selector
			c.openSessionMenu()
			return
		}
		c.appendTranscript(c.resumeLines(fields)...)
	case "/sessions":
		c.openSessionMenu()
	case "/clone":
		c.appendTranscript(c.cloneSessionLines(fields)...)
	case "/copy":
		c.appendTranscript(c.copyLastAssistantLines(fields[1:]...)...)
	case "/retry":
		c.showQueueCommand(c.retryCommand(fields))
	case "/queue":
		c.showQueueCommand(c.queueCommand(fields))
	case "/attachments":
		c.appendTranscript(c.pendingMediaLines(fields)...)
	case "/detach":
		c.appendTranscript(c.detachMediaLines(fields)...)
	case "/attach":
		c.appendTranscript(c.attachCommand(text, fields)...)
	case "/paste-image", "/paste":
		c.appendTranscript(c.pasteImageCommand(text, fields)...)
	case "/login":
		c.handleLoginCommand(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), fields[0])))
	case "/logout":
		if len(fields) == 1 {
			c.handleLogoutCommand()
		} else {
			c.appendTranscript(c.logoutLines(fields)...) // gi: /logout <provider>
		}
	case "/reload":
		c.appendTranscript(c.reloadLines()...)
	case "/tools":
		c.transcript = append(c.transcript, c.toolCommand(fields)...)
	case "/skills":
		query := ""
		if len(fields) > 1 {
			query = strings.Join(fields[1:], " ")
		}
		c.transcript = append(c.transcript, c.skillLines(query)...)
	case "/model":
		if len(fields) == 1 {
			c.openModelMenu()
		} else {
			c.transcript = append(c.transcript, c.modelCommand(fields)...)
		}
	case "/scoped-models":
		if len(fields) == 1 {
			c.openScopedModels() // Pi's selector; gi keeps its subcommands
			return
		}
		c.transcript = append(c.transcript, c.scopedModelsCommand(fields)...)
	case "/thinking":
		if len(fields) == 1 {
			c.openThinkingMenu()
		} else {
			c.transcript = append(c.transcript, c.thinkingCommand(fields)...)
		}
	case "/compact":
		if len(fields) == 2 && fields[1] == "info" {
			c.appendTranscript(c.compactLines()...)
		} else {
			// Pi: /compact [instructions] focuses the summary.
			c.startCompactionWithInstructions(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), fields[0])))
			return
		}
	case "/scrollback":
		c.appendTranscript(c.scrollbackCommand(fields)...)
	case "/history-limit":
		c.appendTranscript(c.historyLimitCommand(fields)...)
	case "/settings":
		c.openSettingsMenu() // Pi's settings selector
		return
	case "/config":
		c.appendTranscript(c.settingsLines()...) // gi's runtime summary
	case "/approvals":
		c.appendTranscript("approvals: no approval gates are configured in gi yet")
	case "/cancel":
		c.appendTranscript(c.cancelCommand())
	case "/export":
		c.appendTranscript(c.exportCommand(text))
	case "/import":
		c.appendTranscript(c.importCommand(text)...)
	case "/share":
		c.appendTranscript(c.shareCommand()...)
	case "/quit", "/exit":
		if c.app != nil {
			c.app.Stop()
		}
		return
	case "/mcp":
		c.appendTranscript(c.mcpCommand(fields)...)
	case "/codemode":
		c.appendTranscript(c.codemodeCommand(fields)...)
	case "/abort":
		if line := c.abortCommand(); line != "" {
			c.appendTranscript(line)
		}
	case "/agents":
		c.transcript = append(c.transcript, c.listAgentLines()...)
	case "/plugins", "/extensions":
		c.transcript = append(c.transcript, c.pluginLines()...)
	case "/tree":
		c.openTreeSelector("")
	case "/fork", "/spawn":
		if fields[0] == "/fork" && len(fields) == 1 {
			c.openForkSelector() // Pi: fork from an earlier user message
			return
		}
		// /spawn [@agentN] (and the legacy /fork @agentN) opens a peer session.
		target := ""
		if len(fields) > 1 {
			target = strings.TrimPrefix(fields[1], "@")
		}
		if target == "" {
			target = c.nextForkAgentID()
		}
		targetSessionID, err := c.engine.ResolveOrCreatePeerSessionID(context.Background(), c.sessionID, target)
		if err != nil {
			c.transcript = append(c.transcript, fmt.Sprintf("error: %v", err))
			break
		}
		c.switchSession(targetSessionID)
		c.transcript = append(c.transcript, fmt.Sprintf("sys: switched to @%s (%s)", target, targetSessionID))
	case "/switch":
		if len(fields) < 2 {
			c.transcript = append(c.transcript, "sys: usage /switch @agent|session_id")
			break
		}
		sess, err := c.resolveSessionRef(fields[1])
		if err != nil {
			c.transcript = append(c.transcript, fmt.Sprintf("error: %v", err))
			break
		}
		c.switchSession(sess.ID)
		c.transcript = append(c.transcript, fmt.Sprintf("sys: switched to @%s (%s)", c.agentIDForSession(sess), sess.ID))
	case "/send":
		if len(fields) < 3 {
			c.transcript = append(c.transcript, "sys: usage /send @agent message")
			break
		}
		target := strings.TrimPrefix(fields[1], "@")
		body := strings.TrimSpace(strings.TrimPrefix(text, fields[0]+" "+fields[1]))
		c.transcript = append(c.transcript, fmt.Sprintf("you → @%s: %s", target, body))
		c.running = true
		scope, model := c.selectionScope(), c.cfg.DefaultModel
		go func() {
			result, err := c.engine.SubmitPeerMessage(context.Background(), scope.id, target, body, "prompt", model, "")
			c.applySessionCompletion(scope, func() {
				if err != nil {
					c.appendTranscript(fmt.Sprintf("error: %v", err))
				} else {
					c.appendTranscript(fmt.Sprintf("sys: delivered to @%s (%s)", target, result.SessionID))
				}
				c.running = false
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
			})
		}()
		return
	case "/where":
		c.transcript = append(c.transcript, c.contextSummary())
	default:
		if lines, handled := c.extensionCommandLines(text, fields); handled {
			c.appendTranscript(lines...)
		} else {
			c.appendTranscript("sys: commands: /help, /hotkeys, /commands [query], /session, /sessions, /new, /name <name>, /resume [index|session_id], /clone [@agentN], /copy [--osc52|--native|--auto|--fallback], /attach <path> [prompt], /attachments, /detach <media:id|all|unresolved>, /reload, /tools [query|active|activate|reset], /skills [query], /skill:name [args], /model [name], /scoped-models [add|remove|set], /thinking [level], /compact, /scrollback [n], /history-limit [n], /settings, /approvals, /queue [page|remove|steer|move], /draft [reload|check|release|restore|discard], /retry [page|check|run|skip|release], /cancel, /agents, /tree, /plugins, /fork [@agentN], /switch @agent|session_id, /send @agent message, /where, !cmd, !!cmd")
		}
	}
	c.running = false
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.stickToBottom = true
	c.scrollTranscriptToBottom()
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) firstUseModelPromptLines() []string {
	if len(c.cfg.EnabledModels) > 0 {
		return []string{fmt.Sprintf("sys: no model selected; choose one with /model <name>. available: %s", strings.Join(c.cfg.EnabledModels, ", "))}
	}
	return []string{"sys: no model selected; choose one with /model <provider/model> before sending your first prompt"}
}

func (c *chatTUI) extensionCommandLines(text string, fields []string) ([]string, bool) {
	if c.engine == nil || len(fields) == 0 {
		return nil, false
	}
	name := strings.TrimPrefix(fields[0], "/")
	rawArgs := strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
	res, handled, err := c.engine.InvokeExtensionCommand(context.Background(), name, rawArgs, c.sessionID, "")
	if !handled {
		return nil, false
	}
	if err != nil {
		return []string{fmt.Sprintf("error: extension command /%s: %v", name, err)}, true
	}
	switch strings.ToLower(strings.TrimSpace(res.Type)) {
	case "noop":
		if strings.TrimSpace(res.Status) != "" {
			c.status = res.Status
		}
		return []string{}, true
	case "error":
		if res.Error != "" {
			return []string{fmt.Sprintf("error: extension command /%s: %s", name, res.Error)}, true
		}
		return append([]string{fmt.Sprintf("error: extension command /%s", name)}, res.Lines...), true
	case "submit":
		prompt := strings.TrimSpace(res.Prompt)
		if prompt == "" {
			return []string{fmt.Sprintf("error: extension command /%s returned empty submit prompt", name)}, true
		}
		c.onSubmit(prompt)
		return []string{fmt.Sprintf("sys: extension command /%s submitted prompt", name)}, true
	default:
		if len(res.Lines) == 0 && strings.TrimSpace(res.Status) != "" {
			return []string{"sys: " + res.Status}, true
		}
		return res.Lines, true
	}
}

// tuiCommands is gi's slash command catalogue, used by /commands and by the
// Pi-style slash autocomplete below the editor.
// piCommands mirrors Pi's BUILTIN_SLASH_COMMANDS: same order, names and
// argument hints, with Pi's descriptions where gi behaves the same. Pi
// built-ins gi does not implement yet (/bug, /changelog, /trust) are
// omitted rather than approximated.
var piCommands = []struct{ name, hint string }{
	{"/settings", "Open settings menu"},
	{"/model <provider/model>", "Select model (opens selector UI)"},
	{"/tree", "Navigate session tree (switch branches)"},
	{"/thinking <level>", "Set thinking level"},
	{"/scoped-models", "Enable/disable models for Ctrl+P cycling"},
	{"/export [path]", "Export session (HTML default, or specify path: .html/.jsonl)"},
	{"/import <path>", "Import and resume a session from a JSONL file"},
	{"/share", "Share session as a secret GitHub gist"},
	{"/copy [--osc52|--native|--auto|--fallback]", "Copy last agent message to clipboard"},
	{"/name <name>", "Set session display name"},
	{"/session", "Show session info and stats"},
	{"/hotkeys", "Show all keyboard shortcuts"},
	{"/fork", "Create a new fork from a previous user message"},
	{"/clone [@agentN]", "Duplicate the current session at the current position"},
	{"/login <provider>", "Configure provider authentication"},
	{"/logout", "Remove provider authentication"},
	{"/new", "Start a new session"},
	{"/compact [instructions]", "Manually compact the session context"},
	{"/resume [index|session_id]", "Resume a different session"},
	{"/reload", "Reload config, skills and context files"},
	{"/quit", "Quit gi"},
}

// giCommands are gi's own commands, listed after Pi's built-ins the way Pi
// lists extension commands after its own.
var giCommands = []struct{ name, hint string }{
	{"/abort", "Abort the running turn (also clears one left by a crash)"},
	{"/queue [page|remove|steer|move]", "Inspect the durable queue; mutate by full turn IDs"},
	{"/retry [page|check|run|skip|release]", "Inspect held failures; guarded retry actions"},
	{"/draft [reload|check|release|restore|discard]", "Inspect or recover durable drafts"},
	{"/attach <path> [prompt]", "Stage up to six media refs for the next prompt"},
	{"/attachments", "List pending media refs and held admissions"},
	{"/detach <media:id|all|unresolved>", "Remove pending media refs"},
	{"/paste-image [prompt]", "Paste a clipboard image, optionally with a prompt"},
	{"/tools [query|active|activate|reset]", "Inspect or change active tools"},
	{"/mcp [login|logout|reconnect [server]]", "MCP servers: status, sign in or out, reconnect"},
	{"/codemode [on|off|only|default|status]", "Toggle the codemode tool for this session"},
	{"/skills [query]", "List discovered skills"},
	{"/skill:name [args]", "Load a discovered SKILL.md"},
	{"/agents", "List configured agents"},
	{"/spawn [@agentN]", "Open a peer agent session"},
	{"/switch @agent|session_id", "Switch the active session"},
	{"/send @agent message", "Send a peer message"},
	{"/where", "Show a context summary"},
	{"/plugins", "Show loaded extensions"},
	{"/scrollback [n]", "Show or set the transcript scrollback limit"},
	{"/history-limit [n]", "Show or set the per-session prompt history limit"},
	{"/cancel", "Cancel the latest active or queued turn"},
	{"/help", "Show grouped help"},
	{"/commands [query]", "Filter the command palette"},
	{"!cmd", "Ask the model to run a shell command"},
	{"!!cmd", "Run a local shell command"},
}

// tuiCommands is the palette/autocomplete catalogue: Pi's built-ins first.
var tuiCommands = append(append([]struct{ name, hint string }{}, piCommands...), giCommands...)

func (c *chatTUI) commandPaletteLines(query string) []string {
	commands := tuiCommands
	q := strings.ToLower(strings.TrimSpace(query))
	lines := []string{"commands: palette"}
	for _, cmd := range commands {
		text := cmd.name + " " + cmd.hint
		if q != "" && !strings.Contains(strings.ToLower(text), q) {
			continue
		}
		lines = append(lines, fmt.Sprintf("- %-36s %s", cmd.name, cmd.hint))
	}
	if c.engine != nil {
		for _, cmd := range c.engine.ExtensionCommandInfos() {
			name := "/" + cmd.Name
			hint := strings.TrimSpace(cmd.Description)
			if hint == "" {
				hint = "extension command"
			}
			text := name + " " + hint + " " + cmd.Usage
			if q != "" && !strings.Contains(strings.ToLower(text), q) {
				continue
			}
			lines = append(lines, fmt.Sprintf("- %-36s %s", name, hint))
		}
	}
	if len(lines) == 1 {
		return []string{fmt.Sprintf("commands: no matches for %q", query)}
	}
	return lines
}

func (c *chatTUI) helpLines() []string {
	return []string{
		"help",
		"enter send · shift-enter newline · esc blur · ctrl-d exit · f6/f7 select block · f8 expand",
		"/commands  all commands",
		"/hotkeys   keyboard shortcuts",
		"/model     choose model · type to filter · ctrl-l cycles",
		"alt-s      session picker · keeps unsent drafts · esc cancels",
		"alt-c      compact context · keeps draft · esc requests stop",
		"alt-i      workspace index · status/reindex · esc closes",
		"alt-m      model picker · session-local · keeps unsent drafts",
		"/session   details for this chat",
		"/where     compact context",
		"/attach    stage session media (up to 6)",
		"/attachments | /detach <media:id|all|unresolved> list/remove pending refs",
		"ctrl-r     search submitted prompts (current input is query)",
		"!cmd       ask model about shell · !!cmd run locally",
	}
}

func (c *chatTUI) logoutLines(fields []string) []string {
	if len(fields) < 2 {
		var stored []string
		for _, status := range inference.ListAuthStatus() {
			if status.Authenticated {
				stored = append(stored, status.ID)
			}
		}
		if len(stored) == 0 {
			return []string{"logout: no stored provider credentials"}
		}
		return []string{"logout: /logout <provider> · stored: " + strings.Join(stored, ", ")}
	}
	provider := strings.TrimSpace(fields[1])
	removed, err := inference.RemoveAuthEntry(provider)
	if err != nil {
		return []string{fmt.Sprintf("error: logout: %v", err)}
	}
	if !removed {
		return []string{fmt.Sprintf("logout: no stored credentials for %q", provider)}
	}
	return []string{fmt.Sprintf("logout: removed credentials for %s", provider)}
}

func (c *chatTUI) completeInputPath(text string, cursor int) (string, int, bool) {
	runes := []rune(text)
	if cursor < 0 || cursor > len(runes) {
		cursor = len(runes)
	}
	start := cursor
	for start > 0 && !isWordSpace(runes[start-1]) {
		start--
	}
	prefix := string(runes[start:cursor])
	if prefix == "" {
		return text, cursor, false
	}
	fileRef := strings.HasPrefix(prefix, "@")
	searchPrefix := prefix
	if fileRef {
		searchPrefix = strings.TrimPrefix(prefix, "@")
		if searchPrefix == "" {
			searchPrefix = "*"
		}
	}
	root := strings.TrimSpace(c.cfg.WorkspaceRoot)
	if root == "" {
		root = "."
	}
	pattern := searchPrefix + "*"
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(root, pattern)
	}
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return text, cursor, false
	}
	sort.Strings(matches)
	match := matches[0]
	if info, err := os.Stat(match); err == nil && info.IsDir() {
		match += string(os.PathSeparator)
	}
	insert := match
	if !filepath.IsAbs(searchPrefix) {
		if rel, err := filepath.Rel(root, match); err == nil && !strings.HasPrefix(rel, "..") {
			insert = rel
			if strings.HasSuffix(match, string(os.PathSeparator)) && !strings.HasSuffix(insert, string(os.PathSeparator)) {
				insert += string(os.PathSeparator)
			}
		}
	}
	if fileRef {
		insert = "@" + insert
	}
	out := string(runes[:start]) + insert + string(runes[cursor:])
	return out, start + utf8.RuneCountInString(insert), true
}

func (c *chatTUI) localShellShortcutLines(command string) []string {
	command = strings.TrimSpace(command)
	if command == "" {
		return []string{"sys: usage: !!command runs a local shell command; !command sends a shell request to the model"}
	}
	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var db *sql.DB
	if c.store != nil {
		db = c.store.DB()
	}
	prepared, err := shellenv.Prepare(ctx, db, c.cfg.ShellPath, command)
	if err != nil {
		return c.bashBlockLines(command, err.Error(), "error", err, startedAt, time.Now())
	}
	cmd, err := gitools.ShellCommand(ctx, prepared.ShellPath, prepared.Command)
	if err != nil {
		return c.bashBlockLines(command, err.Error(), "error", err, startedAt, time.Now())
	}
	if root := strings.TrimSpace(c.cfg.WorkspaceRoot); root != "" {
		cmd.Dir = root
	}
	if prepared.Env != nil {
		cmd.Env = prepared.Env
	}
	out, err := cmd.CombinedOutput()
	status := "ok"
	if err != nil {
		status = "error"
	}
	return c.bashBlockLines(command, string(out), status, err, startedAt, time.Now())
}

// bashBlockLines builds a PiSwift-style bash execution transcript block: a
// header with the command, the full output retained in the block body (capped),
// and a footer pointing at a full-output file when the output is truncated.
func (c *chatTUI) bashBlockLines(command, output, status string, runErr error, startedAt, endedAt time.Time) []string {
	const maxBodyLines = 500
	text := strings.ReplaceAll(plainTerminalOutput(output), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.TrimRight(text, "\n")
	var rawLines []string
	if strings.TrimSpace(text) == "" {
		rawLines = []string{"(no output)"}
	} else {
		rawLines = strings.Split(text, "\n")
	}
	footer := ""
	bodyLines := rawLines
	if len(rawLines) > maxBodyLines {
		if path, perr := c.writeBashFullOutput(command, output); perr == nil && path != "" {
			footer = "output truncated · full output: " + path
		} else {
			footer = fmt.Sprintf("output truncated · %d lines hidden", len(rawLines)-maxBodyLines)
		}
		bodyLines = rawLines[len(rawLines)-maxBodyLines:]
	}
	if runErr != nil {
		bodyLines = append(bodyLines, "error: "+truncate(runErr.Error(), 160))
	}
	meta := transcriptBlockMeta{
		Key:       fmt.Sprintf("bash:%d", time.Now().UnixNano()),
		Kind:      "bash",
		Title:     "$ " + command,
		Status:    status,
		StartedAt: startedAt.Format(time.RFC3339Nano),
		EndedAt:   endedAt.Format(time.RFC3339Nano),
		Footer:    footer,
	}
	lines := []string{encodeTranscriptBlockMarker(meta)}
	for _, line := range bodyLines {
		lines = append(lines, "│ "+line)
	}
	return lines
}

func (c *chatTUI) writeBashFullOutput(command, output string) (string, error) {
	dir := strings.TrimSpace(c.cfg.WorkspaceRoot)
	if dir == "" {
		dir = os.TempDir()
	} else {
		dir = filepath.Join(dir, ".gi-run", "bash-output")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, fmt.Sprintf("bash-%d.txt", time.Now().UnixNano()))
	header := "$ " + command + "\n\n"
	if err := os.WriteFile(path, []byte(header+output), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (c *chatTUI) reloadLines() []string {
	workspace := c.cfg.WorkspaceRoot
	if strings.TrimSpace(workspace) == "" {
		workspace = config.DefaultWorkspaceRoot()
	}
	reloaded := config.Load(workspace)
	c.cfg = reloaded
	if c.input != nil {
		c.input.placeholder = "Send a message…"
	}
	lines := []string{
		"reload: config refreshed",
		fmt.Sprintf("- workspace=%s", c.cfg.WorkspaceRoot),
		fmt.Sprintf("- provider=%s model=%s thinking=%s", c.cfg.DefaultProvider, c.cfg.DefaultModel, c.cfg.DefaultThinkingLevel),
		fmt.Sprintf("- discovery: skills=%d tools=%d", len(c.cfg.Discovery.Skills), len(c.cfg.Discovery.Tools)),
	}
	discoveredExtensions := gitools.DiscoverExtensionScripts(c.cfg.WorkspaceRoot)
	mountedExtensions := 0
	if c.engine != nil {
		mountedExtensions = len(c.engine.ExtensionInfos())
	}
	lines = append(lines, fmt.Sprintf("- extensions: discovered=%d mounted=%d", len(discoveredExtensions), mountedExtensions))
	if len(discoveredExtensions) != mountedExtensions {
		lines = append(lines, "- note: extension set changed; safe handler unload/reload is deferred, restart to remount extensions")
	} else {
		lines = append(lines, "- note: extension handlers remain mounted; /reload refreshes config and skill/tool discovery safely")
	}
	return lines
}

func (c *chatTUI) copyLastAssistantLines(args ...string) []string {
	messages, err := c.store.ListMessages(context.Background(), c.sessionID)
	if err != nil {
		return []string{fmt.Sprintf("error: copy last assistant message: %v", err)}
	}
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != "assistant" || strings.TrimSpace(messages[i].Content) == "" {
			continue
		}
		// Copy the persisted source, not a rendered or trimmed projection.
		return c.copyContentLines(messages[i].Content, "last assistant message", args...)
	}
	return []string{"copy: no assistant message found"}
}

// copyContentLines copies content (what it is, for the fallback) with the
// configured (or given) clipboard mode and reports how.
func (c *chatTUI) copyContentLines(content, what string, args ...string) []string {
	mode, persist, usage := c.copyModeFromArgs(args)
	if usage != "" {
		return []string{usage}
	}
	if persist {
		if err := config.PersistClipboardMode(c.cfg.WorkspaceRoot, mode); err != nil {
			return []string{fmt.Sprintf("warn: failed to persist clipboard mode: %v", err)}
		}
		c.cfg.TUIClipboardMode = mode
	}
	if mode == "osc52" {
		if err := c.writeOSC52(content); err != nil {
			return c.copyFallbackLines(content, what, fmt.Sprintf("OSC 52 failed: %v", err))
		}
		return []string{fmt.Sprintf("copy: sent %d bytes using OSC 52", len(content))}
	}
	if mode == "native" || mode == "auto" {
		if err := c.copyNative(content); err == nil {
			return []string{fmt.Sprintf("copy: sent %d bytes using native clipboard helper", len(content))}
		} else if mode == "native" {
			return c.copyFallbackLines(content, what, fmt.Sprintf("native clipboard failed: %v", err))
		}
	}
	return c.copyFallbackLines(content, what, "clipboard unavailable")
}

func (c *chatTUI) copyModeFromArgs(args []string) (mode string, persist bool, usage string) {
	mode = strings.TrimSpace(c.cfg.TUIClipboardMode)
	if mode == "" {
		mode = "off"
	}
	for i := 0; i < len(args); i++ {
		switch strings.ToLower(strings.TrimSpace(args[i])) {
		case "--osc52":
			mode = "osc52"
		case "--native":
			mode = "native"
		case "--auto":
			mode = "auto"
		case "--fallback", "--off":
			mode = "off"
		case "--mode":
			if i+1 >= len(args) {
				return "", false, "copy: usage /copy [--osc52|--native|--auto|--fallback|--mode <off|osc52|native|auto>] [--persist]"
			}
			i++
			mode = strings.ToLower(strings.TrimSpace(args[i]))
		case "--persist":
			persist = true
		default:
			return "", false, "copy: usage /copy [--osc52|--native|--auto|--fallback|--mode <off|osc52|native|auto>] [--persist]"
		}
	}
	switch mode {
	case "osc52", "native", "auto":
		return mode, persist, ""
	default:
		return "off", persist, ""
	}
}

func (c *chatTUI) copyFallbackLines(content, what, reason string) []string {
	lines := []string{fmt.Sprintf("copy: %s; %s follows (%d chars)", reason, what, len(content))}
	lines = append(lines, prefixMultiline("copy", content)...)
	return lines
}

const osc52PayloadLimit = 64 * 1024

func osc52Sequence(content string) (string, error) {
	if len(content) > osc52PayloadLimit {
		return "", fmt.Errorf("payload too large (%d > %d bytes)", len(content), osc52PayloadLimit)
	}
	return "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(content)) + "\x07", nil
}

func (c *chatTUI) writeOSC52(content string) error {
	seq, err := osc52Sequence(content)
	if err != nil {
		return err
	}
	w := c.osc52Writer
	if w == nil {
		w = os.Stdout
	}
	_, err = io.WriteString(w, seq)
	return err
}

func (c *chatTUI) copyNative(content string) error {
	helper, ok := selectNativeClipboardHelper(runtime.GOOS, os.Getenv("WAYLAND_DISPLAY") != "", os.Getenv("DISPLAY") != "", c.lookupClipboardHelper)
	if !ok {
		return fmt.Errorf("no native clipboard helper found")
	}
	runner := c.clipboardRun
	if runner == nil {
		runner = runClipboardHelper
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return runner(ctx, helper.name, helper.args, content)
}

type clipboardHelper struct {
	name string
	args []string
}

func (c *chatTUI) lookupClipboardHelper(name string) (string, error) {
	if c.clipboardLookPath != nil {
		return c.clipboardLookPath(name)
	}
	return exec.LookPath(name)
}

func selectNativeClipboardHelper(goos string, wayland, x11 bool, lookPath func(string) (string, error)) (clipboardHelper, bool) {
	candidates := []clipboardHelper{}
	switch goos {
	case "darwin":
		candidates = append(candidates, clipboardHelper{name: "pbcopy"})
	case "windows":
		candidates = append(candidates, clipboardHelper{name: "clip.exe"})
	default:
		if wayland {
			candidates = append(candidates, clipboardHelper{name: "wl-copy"})
		}
		if x11 {
			candidates = append(candidates, clipboardHelper{name: "xclip", args: []string{"-selection", "clipboard"}}, clipboardHelper{name: "xsel", args: []string{"--clipboard", "--input"}})
		}
		candidates = append(candidates, clipboardHelper{name: "clip.exe"})
	}
	for _, candidate := range candidates {
		if _, err := lookPath(candidate.name); err == nil {
			return candidate, true
		}
	}
	return clipboardHelper{}, false
}

func runClipboardHelper(ctx context.Context, name string, args []string, content string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	_, writeErr := io.WriteString(stdin, content)
	closeErr := stdin.Close()
	waitErr := cmd.Wait()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return waitErr
}

func (c *chatTUI) cloneSessionLines(fields []string) []string {
	target := ""
	if len(fields) > 1 {
		target = strings.TrimPrefix(strings.TrimSpace(fields[1]), "@")
	}
	if target == "" {
		target = c.nextForkAgentID()
	}
	newID := store.NowID("session")
	cloned, err := c.store.CloneSession(context.Background(), c.sessionID, newID, "@"+target, target)
	if err != nil {
		return []string{fmt.Sprintf("error: clone session: %v", err)}
	}
	c.switchSession(cloned.ID)
	return []string{fmt.Sprintf("sys: cloned to @%s (%s)", c.agentIDForSession(cloned), cloned.ID)}
}

func (c *chatTUI) newSessionLines() []string {
	id := store.NowID("session")
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", id)
	state := map[string]any{"status": "idle", "queue_count": 0, "model": c.cfg.DefaultModel, "provider": c.cfg.DefaultProvider, "thinking_level": c.cfg.DefaultThinkingLevel}
	sess, _, err := c.store.ResolveOrCreateMainSessionFromAllocation(context.Background(), store.ResolveOrCreateSessionFromAllocationInput{ID: id, Title: "@agent", State: state, Allocation: alloc})
	if err != nil {
		return []string{fmt.Sprintf("error: create session: %v", err)}
	}
	c.switchSession(sess.ID)
	return []string{fmt.Sprintf("sys: new session @%s (%s)", c.agentIDForSession(sess), sess.ID)}
}

func (c *chatTUI) nameSessionLines(text string, fields []string) []string {
	name := strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
	if name == "" { // Pi: /name without an argument shows the current name
		if sess, err := c.store.GetSession(context.Background(), c.sessionID); err == nil && strings.TrimSpace(sess.Title) != "" {
			return []string{fmt.Sprintf("sys: session name: %s", sess.Title)}
		}
		return []string{"sys: session has no name; /name <name> sets one"}
	}
	if err := c.store.UpdateSessionTitle(context.Background(), c.sessionID, name); err != nil {
		return []string{fmt.Sprintf("error: rename session: %v", err)}
	}
	return []string{fmt.Sprintf("sys: session renamed to %s", name)}
}

func (c *chatTUI) resumeLines(fields []string) []string {
	sessions, err := c.store.ListSessions(context.Background())
	if err != nil {
		return []string{fmt.Sprintf("error: list sessions: %v", err)}
	}
	if len(fields) > 1 {
		arg := strings.TrimSpace(fields[1])
		var target *store.Session
		if idx, err := strconv.Atoi(arg); err == nil {
			if idx < 1 || idx > len(sessions) {
				return []string{fmt.Sprintf("error: resume index out of range: %d", idx)}
			}
			target = &sessions[idx-1]
		} else {
			resolved, err := c.resolveSessionRef(arg)
			if err != nil {
				return []string{fmt.Sprintf("error: %v", err)}
			}
			target = resolved
		}
		c.switchSession(target.ID)
		return []string{fmt.Sprintf("sys: resumed @%s (%s)", c.agentIDForSession(target), target.ID)}
	}
	if len(sessions) == 0 {
		return []string{"resume: no sessions"}
	}
	limit := len(sessions)
	if limit > 10 {
		limit = 10
	}
	lines := []string{"resume: recent sessions"}
	for i := 0; i < limit; i++ {
		sess := sessions[i]
		messages, _ := c.store.ListMessages(context.Background(), sess.ID)
		turns, _ := c.store.ListTurns(context.Background(), sess.ID)
		status, _ := sess.State["status"].(string)
		if status == "" {
			status = "idle"
		}
		if c.compactOutput() {
			lines = append(lines, fmt.Sprintf("%d. @%s %s [%s] · %s · m=%d t=%d", i+1, c.agentIDForSession(&sess), compactText(sess.Title, 18), compactID(sess.ID), status, len(messages), len(turns)))
		} else {
			lines = append(lines, fmt.Sprintf("%d. @%s %s (%s) · %s · messages=%d turns=%d", i+1, c.agentIDForSession(&sess), sess.Title, sess.ID, status, len(messages), len(turns)))
		}
	}
	lines = append(lines, "resume: use /resume <index|session_id>")
	return lines
}

func (c *chatTUI) sessionLines() []string {
	sess, err := c.store.GetSession(context.Background(), c.sessionID)
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
	messages, _ := c.store.ListMessages(context.Background(), c.sessionID)
	turns, _ := c.store.ListTurns(context.Background(), c.sessionID)
	queuedTurns, _ := c.store.CountQueuedTurns(context.Background(), c.sessionID)
	steeringDepth, _ := c.store.SteeringQueueLength(context.Background(), c.sessionID)
	activeTurnID, claimToken, _ := c.store.GetSessionActiveTurn(context.Background(), c.sessionID)
	model := c.cfg.DefaultModel
	provider := c.cfg.DefaultProvider
	thinking := c.cfg.DefaultThinkingLevel
	status := "idle"
	queueCount := queuedTurns
	choice := inference.SessionModel(sess.State, inference.SessionModelChoice{Model: model, Provider: provider, Thinking: thinking})
	model, provider, thinking = choice.Model, choice.Provider, choice.Thinking
	if v, ok := sess.State["status"].(string); ok && v != "" {
		status = v
	}
	if v := intFromAny(sess.State["queue_count"]); v > queueCount {
		queueCount = v
	}
	parent := sess.ParentSessionID
	if parent == "" {
		parent = "root"
	}
	active := "none"
	if activeTurnID != "" {
		active = activeTurnID
		if claimToken != "" {
			active += " (claimed)"
		}
	}
	return []string{
		"session: current",
		fmt.Sprintf("- id=%s", sess.ID),
		fmt.Sprintf("- title=%s agent=@%s parent=%s", sess.Title, c.agentIDForSession(sess), parent),
		fmt.Sprintf("- messages=%d turns=%d queued_turns=%d steering=%d", len(messages), len(turns), queueCount, steeringDepth),
		fmt.Sprintf("- status=%s running=%v active_turn=%s", status, c.running, active),
		fmt.Sprintf("- model=%s provider=%s thinking=%s", model, provider, thinking),
	}
}

func splitModelRef(defaultProvider, label string) (string, string) {
	label = strings.TrimSpace(label)
	if label == "" {
		return strings.TrimSpace(defaultProvider), ""
	}
	if slash := strings.Index(label, "/"); slash > 0 {
		return strings.TrimSpace(label[:slash]), strings.TrimSpace(label[slash+1:])
	}
	return strings.TrimSpace(defaultProvider), label
}

func canonicalModelRef(defaultProvider, label string) string {
	provider, id := splitModelRef(defaultProvider, label)
	if id == "" {
		return ""
	}
	if provider == "" {
		return id
	}
	return provider + "/" + id
}

func (c *chatTUI) availableModelChoices() []string {
	choices := make([]string, 0, len(c.cfg.EnabledModels)+8)
	seen := map[string]bool{}
	appendChoice := func(label string) {
		label = strings.TrimSpace(label)
		if label == "" {
			return
		}
		key := canonicalModelRef(c.modelDefaults().Provider, label)
		if key == "" {
			key = label
		}
		if seen[key] {
			return
		}
		seen[key] = true
		choices = append(choices, key)
	}
	for _, model := range c.cfg.EnabledModels {
		appendChoice(model)
	}
	modelOptions := c.sessionModelCatalogue()
	for _, option := range modelOptions {
		appendChoice(option.Label)
	}
	appendChoice(c.cfg.DefaultModel)
	return choices
}

func (c *chatTUI) modelCommand(fields []string) []string {
	if len(fields) == 1 {
		return c.modelListLines()
	}
	model := strings.TrimSpace(strings.Join(fields[1:], " "))
	if model == "" {
		return []string{"sys: usage /model <model|index>"}
	}
	if idx, err := strconv.Atoi(model); err == nil {
		choices := c.availableModelChoices()
		if idx > 0 && idx <= len(choices) {
			model = choices[idx-1]
		}
	}
	if err := c.chooseSessionModel(model); err != nil {
		return []string{fmt.Sprintf("error: select model: %v", err)}
	}
	return []string{fmt.Sprintf("model: %s", c.sessionModelLabel())}
}

func (c *chatTUI) modelListLines() []string {
	current := strings.TrimSpace(c.cfg.DefaultModel)
	if current == "" {
		current = "(none selected)"
	}
	provider := strings.TrimSpace(c.cfg.DefaultProvider)
	if provider == "" {
		provider = "(inferred from model when possible)"
	}
	thinking := strings.TrimSpace(c.cfg.DefaultThinkingLevel)
	if thinking == "" {
		thinking = "low"
	}
	modelWidth := 52
	if c.compactOutput() {
		modelWidth = 34
	}
	lines := []string{fmt.Sprintf("model %s · %s", compactMaybe(current, c.compactOutput(), modelWidth), thinking)}
	if provider != "" && provider != "(inferred from model when possible)" {
		lines[0] += fmt.Sprintf(" · %s", compactMaybe(provider, c.compactOutput(), 18))
	}
	choices := c.availableModelChoices()
	if len(choices) == 0 {
		return append(lines, "no available models · /model <provider/model>")
	}
	for i, model := range choices {
		marker := " "
		if canonicalModelRef(c.cfg.DefaultProvider, model) == canonicalModelRef(c.cfg.DefaultProvider, c.cfg.DefaultModel) {
			marker = "›"
		}
		lines = append(lines, fmt.Sprintf("%s %d  %s", marker, i+1, compactMaybe(model, c.compactOutput(), modelWidth)))
	}
	if len(c.cfg.EnabledModels) > 0 {
		lines = append(lines, fmt.Sprintf("enabled: %d · /scoped-models list to manage pinned models", len(c.cfg.EnabledModels)))
	}
	lines = append(lines, "/model <n> to switch · ctrl-l cycles enabled models")
	return lines
}

func (c *chatTUI) scopedModelsCommand(fields []string) []string {
	if len(fields) == 1 || strings.EqualFold(fields[1], "list") {
		lines := []string{"scoped-models: enabled models:"}
		if len(c.cfg.EnabledModels) == 0 {
			return append(lines, "- none configured", "- usage: /scoped-models add <model> | remove <model|index> | set <model> [model...]")
		}
		for i, model := range c.cfg.EnabledModels {
			marker := " "
			if model == c.cfg.DefaultModel {
				marker = "*"
			}
			lines = append(lines, fmt.Sprintf("%s%d. %s", marker, i+1, model))
		}
		lines = append(lines, "- usage: /scoped-models add <model> | remove <model|index> | set <model> [model...]")
		return lines
	}
	action := strings.ToLower(strings.TrimSpace(fields[1]))
	args := fields[2:]
	models := append([]string(nil), c.cfg.EnabledModels...)
	switch action {
	case "add":
		if len(args) == 0 {
			return []string{"sys: usage /scoped-models add <model> [model...]"}
		}
		for _, model := range args {
			model = strings.TrimSpace(model)
			if model != "" && !containsString(models, model) {
				models = append(models, model)
			}
		}
	case "remove", "rm":
		if len(args) != 1 {
			return []string{"sys: usage /scoped-models remove <model|index>"}
		}
		needle := strings.TrimSpace(args[0])
		removed := false
		if idx, err := strconv.Atoi(needle); err == nil && idx > 0 && idx <= len(models) {
			models = append(models[:idx-1], models[idx:]...)
			removed = true
		} else {
			filtered := models[:0]
			for _, model := range models {
				if model == needle {
					removed = true
					continue
				}
				filtered = append(filtered, model)
			}
			models = filtered
		}
		if !removed {
			return []string{fmt.Sprintf("sys: scoped model not found: %s", needle)}
		}
	case "set":
		if len(args) == 0 {
			return []string{"sys: usage /scoped-models set <model> [model...]"}
		}
		models = nil
		for _, model := range args {
			model = strings.TrimSpace(model)
			if model != "" && !containsString(models, model) {
				models = append(models, model)
			}
		}
	default:
		return []string{"sys: usage /scoped-models [list|add|remove|set]"}
	}
	c.cfg.EnabledModels = models
	if c.cfg.DefaultModel != "" && !containsString(models, c.cfg.DefaultModel) && len(models) > 0 {
		c.cfg.DefaultModel = models[0]
	}
	lines := []string{fmt.Sprintf("sys: scoped models updated (%d enabled)", len(models))}
	if err := config.PersistModelSelection(c.cfg.WorkspaceRoot, c.cfg.DefaultProvider, c.cfg.DefaultModel, c.cfg.DefaultThinkingLevel, c.cfg.EnabledModels); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist scoped models: %v", err))
	}
	return append(lines, c.modelListLines()...)
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func (c *chatTUI) openThinkingMenu() {
	choices := c.thinkingMenuLevels()
	selected := 0
	current := c.effectiveThinking(c.cfg.DefaultProvider, c.cfg.DefaultModel, c.cfg.DefaultThinkingLevel)
	for i, level := range choices {
		if level == current {
			selected = i
			break
		}
	}
	c.modelMenuOpen = true
	c.modelMenuKind = "thinking"
	c.modelMenuValues = nil
	c.modelMenuAll = choices
	c.modelMenuQuery = ""
	c.modelMenuChoices = choices
	c.modelMenuSelected = selected
	c.modelMenuScroll = 0
	c.ensureModelMenuSelectionVisible()
	c.inputActive = false
	if c.app != nil {
		c.app.BlurFocused()
		c.app.MarkDirty()
	}
}

func (c *chatTUI) thinkingCommand(fields []string) []string {
	if len(fields) == 1 {
		return []string{fmt.Sprintf("sys: thinking: %s", c.cfg.DefaultThinkingLevel)}
	}
	level := strings.ToLower(strings.TrimSpace(fields[1]))
	if level == "" {
		return []string{"sys: usage /thinking <off|minimal|low|medium|high|xhigh|max>"}
	}
	// Like Pi, clamp to what the model supports and bind the level to the
	// session's model so inference applies it (not just the footer).
	model := c.sessionModelLabel()
	if effective, known := inference.EffectiveThinking(model, level); known {
		if effective == "" {
			return []string{fmt.Sprintf("sys: %s does not support thinking", model)}
		}
		level = effective
	}
	c.cfg.DefaultThinkingLevel = level
	lines := []string{fmt.Sprintf("sys: thinking set to %s", level)}
	if c.store == nil || c.sessionID == "" {
		return lines
	}
	ctx := context.Background()
	if session, err := c.store.GetSession(ctx, c.sessionID); err == nil && inference.ValidateThinking(model, level) == nil {
		if err := c.store.SelectSessionThinking(ctx, c.sessionID, store.SessionThinkingToken(c.sessionID, session.State), model, level); err != nil {
			lines = append(lines, fmt.Sprintf("warn: failed to select session thinking level: %v", err))
		}
		return lines
	}
	if err := c.store.TouchSessionState(ctx, c.sessionID, map[string]any{"thinking_level": level}); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist thinking level in session state: %v", err))
	}
	return lines
}

func (c *chatTUI) currentScrollbackLimit() int {
	if c.cfg.ScrollbackLimit > 0 {
		return c.cfg.ScrollbackLimit
	}
	return 1000
}

func (c *chatTUI) pruneTranscript(lines []string) []string {
	limit := c.currentScrollbackLimit()
	if limit <= 0 || len(lines) <= limit {
		return lines
	}
	return append([]string(nil), lines[len(lines)-limit:]...)
}

func (c *chatTUI) appendTranscript(lines ...string) {
	if len(lines) == 0 {
		return
	}
	c.transcript = append(c.transcript, lines...)
	c.applyTranscriptLimit()
}

func (c *chatTUI) applyTranscriptLimit() {
	limit := c.currentScrollbackLimit()
	if limit <= 0 || len(c.transcript) <= limit {
		return
	}
	drop := len(c.transcript) - limit
	c.transcript = append([]string(nil), c.transcript[drop:]...)
	c.regularPrinted = max(0, c.regularPrinted-drop)
	if c.draftLineIndex >= 0 {
		oldIndex := c.draftLineIndex
		c.draftLineIndex -= drop
		if c.draftLineIndex < 0 || drop >= oldIndex+c.draftLineCount {
			c.draftLineIndex = -1
			c.draftLineCount = 0
		}
	}
	c.reindexTranscriptBlocks()
}

func (c *chatTUI) compactLines() []string {
	messages, _ := c.store.ListMessages(context.Background(), c.sessionID)
	turns, _ := c.store.ListTurns(context.Background(), c.sessionID)
	settings := c.cfg.Compaction
	return []string{
		fmt.Sprintf("compact: enabled=%v threshold_tokens=%d keep_recent_tokens=%d reserve_tokens=%d strategy=%s", settings.Enabled, settings.ThresholdTokens, settings.KeepRecentTokens, settings.ReserveTokens, settings.Strategy),
		fmt.Sprintf("compact: messages=%d turns=%d; /compact runs maintenance, Alt-C preserves draft, Escape stops", len(messages), len(turns)),
	}
}

func (c *chatTUI) settingsLines() []string {
	activeTools := "(none)"
	if c.engine != nil {
		if names := c.engine.ActiveTools(); len(names) > 0 {
			activeTools = strings.Join(names, ", ")
		}
	}
	enabledModels := "(none)"
	if len(c.cfg.EnabledModels) > 0 {
		enabledModels = strings.Join(c.cfg.EnabledModels, ", ")
	}
	clipboardMode := c.cfg.TUIClipboardMode
	if clipboardMode == "" {
		clipboardMode = "default (selection: osc52; /copy: off)"
	}
	themeSetting := c.cfg.Theme
	if themeSetting == "" {
		themeSetting = "(auto)"
	}
	wheel := "auto"
	if c.cfg.TUIWheelScrollLines > 0 {
		wheel = strconv.Itoa(c.cfg.TUIWheelScrollLines)
	}
	thinking := c.effectiveThinking(c.cfg.DefaultProvider, c.cfg.DefaultModel, c.cfg.DefaultThinkingLevel)
	if thinking == "" {
		thinking = "off (model does not support reasoning)"
	}
	retry := c.cfg.Retry.Policy()
	return []string{
		"settings: live runtime summary (not a settings-file dump; secrets omitted)",
		"settings: runtime",
		fmt.Sprintf("- workspace: %s", c.cfg.WorkspaceRoot),
		fmt.Sprintf("- max_iterations: %d", c.cfg.MaxIterations),
		fmt.Sprintf("- active_tools: %s", activeTools),
		"settings: model",
		fmt.Sprintf("- provider: %s", c.cfg.DefaultProvider),
		fmt.Sprintf("- model: %s", c.cfg.DefaultModel),
		fmt.Sprintf("- thinking: %s", thinking),
		fmt.Sprintf("- thinking_configured: %s", c.cfg.DefaultThinkingLevel),
		fmt.Sprintf("- enabled_models: %s", enabledModels),
		"settings: editor",
		fmt.Sprintf("- theme: %s", piActiveTheme),
		fmt.Sprintf("- theme_configured: %s", themeSetting),
		fmt.Sprintf("- fullscreen_wheel_scroll_lines: %s", wheel),
		fmt.Sprintf("- scrollback_limit: %d", c.currentScrollbackLimit()),
		fmt.Sprintf("- clipboard_mode: %s", clipboardMode),
		fmt.Sprintf("- history_limit: %d", c.currentHistoryLimit()),
		"- shortcuts: /hotkeys",
		"settings: session",
		fmt.Sprintf("- session_id: %s", c.sessionID),
		fmt.Sprintf("- running: %v", c.running),
		"settings: discovery",
		fmt.Sprintf("- tools_discovery: %d", len(c.cfg.Discovery.Tools)),
		fmt.Sprintf("- skills_discovery: %d", len(c.cfg.Discovery.Skills)),
		"settings: compaction",
		fmt.Sprintf("- enabled: %v", c.cfg.Compaction.Enabled),
		fmt.Sprintf("- context_window: %d", c.cfg.Compaction.ContextWindow),
		fmt.Sprintf("- threshold_tokens: %d", c.cfg.Compaction.ThresholdTokens),
		fmt.Sprintf("- keep_recent_tokens: %d", c.cfg.Compaction.KeepRecentTokens),
		fmt.Sprintf("- reserve_tokens: %d", c.cfg.Compaction.ReserveTokens),
		fmt.Sprintf("- strategy: %s", c.cfg.Compaction.Strategy),
		"settings: provider retry",
		fmt.Sprintf("- enabled: %v", retry.Enabled),
		fmt.Sprintf("- max_retries: %d", retry.MaxRetries),
		fmt.Sprintf("- base_delay_ms: %d", retry.BaseDelayMS),
		fmt.Sprintf("- max_agent_delay_ms: %d", retry.MaxDelayMS),
		"settings: peering",
		fmt.Sprintf("- enabled: %v", c.cfg.Peering.Enabled),
		fmt.Sprintf("- hostname: %s", c.cfg.Peering.Hostname),
		fmt.Sprintf("- state_dir: %s", c.cfg.Peering.StateDir),
		fmt.Sprintf("- auth_key_env: %s", c.cfg.Peering.AuthKeyEnv),
		fmt.Sprintf("- auth_key_keychain: %s", c.cfg.Peering.AuthKeyKeychain),
	}
}

func (c *chatTUI) scrollbackCommand(fields []string) []string {
	if len(fields) == 1 {
		return []string{fmt.Sprintf("sys: scrollback limit: %d", c.currentScrollbackLimit())}
	}
	limit, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	if err != nil || limit <= 0 {
		return []string{"sys: usage /scrollback <positive-limit>"}
	}
	c.cfg.ScrollbackLimit = limit
	c.applyTranscriptLimit()
	lines := []string{fmt.Sprintf("sys: scrollback limit set to %d", limit)}
	if err := config.PersistScrollbackLimit(c.cfg.WorkspaceRoot, limit); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist scrollback limit: %v", err))
	}
	return lines
}

func (c *chatTUI) historyLimitCommand(fields []string) []string {
	if len(fields) == 1 {
		return []string{fmt.Sprintf("sys: history limit: %d", c.currentHistoryLimit())}
	}
	limit, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	if err != nil || limit <= 0 {
		return []string{"sys: usage /history-limit <positive-limit>"}
	}
	c.cfg.TUIHistoryLimit = limit
	c.applyHistoryLimit()
	lines := []string{fmt.Sprintf("sys: history limit set to %d", limit)}
	if err := config.PersistTUIHistoryLimit(c.cfg.WorkspaceRoot, limit); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist history limit: %v", err))
	}
	return lines
}

// abortCommand stops the session's active work. A live run is aborted like
// Escape (queued text returns to the editor). An active turn no worker is
// running — left 'running' by a crash or killed process, which makes the
// session look busy (e.g. /compact unavailable) while the TUI is idle — is
// finalized as aborted instead of being replayed on the next start.
func (c *chatTUI) abortCommand() string {
	if c.compaction.active {
		c.stopCompaction()
		return "sys: compaction cancellation requested"
	}
	if c.engine == nil || c.store == nil || c.sessionID == "" {
		return "sys: nothing to abort"
	}
	ctx, cancel := c.draftContext()
	activeID, _, err := c.store.GetSessionActiveTurn(ctx, c.sessionID)
	cancel()
	if errors.Is(err, sql.ErrNoRows) || (err == nil && activeID == "") {
		return "sys: nothing to abort"
	}
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	if c.engine.LiveActiveTurn(c.sessionID) == activeID {
		c.restoreQueuedDraftForActive(activeID)
		return ""
	}
	ctx, cancel = c.draftContext()
	aborted, err := c.engine.AbortStaleActiveTurn(ctx, c.sessionID)
	cancel()
	switch {
	case errors.Is(err, turn.ErrTurnOwnedElsewhere):
		return fmt.Sprintf("sys: %s is running in another gi process; abort it there", activeID)
	case err != nil:
		return fmt.Sprintf("error: abort %s: %v", activeID, err)
	case aborted == "":
		return "sys: nothing to abort"
	}
	c.running = false
	return fmt.Sprintf("sys: aborted interrupted turn %s (no worker was running it)", aborted)
}

func (c *chatTUI) cancelCommand() string {
	if c.compaction.active {
		c.stopCompaction()
		return "sys: compaction cancellation requested"
	}
	turns, err := c.store.ListTurns(context.Background(), c.sessionID)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	for i := len(turns) - 1; i >= 0; i-- {
		if turns[i].Status == "running" || turns[i].Status == "queued" || turns[i].Status == "cancelling" {
			if err := c.engine.CancelTurn(context.Background(), c.sessionID, turns[i].ID); err != nil {
				return fmt.Sprintf("error: %v", err)
			}
			c.running = false
			return fmt.Sprintf("sys: cancellation requested for %s", turns[i].ID)
		}
	}
	c.running = false
	return "sys: no running or queued turn to cancel"
}

func (c *chatTUI) toolCommand(fields []string) []string {
	if len(fields) > 1 {
		switch fields[1] {
		case "active":
			active := c.engine.ActiveTools()
			if len(active) == 0 {
				return []string{"tools: active: (none)"}
			}
			return []string{fmt.Sprintf("tools: active: %s", strings.Join(active, ", "))}
		case "activate":
			if len(fields) < 3 {
				return []string{"tools: usage /tools activate <tool> [tool...]"}
			}
			names := make([]any, 0, len(fields)-2)
			var lines []string
			for _, name := range fields[2:] {
				name = strings.TrimSpace(name)
				if name == "codemode" {
					// Codemode is a per-session toggle (#25), not part of the
					// engine-wide active set.
					lines = append(lines, c.codemodeCommand([]string{"/codemode", "on"})...)
				} else if name != "" {
					names = append(names, name)
				}
			}
			if len(names) == 0 {
				return lines
			}
			out, err := c.engine.ExecuteToolsMeta(map[string]any{"activate": names})
			if err != nil {
				return append(lines, fmt.Sprintf("error: %v", err))
			}
			return append(lines, prefixMultiline("tools", out)...)
		case "reset":
			out, err := c.engine.ExecuteToolsMeta(map[string]any{"reset_active": true})
			if err != nil {
				return []string{fmt.Sprintf("error: %v", err)}
			}
			lines := prefixMultiline("tools", out)
			if c.engine != nil && c.sessionID != "" {
				// Codemode returns to its settings default too.
				lines = append(lines, c.codemodeCommand([]string{"/codemode", "default"})...)
			}
			return lines
		}
	}
	query := ""
	if len(fields) > 1 {
		query = strings.Join(fields[1:], " ")
	}
	args := map[string]any{"include_inactive": true}
	if strings.TrimSpace(query) != "" {
		args["query"] = strings.TrimSpace(query)
	}
	out, err := c.engine.ExecuteToolsMeta(args)
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
	return prefixMultiline("tools", out)
}

func (c *chatTUI) skillCommandLines(text string) []string {
	body := strings.TrimSpace(strings.TrimPrefix(text, "/skill:"))
	if body == "" {
		return []string{"sys: usage /skill:<name> [args]"}
	}
	name := body
	args := ""
	if fields := strings.Fields(body); len(fields) > 0 {
		name = fields[0]
		args = strings.TrimSpace(strings.TrimPrefix(body, name))
	}
	out, err := skills.ExecuteMeta(c.cfg.WorkspaceRoot, map[string]any{"name": name})
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
	lines := []string{fmt.Sprintf("skill:%s loaded", name)}
	if args != "" {
		lines = append(lines, fmt.Sprintf("- args: %s", args))
	}
	lines = append(lines, prefixMultiline("skill", out)...)
	return lines
}

func (c *chatTUI) skillLines(query string) []string {
	args := map[string]any{}
	if strings.TrimSpace(query) != "" {
		args["query"] = strings.TrimSpace(query)
	}
	out, err := skills.ExecuteMeta(c.cfg.WorkspaceRoot, args)
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
	return prefixMultiline("skills", out)
}

func prefixMultiline(prefix, text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{prefix + ": (empty)"}
	}
	parts := strings.Split(text, "\n")
	out := make([]string, 0, len(parts))
	for i, line := range parts {
		if i == 0 {
			out = append(out, prefix+": "+line)
		} else {
			out = append(out, "  "+line)
		}
	}
	return out
}

func (c *chatTUI) pluginLines() []string {
	lines := []string{"plugins: extensions:"}
	extensions := c.engine.ExtensionInfos()
	if len(extensions) == 0 {
		lines = append(lines, "- none loaded")
	} else {
		for _, ext := range extensions {
			line := fmt.Sprintf("- %s %s %s", ext.Status, ext.Engine, ext.Path)
			if ext.Error != "" {
				line += ": " + truncate(ext.Error, 120)
			}
			lines = append(lines, line)
		}
	}
	commands := c.engine.ExtensionCommandInfos()
	conflicts := c.engine.ExtensionCommandConflicts()
	lines = append(lines, fmt.Sprintf("plugins: commands: %d", len(commands)))
	for _, cmd := range commands {
		usage := strings.TrimSpace(cmd.Usage)
		if usage == "" {
			usage = "/" + cmd.Name
		}
		lines = append(lines, fmt.Sprintf("- /%s from %s (%s) · %s", cmd.Name, gitools.FirstNonEmpty(cmd.Source, "extension"), gitools.FirstNonEmpty(cmd.Engine, "unknown"), usage))
	}
	for _, cmd := range conflicts {
		lines = append(lines, fmt.Sprintf("- conflict /%s from %s (%s)", cmd.Name, gitools.FirstNonEmpty(cmd.Source, "extension"), gitools.FirstNonEmpty(cmd.Engine, "unknown")))
	}
	hooks := c.engine.HookInfos()
	lines = append(lines, fmt.Sprintf("plugins: hooks: %d", len(hooks)))
	for _, hook := range hooks {
		lines = append(lines, fmt.Sprintf("- %s from %s #%d", hook.Name, hook.Source, hook.ID))
	}
	return lines
}

func (c *chatTUI) Render(app *gotui.App) *gotui.Element {
	if c.regularMode {
		return c.renderRegular(app)
	}
	w, h := app.Size()
	padding := 0 // Pi draws transcript, editor and footer edge to edge.
	contentWidth := w - (padding * 2)
	if contentWidth < 20 {
		contentWidth = 20
	}

	root := gotui.New(
		gotui.WithDirection(gotui.Column),
		gotui.WithWidthPercent(100),
		gotui.WithHeightPercent(100),
		gotui.WithPadding(padding),
		gotui.WithGap(0),
	)

	c.ensureInput()
	c.input.width = contentWidth
	c.input.suspended = c.workspaceIndex.active || c.modelMenuOpen
	footerLines := c.footerLines(contentWidth)
	pendingLines := c.pendingDockLines(contentWidth, h)
	// Pi's widget container above the editor always starts with Spacer(1),
	// leaving one blank row between the transcript and the input.
	widgetLines := append([]string{""}, c.extensionWidgetLines()...)
	if c.editorAskActive {
		widgetLines = append(widgetLines, "? "+c.editorAskPrompt+"  (Enter submit · Esc cancel)")
	}
	menuHeight := c.modelMenuHeight() + c.workspaceIndexHeight() + c.slashMenuHeight()
	c.boundEditor(h, padding, len(footerLines), len(widgetLines)+len(pendingLines), menuHeight, false)
	activeInput := c.input
	inputSlot := 0
	if c.search.active {
		c.refreshTranscriptSearch(contentWidth)
		activeInput = c.search.input
		activeInput.width = contentWidth
		activeInput.maxLines = c.input.maxLines
		inputSlot = 1
	}
	inputHeight := activeInput.Render(app).HeightForWidth(contentWidth)
	if inputHeight < 1 {
		inputHeight = 1
	}
	// Pi's model selector replaces the editor (and its borders) while open.
	piSelector := c.modelMenuOpen && c.modelMenuKind != "session-rename"
	editorRows := inputHeight + 2
	if piSelector {
		editorRows = 0
	}
	reservedHeight := (padding * 2) + len(footerLines) + len(pendingLines) + len(widgetLines) + editorRows + menuHeight
	transcriptHeight := h - reservedHeight
	if transcriptHeight < 4 && !piSelector {
		transcriptHeight = 4
	}
	transcriptHeight = max(0, transcriptHeight)
	transcriptOptions := []gotui.Option{
		gotui.WithWidthPercent(100),
		gotui.WithHeight(transcriptHeight),
		gotui.WithScrollable(gotui.ScrollVertical),
		// Pi has no transcript scrollbar; the full width belongs to content.
		gotui.WithScrollbarHidden(true),
		gotui.WithScrollOffset(0, c.transcriptScroll),
		gotui.WithDirection(gotui.Column),
	}
	c.validateTranscriptSelection(contentWidth, transcriptHeight)
	transcript := gotui.New(transcriptOptions...)
	c.transcriptRef.Set(transcript)
	c.transcriptRegion = transcript
	c.transcriptBlockRefs = nil
	if c.transcriptExpanded == nil {
		c.transcriptExpanded = map[string]bool{}
	}
	blocks := c.transcriptBlocks()
	if c.selectedTranscriptBlock == "" {
		for i := len(blocks) - 1; i >= 0; i-- {
			if blocks[i].Key != "" && (len(blocks[i].Body) > 0 || blocks[i].Subheader != "") {
				c.selectedTranscriptBlock = blocks[i].Key
				break
			}
		}
		blocks = c.transcriptBlocks()
	}
	if c.textSelection.active {
		c.renderTranscriptSelectionRows(transcript)
	} else if c.search.active {
		c.renderTranscriptSearchRows(transcript)
	} else {
		// Only blocks on screen are laid out; spacers keep the scroll
		// geometry (#34).
		c.addTranscriptWindow(transcript, blocks, contentWidth, transcriptHeight)
	}
	if c.stickToBottom {
		// Resolve the bottom after layout, when wrapping/expansion is known.
		transcript.ScrollToBottom()
	}
	root.AddChild(transcript)
	if len(pendingLines) > 0 {
		root.AddChild(c.renderLineBlock(pendingLines, piFg(piDim)))
	}
	if c.modelMenuOpen && !piSelector {
		root.AddChild(c.renderModelMenu(contentWidth))
	}
	if c.workspaceIndex.active {
		root.AddChild(c.renderWorkspaceIndex(contentWidth))
	}

	if len(widgetLines) > 0 {
		root.AddChild(c.renderLineBlock(widgetLines, gotui.NewStyle()))
	}

	// The editor must lay out before its borders so overflow counts are current.
	inputEl := app.MountPersistent(c, inputSlot, func() gotui.Component { return activeInput })
	c.inputRegion = inputEl
	switch {
	case piSelector:
		root.AddChild(c.renderModelMenu(contentWidth))
		root.AddChild(c.renderFooter(contentWidth))
		return root
	case c.textSelection.active:
		root.AddChild(borderElement([]gotui.TextSpan{{Text: c.selectionSeparator(contentWidth), Style: piFg(c.editorBorderColor())}}))
	case c.search.active:
		root.AddChild(borderElement([]gotui.TextSpan{{Text: c.transcriptSearchLabel(contentWidth), Style: piFg(c.editorBorderColor())}}))
	default:
		root.AddChild(c.renderEditorTopBorder(activeInput, contentWidth))
	}
	root.AddChild(inputEl)
	root.AddChild(c.renderEditorBottomBorder(activeInput, contentWidth))
	if c.slash.active && !c.search.active {
		root.AddChild(c.renderSlashMenu(contentWidth))
	}

	root.AddChild(c.renderFooter(contentWidth))

	return root
}

func abbreviateHomePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	home = strings.TrimSpace(home)
	if home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	prefix := home + string(os.PathSeparator)
	if strings.HasPrefix(path, prefix) {
		return "~" + string(os.PathSeparator) + strings.TrimPrefix(path, prefix)
	}
	return path
}

func (c *chatTUI) footerPathLineForWidth(width int) string {
	workspace := strings.TrimSpace(c.cfg.WorkspaceRoot)
	if workspace == "" {
		workspace = "."
	}
	if branch := c.gitBranchName(workspace); branch != "" {
		workspace += " (" + branch + ")"
	}
	workspace = abbreviateHomePath(workspace)
	if width > 0 {
		return compactMaybe(workspace, true, width)
	}
	return workspace
}

// setExtensionStatus is a backend-safe TUI extension slot: extensions can set a
// keyed status segment that renders as an extra dim footer line. It can never
// add top chrome; cleared keys (empty text) are removed.
func (c *chatTUI) setExtensionStatus(key, text string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	if c.extensionStatuses == nil {
		c.extensionStatuses = map[string]string{}
	}
	text = strings.TrimSpace(sanitizeStatusText(text))
	if text == "" {
		delete(c.extensionStatuses, key)
		return
	}
	c.extensionStatuses[key] = text
}

func (c *chatTUI) extensionStatusLines() []string {
	if len(c.extensionStatuses) == 0 {
		return nil
	}
	keys := make([]string, 0, len(c.extensionStatuses))
	for k := range c.extensionStatuses {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, c.extensionStatuses[k])
	}
	return lines
}

func sanitizeStatusText(text string) string {
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return strings.TrimSpace(text)
}

// setExtensionWidget is a TUI extension slot that renders a keyed multi-line
// widget between the transcript and the editor. It stays inside the bottom band
// (above the editor, below the transcript) and can never add top chrome.
func (c *chatTUI) setExtensionWidget(key string, lines []string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	if c.extensionWidgets == nil {
		c.extensionWidgets = map[string][]string{}
	}
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		cleaned = append(cleaned, sanitizeStatusText(line))
	}
	if len(cleaned) == 0 {
		delete(c.extensionWidgets, key)
		return
	}
	c.extensionWidgets[key] = cleaned
}

func (c *chatTUI) extensionWidgetLines() []string {
	if len(c.extensionWidgets) == 0 {
		return nil
	}
	keys := make([]string, 0, len(c.extensionWidgets))
	for k := range c.extensionWidgets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var lines []string
	for _, k := range keys {
		lines = append(lines, c.extensionWidgets[k]...)
	}
	return lines
}

// setExtensionToolRender is a custom tool-renderer slot: extensions can choose
// how a named tool's block body renders without writing top chrome. Supported
// modes: "full" (default), "compact" (first body line), "hidden" (header only).
func (c *chatTUI) setExtensionToolRender(tool, mode string) {
	tool = strings.TrimSpace(tool)
	if tool == "" {
		return
	}
	mode = strings.ToLower(strings.TrimSpace(mode))
	if c.extensionToolModes == nil {
		c.extensionToolModes = map[string]string{}
	}
	switch mode {
	case "", "full", "default":
		delete(c.extensionToolModes, tool)
	case "compact", "hidden":
		c.extensionToolModes[tool] = mode
	default:
		delete(c.extensionToolModes, tool)
	}
}

func (c *chatTUI) applyToolRenderMode(tool string, body []string) ([]string, bool) {
	mode := c.extensionToolModes[strings.TrimSpace(tool)]
	switch mode {
	case "hidden":
		return nil, false
	case "compact":
		if len(body) > 1 {
			return body[:1], false
		}
		return body, false
	}
	return body, true
}

// setEditorAsk is the editor-replacement extension slot (PiSwift setEditorComponent
// / hook input). It puts the editor into a bounded ask mode: the bottom band shows
// a prompt above the editor, the input is prefilled, and the next submit captures
// the answer instead of sending it to the model. It never adds top chrome.
func (c *chatTUI) setEditorAsk(key, prompt, prefill string) {
	key = strings.TrimSpace(key)
	prompt = sanitizeStatusText(prompt)
	if prompt == "" {
		c.cancelEditorAsk()
		return
	}
	c.ensureInput()
	if !c.editorAskActive {
		c.editorAskPrevPlaceholder = c.input.placeholder
		c.editorAskPrevText = c.input.ExpandedText()
		c.editorAskPrevCursor = c.input.expandedCursor()
	}
	c.editorAskActive = true
	c.editorAskKey = key
	c.editorAskPrompt = prompt
	c.input.placeholder = prompt
	c.input.SetText(prefill)
	c.focusInput()
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) completeEditorAsk(answer string) {
	if h := c.editorAskHandler; h != nil {
		c.editorAskHandler = nil
		c.exitEditorAsk()
		h(answer, false)
		return
	}
	key := c.editorAskKey
	c.exitEditorAsk()
	label := key
	if label == "" {
		label = "answer"
	}
	c.appendTranscript(fmt.Sprintf("sys: %s: %s", label, answer))
	if c.engine != nil && c.engine.Topics() != nil {
		c.engine.Topics().Publish(topics.Envelope{Topic: "extension.editor_result", SessionID: c.sessionID, Payload: map[string]any{"key": key, "answer": answer}})
	}
}

func (c *chatTUI) cancelEditorAsk() {
	if !c.editorAskActive {
		return
	}
	if h := c.editorAskHandler; h != nil {
		c.editorAskHandler = nil
		c.exitEditorAsk()
		h("", true)
		return
	}
	c.exitEditorAsk()
	c.appendTranscript("sys: prompt cancelled")
}

func (c *chatTUI) exitEditorAsk() {
	c.editorAskActive = false
	c.editorAskKey = ""
	c.editorAskPrompt = ""
	if c.input != nil {
		c.input.placeholder = c.editorAskPrevPlaceholder
		if c.input.placeholder == "" {
			c.input.placeholder = "Send a message\u2026"
		}
		c.input.SetText(c.editorAskPrevText)
		c.input.cursorPos = min(c.editorAskPrevCursor, utf8.RuneCountInString(c.input.Text()))
		c.editorAskPrevText, c.editorAskPrevCursor = "", 0
	}
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func widgetPayloadLines(payload map[string]any) []string {
	switch v := payload["lines"].(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	if text, ok := payload["text"].(string); ok && text != "" {
		return strings.Split(text, "\n")
	}
	return nil
}

func formatTokenCount(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%dK", n/1000)
	}
	return strconv.Itoa(n)
}

func formatContextUsage(tokens, window int) string {
	if tokens <= 0 {
		return "ctx 0"
	}
	if window > 0 {
		pct := int(float64(tokens) * 100 / float64(window))
		if pct < 0 {
			pct = 0
		}
		return fmt.Sprintf("ctx %s/%s %d%%", formatTokenCount(tokens), formatTokenCount(window), pct)
	}
	return fmt.Sprintf("ctx %s", formatTokenCount(tokens))
}

// gitBranchName finds the repository containing workspace, as Pi's footer
// does from its cwd: walk up to the nearest .git (directory, or a worktree/
// submodule file pointing at the real git dir) and read HEAD.
func (c *chatTUI) gitBranchName(workspace string) string {
	if workspace == "" {
		return ""
	}
	dir, err := filepath.Abs(workspace)
	if err != nil {
		return ""
	}
	for {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			gitDir := gitPath
			if !info.IsDir() {
				raw, err := os.ReadFile(gitPath)
				if err != nil {
					return ""
				}
				ref := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(raw)), "gitdir:"))
				if ref == "" {
					return ""
				}
				if !filepath.IsAbs(ref) {
					ref = filepath.Join(dir, ref)
				}
				gitDir = ref
			}
			raw, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
			if err != nil {
				return ""
			}
			head := strings.TrimSpace(string(raw))
			const prefix = "ref: refs/heads/"
			if strings.HasPrefix(head, prefix) {
				return strings.TrimPrefix(head, prefix)
			}
			// Pi reports "detached" for a detached HEAD.
			return "detached"
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// Pi has no outer frame padding: rows run edge to edge and components pad
// themselves (messages and tool bands by one column).
func (c *chatTUI) currentPadding() int { return 0 }

func (c *chatTUI) compactOutput() bool { return c.currentContentWidth() < 72 }

func compactID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 16 {
		return id
	}
	return id[:8] + "…" + id[len(id)-5:]
}

func compactText(text string, max int) string {
	text = strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if text == "" {
		return "(untitled)"
	}
	return truncate(text, max)
}

func compactMaybe(text string, compact bool, max int) string {
	if !compact {
		return text
	}
	return compactText(text, max)
}

func (c *chatTUI) currentContentWidth() int {
	if c.outputWidth > 0 {
		return c.outputWidth
	}
	if c.app == nil {
		return 78
	}
	w, _ := c.app.Size()
	padding := 0 // Pi draws transcript, editor and footer edge to edge.
	contentWidth := w - (padding * 2)
	if contentWidth < 20 {
		contentWidth = 20
	}
	return contentWidth
}

func (c *chatTUI) horizontalRule(width int) string {
	if width < 8 {
		width = 8
	}
	return strings.Repeat("─", width)
}

func (c *chatTUI) wrapLines(text string, width int) []string {
	if width < 1 {
		width = 1
	}
	out := []string{}
	for _, raw := range strings.Split(text, "\n") {
		runes := []rune(raw)
		if len(runes) == 0 {
			out = append(out, "")
			continue
		}
		for len(runes) > width {
			out = append(out, string(runes[:width]))
			runes = runes[width:]
		}
		out = append(out, string(runes))
	}
	if len(out) == 0 {
		return []string{""}
	}
	return out
}

func (c *chatTUI) renderLineBlock(lines []string, style gotui.Style) *gotui.Element {
	block := gotui.New(
		gotui.WithDirection(gotui.Column),
		gotui.WithWidthPercent(100),
		gotui.WithHeight(len(lines)),
	)
	for _, line := range lines {
		block.AddChild(c.renderInlineStyledLine(line, style))
	}
	return block
}

type tuiInlineSegment struct {
	Text   string
	Code   bool
	Styles []string // Markdown element styles, outermost first
}

// parseTUIInlineSegments splits a projected line at its (possibly nested)
// inline-code and Markdown style markers.
func parseTUIInlineSegments(line string) []tuiInlineSegment {
	if !strings.Contains(line, "\x00gi-") {
		return []tuiInlineSegment{{Text: line}}
	}
	var segments []tuiInlineSegment
	var stack []string
	var text strings.Builder
	flush := func() {
		if text.Len() == 0 {
			return
		}
		seg := tuiInlineSegment{Text: text.String()}
		for _, name := range stack {
			if name == "code" {
				seg.Code = true
			} else {
				seg.Styles = append(seg.Styles, name)
			}
		}
		segments = append(segments, seg)
		text.Reset()
	}
	for i := 0; i < len(line); {
		if marker, open, name, ok := markerToken(line[i:]); ok {
			flush()
			if open {
				stack = append(stack, name)
			} else if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			i += len(marker)
			continue
		}
		text.WriteByte(line[i])
		i++
	}
	flush()
	if len(segments) == 0 {
		return []tuiInlineSegment{{Text: ""}}
	}
	return segments
}

func (c *chatTUI) renderInlineStyledLine(line string, style gotui.Style) *gotui.Element {
	// go-tui word wrapping discards leading whitespace in rich text. Keep it
	// as element padding so fenced code indentation survives the PTY renderer.
	leading := len(line) - len(strings.TrimLeft(line, " "))
	// Table rows are already wrapped by the Markdown renderer. Preserve their
	// literal ASCII padding by disabling a second rich-text word-wrap pass.
	// Replacing it with NBSP breaks terminal grapheme handling (notably flags
	// in tmux), so frame diffs can leave stale border cells while scrolling.
	preformatted := leading > 0 || strings.HasPrefix(line, "|") || strings.HasPrefix(line, "+") || strings.HasPrefix(line, "│") || strings.HasPrefix(line, "┌") || strings.HasPrefix(line, "├") || strings.HasPrefix(line, "└")
	line = line[leading:]
	segments := parseTUIInlineSegments(line)
	options := []gotui.Option{gotui.WithWidthPercent(100)}
	if preformatted {
		options = append(options, gotui.WithWrap(false))
	}
	if leading > 0 {
		options = append(options, gotui.WithPaddingTRBL(0, 0, 0, leading))
	}
	if len(segments) == 1 && !segments[0].Code && len(segments[0].Styles) == 0 {
		spans := transcriptLinkSpans(segments[0].Text, style)
		return gotui.New(append(options, gotui.WithRichText(spans...))...)
	}
	// A row of child Elements lays each styled fragment out independently. At
	// narrow widths it can put the highlighted word on a different line and
	// drop following text. One rich-text element wraps the entire styled run.
	spans := make([]gotui.TextSpan, 0, len(segments))
	for _, seg := range segments {
		if seg.Text == "" {
			continue
		}
		segStyle := style
		for _, name := range seg.Styles {
			segStyle = piMarkdownStyle(segStyle, name)
		}
		if seg.Code {
			segStyle = style.Foreground(piMdCode)
			// go-tui's rich-text word wrapper collapses ASCII spaces in
			// spans, including the space following an ANSI style change.
			// Non-breaking spaces keep code's exact visual width and prevent
			// the following word from being pulled into the code span.
			text := seg.Text
			if !preformatted {
				text = strings.ReplaceAll(text, " ", "\u00a0")
			}
			spans = append(spans, gotui.TextSpan{Text: text, Style: segStyle})
		} else {
			parts := transcriptLinkSpans(seg.Text, segStyle)
			spans = append(spans, parts...)
		}
	}
	return gotui.New(append(options, gotui.WithRichText(spans...))...)
}

func (c *chatTUI) buildTranscriptRenderableBlocks(lines []string) []transcriptRenderableBlock {
	blocks := make([]transcriptRenderableBlock, 0, len(lines))
	seenErrorBlocks := map[string]bool{}
	c.ensureTranscriptBlockState()
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if meta, ok := parseTranscriptBlockMarker(line); ok {
			if meta.Kind == startupHeaderKind {
				// gi's startup header; Body carries its content signature so
				// cached renders follow MCP/extension/skill changes.
				blocks = append(blocks, transcriptRenderableBlock{Kind: startupHeaderKind, Key: meta.Key, Expanded: c.transcriptExpanded[meta.Key], Expandable: true, Body: c.startupSignature()})
				continue
			}
			body := make([]string, 0, 4)
			j := i + 1
			for j < len(lines) && strings.HasPrefix(lines[j], "│ ") {
				part := strings.TrimPrefix(lines[j], "│ ")
				if meta.Kind == "tool" || meta.Kind == "bash" {
					part = plainTerminalOutput(part)
				}
				body = append(body, part)
				j++
			}
			if meta.MarkdownSource != "" && (meta.Kind == "user" || meta.Kind == "assistant") {
				block := transcriptRenderableBlock{Kind: meta.Kind, MarkdownSource: meta.MarkdownSource, Body: body, Expanded: true}
				if len(body) > 0 {
					block.Header = body[0]
					block.Body = body[1:]
				}
				block.HeaderStyle, block.BodyStyle, block.HintStyle, _, _ = transcriptBlockPalette(meta.Kind, "", false)
				blocks = append(blocks, block)
				i = j - 1
				continue
			}
			expanded := c.transcriptExpanded[meta.Key]
			headStyle, bodyStyle, hintStyle, borderStyle, border := transcriptBlockPalette(meta.Kind, meta.Status, c.selectedTranscriptBlock == meta.Key)
			header := plainTerminalOutput(meta.Title)
			subheader := ""
			if meta.Kind == "tool" {
				// Header is Pi's call line; timing/args are rendered by the tool shell.
			} else if meta.Kind == "thinking" || meta.Kind == "thinking_indicator" {
				subheader = ""
			} else {
				timeBits := []string{}
				if clock := formatBlockClock(meta.StartedAt); clock != "" {
					timeBits = append(timeBits, clock)
				}
				if strings.TrimSpace(meta.EndedAt) != "" {
					if elapsed := formatBlockElapsed(meta.StartedAt, meta.EndedAt); elapsed != "" {
						timeBits = append(timeBits, elapsed)
					}
				}
				if meta.Detail != "" {
					timeBits = append(timeBits, meta.Detail)
				}
				subheader = strings.Join(timeBits, " · ")
			}
			expandable := len(body) > 2
			selectedHint := borderStyle
			previewLimit := 0
			previewTail := false
			if meta.Kind == "thinking" || meta.Kind == "thinking_indicator" {
				expandable = false
				selectedHint = ""
			}
			if meta.Kind == "tool" {
				body, expandable = c.applyToolRenderMode(meta.Title, body)
				previewLimit = toolPreviewLines
				if isShellTool(meta.Title) {
					previewLimit, previewTail = toolShellPreviewLines, true
				}
				expandable = expandable && len(body) > previewLimit
				if meta.Title == "read" && meta.ToolPath != "" && meta.Status == "ok" {
					expandable = len(body) > 0 && c.extensionToolModes[meta.Title] == ""
				}
				if meta.Title == "edit" && (meta.EditDiff != nil || meta.EditError != "") {
					expandable = false
				}
				if meta.Title == "write" && meta.ToolContent != nil {
					expandable = c.extensionToolModes[meta.Title] == "" && len(strings.Split(*meta.ToolContent, "\n")) > previewLimit
				}
				if meta.Title == "codemode" {
					// Pi's codemode previews: 10 script lines, 8 calls, 5 output lines.
					code := 0
					if meta.ToolContent != nil {
						code = len(strings.Split(strings.TrimRight(*meta.ToolContent, "\n"), "\n"))
					}
					expandable = code > codemodeCodePreviewLines || len(meta.Calls) > codemodeCallPreviewCount || len(body) > codemodeOutputPreviewLines
				}
			}
			if meta.Kind == "bash" {
				previewLimit = bashPreviewLines
				previewTail = true
				expandable = len(body) > bashPreviewLines
			}
			blocks = append(blocks, transcriptRenderableBlock{Key: meta.Key, Kind: meta.Kind, MarkdownSource: meta.MarkdownSource, Header: header, Subheader: subheader, Body: body, Expandable: expandable, Expanded: expanded, PreviewLimit: previewLimit, PreviewTail: previewTail, Footer: strings.TrimSpace(meta.Footer), Status: meta.Status, Selected: c.selectedTranscriptBlock == meta.Key, Border: gotui.BorderRounded, BorderStyle: border, HeaderStyle: headStyle, BodyStyle: bodyStyle, HintStyle: hintStyle, SelectedHint: selectedHint, ToolPath: meta.ToolPath, ToolContent: meta.ToolContent, ToolRange: meta.ToolRange, ToolNotice: meta.ToolNotice, EditDiff: meta.EditDiff, EditError: meta.EditError, Calls: meta.Calls, FullOutputPath: meta.FullOutputPath, ToolArg: strings.TrimSpace(meta.Detail), StartedAt: meta.StartedAt, EndedAt: meta.EndedAt})
			i = j - 1
			continue
		}
		switch {
		case strings.HasPrefix(line, "local$ "):
			body := make([]string, 0, 4)
			j := i + 1
			for j < len(lines) {
				if strings.HasPrefix(lines[j], "│ ") {
					body = append(body, plainTerminalOutput(strings.TrimPrefix(lines[j], "│ ")))
					j++
					continue
				}
				if strings.HasPrefix(lines[j], "error:") {
					body = append(body, lines[j])
					j++
					continue
				}
				break
			}
			key := fmt.Sprintf("local:%d:%s", i, line)
			headStyle, bodyStyle, hintStyle, borderStyle, border := transcriptBlockPalette("local", "info", c.selectedTranscriptBlock == key)
			blocks = append(blocks, transcriptRenderableBlock{Key: key, Kind: "local", Header: line, Body: body, Expandable: len(body) > 2, Expanded: c.transcriptExpanded[key], Selected: c.selectedTranscriptBlock == key, Border: gotui.BorderRounded, BorderStyle: border, HeaderStyle: headStyle, BodyStyle: bodyStyle, HintStyle: hintStyle, SelectedHint: borderStyle})
			i = j - 1
		case isTUIErrorLine(line):
			key := tuiErrorDedupKey(line)
			if seenErrorBlocks[key] {
				continue
			}
			seenErrorBlocks[key] = true
			body := []string{trimTUIErrorLine(line)}
			blockKey := "error:" + key
			headStyle, bodyStyle, hintStyle, selectedHint, borderStyle := transcriptBlockPalette("error", "error", c.selectedTranscriptBlock == blockKey)
			blocks = append(blocks, transcriptRenderableBlock{Key: blockKey, Kind: "error", Header: "Error", Body: body, Selected: c.selectedTranscriptBlock == blockKey, Border: gotui.BorderRounded, BorderStyle: borderStyle, HeaderStyle: headStyle, BodyStyle: bodyStyle, HintStyle: hintStyle, SelectedHint: selectedHint})
		default:
			kind := "plain"
			switch {
			case strings.HasPrefix(line, "sys:"):
				kind = "system"
			case strings.HasPrefix(line, "you: ") || strings.HasPrefix(line, "you [queued]: "):
				kind = "user"
			case strings.HasPrefix(line, c.cfg.AssistantName+":"):
				kind = "assistant"
			}
			headStyle, bodyStyle, hintStyle, _, _ := transcriptBlockPalette(kind, "", false)
			var body []string
			start := i
			if kind == "user" || kind == "assistant" {
				// Markdown projection indents continuation rows by its speaker
				// prefix. Keep one message band, not padding around every row.
				prefix := "you: "
				if strings.HasPrefix(line, "you [queued]: ") {
					prefix = "you [queued]: "
				}
				if kind == "assistant" {
					prefix = c.cfg.AssistantName + ": "
				}
				indent := strings.Repeat(" ", utf8.RuneCountInString(prefix))
				for i+1 < len(lines) && strings.HasPrefix(lines[i+1], indent) {
					i++
					body = append(body, lines[i])
				}
			}
			blocks = append(blocks, transcriptRenderableBlock{Key: fmt.Sprintf("%s:%d:%s", kind, start, line), Kind: kind, Header: line, Body: body, HeaderStyle: headStyle, BodyStyle: bodyStyle, HintStyle: hintStyle})
		}
	}
	return blocks
}

func shouldRenderTranscriptStatusLabel(kind, status string) bool {
	if status == "" {
		return false
	}
	if (kind == "tool" || kind == "bash") && status == "ok" {
		return false
	}
	return true
}

func isTUIErrorLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "error:") {
		return true
	}
	if strings.HasPrefix(lower, "sys:") {
		bodyLower := strings.ToLower(strings.TrimSpace(trimmed[4:]))
		return strings.Contains(bodyLower, "inference error") || strings.Contains(bodyLower, "max retries exceeded")
	}
	return strings.Contains(lower, "max retries exceeded")
}

func trimTUIErrorLine(line string) string {
	trimmed := strings.TrimSpace(line)
	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "sys:"):
		return strings.TrimSpace(trimmed[4:])
	case strings.HasPrefix(lower, "error:"):
		msg := strings.TrimSpace(trimmed[6:])
		if msg != "" {
			return msg
		}
	}
	return trimmed
}

func tuiErrorDedupKey(line string) string {
	msg := strings.ToLower(trimTUIErrorLine(line))
	msg = strings.Join(strings.Fields(msg), " ")
	if strings.Contains(msg, "max retries exceeded") {
		return "max retries exceeded"
	}
	if strings.HasPrefix(msg, "inference error:") {
		msg = strings.TrimSpace(strings.TrimPrefix(msg, "inference error:"))
	}
	if msg == "" {
		return "error"
	}
	return msg
}

// diffLineStyle returns Pi's toolDiff* color for unified-diff lines in tool/bash
// output: green for additions, red for removals, dim for hunk/diff headers.
func diffLineStyle(line string) (gotui.Style, bool) {
	trimmed := strings.TrimLeft(line, " ")
	switch {
	case strings.HasPrefix(trimmed, "+++") || strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "@@"):
		return piFg(piDiffContext), true
	case strings.HasPrefix(trimmed, "+"):
		return piFg(piDiffAdded), true
	case strings.HasPrefix(trimmed, "-"):
		return piFg(piDiffRemoved), true
	}
	return gotui.Style{}, false
}

// transcriptBlockPalette maps transcript blocks to Pi's dark theme roles:
// tool titles use toolTitle (bold), output toolOutput, bash commands bashMode,
// thinking thinkingText (italic), custom/hook messages customMessageLabel and
// customMessageText, status lines dim, errors error. Assistant Markdown uses
// the terminal's default foreground, as Pi's Markdown component does.
func transcriptBlockPalette(kind, status string, selected bool) (gotui.Style, gotui.Style, gotui.Style, string, gotui.Style) {
	head := gotui.NewStyle()
	body := gotui.NewStyle()
	hint := piFg(piMuted)
	border := piFg(piBorderMuted)
	selectedHint := "F6/F7 select · F8 toggle · click to expand"
	if selected {
		border = border.Bold()
		selectedHint = "selected · F6/F7 move · F8 toggle · click to expand"
	}
	failed := status == "error" || status == "failed"
	switch kind {
	case "user":
		head, body = piFg(piUserText), piFg(piUserText)
	case "tool":
		head, body = piFg(piToolTitle).Bold(), piFg(piToolOutput)
	case "bash", "local":
		head, body = piFg(piBashMode).Bold(), piFg(piMuted)
	case "thought", "thinking", "thinking_indicator":
		head, body = piFg(piThinkingText).Italic(), piFg(piThinkingText).Italic()
	case "hook", "route", "dispatcher", "subturn", "compact":
		head, body = piFg(piCustomLabel).Bold(), piFg(piCustomText)
	case "error":
		head, body = piFg(piError).Bold(), piFg(piError)
	case "system":
		head, body = piFg(piDim), piFg(piDim)
	}
	if failed && kind != "user" {
		head = piFg(piError).Bold()
	} else if status == "skipped" {
		head = piFg(piWarning).Bold()
	}
	return head, body, hint, selectedHint, border
}

func brailleSpinnerFrame(t time.Time) string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	if t.IsZero() {
		t = time.Now()
	}
	idx := int((t.UnixMilli() / 80) % int64(len(frames)))
	if idx < 0 {
		idx = 0
	}
	return frames[idx]
}

func (c *chatTUI) renderTranscriptBlock(block transcriptRenderableBlock) *gotui.Element {
	if block.Kind == "thinking_indicator" {
		// Pi shows turn activity in the editor's top border, not the transcript.
		return gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(0))
	}
	if block.Kind == "tool" && block.Header == "edit" {
		// Pi's edit box takes its colour from the preview; a result error the
		// preview did not show sits below it.
		banded := block
		banded.Status = editBandStatus(block)
		box := padTranscriptBlock(c.renderTranscriptBlockContent(block), banded)
		below := c.editResultBelow(block)
		if below == nil {
			return box
		}
		wrapper := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100))
		wrapper.AddChild(box)
		wrapper.AddChild(below)
		return wrapper
	}
	return padTranscriptBlock(c.renderTranscriptBlockContent(block), block)
}

func (c *chatTUI) renderTranscriptBlockContent(block transcriptRenderableBlock) *gotui.Element {
	switch block.Kind {
	case startupHeaderKind:
		return c.renderStartupHeader(block.Expanded)
	case "tool":
		return c.renderPiToolBlock(block)
	case "bash":
		return c.renderPiBashBlock(block)
	}
	if block.Kind == "user" || block.Kind == "assistant" {
		if block.MarkdownSource != "" {
			message := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100))
			// Account for message-band padding.
			for _, line := range renderMarkdownTranscript("", block.MarkdownSource, c.transcriptBlockContentWidth(block.Kind)) {
				message.AddChild(c.renderInlineStyledLine(line, block.BodyStyle))
			}
			return message
		}
		message := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100))
		// Speaker prefixes remain in the transcript for message boundaries,
		// but are not part of the visible Markdown. Remove only the projector's
		// continuation indent; preserve real indentation inside code blocks.
		prefix := "you: "
		if block.Kind == "assistant" {
			prefix = c.cfg.AssistantName + ": "
		} else if strings.HasPrefix(block.Header, "you [queued]: ") {
			prefix = "you [queued]: "
		}
		prefixed := strings.HasPrefix(block.Header, prefix)
		if !prefixed {
			prefix = ""
		}
		message.AddChild(c.renderInlineStyledLine(strings.TrimPrefix(block.Header, prefix), block.HeaderStyle))
		indent := strings.Repeat(" ", utf8.RuneCountInString(prefix))
		for _, line := range block.Body {
			message.AddChild(c.renderInlineStyledLine(strings.TrimPrefix(line, indent), block.BodyStyle))
		}
		return message
	}
	if block.Kind == "thought" {
		container := gotui.New(
			gotui.WithDirection(gotui.Column),
			gotui.WithWidthPercent(100),
		)
		ref := gotui.NewRef()
		ref.Set(container)
		c.transcriptBlockRefs = append(c.transcriptBlockRefs, transcriptBlockHitTarget{Key: block.Key, Ref: ref})
		body := block.Body
		if block.MarkdownSource != "" {
			// Reproject source at the padded inner width. Wrapping at the outer
			// width first leaves orphan words when rich text wraps a second time.
			body = renderMarkdownTranscript("", block.MarkdownSource, c.transcriptBlockContentWidth(block.Kind))
		}
		if len(body) == 0 {
			container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(fmt.Sprintf("%s Thinking...", brailleSpinnerFrame(time.Now()))), gotui.WithTextStyle(block.BodyStyle)))
			return container
		}
		if c.cfg.HideThinkingBlock { // Pi's hideThinkingBlock label
			container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText("Thinking..."), gotui.WithTextStyle(block.BodyStyle.Italic())))
			return container
		}
		for _, line := range body {
			container.AddChild(c.renderInlineStyledLine(line, block.BodyStyle))
		}
		return container
	}
	if len(block.Body) == 0 && block.Subheader == "" && block.Kind != "local" {
		headerText := block.Header
		if block.Status == "running" {
			headerText = fmt.Sprintf("%s %s", brailleSpinnerFrame(time.Now()), headerText)
		} else if shouldRenderTranscriptStatusLabel(block.Kind, strings.TrimSpace(block.Status)) {
			headerText = fmt.Sprintf("%s [%s]", headerText, block.Status)
		}
		return c.renderInlineStyledLine(headerText, block.HeaderStyle)
	}
	container := gotui.New(
		gotui.WithDirection(gotui.Column),
		gotui.WithWidthPercent(100),
	)
	ref := gotui.NewRef()
	ref.Set(container)
	c.transcriptBlockRefs = append(c.transcriptBlockRefs, transcriptBlockHitTarget{Key: block.Key, Ref: ref})
	statusLabel := strings.TrimSpace(block.Status)
	headerText := block.Header
	if statusLabel == "running" {
		headerText = fmt.Sprintf("%s %s", brailleSpinnerFrame(time.Now()), headerText)
	} else if shouldRenderTranscriptStatusLabel(block.Kind, statusLabel) {
		headerText = fmt.Sprintf("%s [%s]", headerText, statusLabel)
	}
	container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(headerText), gotui.WithTextStyle(block.HeaderStyle)))
	if block.Subheader != "" {
		container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(block.Subheader), gotui.WithTextStyle(block.HintStyle)))
	}
	visibleBody := block.Body
	hiddenCount := 0
	previewLimit := 2
	if block.PreviewLimit > 0 {
		previewLimit = block.PreviewLimit
	}
	if block.Expandable && !block.Expanded {
		if len(visibleBody) > previewLimit {
			hiddenCount = len(visibleBody) - previewLimit
			if block.PreviewTail {
				visibleBody = visibleBody[len(visibleBody)-previewLimit:]
			} else {
				visibleBody = visibleBody[:previewLimit]
			}
		}
	}
	for _, line := range visibleBody {
		style := block.BodyStyle
		if block.Kind == "tool" || block.Kind == "bash" {
			if s, ok := diffLineStyle(line); ok {
				style = s
			}
		}
		container.AddChild(c.renderInlineStyledLine(line, style))
	}
	if block.Expandable && hiddenCount > 0 {
		hint := fmt.Sprintf("… %d more line(s) · F8 expand", hiddenCount)
		container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(hint), gotui.WithTextStyle(block.HintStyle)))
	} else if block.Expandable && block.Expanded && block.PreviewTail {
		container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText("F8 collapse"), gotui.WithTextStyle(block.HintStyle)))
	}
	if strings.TrimSpace(block.Footer) != "" {
		container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(block.Footer), gotui.WithTextStyle(block.HintStyle)))
	}
	return container
}

func (c *chatTUI) transcriptRenderWidth() int {
	width := c.currentContentWidth()
	if width <= 0 {
		return 80
	}
	return width
}

func (c *chatTUI) loadTranscript() []string {
	if c.store == nil {
		return append([]string(nil), c.transcript...)
	}
	msgs, err := c.store.ListMessages(context.Background(), c.sessionID)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(msgs))
	// Calls precede results. Scope IDs to their turn because providers can reuse them.
	calls := map[string]any{}
	for _, m := range msgs {
		if kind, _ := m.Payload["kind"].(string); kind == "tool_calls" {
			turnID, _ := m.Payload["turn_id"].(string)
			raw, _ := json.Marshal(m.Payload["tool_calls"])
			var entries []struct {
				ID        string         `json:"id"`
				Arguments map[string]any `json:"arguments"`
			}
			if json.Unmarshal(raw, &entries) == nil {
				for _, call := range entries {
					calls[turnID+"\x00"+call.ID] = call.Arguments
				}
			}
		}
		if m.Role == "tool_result" {
			turnID, _ := m.Payload["turn_id"].(string)
			callID, _ := m.Payload["tool_call_id"].(string)
			arguments := calls[turnID+"\x00"+callID]
			if arguments == nil {
				arguments = m.Payload["arguments"]
			}
			out = append(out, c.renderToolResultWithArguments(m, arguments)...)
		} else {
			out = append(out, c.renderMessageLines(m, c.transcriptRenderWidth())...)
		}
	}
	return c.pruneTranscript(out)
}

func (c *chatTUI) renderToolCallSummaryLines(m store.Message, width int) []string {
	return nil
}

func (c *chatTUI) renderMessageLines(m store.Message, width int) []string {
	kind, _ := m.Payload["kind"].(string)
	if m.Role == "tool_result" || kind == "tool_result" {
		return c.renderToolResultLines(m)
	}
	if kind == "tool_calls" {
		return c.renderToolCallSummaryLines(m, width)
	}
	if kind == "compaction" {
		return c.renderCompactionMessageLines(m)
	}
	if kind == turn.BranchSummaryKind {
		return c.renderBranchSummaryLines(m)
	}
	prefix := "you: "
	switch m.Role {
	case "assistant":
		prefix = c.cfg.AssistantName + ": "
	case "system":
		prefix = "sys: "
	}
	if m.Role == "user" || m.Role == "assistant" || looksLikeMarkdown(m.Content) {
		return renderChatMarkdown(m.Role, prefix, m.Content, width)
	}
	return []string{c.renderMessageLine(m)}
}

func (c *chatTUI) renderToolResultLines(m store.Message) []string {
	return c.renderToolResultWithArguments(m, m.Payload["arguments"])
}

func (c *chatTUI) renderToolResultWithArguments(m store.Message, arguments any) []string {
	toolName, _ := m.Payload["tool_name"].(string)
	if toolName == "" {
		toolName = "tool"
	}
	isErr, _ := m.Payload["is_error"].(bool)
	status := "ok"
	if isErr {
		status = "error"
	}
	meta := transcriptBlockMeta{Key: "msg:" + m.ID, Kind: "tool", Title: toolName, Status: status, StartedAt: strings.TrimSpace(m.CreatedAt), EndedAt: strings.TrimSpace(m.CreatedAt)}
	meta.Detail = toolInvocationText(toolName, arguments)
	setFileToolArguments(&meta, arguments)
	setCodemodeArguments(&meta, arguments)
	content := m.Content
	if toolName == "codemode" {
		meta.Calls, meta.FullOutputPath = codemodeDetails(m.Payload["details"])
		content = stripScriptHeader(content)
	}
	setFileToolDetails(&meta, m.Payload["details"], strings.TrimSpace(plainTerminalOutput(content)))
	lines := []string{encodeTranscriptBlockMarker(meta)}
	for _, line := range toolOutputBodyLines(content) {
		if toolName != "read" {
			line = truncate(line, 200)
		}
		lines = append(lines, "│ "+line)
	}
	return lines
}

func (c *chatTUI) renderCompactionMessageLines(m store.Message) []string {
	tokens := toInt(m.Payload["tokens_before"], 0)
	return []string{encodeTranscriptBlockMarker(transcriptBlockMeta{Key: "msg:" + m.ID, Kind: "compact", Title: "Context compacted", Status: "ok", StartedAt: strings.TrimSpace(m.CreatedAt), EndedAt: strings.TrimSpace(m.CreatedAt)}), fmt.Sprintf("│ summary=%s", truncate(m.Content, 160)), fmt.Sprintf("│ tokens_before=%d", tokens)}
}

// renderBranchSummaryLines shows a /tree branch summary as a block (Pi's
// BranchSummaryMessageComponent: collapsed until expanded).
func (c *chatTUI) renderBranchSummaryLines(m store.Message) []string {
	lines := []string{encodeTranscriptBlockMarker(transcriptBlockMeta{Key: "msg:" + m.ID, Kind: "compact", Title: "Branch summary", Status: "ok", StartedAt: strings.TrimSpace(m.CreatedAt), EndedAt: strings.TrimSpace(m.CreatedAt)})}
	for _, line := range strings.Split(m.Content, "\n") {
		lines = append(lines, "│ "+line)
	}
	return lines
}

func (c *chatTUI) renderMessageLine(m store.Message) string {
	kind, _ := m.Payload["kind"].(string)
	switch {
	case kind == "compaction":
		tokens := toInt(m.Payload["tokens_before"], 0)
		return fmt.Sprintf("compact: %s (tokens_before=%d)", truncate(m.Content, 160), tokens)
	case m.Role == "tool_result" || kind == "tool_result":
		toolName, _ := m.Payload["tool_name"].(string)
		if toolName == "" {
			toolName = "tool"
		}
		isErr, _ := m.Payload["is_error"].(bool)
		status := "ok"
		if isErr {
			status = "error"
		}
		return fmt.Sprintf("tool[%s/%s]: %s", toolName, status, foldedContentSummary(plainTerminalOutput(m.Content), 160))
	}
	prefix := "you"
	switch m.Role {
	case "assistant":
		prefix = c.cfg.AssistantName
	case "system":
		prefix = "sys"
	}
	return fmt.Sprintf("%s: %s", prefix, truncate(m.Content, 200))
}

func foldedContentSummary(content string, maxLen int) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return "(empty)"
	}
	lines := strings.Split(trimmed, "\n")
	first := strings.TrimSpace(lines[0])
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			first = strings.TrimSpace(line)
			break
		}
	}
	first = strings.Join(strings.Fields(first), " ")
	if len(lines) <= 1 {
		return truncate(first, maxLen)
	}
	summary := fmt.Sprintf("%d lines · %s", len(lines), first)
	return truncate(summary, maxLen)
}

func toInt(value any, fallback int) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	default:
		return fallback
	}
}

func (c *chatTUI) transcriptViewportHeight() int {
	if c.transcriptRef == nil || c.transcriptRef.El() == nil {
		return 4
	}
	_, h := c.transcriptRef.El().ViewportSize()
	if h <= 0 {
		return 4
	}
	return h
}

func (c *chatTUI) scrollTranscript(delta int) {
	maxScroll := c.transcriptMaxScroll()
	if c.transcriptRef != nil && c.transcriptRef.El() != nil {
		_, c.transcriptScroll = c.transcriptRef.El().ScrollOffset()
	}
	c.transcriptScroll += delta
	if c.transcriptScroll < 0 {
		c.transcriptScroll = 0
	}
	if c.transcriptScroll > maxScroll {
		c.transcriptScroll = maxScroll
	}
	c.stickToBottom = c.transcriptScroll >= maxScroll
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) pageTranscript(delta int) {
	// Pi keeps four rows of context when paging.
	step := c.transcriptViewportHeight() - 4
	if step < 1 {
		step = 1
	}
	c.scrollTranscript(delta * step)
}

func (c *chatTUI) scrollTranscriptToTop() {
	c.setTranscriptPosition(0)
	c.stickToBottom = false
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) scrollTranscriptToBottom() {
	c.setTranscriptPosition(c.transcriptMaxScroll())
	c.stickToBottom = true
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) visibleTranscript() []string {
	lines := c.pruneTranscript(append([]string(nil), c.transcript...))
	// Fullscreen shows gi's startup header first (Pi's header container);
	// regular mode prints it once to scrollback instead.
	if !c.regularMode && c.startupHeaderShown() {
		return append([]string{startupHeaderMarker()}, lines...)
	}
	if len(lines) == 0 {
		return []string{"(no messages yet)"}
	}
	return lines
}

type tuiContextSummary struct {
	sessionTitle  string
	agentID       string
	parent        string
	model         string
	provider      string
	thinking      string
	status        string
	messageCount  int
	turnCount     int
	queuedTurns   int
	steeringDepth int
	contextTokens int
	contextWindow int
	inputTokens   int
	outputTokens  int
	cacheRead     int
	cacheWrite    int
	costTotal     float64
}

func (c *chatTUI) latestUsageTokens(turns []store.Turn) (input, output, total int) {
	if c.store == nil || len(turns) == 0 {
		return 0, 0, 0
	}
	ctx := context.Background()
	for i := len(turns) - 1; i >= 0; i-- {
		events, err := c.store.ListTurnEvents(ctx, turns[i].ID)
		if err != nil {
			continue
		}
		for j := len(events) - 1; j >= 0; j-- {
			usage, _ := events[j].Payload["usage"].(map[string]any)
			if usage == nil {
				continue
			}
			input = intFromAny(usage["input"])
			output = intFromAny(usage["output"])
			total = intFromAny(usage["total"])
			if total == 0 {
				total = intFromAny(usage["totalTokens"])
			}
			return input, output, total
		}
	}
	return 0, 0, 0
}

func (c *chatTUI) applyModelContextWindow(data *tuiContextSummary) {
	if data == nil {
		return
	}
	if contextWindow := inference.ResolveModelContextWindow(data.provider, data.model); contextWindow > 0 {
		data.contextWindow = contextWindow
	}
}

func (c *chatTUI) contextSummaryData() tuiContextSummary {
	data := tuiContextSummary{sessionTitle: c.sessionID, agentID: "agent", parent: "root", model: c.cfg.DefaultModel, provider: c.cfg.DefaultProvider, thinking: c.cfg.DefaultThinkingLevel, status: "idle", contextWindow: c.cfg.Compaction.ContextWindow, contextTokens: c.lastContextTokens, inputTokens: c.lastInputTokens, outputTokens: c.lastOutputTokens, cacheRead: c.lastCacheRead, cacheWrite: c.lastCacheWrite, costTotal: c.lastCostTotal}
	c.applyModelContextWindow(&data)
	if c.store == nil || c.sessionID == "" {
		return data
	}
	session, err := c.store.GetSession(context.Background(), c.sessionID)
	if err != nil {
		return data
	}
	messages, _ := c.store.ListMessages(context.Background(), c.sessionID)
	turns, _ := c.store.ListTurns(context.Background(), c.sessionID)
	data.messageCount = len(messages)
	data.turnCount = len(turns)
	data.queuedTurns, _ = c.store.CountQueuedTurns(context.Background(), c.sessionID)
	data.steeringDepth, _ = c.store.SteeringQueueLength(context.Background(), c.sessionID)
	if data.inputTokens == 0 && data.outputTokens == 0 {
		data.inputTokens, data.outputTokens, _ = c.latestUsageTokens(turns)
	}
	data.contextTokens = 0
	if measured, err := c.store.LatestContextMeasurement(context.Background(), c.sessionID); err == nil && measured != nil {
		data.contextTokens = measured.Tokens
	}
	data.sessionTitle = session.Title
	data.agentID = c.agentIDForSession(session)
	if session.ParentSessionID != "" {
		data.parent = session.ParentSessionID
	}
	state := session.State
	choice := inference.SessionModel(state, inference.SessionModelChoice{Model: data.model, Provider: data.provider, Thinking: data.thinking})
	data.model, data.provider, data.thinking = choice.Model, choice.Provider, choice.Thinking
	if v, ok := state["status"].(string); ok && v != "" {
		data.status = v
	}
	c.applyModelContextWindow(&data)
	return data
}

func (c *chatTUI) contextSummaryLines(width int) []string {
	data := c.contextSummaryData()
	line := fmt.Sprintf("@%s · %s · %s · m%d/t%d", data.agentID, data.model, data.thinking, data.messageCount, data.turnCount)
	if data.contextTokens > 0 {
		line += " · " + formatContextUsage(data.contextTokens, data.contextWindow)
	}
	if data.queuedTurns > 0 || data.steeringDepth > 0 {
		line += fmt.Sprintf(" · q%d/s%d", data.queuedTurns, data.steeringDepth)
	}
	if data.sessionTitle != "" && data.sessionTitle != "@"+data.agentID {
		line = fmt.Sprintf("%s · %s", data.sessionTitle, line)
	}
	return c.wrapLines(line, width)
}

func (c *chatTUI) contextSummary() string {
	return strings.Join(c.contextSummaryLines(9999), "\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func Main() {
	dbPath := flag.String("db", config.DefaultTUIDBPath(), "SQLite database path")
	workspace := flag.String("workspace", config.DefaultWorkspaceRoot(), "Workspace root")
	model := flag.String("model", "", "Override default model")
	mode := flag.String("tui-mode", "", "Terminal rendering: fullscreen or regular (native scrollback); default: tuiMode setting, else fullscreen")
	flag.Parse()
	if err := RunMode(*dbPath, *workspace, *model, *mode); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// resolveTUIMode picks the terminal mode like Pi: the --tui-mode flag, else
// the tuiMode setting, else fullscreen (Pi 1.0's default).
func resolveTUIMode(flagValue, setting string) string {
	if flagValue != "" {
		return flagValue
	}
	switch setting {
	case "regular", "fullscreen":
		return setting
	case "":
	default:
		log.Printf("tuiMode %q is not regular or fullscreen; using fullscreen", setting)
	}
	return "fullscreen"
}

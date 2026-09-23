package tui

import (
	"context"
	"fmt"
	"sync"
	"time"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/search/chunking"
	"github.com/rcarmo/gi/internal/search/indexer"
	searchstore "github.com/rcarmo/gi/internal/search/store"
)

var terminalIndexScopes = []string{"all", "notes", "skills"}

type terminalIndexResult struct {
	epoch, generation uint64
	status            searchstore.ScopeStatus
	err               error
}

type terminalIndex struct {
	active, busy                  bool
	scope, action                 int
	epoch                         uint64
	status                        searchstore.ScopeStatus
	err                           error
	savedScroll                   int
	resized                       bool
	savedFollow, savedInputActive bool
	scheduler                     *indexer.Scheduler
	configs                       map[string]searchstore.ScopeConfig
	initErr                       error
	ctx                           context.Context
	cancel                        context.CancelFunc
	wg                            sync.WaitGroup
	results                       chan terminalIndexResult
}

// Lifetime begins without a scan. All state below is UI-thread-owned except
// the bounded result channel and WaitGroup; background work captures its inputs.
func (c *chatTUI) initWorkspaceIndex() {
	p := &c.workspaceIndex
	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.results = make(chan terminalIndexResult, 1)
	if c.store == nil {
		p.initErr = fmt.Errorf("Index unavailable")
		return
	}
	p.configs = map[string]searchstore.ScopeConfig{}
	var configs []searchstore.ScopeConfig
	for _, scope := range terminalIndexScopes {
		s := c.cfg.WorkspaceIndex
		cfg, err := searchstore.ConfiguredScopeConfig(c.cfg.WorkspaceRoot, scope, s.ExtraRoots, s.ExtraExtensions, s.OptionalRoots, chunking.LineVersion)
		if err != nil {
			p.initErr = err
			return
		}
		configs = append(configs, cfg)
		p.configs[scope] = cfg
	}
	p.scheduler, p.initErr = indexer.NewScheduler(p.ctx, searchstore.NewRefreshStore(c.store.DB()), configs)
}
func (c *chatTUI) stopWorkspaceIndex() {
	if c.workspaceIndex.active && c.regularMode && c.app != nil {
		c.app.Terminal().ExitAltScreen()
		c.workspaceIndex.active = false
	}
	p := &c.workspaceIndex
	if p.cancel != nil {
		p.cancel()
	}
	if p.scheduler != nil {
		p.scheduler.Close()
	}
	p.wg.Wait()
}
func (c *chatTUI) openWorkspaceIndex() {
	if c.modelMenuOpen || c.search.active || c.editorAskActive || c.workspaceIndex.active {
		return
	}
	c.ensureInput()
	p := &c.workspaceIndex
	if p.ctx == nil {
		c.initWorkspaceIndex()
	}
	p.active = true
	p.epoch++
	p.scope = 0
	p.action = 0
	p.status = searchstore.ScopeStatus{}
	p.err = p.initErr
	p.savedScroll = c.transcriptScroll
	if c.transcriptRef != nil && c.transcriptRef.El() != nil {
		_, p.savedScroll = c.transcriptRef.El().ScrollOffset()
	}
	p.savedFollow = c.stickToBottom
	p.savedInputActive = c.inputActive
	c.stickToBottom = false
	c.clearTranscriptSelection()
	c.inputActive = false
	c.input.Blur()
	if c.app != nil {
		if c.regularMode {
			// Keep go-tui's inline renderer: its full-screen RenderFull calls
			// Terminal.Clear (ESC[3J), which erases main-screen history in tmux.
			p.resized = false
			c.app.Terminal().EnterAltScreen()
			w, h := c.app.Size()
			c.app.Dispatch(gotui.ResizeEvent{Width: w, Height: h})
		}
		c.app.BlurFocused()
		c.app.MarkDirty()
	}
	if !p.busy {
		c.requestWorkspaceIndex(false)
	}
}
func (c *chatTUI) closeWorkspaceIndex() {
	p := &c.workspaceIndex
	if !p.active {
		return
	}
	p.active = false
	p.epoch++
	if c.regularMode && c.app != nil {
		c.app.Terminal().ExitAltScreen()
		w, h := c.app.Size()
		c.app.Dispatch(gotui.ResizeEvent{Width: w, Height: h})
		// Width changes invalidate go-tui's inline history geometry. Establish
		// it on the restored main screen before the editor can grow again.
		if p.resized {
			c.app.PrintAboveln("sys: terminal resized to %dx%d", w, h)
		}
	}
	c.transcriptScroll, c.stickToBottom = p.savedScroll, p.savedFollow
	c.inputActive = p.savedInputActive
	if c.inputActive {
		c.input.Focus()
		if c.app != nil {
			c.app.FocusNext()
		}
	}
	if c.app != nil {
		c.app.MarkDirty()
	}
}
func (c *chatTUI) requestWorkspaceIndex(refresh bool) {
	p := &c.workspaceIndex
	if p.busy || p.initErr != nil || !p.active {
		return
	}
	scope := terminalIndexScopes[p.scope]
	cfg := p.configs[scope]
	epoch, generation := p.epoch, c.sessionGeneration
	p.busy = true
	p.err = nil
	ctx, scheduler, results := p.ctx, p.scheduler, p.results
	storage := searchstore.NewRefreshStore(c.store.DB())
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		var status searchstore.ScopeStatus
		var err error
		if refresh {
			var ticket *indexer.RefreshTicket
			ticket, err = scheduler.Request(scope)
			if err == nil {
				status, err = ticket.Wait(ctx)
			}
		} else {
			op, cancel := context.WithTimeout(ctx, 5*time.Second)
			status, err = storage.Status(op, cfg)
			cancel()
		}
		select {
		case results <- terminalIndexResult{epoch, generation, status, err}:
		case <-ctx.Done():
		}
	}()
	if c.app != nil {
		c.app.MarkDirty()
	}
}
func (c *chatTUI) handleWorkspaceIndexResult(r terminalIndexResult) {
	p := &c.workspaceIndex
	p.busy = false
	if !p.active || r.epoch != p.epoch || r.generation != c.sessionGeneration {
		// Reopening while an earlier request completed displays a fresh read, never
		// the old operation's status/error or a second automatic refresh.
		if p.active {
			c.requestWorkspaceIndex(false)
		}
		return
	}
	p.status, p.err = r.status, r.err
	if c.app != nil {
		c.app.MarkDirty()
	}
}
func (c *chatTUI) workspaceIndexKeys() gotui.KeyMap {
	return gotui.KeyMap{
		gotui.OnPreemptStop(gotui.KeyCtrlC, func(gotui.KeyEvent) {
			if c.app != nil {
				c.app.Stop()
			}
		}),
		gotui.OnPreemptStop(gotui.KeyEscape, func(gotui.KeyEvent) { c.closeWorkspaceIndex() }),
		gotui.OnPreemptStop(gotui.KeyLeft, func(gotui.KeyEvent) { c.changeIndexScope(-1) }),
		gotui.OnPreemptStop(gotui.KeyRight, func(gotui.KeyEvent) { c.changeIndexScope(1) }),
		gotui.OnPreemptStop(gotui.KeyUp, func(gotui.KeyEvent) { c.workspaceIndex.action = 0 }),
		gotui.OnPreemptStop(gotui.KeyDown, func(gotui.KeyEvent) { c.workspaceIndex.action = 1 }),
		gotui.OnPreemptStop(gotui.KeyEnter, func(gotui.KeyEvent) { c.requestWorkspaceIndex(c.workspaceIndex.action == 1) }),
		gotui.OnPreemptStop(gotui.AnyKey, func(gotui.KeyEvent) {}),
	}
}
func (c *chatTUI) changeIndexScope(delta int) {
	p := &c.workspaceIndex
	if p.busy {
		return
	}
	p.scope = (p.scope + delta + len(terminalIndexScopes)) % len(terminalIndexScopes)
	p.epoch++
	p.action = 0
	p.status = searchstore.ScopeStatus{}
	p.err = p.initErr
	c.requestWorkspaceIndex(false)
}
func (c *chatTUI) workspaceIndexLines(width int) []string {
	p := &c.workspaceIndex
	if !p.active {
		return nil
	}
	state := p.status.State
	if state == "" {
		state = "unknown"
	}
	if p.busy {
		state = "working"
	}
	detail := fmt.Sprintf("%d files · generation %d", p.status.IndexedFileCount, p.status.Generation)
	if p.status.LastError != "" {
		detail = p.status.LastError
	}
	if p.err != nil {
		state = "error"
		detail = p.err.Error()
	}
	actions := []string{"  Status", "  Reindex"}
	actions[p.action] = "› " + actions[p.action][2:]
	lines := []string{fmt.Sprintf("Index · %s · ←/→ scope · Esc close", terminalIndexScopes[p.scope]), "State: " + state, detail, actions[0], actions[1]}
	for i := range lines {
		lines[i] = selectorText(lines[i], width)
	}
	return lines
}
func (c *chatTUI) workspaceIndexHeight() int {
	if c.workspaceIndex.active {
		return 5
	}
	return 0
}
func (c *chatTUI) renderWorkspaceIndex(width int) *gotui.Element {
	return c.renderLineBlock(c.workspaceIndexLines(width), gotui.NewStyle().Foreground(gotui.Cyan))
}

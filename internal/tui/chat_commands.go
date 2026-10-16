package tui

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/shellenv"
	gitools "github.com/rcarmo/gi/internal/tools"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

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
		Key:        fmt.Sprintf("bash:%d", time.Now().UnixNano()),
		Kind:       "bash",
		Title:      "$ " + command,
		Status:     status,
		StartedAt:  startedAt.Format(time.RFC3339Nano),
		EndedAt:    endedAt.Format(time.RFC3339Nano),
		DurationMS: recordedToolDuration(endedAt.Sub(startedAt).Milliseconds()),
		Footer:     footer,
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

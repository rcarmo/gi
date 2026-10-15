package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
)

// Pi's FooterComponent: a dim pwd line ("~/path (branch) • session name") and
// a dim stats line with cumulative session usage on the left and the model on
// the right. Extension statuses share a single third line. Context usage turns
// warning above 70% and error above 90%.

type footerRow struct {
	text string
	// Optional colored span [start,end) in bytes (context percentage).
	accentStart, accentEnd int
	accent                 gotui.Color
}

type sessionUsageTotals struct {
	input, output, cacheRead, cacheWrite int
	cost                                 float64
}

type footerUsageCache struct {
	sessionID string
	turns     map[string]sessionUsageTotals // finished turns only
}

type footerAuthCache struct {
	at       time.Time
	statuses []inference.AuthStatus
}

// piFormatTokens matches Pi's formatTokens.
func piFormatTokens(count int) string {
	switch {
	case count < 1000:
		return fmt.Sprint(count)
	case count < 10000:
		return fmt.Sprintf("%.1fk", float64(count)/1000)
	case count < 1000000:
		return fmt.Sprintf("%dk", int(float64(count)/1000+0.5))
	case count < 10000000:
		return fmt.Sprintf("%.1fM", float64(count)/1000000)
	}
	return fmt.Sprintf("%dM", int(float64(count)/1000000+0.5))
}

func piFormatCwd(cwd string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return cwd
	}
	rel, err := filepath.Rel(filepath.Clean(home), filepath.Clean(cwd))
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return cwd
	}
	if rel == "." {
		return "~"
	}
	return "~" + string(filepath.Separator) + rel
}

func (c *chatTUI) footerAuthStatuses() []inference.AuthStatus {
	if c.footerAuth == nil || time.Since(c.footerAuth.at) > 5*time.Second {
		c.footerAuth = &footerAuthCache{at: time.Now(), statuses: inference.ListAuthStatus()}
	}
	return c.footerAuth.statuses
}

// sessionUsage sums each turn's final inference usage, as Pi sums usage over
// all session entries. Finished turns are cached; only live turns are rescanned.
func (c *chatTUI) sessionUsage(turns []store.Turn) sessionUsageTotals {
	var total sessionUsageTotals
	if c.store == nil {
		return total
	}
	if c.footerUsage == nil || c.footerUsage.sessionID != c.sessionID {
		c.footerUsage = &footerUsageCache{sessionID: c.sessionID, turns: map[string]sessionUsageTotals{}}
	}
	for _, turn := range turns {
		usage, ok := c.footerUsage.turns[turn.ID]
		if !ok {
			usage = c.turnUsage(turn.ID)
			if strings.TrimSpace(turn.FinishedAt) != "" || turn.Status == "completed" || turn.Status == "failed" || turn.Status == "cancelled" {
				c.footerUsage.turns[turn.ID] = usage
			}
		}
		total.input += usage.input
		total.output += usage.output
		total.cacheRead += usage.cacheRead
		total.cacheWrite += usage.cacheWrite
		total.cost += usage.cost
	}
	return total
}

func (c *chatTUI) turnUsage(turnID string) sessionUsageTotals {
	events, err := c.store.ListTurnEvents(context.Background(), turnID)
	if err != nil {
		return sessionUsageTotals{}
	}
	for i := len(events) - 1; i >= 0; i-- {
		usage, _ := events[i].Payload["usage"].(map[string]any)
		if usage == nil {
			continue
		}
		return sessionUsageTotals{input: intFromAny(usage["input"]), output: intFromAny(usage["output"]), cacheRead: intFromAny(usage["cache_read"]), cacheWrite: intFromAny(usage["cache_write"]), cost: floatFromAny(usage["cost_total"])}
	}
	return sessionUsageTotals{}
}

// footerCacheTTL bounds how stale the footer may be between explicit
// invalidations. Rendering happens on every keystroke and wheel event; the
// footer's store reads must not run on each of those frames.
const footerCacheTTL = time.Second

type footerCache struct {
	key  string
	at   time.Time
	rows []footerRow
}

func (c *chatTUI) invalidateFooter() { c.footerCached = nil }

func (c *chatTUI) footerRows(width int) []footerRow {
	key := fmt.Sprintf("%s|%d|%s|%s|%s|%v|%v|%d|%s|%v|%v", c.sessionID, width, c.cfg.DefaultProvider, c.cfg.DefaultModel, c.cfg.DefaultThinkingLevel, c.running, c.compaction.active, c.lastContextTokens, strings.Join(c.extensionStatusLines(), "\x00"), c.compaction.noticeUntil.After(time.Now()), c.lastCostTotal)
	if fc := c.footerCached; fc != nil && fc.key == key && time.Since(fc.at) < footerCacheTTL {
		return fc.rows
	}
	rows := c.computeFooterRows(width)
	c.footerCached = &footerCache{key: key, at: time.Now(), rows: rows}
	return rows
}

// footerData reads only what Pi's footer shows: the session's model choice
// and title, cumulative usage, and the latest context measurement.
func (c *chatTUI) footerData() tuiContextSummary {
	data := tuiContextSummary{sessionTitle: c.sessionID, model: c.cfg.DefaultModel, provider: c.cfg.DefaultProvider, thinking: c.cfg.DefaultThinkingLevel, contextWindow: c.cfg.Compaction.ContextWindow, contextTokens: c.lastContextTokens, inputTokens: c.lastInputTokens, outputTokens: c.lastOutputTokens, cacheRead: c.lastCacheRead, cacheWrite: c.lastCacheWrite, costTotal: c.lastCostTotal}
	c.applyModelContextWindow(&data)
	if c.store == nil || c.sessionID == "" {
		return data
	}
	if session, err := c.store.GetSession(context.Background(), c.sessionID); err == nil {
		data.sessionTitle = session.Title
		c.programSessionName = session.Title
		choice := inference.SessionModel(session.State, inference.SessionModelChoice{Model: data.model, Provider: data.provider, Thinking: data.thinking})
		data.model, data.provider, data.thinking = choice.Model, choice.Provider, choice.Thinking
		c.applyModelContextWindow(&data)
	}
	return data
}

func (c *chatTUI) computeFooterRows(width int) []footerRow {
	data := c.footerData()
	usage := sessionUsageTotals{input: data.inputTokens, output: data.outputTokens, cacheRead: data.cacheRead, cacheWrite: data.cacheWrite, cost: data.costTotal}
	var hitRate *float64
	sessionName := ""
	if c.store != nil && c.sessionID != "" {
		turns, _ := c.store.ListTurns(context.Background(), c.sessionID)
		usage = c.sessionUsage(turns)
		data.contextTokens = 0
		if m, err := c.store.LatestContextMeasurement(context.Background(), c.sessionID); err == nil && m != nil {
			data.contextTokens = m.Tokens
			if prompt := m.Input + m.CacheRead + m.CacheWrite; prompt > 0 {
				rate := float64(m.CacheRead) * 100 / float64(prompt)
				hitRate = &rate
			}
		}
		if title := strings.TrimSpace(data.sessionTitle); title != "" && title != c.sessionID && !strings.HasPrefix(title, "@") {
			sessionName = title
		}
	}

	workspace := strings.TrimSpace(c.cfg.WorkspaceRoot)
	if workspace == "" {
		workspace = "."
	}
	pwd := piFormatCwd(workspace)
	if branch := c.gitBranchName(workspace); branch != "" {
		pwd += " (" + branch + ")"
	}
	if sessionName != "" {
		pwd += " • " + sessionName
	}
	rows := []footerRow{{text: truncateWithEllipsis(pwd, width, "...")}}

	parts := []string{}
	if usage.input > 0 {
		parts = append(parts, "↑"+piFormatTokens(usage.input))
	}
	if usage.output > 0 {
		parts = append(parts, "↓"+piFormatTokens(usage.output))
	}
	if usage.cacheRead > 0 {
		parts = append(parts, "R"+piFormatTokens(usage.cacheRead))
	}
	if usage.cacheWrite > 0 {
		parts = append(parts, "W"+piFormatTokens(usage.cacheWrite))
	}
	if (usage.cacheRead > 0 || usage.cacheWrite > 0) && hitRate != nil {
		parts = append(parts, fmt.Sprintf("CH%.1f%%", *hitRate))
	}
	subscription := false
	providers := 0
	for _, s := range c.footerAuthStatuses() {
		if s.Authenticated {
			providers++
			if s.ID == data.provider && s.Kind == "oauth" {
				subscription = true
			}
		}
	}
	if usage.cost > 0 || subscription {
		cost := fmt.Sprintf("$%.3f", usage.cost)
		if subscription {
			cost += " (sub)"
		}
		parts = append(parts, cost)
	}
	auto := ""
	if c.cfg.Compaction.Enabled {
		auto = " (auto)"
	}
	percent := 0.0
	if data.contextWindow > 0 {
		percent = float64(data.contextTokens) * 100 / float64(data.contextWindow)
	}
	contextText := fmt.Sprintf("%.1f%%/%s%s", percent, piFormatTokens(data.contextWindow), auto)
	parts = append(parts, contextText)
	statsLeft := strings.Join(parts, " ")
	accentStart, accentEnd := len(statsLeft)-len(contextText), len(statsLeft)
	var accent gotui.Color
	switch {
	case percent > 90:
		accent = piError
	case percent > 70:
		accent = piWarning
	default:
		accentEnd = accentStart
	}
	statsLeftWidth := gotui.StringWidth(statsLeft)
	if width > 0 && statsLeftWidth > width {
		statsLeft = truncateWithEllipsis(statsLeft, width, "...")
		statsLeftWidth = gotui.StringWidth(statsLeft)
		accentStart, accentEnd = min(accentStart, len(statsLeft)), min(accentEnd, len(statsLeft))
	}

	model := strings.TrimSpace(data.model)
	if provider := strings.TrimSpace(data.provider); provider != "" {
		model = strings.TrimPrefix(model, provider+"/")
	}
	if model == "" {
		model = "no-model"
	}
	right := model
	switch thinking := c.effectiveThinking(data.provider, data.model, data.thinking); thinking {
	case "":
	case "off":
		right += " • thinking off"
	default:
		right += " • " + thinking
	}
	const minPadding = 2
	if providers > 1 && strings.TrimSpace(data.provider) != "" {
		if withProvider := "(" + data.provider + ") " + right; statsLeftWidth+minPadding+gotui.StringWidth(withProvider) <= width {
			right = withProvider
		}
	}
	line := statsLeft
	rightWidth := gotui.StringWidth(right)
	if statsLeftWidth+minPadding+rightWidth <= width {
		line += strings.Repeat(" ", width-statsLeftWidth-rightWidth) + right
	} else if avail := width - statsLeftWidth - minPadding; avail > 0 {
		r := truncateCells(right, avail)
		line += strings.Repeat(" ", max(0, width-statsLeftWidth-gotui.StringWidth(r))) + r
	}
	rows = append(rows, footerRow{text: line, accentStart: accentStart, accentEnd: accentEnd, accent: accent})

	if statuses := c.extensionStatusLines(); len(statuses) > 0 {
		rows = append(rows, footerRow{text: truncateWithEllipsis(strings.Join(statuses, " "), width, "...")})
	}
	// Gi-only: transient compaction/selection feedback (Pi reports these as
	// chat status rows); shown for four seconds under the footer.
	if time.Now().Before(c.compaction.noticeUntil) {
		rows = append(rows, footerRow{text: truncateWithEllipsis(sanitizeStatusText(c.compaction.notice), width, "...")})
	}
	return rows
}

func truncateWithEllipsis(s string, width int, ellipsis string) string {
	if width <= 0 || gotui.StringWidth(s) <= width {
		return s
	}
	return truncateCells(s, max(0, width-gotui.StringWidth(ellipsis))) + ellipsis
}

func (c *chatTUI) footerLines(width int) []string {
	rows := c.footerRows(width)
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.text
	}
	return out
}

func (c *chatTUI) renderFooter(width int) *gotui.Element {
	rows := c.footerRows(width)
	block := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100), gotui.WithHeight(len(rows)))
	dim := piFg(piDim)
	for _, r := range rows {
		spans := []gotui.TextSpan{{Text: r.text, Style: dim}}
		if r.accentEnd > r.accentStart {
			spans = []gotui.TextSpan{{Text: r.text[:r.accentStart], Style: dim}, {Text: r.text[r.accentStart:r.accentEnd], Style: piFg(r.accent)}}
			if rest := r.text[r.accentEnd:]; rest != "" {
				spans = append(spans, gotui.TextSpan{Text: rest, Style: dim})
			}
		}
		block.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(1), gotui.WithWrap(false), gotui.WithRichText(spans...)))
	}
	return block
}

package tui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	gotui "github.com/grindlemire/go-tui"
)

// Tool output rendering, ported from Pi's ToolExecutionComponent (agent tool
// calls on a status-colored background band) and BashExecutionComponent (user
// shell commands between horizontal borders).

var (
	piToolPendingBg = piRGB(52, 56, 58)
	piToolSuccessBg = piRGB(37, 65, 49)
	piToolErrorBg   = piRGB(91, 40, 42)
)

const piExpandKey = "ctrl+o"

func isShellTool(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "bash", "shell", "sh", "exec", "powershell":
		return true
	}
	return false
}

// toolBandColor is Pi's tool background: pending while running, success or
// error once the result arrives.
func toolBandColor(status string) gotui.Color {
	switch status {
	case "ok":
		return piToolSuccessBg
	case "error", "failed":
		return piToolErrorBg
	}
	return piToolPendingBg
}

// piFormatDuration matches Pi's bash renderer formatDuration.
func piFormatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	total := int(d.Seconds())
	minutes, seconds := total/60, total%60
	if minutes < 60 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%dh %dm %ds", minutes/60, minutes%60, seconds)
}

func blockDuration(startedAt, endedAt string) (time.Duration, bool) {
	start, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(startedAt))
	if err != nil {
		return 0, false
	}
	end := time.Now()
	if strings.TrimSpace(endedAt) != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(endedAt)); err == nil {
			end = parsed
		}
	}
	return end.Sub(start), true
}

func textRow(spans ...gotui.TextSpan) *gotui.Element {
	return gotui.New(gotui.WithWidthPercent(100), gotui.WithRichText(spans...))
}

func blankRow() *gotui.Element {
	return gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(1))
}

// expandHint renders Pi's keyHint: key in dim, text in muted.
func expandHint(prefix, action string) []gotui.TextSpan {
	return []gotui.TextSpan{
		{Text: prefix, Style: piFg(piMuted)},
		{Text: piExpandKey, Style: piFg(piDim)},
		{Text: " " + action + ")", Style: piFg(piMuted)},
	}
}

func (c *chatTUI) registerBlockTarget(key string, el *gotui.Element) {
	ref := gotui.NewRef()
	ref.Set(el)
	c.transcriptBlockRefs = append(c.transcriptBlockRefs, transcriptBlockHitTarget{Key: key, Ref: ref})
}

// toolCallSpans is Pi's renderCall: "$ command" for shell tools, otherwise the
// tool name in bold toolTitle followed by its argument in accent.
func toolCallSpans(block transcriptRenderableBlock) []gotui.TextSpan {
	title := piFg(piToolTitle).Bold()
	name := strings.TrimSpace(block.Header)
	arg := strings.TrimSpace(block.ToolArg)
	if isShellTool(name) {
		if arg == "" {
			return []gotui.TextSpan{{Text: "$ ", Style: title}, {Text: "...", Style: piFg(piMuted)}}
		}
		return []gotui.TextSpan{{Text: "$ " + arg, Style: title}}
	}
	spans := []gotui.TextSpan{{Text: name, Style: title}}
	if arg != "" {
		spans = append(spans, gotui.TextSpan{Text: " " + arg, Style: piFg(piAccent)})
	}
	return spans
}

func toolOutputStyle(line string) gotui.Style {
	if s, ok := diffLineStyle(line); ok {
		return s
	}
	if strings.HasPrefix(line, "error=") {
		return piFg(piError)
	}
	if strings.HasPrefix(line, "reason=") {
		return piFg(piWarning)
	}
	return piFg(piToolOutput)
}

func (c *chatTUI) renderPiToolBlock(block transcriptRenderableBlock) *gotui.Element {
	container := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100))
	c.registerBlockTarget(block.Key, container)
	isFileTool := block.Header == "read" || block.Header == "write" || block.Header == "edit"
	if isFileTool {
		container.AddChild(textRow(c.fileToolCallSpans(block)...))
	} else {
		container.AddChild(textRow(toolCallSpans(block)...))
	}
	if block.Header == "edit" {
		c.renderEditBlock(block, container)
		return container
	}
	if block.Header == "codemode" {
		c.renderCodemodeBlock(block, container)
		return container
	}
	body := block.Body
	if len(body) == 1 && body[0] == "(empty)" {
		body = nil
	}
	shell := isShellTool(block.Header)
	limit := block.PreviewLimit
	if limit <= 0 {
		limit = toolPreviewLines
	}
	fileTool := block.Header == "read" && block.ToolPath != "" || block.Header == "write" && block.ToolContent != nil
	if fileTool && c.extensionToolModes[block.Header] != "hidden" {
		if block.Header == "read" {
			if block.Expanded || block.Status != "ok" {
				c.appendFileToolPreview(container, block, strings.Join(body, "\n"), block.Status == "ok")
				if block.ToolNotice != "" {
					container.AddChild(textRow(gotui.TextSpan{Text: block.ToolNotice, Style: piFg(piWarning)}))
				}
			}
		} else {
			if block.ToolContent != nil {
				c.appendFileToolPreview(container, block, *block.ToolContent, true)
			}
			if block.Status == "error" || block.Status == "failed" || block.Status == "skipped" {
				c.appendErrorText(container, strings.Join(body, "\n"))
			}
		}
	}
	if !fileTool && len(body) > 0 {
		container.AddChild(blankRow())
		visible, hidden := body, 0
		if !block.Expanded && len(body) > limit {
			hidden = len(body) - limit
			if block.PreviewTail {
				visible = body[hidden:]
			} else {
				visible = body[:limit]
			}
		}
		if hidden > 0 && block.PreviewTail {
			container.AddChild(textRow(expandHint(fmt.Sprintf("... (%d earlier lines, ", hidden), "to expand")...))
		}
		for _, line := range visible {
			container.AddChild(c.renderInlineStyledLine(line, toolOutputStyle(line)))
		}
		if hidden > 0 && !block.PreviewTail {
			container.AddChild(textRow(expandHint(fmt.Sprintf("... (%d more lines, ", hidden), "to expand")...))
		}
	}
	if footer := strings.TrimSpace(block.Footer); footer != "" {
		container.AddChild(blankRow())
		container.AddChild(textRow(gotui.TextSpan{Text: "[" + footer + "]", Style: piFg(piWarning)}))
	}
	if shell {
		var d time.Duration
		label := "Took"
		ok := block.DurationMS != nil
		if ok {
			d = time.Duration(*block.DurationMS) * time.Millisecond
		} else if block.Status == "running" && strings.TrimSpace(block.EndedAt) == "" {
			d, ok = blockDuration(block.StartedAt, "")
			label = "Elapsed"
		}
		if ok {
			container.AddChild(blankRow())
			container.AddChild(textRow(gotui.TextSpan{Text: label + " " + piFormatDuration(d), Style: piFg(piMuted)}))
		}
	}
	return container
}

var exitStatusPattern = regexp.MustCompile(`exit status (\d+)`)

// renderPiBashBlock ports BashExecutionComponent for local `!!` commands.
// Gi's local commands are not sent to the model, as Pi's `!!`, so the borders
// use dim while the command keeps bashMode.
func (c *chatTUI) renderPiBashBlock(block transcriptRenderableBlock) *gotui.Element {
	container := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100))
	c.registerBlockTarget(block.Key, container)
	width := max(1, c.currentContentWidth())
	border := func() *gotui.Element {
		return gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(1), gotui.WithWrap(false), gotui.WithText(strings.Repeat("─", width)), gotui.WithTextStyle(piFg(piDim)))
	}
	padded := func(el *gotui.Element) *gotui.Element {
		row := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100), gotui.WithPaddingTRBL(0, 1, 0, 1))
		row.AddChild(el)
		return row
	}
	container.AddChild(border())
	container.AddChild(padded(textRow(gotui.TextSpan{Text: block.Header, Style: piFg(piBashMode).Bold()})))
	var output []string
	exitCode := ""
	for _, line := range block.Body {
		if strings.HasPrefix(line, "error: ") {
			if m := exitStatusPattern.FindStringSubmatch(line); m != nil {
				exitCode = m[1]
				continue
			}
		}
		if line == "(no output)" {
			continue
		}
		output = append(output, line)
	}
	limit := block.PreviewLimit
	if limit <= 0 {
		limit = bashPreviewLines
	}
	hidden := 0
	visible := output
	if !block.Expanded && len(output) > limit {
		hidden = len(output) - limit
		visible = output[hidden:]
	}
	if len(visible) > 0 {
		container.AddChild(blankRow())
		for _, line := range visible {
			container.AddChild(padded(c.renderInlineStyledLine(line, piFg(piMuted))))
		}
	}
	var status []*gotui.Element
	if len(output) > limit && !block.Static {
		if block.Expanded {
			status = append(status, textRow(gotui.TextSpan{Text: "(", Style: piFg(piMuted)}, gotui.TextSpan{Text: piExpandKey, Style: piFg(piDim)}, gotui.TextSpan{Text: " to collapse)", Style: piFg(piMuted)}))
		} else {
			status = append(status, textRow(expandHint(fmt.Sprintf("... %d more lines (", hidden), "to expand")...))
		}
	}
	if exitCode != "" {
		status = append(status, textRow(gotui.TextSpan{Text: "(exit " + exitCode + ")", Style: piFg(piError)}))
	}
	if footer := strings.TrimSpace(block.Footer); footer != "" {
		footer = strings.Replace(footer, "output truncated · full output: ", "Output truncated. Full output: ", 1)
		status = append(status, textRow(gotui.TextSpan{Text: footer, Style: piFg(piWarning)}))
	}
	if len(status) > 0 {
		container.AddChild(blankRow())
		for _, row := range status {
			container.AddChild(padded(row))
		}
	}
	container.AddChild(border())
	return container
}

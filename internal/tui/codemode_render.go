package tui

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	gotui "github.com/grindlemire/go-tui"
)

// Codemode presentation, ported from Pi's codemode renderer
// (pi-coding-agent extensions/codemode/renderer.js). The call shows the
// script (JavaScript-highlighted, 10-line preview); the result lists the
// nested tool calls with their status as they run, followed by the script
// output without its "Script completed" header (5-line preview). Nested
// calls are rows of the codemode block, never separate tool blocks.

const (
	codemodeCodePreviewLines   = 10
	codemodeCallPreviewCount   = 8
	codemodeOutputPreviewLines = 5
	codemodeCollapsedArgsChars = 80
)

var codemodeScriptHeader = regexp.MustCompile(`^Script (completed|failed)\nWall time [\d.]+ seconds\nOutput:\n`)

// codemodeCallRow is one nested call (the engine's codemode call record).
type codemodeCallRow struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Args       string  `json:"args"`
	Status     string  `json:"status"`
	DurationMs float64 `json:"durationMs,omitempty"`
	Error      string  `json:"error,omitempty"`
	Cost       float64 `json:"cost,omitempty"`
	StartedAt  string  `json:"startedAt,omitempty"` // live rows: when the call started
}

// codemodeDetails reads a codemode result's details (live event structs or
// stored JSON alike).
func codemodeDetails(details any) ([]codemodeCallRow, string) {
	if details == nil {
		return nil, ""
	}
	raw, err := json.Marshal(details)
	if err != nil {
		return nil, ""
	}
	var d struct {
		Calls          []codemodeCallRow `json:"calls"`
		FullOutputPath string            `json:"fullOutputPath"`
	}
	_ = json.Unmarshal(raw, &d)
	return d.Calls, d.FullOutputPath
}

// stripScriptHeader drops "Script completed\nWall time …\nOutput:\n"; input
// rejected before running (invalid options) has no header.
func stripScriptHeader(text string) string {
	return codemodeScriptHeader.ReplaceAllString(text, "")
}

// setCodemodeArguments keeps the script for the call view.
func setCodemodeArguments(meta *transcriptBlockMeta, args any) {
	if meta.Title != "codemode" {
		return
	}
	if m, ok := args.(map[string]any); ok {
		if code, ok := m["code"].(string); ok {
			meta.ToolContent = &code
		}
	}
}

// updateCodemodeCall turns a nested call's runtime event into a row of its
// parent codemode block.
func (c *chatTUI) updateCodemodeCall(parentID, typ string, payload map[string]any, ts time.Time) {
	turnID, _ := payload["turn_id"].(string)
	key := c.transcriptToolBlocks[c.toolRuntimeBlockKey(map[string]any{"tool_call_id": parentID, "turn_id": turnID}, "codemode")]
	span, ok := c.transcriptBlockSpans[key]
	if !ok || span.HeaderIndex < 0 || span.HeaderIndex >= len(c.transcript) {
		return
	}
	meta, ok := parseTranscriptBlockMarker(c.transcript[span.HeaderIndex])
	if !ok || meta.EndedAt != "" {
		return
	}
	id, _ := payload["tool_call_id"].(string)
	name, _ := payload["tool"].(string)
	idx := -1
	for i, row := range meta.Calls {
		if row.ID == id {
			idx = i
		}
	}
	switch typ {
	case "tool_started":
		if idx >= 0 {
			return
		}
		args := ""
		if a, ok := payload["arguments"]; ok && a != nil {
			if raw, err := json.Marshal(a); err == nil {
				args = string(raw)
			}
		}
		meta.Calls = append(meta.Calls, codemodeCallRow{ID: id, Name: name, Args: args, Status: "running", StartedAt: normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)})
	case "tool_finished", "tool_failed":
		if idx < 0 {
			return
		}
		row := &meta.Calls[idx]
		if row.Status != "running" {
			return
		}
		if ms := recordedToolDuration(payload["duration_ms"]); ms != nil {
			row.DurationMs = float64(*ms)
		}
		row.Status = "ok"
		if typ == "tool_failed" {
			row.Status = "error"
			row.Error, _ = payload["error"].(string)
		}
	default:
		return
	}
	c.replaceTranscriptBlock(meta, c.readTranscriptBlockBody(key))
}

func formatCodemodeDuration(ms float64) string {
	if ms <= 0 {
		return ""
	}
	if ms < 1000 {
		return fmt.Sprintf("%dms", int(ms+0.5))
	}
	return fmt.Sprintf("%.1fs", ms/1000)
}

// formatCodemodeCost: cents for larger amounts, two significant digits for
// fractions of a cent (Pi's formatCost).
func formatCodemodeCost(cost float64) string {
	if cost >= 0.01 {
		return fmt.Sprintf("$%.2f", cost)
	}
	return "$" + strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2g", cost), "0"), ".")
}

func codemodeStatusIcon(status string) gotui.TextSpan {
	switch status {
	case "running":
		return gotui.TextSpan{Text: "…", Style: piFg(piWarning)}
	case "ok":
		return gotui.TextSpan{Text: "✓", Style: piFg(piSuccess)}
	case "error":
		return gotui.TextSpan{Text: "✗", Style: piFg(piError)}
	}
	return gotui.TextSpan{Text: "⊘", Style: piFg(piMuted)}
}

// renderCodemodeBlock is Pi's renderCall plus renderResult.
func (c *chatTUI) renderCodemodeBlock(block transcriptRenderableBlock, container *gotui.Element) {
	// Call: the script (it includes the // @options: line).
	if block.ToolContent != nil {
		code := strings.TrimRight(strings.ReplaceAll(*block.ToolContent, "\r", ""), " \t\n")
		code = strings.ReplaceAll(code, "\t", "   ")
		if code != "" {
			lines, base := fileToolSegments(code, "codemode.js", true)
			visible, hidden := lines, 0
			if !block.Expanded && len(lines) > codemodeCodePreviewLines {
				visible, hidden = lines[:codemodeCodePreviewLines], len(lines)-codemodeCodePreviewLines
			}
			for _, line := range visible {
				for _, row := range fileToolRows(line, c.transcriptBlockContentWidth("tool"), base) {
					container.AddChild(row)
				}
			}
			if hidden > 0 {
				container.AddChild(textRow(expandHint(fmt.Sprintf("... (%d more lines, ", hidden), "to expand")...))
			}
		}
	}
	// Result: nested calls, newest last.
	if calls := block.Calls; len(calls) > 0 {
		shown := calls
		if !block.Expanded && len(calls) > codemodeCallPreviewCount {
			shown = calls[len(calls)-codemodeCallPreviewCount:]
		}
		container.AddChild(blankRow())
		if len(shown) < len(calls) {
			container.AddChild(textRow(expandHint(fmt.Sprintf("... (%d earlier calls, ", len(calls)-len(shown)), "to expand")...))
		}
		total, priced := 0.0, 0
		for _, call := range calls {
			if call.Cost > 0 {
				total += call.Cost
				priced++
			}
		}
		for _, call := range shown {
			args := call.Args
			if !block.Expanded && len([]rune(args)) > codemodeCollapsedArgsChars {
				args = string([]rune(args)[:codemodeCollapsedArgsChars-3]) + "..."
			}
			spans := []gotui.TextSpan{codemodeStatusIcon(call.Status), {Text: " " + call.Name, Style: piFg(piText).Bold()}}
			if args != "" {
				spans = append(spans, gotui.TextSpan{Text: " " + args, Style: piFg(piMuted)})
			}
			if d := formatCodemodeDuration(call.DurationMs); d != "" {
				spans = append(spans, gotui.TextSpan{Text: " " + d, Style: piFg(piDim)})
			}
			if call.Cost > 0 {
				spans = append(spans, gotui.TextSpan{Text: " " + formatCodemodeCost(call.Cost), Style: piFg(piDim)})
			}
			container.AddChild(textRow(spans...))
			if block.Expanded && call.Error != "" {
				for _, line := range strings.Split(call.Error, "\n") {
					container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithPaddingTRBL(0, 0, 0, 4), gotui.WithRichText(gotui.TextSpan{Text: line, Style: piFg(piError)})))
				}
			}
		}
		if priced > 1 {
			container.AddChild(textRow(gotui.TextSpan{Text: "Model calls: " + formatCodemodeCost(total), Style: piFg(piMuted)}))
		}
	}
	// Output, without the "Script completed" header.
	body := block.Body
	if len(body) == 1 && body[0] == "(empty)" {
		body = nil
	}
	if len(body) == 0 || block.Status == "running" {
		return
	}
	style := toolOutputStyle("")
	if block.Status == "error" {
		style = piFg(piError)
	}
	container.AddChild(blankRow())
	visible, hidden := body, 0
	if !block.Expanded && len(body) > codemodeOutputPreviewLines {
		visible, hidden = body[:codemodeOutputPreviewLines], len(body)-codemodeOutputPreviewLines
	}
	for _, line := range visible {
		container.AddChild(c.renderInlineStyledLine(strings.ReplaceAll(line, "\t", "   "), style))
	}
	if hidden > 0 {
		container.AddChild(textRow(expandHint(fmt.Sprintf("... (%d more lines, ", hidden), "to expand")...))
	}
	if !block.Expanded && block.FullOutputPath != "" {
		container.AddChild(textRow(gotui.TextSpan{Text: "Full output: " + block.FullOutputPath, Style: piFg(piMuted)}))
	}
}

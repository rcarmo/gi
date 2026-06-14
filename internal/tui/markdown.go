package tui

import (
	"bytes"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

var tuiMarkdown = goldmark.New(goldmark.WithExtensions(extension.GFM))

const (
	markdownInlineCodeStart = "\x00gi-code-start\x00"
	markdownInlineCodeEnd   = "\x00gi-code-end\x00"
)

func looksLikeMarkdown(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	markers := []string{"# ", "## ", "### ", "- ", "* ", "1. ", "```", "|", "> ", "**", "__", "`"}
	for _, marker := range markers {
		if strings.Contains(trimmed, marker) {
			return true
		}
	}
	return strings.Contains(trimmed, "\n")
}

func renderMarkdownTranscript(prefix, markdown string, width int) []string {
	contentWidth := width - utf8.RuneCountInString(prefix)
	if contentWidth < 20 {
		contentWidth = 20
	}
	source := []byte(markdown)
	root := tuiMarkdown.Parser().Parse(text.NewReader(source))
	renderer := &markdownProjector{source: source, width: contentWidth}
	body := renderer.renderBlocks(root, 0)
	if len(body) == 0 {
		body = []string{""}
	}
	indent := strings.Repeat(" ", utf8.RuneCountInString(prefix))
	out := make([]string, 0, len(body))
	for i, line := range body {
		if i == 0 {
			out = append(out, prefix+line)
		} else {
			out = append(out, indent+line)
		}
	}
	return out
}

type markdownProjector struct {
	source []byte
	width  int
}

func (m *markdownProjector) renderBlocks(node gast.Node, depth int) []string {
	var lines []string
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *gast.Heading:
			text := strings.TrimSpace(m.renderInlineChildren(n))
			if text == "" {
				continue
			}
			lines = append(lines, strings.ToUpper(text))
			underline := strings.Repeat("=", min(max(utf8.RuneCountInString(text), 3), m.width))
			if n.Level > 1 {
				underline = strings.Repeat("-", min(max(utf8.RuneCountInString(text), 3), m.width))
			}
			lines = append(lines, underline, "")
		case *gast.Paragraph:
			text := strings.TrimSpace(m.renderInlineChildren(n))
			if text != "" {
				lines = append(lines, wrapParagraph(text, m.width)...)
				lines = append(lines, "")
			}
		case *gast.TextBlock:
			text := strings.TrimSpace(m.renderInlineChildren(n))
			if text != "" {
				lines = append(lines, wrapParagraph(text, m.width)...)
			}
		case *gast.FencedCodeBlock:
			lang := strings.TrimSpace(string(n.Language(m.source)))
			label := "[code]"
			if lang != "" {
				label = "[code:" + lang + "]"
			}
			if n.Lines().Len() > 0 {
				label += " " + strconv.Itoa(n.Lines().Len()) + " lines"
			}
			lines = append(lines, label)
			for i := 0; i < n.Lines().Len(); i++ {
				segment := n.Lines().At(i)
				codeLine := strings.TrimRight(string(segment.Value(m.source)), "\r\n")
				if codeLine == "" {
					lines = append(lines, "    ")
					continue
				}
				lines = append(lines, wrapPreformattedWithPrefix(codeLine, m.width, "    ")...)
			}
			lines = append(lines, "")
		case *gast.CodeBlock:
			label := "[code]"
			if n.Lines().Len() > 0 {
				label += " " + strconv.Itoa(n.Lines().Len()) + " lines"
			}
			lines = append(lines, label)
			for i := 0; i < n.Lines().Len(); i++ {
				segment := n.Lines().At(i)
				codeLine := strings.TrimRight(string(segment.Value(m.source)), "\r\n")
				lines = append(lines, wrapPreformattedWithPrefix(codeLine, m.width, "    ")...)
			}
			lines = append(lines, "")
		case *gast.Blockquote:
			quoted := m.renderBlocks(n, depth+1)
			for _, line := range quoted {
				if line == "" {
					lines = append(lines, "")
					continue
				}
				lines = append(lines, wrapWithPrefix(line, m.width, "> ")...)
			}
			lines = append(lines, "")
		case *gast.List:
			itemIndex := n.Start
			if itemIndex == 0 {
				itemIndex = 1
			}
			for item := n.FirstChild(); item != nil; item = item.NextSibling() {
				itemLines := m.renderBlocks(item, depth+1)
				itemLines = trimBlankEdges(itemLines)
				if len(itemLines) == 0 {
					continue
				}
				prefix := "• "
				if n.IsOrdered() {
					prefix = strconv.Itoa(itemIndex) + ". "
					itemIndex++
				}
				lines = append(lines, wrapWithPrefix(itemLines[0], m.width, prefix)...)
				contPrefix := strings.Repeat(" ", utf8.RuneCountInString(prefix))
				for _, extra := range itemLines[1:] {
					if extra == "" {
						continue
					}
					lines = append(lines, wrapWithPrefix(extra, m.width, contPrefix)...)
				}
			}
			lines = append(lines, "")
		case *extast.Table:
			lines = append(lines, m.renderTable(n)...)
			lines = append(lines, "")
		case *gast.ThematicBreak:
			lines = append(lines, strings.Repeat("-", min(max(10, m.width/2), m.width)), "")
		default:
			if child.HasChildren() {
				lines = append(lines, m.renderBlocks(child, depth+1)...)
			}
		}
	}
	return trimBlankEdges(lines)
}

func (m *markdownProjector) renderTable(table *extast.Table) []string {
	var headers []string
	var rows [][]string
	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		var cells []string
		for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
			cells = append(cells, strings.TrimSpace(m.renderInlineChildren(cell)))
		}
		if len(cells) == 0 {
			continue
		}
		if row.Kind() == extast.KindTableHeader {
			headers = cells
			continue
		}
		rows = append(rows, cells)
	}
	if len(headers) == 0 && len(rows) > 0 {
		headers = make([]string, len(rows[0]))
		for i := range headers {
			headers[i] = "Col" + strconv.Itoa(i+1)
		}
	}
	if m.tableFits(headers, rows) {
		return m.renderTableGrid(headers, rows)
	}
	return m.renderTableStacked(headers, rows)
}

func (m *markdownProjector) tableFits(headers []string, rows [][]string) bool {
	if len(headers) == 0 {
		return false
	}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && utf8.RuneCountInString(cell) > widths[i] {
				widths[i] = utf8.RuneCountInString(cell)
			}
		}
	}
	total := 1
	for _, w := range widths {
		total += w + 3
	}
	return total <= m.width
}

func (m *markdownProjector) renderTableGrid(headers []string, rows [][]string) []string {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && utf8.RuneCountInString(cell) > widths[i] {
				widths[i] = utf8.RuneCountInString(cell)
			}
		}
	}
	formatRow := func(cells []string) string {
		parts := make([]string, len(widths))
		for i := range widths {
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			pad := widths[i] - utf8.RuneCountInString(cell)
			if pad < 0 {
				pad = 0
			}
			parts[i] = " " + cell + strings.Repeat(" ", pad) + " "
		}
		return "|" + strings.Join(parts, "|") + "|"
	}
	sepParts := make([]string, len(widths))
	for i, w := range widths {
		sepParts[i] = strings.Repeat("-", w+2)
	}
	out := []string{formatRow(headers), "+" + strings.Join(sepParts, "+") + "+"}
	for _, row := range rows {
		out = append(out, formatRow(row))
	}
	return out
}

func (m *markdownProjector) renderTableStacked(headers []string, rows [][]string) []string {
	var out []string
	for idx, row := range rows {
		if idx > 0 {
			out = append(out, strings.Repeat("-", min(max(8, m.width/3), m.width)))
		}
		for i, cell := range row {
			header := "Col" + strconv.Itoa(i+1)
			if i < len(headers) && headers[i] != "" {
				header = headers[i]
			}
			out = append(out, wrapWithPrefix(cell, m.width, header+": ")...)
		}
	}
	if len(out) == 0 && len(headers) > 0 {
		out = append(out, wrapParagraph(strings.Join(headers, " | "), m.width)...)
	}
	return out
}

func (m *markdownProjector) renderInlineChildren(node gast.Node) string {
	var b strings.Builder
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		b.WriteString(m.renderInline(child))
	}
	return b.String()
}

func (m *markdownProjector) renderInline(node gast.Node) string {
	switch n := node.(type) {
	case *gast.Text:
		text := string(n.Text(m.source))
		if n.HardLineBreak() {
			return text + "\n"
		}
		if n.SoftLineBreak() {
			return text + " "
		}
		return text
	case *gast.String:
		return string(n.Value)
	case *gast.CodeSpan:
		return markdownInlineCodeStart + strings.TrimSpace(m.renderInlineChildren(n)) + markdownInlineCodeEnd
	case *gast.Emphasis:
		return m.renderInlineChildren(n)
	case *extast.Strikethrough:
		return m.renderInlineChildren(n)
	case *extast.TaskCheckBox:
		if n.IsChecked {
			return "☑ "
		}
		return "☐ "
	case *gast.Link:
		label := strings.TrimSpace(m.renderInlineChildren(n))
		dest := string(n.Destination)
		if label == "" || label == dest {
			return dest
		}
		return label + " (" + dest + ")"
	case *gast.AutoLink:
		return string(n.URL(m.source))
	default:
		return m.renderInlineChildren(node)
	}
}

func wrapParagraph(text string, width int) []string {
	if width < 1 {
		width = 1
	}
	var lines []string
	for _, para := range strings.Split(text, "\n") {
		words := markdownParagraphTokens(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		current := words[0]
		currentWidth := markdownRenderedWidth(current)
		for _, word := range words[1:] {
			wordWidth := markdownRenderedWidth(word)
			if currentWidth+1+wordWidth <= width {
				current += " " + word
				currentWidth += 1 + wordWidth
				continue
			}
			lines = append(lines, current)
			if wordWidth > width && !strings.Contains(word, markdownInlineCodeStart) {
				parts := wrapLongRunes(word, width)
				lines = append(lines, parts[:len(parts)-1]...)
				current = parts[len(parts)-1]
				currentWidth = markdownRenderedWidth(current)
			} else {
				current = word
				currentWidth = wordWidth
			}
		}
		lines = append(lines, current)
	}
	return lines
}

func markdownParagraphTokens(text string) []string {
	var tokens []string
	for i := 0; i < len(text); {
		for i < len(text) {
			r, size := utf8.DecodeRuneInString(text[i:])
			if !isMarkdownTokenSpace(r) {
				break
			}
			i += size
		}
		if i >= len(text) {
			break
		}
		if strings.HasPrefix(text[i:], markdownInlineCodeStart) {
			end := strings.Index(text[i+len(markdownInlineCodeStart):], markdownInlineCodeEnd)
			if end >= 0 {
				endPos := i + len(markdownInlineCodeStart) + end + len(markdownInlineCodeEnd)
				tokens = append(tokens, text[i:endPos])
				i = endPos
				continue
			}
		}
		start := i
		for i < len(text) {
			if strings.HasPrefix(text[i:], markdownInlineCodeStart) && i > start {
				break
			}
			r, size := utf8.DecodeRuneInString(text[i:])
			if isMarkdownTokenSpace(r) {
				break
			}
			i += size
		}
		if i > start {
			tokens = append(tokens, text[start:i])
		}
	}
	return tokens
}

func isMarkdownTokenSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func markdownRenderedWidth(s string) int {
	return utf8.RuneCountInString(stripMarkdownInlineStyleMarkers(s))
}

func stripMarkdownInlineStyleMarkers(s string) string {
	s = strings.ReplaceAll(s, markdownInlineCodeStart, "")
	s = strings.ReplaceAll(s, markdownInlineCodeEnd, "")
	return s
}

func wrapWithPrefix(text string, width int, prefix string) []string {
	prefixWidth := utf8.RuneCountInString(prefix)
	contentWidth := width - prefixWidth
	if contentWidth < 8 {
		contentWidth = 8
	}
	wrapped := wrapParagraph(text, contentWidth)
	out := make([]string, 0, len(wrapped))
	for i, line := range wrapped {
		if i == 0 {
			out = append(out, prefix+line)
		} else {
			out = append(out, strings.Repeat(" ", prefixWidth)+line)
		}
	}
	return out
}

func wrapPreformattedWithPrefix(text string, width int, prefix string) []string {
	prefixWidth := utf8.RuneCountInString(prefix)
	contentWidth := width - prefixWidth
	if contentWidth < 8 {
		contentWidth = 8
	}
	runes := []rune(text)
	if len(runes) == 0 {
		return []string{prefix}
	}
	out := make([]string, 0, (len(runes)/contentWidth)+1)
	for len(runes) > contentWidth {
		out = append(out, prefix+string(runes[:contentWidth]))
		runes = runes[contentWidth:]
		prefix = strings.Repeat(" ", prefixWidth)
	}
	out = append(out, prefix+string(runes))
	return out
}

func wrapLongRunes(word string, width int) []string {
	runes := []rune(word)
	if len(runes) == 0 {
		return []string{""}
	}
	var out []string
	for len(runes) > width {
		out = append(out, string(runes[:width]))
		runes = runes[width:]
	}
	out = append(out, string(runes))
	return out
}

func trimBlankEdges(lines []string) []string {
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	if start >= end {
		return nil
	}
	trimmed := append([]string(nil), lines[start:end]...)
	// collapse repeated blank lines inside the slice
	var out []string
	blank := false
	for _, line := range trimmed {
		isBlank := strings.TrimSpace(line) == ""
		if isBlank && blank {
			continue
		}
		out = append(out, line)
		blank = isBlank
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func markdownToPlain(text string) string {
	var buf bytes.Buffer
	if err := tuiMarkdown.Convert([]byte(text), &buf); err != nil {
		return text
	}
	return buf.String()
}

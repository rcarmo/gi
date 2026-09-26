package tui

import (
	"strings"
	"unicode"

	gotui "github.com/grindlemire/go-tui"
)

// A soft-wrap run is admitted only when the renderer's visible cells reproduce
// one source paragraph. Never infer continuation from a full-width row alone.
type transcriptSearchCell struct{ row, start, end int }
type transcriptSearchRun struct {
	text  string
	cells []transcriptSearchCell
}

func transcriptWrapRuns(root *gotui.Element, rows []transcriptSearchRow) []transcriptSearchRun {
	var runs []transcriptSearchRun
	var visit func(*gotui.Element)
	visit = func(el *gotui.Element) {
		if len(el.Children()) > 0 {
			for _, child := range el.Children() {
				visit(child)
			}
			return
		}
		if !el.Wrap() || el.TextAlign() != gotui.TextAlignLeft {
			return
		}
		rect := el.ContentRect()
		if rect.Height < 2 || rect.Width < 1 {
			return
		}
		source := el.Text()
		if spans := el.RichText(); len(spans) > 0 {
			var b strings.Builder
			for _, s := range spans {
				b.WriteString(s.Text)
			}
			source = b.String()
		}
		// Joining explicit newlines is not soft-wrap search. Multi-paragraph leaves
		// retain row-local matching until source break provenance is represented.
		if strings.Contains(source, "\n") {
			return
		}
		expected := strings.Join(strings.Fields(source), " ")
		if expected == "" {
			return
		}
		var run transcriptSearchRun
		var textBuilder strings.Builder
		offset, firstRow, lastRow := 0, -1, -1
		for row := rect.Y; row < rect.Y+rect.Height && row < len(rows); row++ {
			if row < 0 {
				return
			}
			col := 0
			for _, span := range rows[row].spans {
				width := gotui.StringWidth(span.Text)
				start, end := col, col+width
				col = end
				if start < rect.X || end > rect.X+rect.Width {
					continue
				}
				text := span.Text
				// Buffer padding must never turn into a searchable space. Interior spaces
				// are recovered from the paragraph only when followed by visible content.
				if strings.TrimFunc(text, unicode.IsSpace) == "" {
					continue
				}
				if offset < len(expected) && expected[offset] == ' ' {
					textBuilder.WriteByte(' ')
					spaceStart := start
					if len(run.cells) > 0 && run.cells[len(run.cells)-1].row == row {
						spaceStart = run.cells[len(run.cells)-1].end
					}
					run.cells = append(run.cells, transcriptSearchCell{row: row, start: spaceStart, end: start})
					offset++
				}
				if !strings.HasPrefix(expected[offset:], text) {
					return
				}
				if firstRow < 0 {
					firstRow = row
				}
				lastRow = row
				textBuilder.WriteString(text)
				for range len(text) {
					run.cells = append(run.cells, transcriptSearchCell{row: row, start: start, end: end})
				}
				offset += len(text)
			}
		}
		if offset == len(expected) && lastRow > firstRow {
			run.text = textBuilder.String()
			runs = append(runs, run)
		}
	}
	visit(root)
	return runs
}

func transcriptWrappedMatches(runs []transcriptSearchRun, needle string) []transcriptSearchMatch {
	var matches []transcriptSearchMatch
	for _, run := range runs {
		// Fold rune-by-rune so a changed UTF-8 byte length still maps to the original
		// full grapheme cell, just like row-local search.
		var folded strings.Builder
		var cells []transcriptSearchCell
		for offset, r := range run.text {
			text := strings.ToLower(string(r))
			folded.WriteString(text)
			for range len(text) {
				cells = append(cells, run.cells[offset])
			}
		}
		text := folded.String()
		for offset := 0; offset < len(text) && needle != ""; {
			i := strings.Index(text[offset:], needle)
			if i < 0 {
				break
			}
			i += offset
			last := i + len(needle) - 1
			parts := []transcriptSearchCell{}
			for _, cell := range cells[i : last+1] {
				if cell.start == cell.end {
					continue
				}
				if len(parts) > 0 && parts[len(parts)-1].row == cell.row {
					p := &parts[len(parts)-1]
					p.start = min(p.start, cell.start)
					p.end = max(p.end, cell.end)
				} else {
					parts = append(parts, cell)
				}
			}
			if len(parts) > 0 {
				var continuation []transcriptSearchCell
				if len(parts) > 1 {
					continuation = parts[1:]
				}
				matches = append(matches, transcriptSearchMatch{row: parts[0].row, start: parts[0].start, end: parts[0].end, continuation: continuation})
			}
			offset = last + 1
		}
	}
	return matches
}

func transcriptMergeWrappedMatches(local []transcriptSearchMatch, runs []transcriptSearchRun, needle string) []transcriptSearchMatch {
	// Only suppress row-local occurrences wholly inside a validated visible run.
	type bounds struct{ start, end int }
	covered := map[int][]bounds{}
	for _, run := range runs {
		byRow := map[int]bounds{}
		for _, cell := range run.cells {
			if cell.start == cell.end {
				continue
			}
			b, ok := byRow[cell.row]
			if !ok {
				b = bounds{cell.start, cell.end}
			} else {
				b.start = min(b.start, cell.start)
				b.end = max(b.end, cell.end)
			}
			byRow[cell.row] = b
		}
		for row, b := range byRow {
			covered[row] = append(covered[row], b)
		}
	}
	var result []transcriptSearchMatch
	for _, match := range local {
		owned := false
		for _, b := range covered[match.row] {
			if match.start >= b.start && match.end <= b.end {
				owned = true
				break
			}
		}
		if !owned {
			result = append(result, match)
		}
	}
	return append(result, transcriptWrappedMatches(runs, needle)...)
}

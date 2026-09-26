package tui

import (
	"strings"
	"time"
	"unicode"

	"github.com/clipperhouse/uax29/v2/words"
	gotui "github.com/grindlemire/go-tui"
)

type transcriptSelectionRange struct{ start, end transcriptPoint }
type transcriptClickSequence struct {
	at    time.Time
	count int
	word  transcriptSelectionRange
	valid bool
}

const transcriptMultiClickInterval = 500 * time.Millisecond

// Match pi-tui's path/kebab joiners on Unicode word boundaries. Map byte
// boundaries back to full rendered grapheme cells, never slice terminal text.
func transcriptWordRange(row int, value transcriptSearchRow, column int) (transcriptSelectionRange, bool) {
	var text strings.Builder
	var cells []transcriptSearchCell
	col := 0
	for _, span := range value.spans {
		width := gotui.StringWidth(span.Text)
		text.WriteString(span.Text)
		for range len(span.Text) {
			cells = append(cells, transcriptSearchCell{row: row, start: col, end: col + width})
		}
		col += width
	}
	source := text.String()
	if source == "" || column < 0 || column >= col {
		return transcriptSelectionRange{}, false
	}
	type segment struct {
		start, end   int
		word, joiner bool
	}
	var parts []segment
	it := words.FromString(source)
	for it.Next() {
		v := it.Value()
		joiner := v == "/" || v == "-"
		word := strings.ContainsFunc(v, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r) || r == '_' })
		parts = append(parts, segment{start: cells[it.Start()].start, end: cells[it.End()-1].end, word: word || joiner, joiner: joiner})
	}
	index := -1
	for i, p := range parts {
		if column >= p.start && column < p.end {
			index = i
			break
		}
	}
	if index < 0 {
		return transcriptSelectionRange{}, false
	}
	start, end := index, index
	join := func(a, b segment) bool { return a.word && b.word && (a.joiner || b.joiner) }
	for start > 0 && join(parts[start-1], parts[start]) {
		start--
	}
	for end+1 < len(parts) && join(parts[end], parts[end+1]) {
		end++
	}
	return transcriptSelectionRange{transcriptPoint{row, parts[start].start}, transcriptPoint{row, parts[end].end}}, true
}
func transcriptLineRange(row int, value transcriptSearchRow) transcriptSelectionRange {
	return transcriptSelectionRange{transcriptPoint{row, 0}, transcriptPoint{row, gotui.StringWidth(value.text)}}
}
func (s *transcriptClickSequence) next(word transcriptSelectionRange, now time.Time) int {
	count := 1
	if s.valid && s.word == word && !now.Before(s.at) && now.Sub(s.at) <= transcriptMultiClickInterval {
		count = s.count%3 + 1
	}
	*s = transcriptClickSequence{at: now, count: count, word: word, valid: true}
	return count
}
func (s *transcriptSelection) extend(point transcriptPoint) {
	if s.granularity == 0 {
		s.end = point
		return
	}
	if point.row < 0 || point.row >= len(s.rows) {
		return
	}
	var target transcriptSelectionRange
	if s.granularity == 2 {
		target = transcriptLineRange(point.row, s.rows[point.row])
	} else {
		var ok bool
		target, ok = transcriptWordRange(point.row, s.rows[point.row], min(point.col, max(0, gotui.StringWidth(s.rows[point.row].text)-1)))
		if !ok {
			return
		}
	}
	before := target.start.row < s.initial.start.row || target.start.row == s.initial.start.row && target.start.col < s.initial.start.col
	if before {
		s.start = s.initial.end
		s.end = target.start
	} else {
		s.start = s.initial.start
		s.end = target.end
	}
}

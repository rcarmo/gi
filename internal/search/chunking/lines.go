package chunking

import (
	"bytes"
	"fmt"
	"unicode/utf8"
)

// LineVersion changes whenever source splitting or location semantics change.
const LineVersion = "utf8-lines-8k-128-v1"
const LineChunkBytes = 8192
const LineChunkLines = 128

// Lines emits contiguous non-overlapping source slices, preserving CRLF and
// final newlines. A long physical line splits only at a UTF-8 boundary. Line
// numbers count '\n' before the boundary, matching the refresh-store contract.
type Lines struct{}

func (Lines) Version() string { return LineVersion }
func (Lines) Chunk(_ string, data []byte) ([]Chunk, error) {
	if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return nil, fmt.Errorf("chunk input must be UTF-8 text")
	}
	if len(data) == 0 {
		return []Chunk{{StartLine: 1, EndLine: 1}}, nil
	}
	var chunks []Chunk
	start, line := 0, 1
	for start < len(data) {
		capEnd := min(start+LineChunkBytes, len(data))
		end := capEnd
		for end < len(data) && !utf8.RuneStart(data[end]) {
			end--
		}
		lastNewline, newlines := 0, 0
		for pos := start; pos < end; pos++ {
			if data[pos] == '\n' {
				lastNewline = pos + 1
				newlines++
				if newlines == LineChunkLines {
					end = pos + 1
					break
				}
			}
		}
		// Prefer a full line except at EOF, where a final unterminated line fits.
		if end < len(data) && lastNewline > start {
			end = lastNewline
		}
		text := string(data[start:end])
		endLine := line + bytes.Count(data[start:end], []byte{'\n'})
		chunks = append(chunks, Chunk{ChunkIndex: len(chunks), StartByte: start, EndByte: end, StartLine: line, EndLine: endLine, Content: text})
		start, line = end, endLine
	}
	return chunks, nil
}

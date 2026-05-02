package chunking

import (
	"bytes"
	"strings"
)

type TextChunker struct {
	MaxBytes int
}

func (c TextChunker) Version() string { return "text-v1" }

func (c TextChunker) Chunk(path string, data []byte) ([]Chunk, error) {
	max := c.MaxBytes
	if max <= 0 {
		max = 2048
	}
	text := string(data)
	parts := splitParagraphs(text)
	var out []Chunk
	var buf strings.Builder
	start := 0
	chunkIndex := 0
	flush := func(end int) {
		content := strings.TrimSpace(buf.String())
		if content == "" {
			buf.Reset()
			start = end
			return
		}
		out = append(out, Chunk{ChunkIndex: chunkIndex, StartByte: start, EndByte: end, TokenEstimate: len(strings.Fields(content)), Content: content})
		chunkIndex++
		buf.Reset()
		start = end
	}
	cursor := 0
	for _, p := range parts {
		if buf.Len() > 0 && buf.Len()+len(p)+2 > max {
			flush(cursor)
		}
		if buf.Len() > 0 {
			buf.WriteString("\n\n")
			cursor += 2
		}
		buf.WriteString(p)
		cursor += len(p)
	}
	flush(len(text))
	return out, nil
}

func splitParagraphs(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	parts := bytes.Split([]byte(text), []byte("\n\n"))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(string(p))
		if s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 && strings.TrimSpace(text) != "" {
		out = append(out, strings.TrimSpace(text))
	}
	return out
}

package chunking

import "strings"

type MarkdownChunker struct {
	TextChunker
}

func (c MarkdownChunker) Version() string { return "markdown-v1" }

func (c MarkdownChunker) Chunk(path string, data []byte) ([]Chunk, error) {
	chunks, err := c.TextChunker.Chunk(path, data)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	for i := range chunks {
		for _, line := range lines {
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "#") {
				chunks[i].Heading = strings.TrimSpace(strings.TrimLeft(trim, "#"))
				break
			}
		}
	}
	return chunks, nil
}

package chunking

import (
	"strings"
)

type CodeChunker struct {
	MaxLines int
}

func (c CodeChunker) Version() string { return "code-v1" }

func (c CodeChunker) Chunk(path string, data []byte) ([]Chunk, error) {
	max := c.MaxLines
	if max <= 0 {
		max = 120
	}
	lines := strings.Split(string(data), "\n")
	var out []Chunk
	for i, chunkIndex := 0, 0; i < len(lines); i += max {
		end := i + max
		if end > len(lines) {
			end = len(lines)
		}
		content := strings.TrimSpace(strings.Join(lines[i:end], "\n"))
		if content == "" {
			continue
		}
		heading := ""
		for _, line := range lines[i:end] {
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "func ") || strings.HasPrefix(trim, "type ") || strings.HasPrefix(trim, "class ") {
				heading = trim
				break
			}
		}
		out = append(out, Chunk{
			ChunkIndex:    chunkIndex,
			StartLine:     i + 1,
			EndLine:       end,
			TokenEstimate: len(strings.Fields(content)),
			Heading:       heading,
			Content:       content,
		})
		chunkIndex++
	}
	return out, nil
}

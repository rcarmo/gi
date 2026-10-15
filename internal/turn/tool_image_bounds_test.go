package turn

import (
	"strings"
	"testing"
)

func TestToolImageReplayAttachmentBounds(t *testing.T) {
	for _, size := range []int{maxToolImageBytes, maxToolImageBytes + 1} {
		t.Run(map[bool]string{true: "accepted", false: "oversized"}[size == maxToolImageBytes], func(t *testing.T) {
			extras := &toolExtras{}
			extras.attachFunc()("image/png", make([]byte, size))
			payload := extras.withReplay(map[string]any{}, "result", false)
			blocks, ok := decodeReplayValue[[]struct {
				Type string `json:"type"`
				Data string `json:"data"`
				Text string `json:"text"`
			}](payload["provider_content_blocks"])
			if !ok || len(blocks) != 2 {
				t.Fatal("missing attachment projection")
			}
			if size == maxToolImageBytes {
				if blocks[1].Type != "image" || len(blocks[1].Data) != ((size+2)/3)*4 {
					t.Fatal("boundary image truncated")
				}
			} else if blocks[1].Type != "text" || blocks[1].Data != "" || !strings.Contains(blocks[1].Text, "too large") {
				t.Fatalf("oversized attachment retained: type=%s text=%s", blocks[1].Type, blocks[1].Text)
			}
		})
	}
}

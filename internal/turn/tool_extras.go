package turn

import (
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"github.com/rcarmo/gi/internal/mcp"
	goai "github.com/rcarmo/go-ai"
)

// maxToolImageBytes bounds one attached image (larger ones are described).
const maxToolImageBytes = 20 << 20

// toolExtras collects what a tool adds to its result beyond text: images
// (Pi: image blocks of the tool result; go-ai replaces them with placeholders
// for models without image input) and tools it loaded (tool_search), which
// become the result's AddedToolNames marker and are declared from the next
// model call.
type toolExtras struct {
	occurrenceID string
	blocks       []goai.ContentBlock
	notes        []string       // transcript lines describing each image
	added        []string       // tools loaded by this call
	details      map[string]any // structured result details for renderers (codemode calls)
	usage        goai.Usage     // model usage made by the tool (codemode models.*)
}

func (t *toolExtras) addUsageFunc() func(goai.Usage) {
	if t == nil {
		return nil
	}
	var mu sync.Mutex
	return func(u goai.Usage) {
		mu.Lock()
		defer mu.Unlock()
		addUsage(&t.usage, u)
	}
}

// addUsage adds u to total (tokens and cost).
func addUsage(total *goai.Usage, u goai.Usage) {
	total.Input += u.Input
	total.Output += u.Output
	total.TotalTokens += u.TotalTokens
	total.CacheRead += u.CacheRead
	total.CacheWrite += u.CacheWrite
	total.Cost.Input += u.Cost.Input
	total.Cost.Output += u.Cost.Output
	total.Cost.Total += u.Cost.Total
}

func (t *toolExtras) setDetailsFunc() func(map[string]any) {
	if t == nil {
		return nil
	}
	return func(d map[string]any) { t.details = d }
}

// withReplay preserves provider-visible result blocks separately from display text.
func (t *toolExtras) withReplay(payload map[string]any, text string, failed bool) map[string]any {
	blocks := []goai.ContentBlock{{Type: "text", Text: text}}
	if t != nil {
		blocks = append(blocks, t.blocks...)
		if !failed && len(t.added) > 0 {
			payload["added_tool_names"] = append([]string(nil), t.added...)
		}
	}
	payload["provider_content_blocks"] = blocks
	return t.withDetails(payload)
}

// withDetails adds the tool's details, if any, to an event or message payload.
func (t *toolExtras) withDetails(payload map[string]any) map[string]any {
	if t != nil && t.details != nil {
		payload["details"] = t.details
	}
	return payload
}

func (t *toolExtras) addToolsFunc() func([]string) {
	if t == nil {
		return nil
	}
	return func(names []string) { t.added = append(t.added, names...) }
}

func (t *toolExtras) attachFunc() func(string, []byte) {
	if t == nil {
		return nil
	}
	return func(mimeType string, data []byte) {
		note := fmt.Sprintf("[image %s, %s]", mimeType, mcp.FormatSize(len(data)))
		t.notes = append(t.notes, note)
		if len(data) > maxToolImageBytes || !strings.HasPrefix(mimeType, "image/") {
			t.blocks = append(t.blocks, goai.ContentBlock{Type: "text", Text: note + " (not attached: too large or not an image)"})
			return
		}
		t.blocks = append(t.blocks, goai.ContentBlock{Type: "image", Data: base64.StdEncoding.EncodeToString(data), MimeType: mimeType})
	}
}

// transcriptSuffix describes attached images in the stored transcript text.
func (t *toolExtras) transcriptSuffix() string {
	if t == nil || len(t.notes) == 0 {
		return ""
	}
	return "\n" + strings.Join(t.notes, "\n")
}

// appendToolResultWithImages appends a tool result whose content is the text
// followed by any attached images.
func appendToolResultWithImages(ctx *goai.Context, call goai.ToolCall, text string, isError bool, images *toolExtras) {
	content := []goai.ContentBlock{{Type: "text", Text: text}}
	msg := goai.Message{Role: goai.RoleToolResult, ToolCallID: call.ID, ToolName: call.Name, Content: content, IsError: isError}
	if images != nil {
		msg.Content = append(msg.Content, images.blocks...)
		if !isError && len(images.added) > 0 {
			msg.AddedToolNames = append([]string(nil), images.added...)
		}
	}
	ctx.Messages = append(ctx.Messages, msg)
}

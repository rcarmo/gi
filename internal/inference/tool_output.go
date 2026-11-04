package inference

import goai "github.com/rcarmo/go-ai"

// The pinned Anthropic encoder omits text:"" but still emits type:"text",
// which the API rejects. An empty tool result is valid; omit only its empty
// text blocks so the encoder uses its supported empty-result representation.
// Apply on every request, including replay, without changing stored history,
// tool-call IDs, error flags, images or nonempty output.
func contextWithoutEmptyToolResultText(input *goai.Context) *goai.Context {
	if input == nil {
		return nil
	}
	var out *goai.Context
	for i, message := range input.Messages {
		if message.Role != goai.RoleToolResult {
			continue
		}
		var content []goai.ContentBlock
		changed := false
		for j, block := range message.Content {
			if block.Type == "text" && block.Text == "" {
				if !changed {
					content = append([]goai.ContentBlock{}, message.Content[:j]...)
					changed = true
				}
			} else if changed {
				content = append(content, block)
			}
		}
		if !changed {
			continue
		}
		if out == nil {
			copyContext := *input
			copyContext.Messages = append([]goai.Message(nil), input.Messages...)
			out = &copyContext
		}
		out.Messages[i].Content = content
	}
	if out != nil {
		return out
	}
	return input
}

package inference

import (
	"encoding/json"
	"reflect"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

func TestEmptyToolOutputNormalisation(t *testing.T) {
	empty := goai.ContentBlock{Type: "text"}
	text := goai.ContentBlock{Type: "text", Text: "actual output"}
	space := goai.ContentBlock{Type: "text", Text: " \n"}
	image := goai.ContentBlock{Type: "image", MimeType: "image/png", Data: "aW1hZ2U="}
	for _, tc := range []struct {
		name        string
		input, want []goai.ContentBlock
	}{
		{"nil", nil, nil},
		{"empty", []goai.ContentBlock{empty}, []goai.ContentBlock{}},
		{"repeated-empty", []goai.ContentBlock{empty, empty}, []goai.ContentBlock{}},
		{"text", []goai.ContentBlock{text}, []goai.ContentBlock{text}},
		{"whitespace", []goai.ContentBlock{space}, []goai.ContentBlock{space}},
		{"mixed", []goai.ContentBlock{text, empty, image, empty, space}, []goai.ContentBlock{text, image, space}},
		{"image-only", []goai.ContentBlock{empty, image}, []goai.ContentBlock{image}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, isError := range []bool{false, true} {
				input := &goai.Context{SystemPrompt: "keep system", Messages: []goai.Message{
					{Role: goai.RoleUser, Content: []goai.ContentBlock{empty}},
					{Role: goai.RoleAssistant, Content: []goai.ContentBlock{empty}},
					{Role: goai.RoleToolResult, ToolCallID: "call", ToolName: "shell", IsError: isError, Content: tc.input},
				}}
				before, _ := json.Marshal(input)
				fixed := contextWithoutEmptyToolResultText(input)
				if !reflect.DeepEqual(fixed.Messages[2].Content, tc.want) {
					t.Fatalf("content=%+v want=%+v", fixed.Messages[2].Content, tc.want)
				}
				result := fixed.Messages[2]
				if result.ToolCallID != "call" || result.ToolName != "shell" || result.IsError != isError || fixed.SystemPrompt != input.SystemPrompt || !reflect.DeepEqual(fixed.Messages[:2], input.Messages[:2]) {
					t.Fatal("changed unrelated message fields")
				}
				after, _ := json.Marshal(input)
				if string(before) != string(after) {
					t.Fatal("mutated history")
				}
				if contextWithoutEmptyToolResultText(fixed) != fixed {
					t.Fatal("normalisation is not idempotent")
				}
				if reflect.DeepEqual(tc.input, tc.want) && fixed != input {
					t.Fatal("copied unchanged context")
				}
			}
		})
	}
	if contextWithoutEmptyToolResultText(nil) != nil {
		t.Fatal("nil context")
	}
}

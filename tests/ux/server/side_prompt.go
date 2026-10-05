package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
)

// Profile the acceptance server from initialisation through shutdown. The
// provider is local, but routing, inference, SQLite and HTTP use production code.
func startSidePromptProfiles() func() {
	dir := os.Getenv("GI_UX_SIDE_PROFILE_DIR")
	if dir == "" {
		return func() {}
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		panic(err)
	}
	cpu, err := os.Create(filepath.Join(dir, "cpu.pprof"))
	if err != nil {
		panic(err)
	}
	if err := pprof.StartCPUProfile(cpu); err != nil {
		panic(err)
	}
	return func() {
		pprof.StopCPUProfile()
		cpu.Close()
		runtime.GC()
		heap, err := os.Create(filepath.Join(dir, "mem.pprof"))
		if err != nil {
			panic(err)
		}
		defer heap.Close()
		if err := pprof.WriteHeapProfile(heap); err != nil {
			panic(err)
		}
	}
}

func serveSidePrompt(w http.ResponseWriter, r *http.Request, body map[string]any, emit func(any)) bool {
	messages, _ := body["messages"].([]any)
	side := false
	var question string
	var history strings.Builder
	for _, raw := range messages {
		message, _ := raw.(map[string]any)
		text, _ := message["content"].(string)
		if message["role"] == "system" && strings.Contains(text, "side conversation") {
			side = true
		}
		if message["role"] == "user" {
			question = text
			history.WriteString(text)
			history.WriteByte('\n')
		}
	}
	if !side {
		return false
	}
	if tools, _ := body["tools"].([]any); len(tools) > 0 {
		http.Error(w, "side tools forbidden", 400)
		return true
	}
	if strings.Contains(question, "fail-side") {
		http.Error(w, "side fixture failure", 400)
		return true
	}
	emit(map[string]any{"id": "side-fixture", "object": "chat.completion.chunk", "choices": []any{map[string]any{"index": 0, "delta": map[string]any{"role": "assistant", "reasoning_content": "side thought", "content": "side answer: " + question + " context: " + history.String()}, "finish_reason": nil}}})
	if strings.Contains(question, "wait-side") {
		<-r.Context().Done()
		return true
	}
	emit(map[string]any{"id": "side-fixture", "object": "chat.completion.chunk", "choices": []any{map[string]any{"index": 0, "delta": map[string]any{}, "finish_reason": "stop"}}, "usage": map[string]any{"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}})
	fmt.Fprint(w, "data: [DONE]\n\n")
	w.(http.Flusher).Flush()
	return true
}

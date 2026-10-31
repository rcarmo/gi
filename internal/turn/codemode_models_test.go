package turn

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func withFakeModels(t *testing.T) {
	t.Helper()
	saved := codemodeModelsBackend
	t.Cleanup(func() { codemodeModelsBackend = saved })
	catalog := map[string]map[string]any{
		"classifier/typesafe/jev-latest": {"type": "classifier", "provider": "typesafe", "id": "jev-latest", "name": "Jev"},
		"image/openrouter/painter":       {"type": "image", "provider": "openrouter", "id": "painter", "name": "Painter"},
	}
	codemodeModelsBackend.modelOfType = func(typ, provider, id string) map[string]any { return catalog[typ+"/"+provider+"/"+id] }
	codemodeModelsBackend.modelsOfType = func(typ, provider string) []map[string]any {
		var out []map[string]any
		for _, m := range catalog {
			if m["type"] == typ {
				out = append(out, m)
			}
		}
		return out
	}
	codemodeModelsBackend.availableOfType = codemodeModelsBackend.modelsOfType
	codemodeModelsBackend.classify = func(_ context.Context, provider, id string, _ json.RawMessage) json.RawMessage {
		raw, _ := json.Marshal(map[string]any{"provider": provider, "model": id, "stopReason": "stop",
			"answers": map[string]any{"sentiment": map[string]any{"type": "choice", "choice": "positive", "confidence": 0.9}},
			"usage":   goai.Usage{Input: 10, TotalTokens: 10, Cost: goai.CostBreakdown{Total: 0.0002}}})
		return raw
	}
	codemodeModelsBackend.generateImages = func(_ context.Context, provider, id string, _ json.RawMessage) json.RawMessage {
		raw, _ := json.Marshal(map[string]any{"provider": provider, "model": id, "stopReason": "stop", "output": []any{map[string]any{"type": "image", "data": "AAAA", "mimeType": "image/png"}}})
		return raw
	}
}

// Scripts reach classifiers and image models through models.*, with Pi's
// argument checks, call rows and usage.
func TestCodemodeModelsAPI(t *testing.T) {
	withFakeModels(t)
	e := codemodeTestEngine(t)
	var details map[string]any
	var usage goai.Usage
	unknown := 0
	rt := tools.ToolRuntime{Store: e.store, ToolCallID: "tc_m", SetDetails: func(d map[string]any) { details = d }, AddUsage: func(u goai.Usage) { usage.Cost.Total += u.Cost.Total }, UnknownUsage: func() { unknown++ }}
	run := func(code string) (string, error) {
		return e.executeCodemode(context.Background(), rt, goai.ToolCall{ID: "tc_m", Name: codemodeToolName, Arguments: map[string]any{"code": code}})
	}
	out, err := run(`const jev = await models.getModelOfType("classifier", "typesafe", "jev-latest");
const r = await models.classify(jev, { state: { message: "great" }, questions: { sentiment: { type: "choice", instructions: "How?", criteria: { positive: "happy", negative: "sad" } } } });
const avail = await models.getAvailableOfType("image");
return { choice: r.answers.sentiment.choice, images: avail.map((m) => m.id) };`)
	if err != nil || !strings.Contains(out, `"choice":"positive"`) || !strings.Contains(out, `"images":["painter"]`) {
		t.Fatalf("%q %v", out, err)
	}
	calls, _ := details["calls"].([]codemodeCallRecord)
	if len(calls) != 1 || calls[0].Name != "models.classify" || calls[0].Args != "typesafe/jev-latest" || calls[0].Status != "ok" || calls[0].Cost != 0.0002 || usage.Cost.Total != 0.0002 {
		t.Fatalf("calls %+v usage %+v", calls, usage)
	}
	for code, want := range map[string]string{
		`await models.getModelOfType("classifier/typesafe/jev-latest")`:                           `models.getModelOfType(type, provider, id) expects three strings, got (a string). The provider and the id are separate arguments`,
		`await models.getModelsOfType("video")`:                                                   `Unknown model type "video". Use "chat", "image", or "classifier".`,
		`await models.classify(undefined, {})`:                                                    `models.classify() expects a classifier model as its first argument, got null. models.getModelOfType() returns undefined for an unknown provider or id. List the classifier models you can use with models.getAvailableOfType("classifier").`,
		`await models.classify({ provider: "openrouter", id: "painter" }, {})`:                    `"openrouter/painter" is an image model, not a classifier model.`,
		`await models.classify({ provider: "typesafe", id: "jev-latest" }, { prompt: 1 })`:        `models.classify() context.state must be an object, got undefined. Expected context:`,
		`await models.generateImages({ provider: "openrouter", id: "painter" }, { prompt: "x" })`: `models.generateImages() context.input must be a non-empty array of blocks, got undefined.`,
	} {
		_, err := run(code)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("%s: %v (want %q)", code, err, want)
		}
	}
	if unknown != 0 {
		t.Fatal("local validation or known counters marked unknown", unknown)
	}
	out, err = run(`const r = await models.generateImages({ provider: "openrouter", id: "painter" }, { input: [{ type: "text", text: "fox" }] }); return r.output.length;`)
	if err != nil || !strings.Contains(out, "Note: models.generateImages() returned 1 image that the script did not show.") {
		t.Fatalf("image note: %q %v", out, err)
	}
	if unknown != 1 {
		t.Fatal("remote model call without usage treated as known zero", unknown)
	}
}

// The codemode description names the models API and its reference, shipped
// in the read-only vfs://reference tree.
func TestCodemodeModelsReference(t *testing.T) {
	e := codemodeTestEngine(t)
	c := &goai.Context{Tools: []goai.Tool{{Name: codemodeToolName}}}
	e.applyCodemodeLoadout(context.Background(), c, "", nil)
	if !strings.Contains(c.Tools[0].Description, "- `models`: classifiers and image generation. Read vfs://reference/codemode-scripts.md first.") {
		t.Fatalf("description:\n%s", c.Tools[0].Description)
	}
	_, content, err := e.store.GetVFSFileContent(context.Background(), "reference", "codemode-scripts.md")
	if err != nil || !strings.Contains(string(content), "## Models") || !strings.Contains(string(content), "models.getAvailableOfType") {
		t.Fatalf("reference: %v", err)
	}
}

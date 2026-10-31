package turn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/codemode"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

// Codemode's models API, ported from Pi's createModelGlobals
// (pi-coding-agent extensions/codemode/execute.js): the typed model
// catalog, classifiers and image models, with Pi's argument checks and
// messages. Model calls are rows of the codemode block, at most four run at
// once per script, and their usage counts toward the turn.

const codemodeMaxConcurrentModelCalls = 4

const classifierContextShape = `{ state: { ... }, questions: { <id>: { type: "choice", instructions, criteria: { <label>: <meaning> } } | { type: "score", instructions, criteria: [<lowest level>, ..., <highest level>] } | { type: "bool", instructions, criteria: { true: <meaning>, false: <meaning> } } } }`

// codemodeModelsBackend runs the models API (tests replace it).
var codemodeModelsBackend = struct {
	modelsOfType    func(typ, provider string) []map[string]any
	availableOfType func(typ, provider string) []map[string]any
	modelOfType     func(typ, provider, id string) map[string]any
	classify        func(ctx context.Context, provider, id string, context json.RawMessage) json.RawMessage
	generateImages  func(ctx context.Context, provider, id string, context json.RawMessage) json.RawMessage
}{inference.ModelsOfType, inference.AvailableOfType, inference.ModelOfType, inference.Classify, inference.GenerateImages}

// jsValue is a decoded script argument with JSON's object key order kept.
type jsValue struct {
	raw json.RawMessage
}

func (v jsValue) isNull() bool { return len(v.raw) == 0 || string(v.raw) == "null" }

func (v jsValue) str() (string, bool) {
	var s string
	if v.isNull() || json.Unmarshal(v.raw, &s) != nil {
		return "", false
	}
	return s, true
}

func (v jsValue) object() (map[string]json.RawMessage, bool) {
	t := bytes.TrimSpace(v.raw)
	if len(t) == 0 || t[0] != '{' {
		return nil, false
	}
	var m map[string]json.RawMessage
	return m, json.Unmarshal(t, &m) == nil
}

func (v jsValue) array() ([]json.RawMessage, bool) {
	t := bytes.TrimSpace(v.raw)
	if len(t) == 0 || t[0] != '[' {
		return nil, false
	}
	var a []json.RawMessage
	return a, json.Unmarshal(t, &a) == nil
}

// orderedKeys returns an object's keys in source order (Object.keys).
func orderedKeys(raw json.RawMessage) []string {
	dec := json.NewDecoder(bytes.NewReader(raw))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return nil
	}
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return keys
		}
		k, _ := tok.(string)
		keys = append(keys, k)
		var skip json.RawMessage
		if dec.Decode(&skip) != nil {
			return keys
		}
	}
	return keys
}

// describeValue is Pi's: "null", "a string", "an array", or an object's keys.
func describeValue(v jsValue) string {
	t := bytes.TrimSpace(v.raw)
	switch {
	case len(t) == 0:
		return "undefined"
	case string(t) == "null":
		return "null"
	case t[0] == '[':
		if a, _ := v.array(); len(a) == 0 {
			return "an empty array"
		}
		return "an array"
	case t[0] == '{':
		keys := orderedKeys(t)
		if len(keys) == 0 {
			return "{}"
		}
		more := ""
		if len(keys) > 6 {
			keys, more = keys[:6], ", ..."
		}
		return "{ " + strings.Join(keys, ", ") + more + " }"
	case t[0] == '"':
		return "a string"
	case string(t) == "true" || string(t) == "false":
		return "a boolean"
	}
	return "a number"
}

// withArticle is Pi's: "an image", "a classifier".
func withArticle(word string) string {
	if word != "" && strings.ContainsRune("aeiou", rune(word[0])) {
		return "an " + word
	}
	return "a " + word
}

func toModelType(v jsValue) (string, error) {
	if s, ok := v.str(); ok && inference.IsModelType(s) {
		return s, nil
	}
	shown := "undefined"
	if !v.isNull() || len(v.raw) > 0 {
		shown = string(bytes.TrimSpace(v.raw))
	}
	return "", fmt.Errorf(`Unknown model type %s. Use "chat", "image", or "classifier".`, shown)
}

func toProvider(v jsValue) (string, error) {
	if v.isNull() {
		return "", nil
	}
	s, ok := v.str()
	if !ok {
		return "", errors.New("provider must be a string")
	}
	return s, nil
}

func scriptArgs(raw json.RawMessage) []jsValue {
	var a []json.RawMessage
	_ = json.Unmarshal(raw, &a)
	out := make([]jsValue, len(a))
	for i, r := range a {
		out[i] = jsValue{r}
	}
	return out
}

func argAtJS(a []jsValue, i int) jsValue {
	if i < len(a) {
		return a[i]
	}
	return jsValue{}
}

func checkClassifierContext(v jsValue) error {
	fail := func(problem string) error {
		return fmt.Errorf(`models.classify() %s. Expected context: %s. See "Classify" in %s.`, problem, classifierContextShape, codemode.ReferencePath)
	}
	ctxObj, ok := v.object()
	if !ok {
		return fail("expects a context object as its second argument, got " + describeValue(v))
	}
	if _, ok := (jsValue{ctxObj["state"]}).object(); !ok {
		return fail("context.state must be an object, got " + describeValue(jsValue{ctxObj["state"]}))
	}
	questions, ok := (jsValue{ctxObj["questions"]}).object()
	if !ok || len(questions) == 0 {
		return fail("context.questions must map question IDs to questions, got " + describeValue(jsValue{ctxObj["questions"]}))
	}
	isStrings := func(values []json.RawMessage) bool {
		if len(values) == 0 {
			return false
		}
		for _, raw := range values {
			if _, ok := (jsValue{raw}).str(); !ok {
				return false
			}
		}
		return true
	}
	for _, id := range orderedKeys(ctxObj["questions"]) {
		at := "context.questions." + id
		q, ok := (jsValue{questions[id]}).object()
		if !ok {
			return fail(at + " must be a question object, got " + describeValue(jsValue{questions[id]}))
		}
		if _, ok := (jsValue{q["instructions"]}).str(); !ok {
			return fail(at + ".instructions must be a string")
		}
		typ, _ := (jsValue{q["type"]}).str()
		criteria := jsValue{q["criteria"]}
		switch typ {
		case "choice":
			m, ok := criteria.object()
			var values []json.RawMessage
			for _, val := range m {
				values = append(values, val)
			}
			if !ok || !isStrings(values) {
				return fail(at + ` is a "choice" question, so criteria must map each label to its meaning`)
			}
		case "score":
			a, ok := criteria.array()
			if !ok || !isStrings(a) {
				return fail(at + ` is a "score" question, so criteria must list the levels as strings, lowest first`)
			}
		case "bool":
			m, ok := criteria.object()
			_, okTrue := (jsValue{m["true"]}).str()
			_, okFalse := (jsValue{m["false"]}).str()
			if !ok || !okTrue || !okFalse {
				return fail(at + ` is a "bool" question, so criteria must be { true: string, false: string }`)
			}
		default:
			shown := "undefined"
			if raw := q["type"]; len(raw) > 0 {
				shown = string(raw)
			}
			return fail(fmt.Sprintf(`%s.type must be "choice", "score", or "bool", got %s`, at, shown))
		}
	}
	return nil
}

func checkImagesContext(v jsValue) error {
	fail := func(problem string) error {
		return fmt.Errorf(`models.generateImages() %s. Expected context: { input: [{ type: "text", text: <prompt> }, ...optional { type: "image", data: <base64>, mimeType } references] }. See "Generate images" in %s.`, problem, codemode.ReferencePath)
	}
	ctxObj, ok := v.object()
	if !ok {
		return fail("expects a context object as its second argument, got " + describeValue(v))
	}
	input, ok := (jsValue{ctxObj["input"]}).array()
	if !ok || len(input) == 0 {
		return fail("context.input must be a non-empty array of blocks, got " + describeValue(jsValue{ctxObj["input"]}))
	}
	for i, raw := range input {
		b, ok := (jsValue{raw}).object()
		typ, _ := (jsValue{b["type"]}).str()
		_, text := (jsValue{b["text"]}).str()
		_, data := (jsValue{b["data"]}).str()
		_, mime := (jsValue{b["mimeType"]}).str()
		if ok && (typ == "text" && text || typ == "image" && data && mime) {
			continue
		}
		return fail(fmt.Sprintf("context.input[%d] must be a text or image block, got %s", i, describeValue(jsValue{raw})))
	}
	return nil
}

func marshalJS(v any) (json.RawMessage, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

// codemodeModelGlobals ports Pi's createModelGlobals.
func (e *Engine) codemodeModelGlobals(rt tools.ToolRuntime, calls *codemodeCallLog, generatedImages *int) []codemode.Global {
	backend := codemodeModelsBackend
	slots := make(chan struct{}, codemodeMaxConcurrentModelCalls)
	callCount := 0
	runModelCall := func(ctx context.Context, name, typ string, args []jsValue, check func(jsValue) error, run func(ctx context.Context, provider, id string, context json.RawMessage) json.RawMessage) (json.RawMessage, error) {
		listHint := fmt.Sprintf(`List the %s models you can use with models.getAvailableOfType("%s").`, typ, typ)
		modelArg := argAtJS(args, 0)
		m, isObj := modelArg.object()
		provider, okP := (jsValue{m["provider"]}).str()
		id, okI := (jsValue{m["id"]}).str()
		if !isObj || !okP || !okI {
			hint := ""
			if modelArg.isNull() {
				hint = " models.getModelOfType() returns undefined for an unknown provider or id."
			}
			return nil, fmt.Errorf("%s() expects %s model as its first argument, got %s.%s %s", name, withArticle(typ), describeValue(modelArg), hint, listHint)
		}
		ref := provider + "/" + id
		if backend.modelOfType(typ, provider, id) == nil {
			for _, other := range inference.ModelTypes {
				if other != typ && backend.modelOfType(other, provider, id) != nil {
					return nil, fmt.Errorf(`"%s" is %s model, not %s model. %s`, ref, withArticle(other), withArticle(typ), listHint)
				}
			}
			return nil, fmt.Errorf(`Unknown %s model "%s". %s`, typ, ref, listHint)
		}
		contextArg := argAtJS(args, 1)
		if err := check(contextArg); err != nil {
			return nil, err
		}
		callCount++
		rec := calls.start(fmt.Sprintf("%s/%s/%d", rt.ToolCallID, name, callCount), name, nil)
		calls.mu.Lock()
		rec.Args = ref
		calls.mu.Unlock()
		select {
		case slots <- struct{}{}:
		case <-ctx.Done():
			calls.finish(rec, ctx.Err(), true)
			return nil, ctx.Err()
		}
		result := run(ctx, provider, id, contextArg.raw)
		<-slots
		var parsed struct {
			StopReason   string      `json:"stopReason"`
			ErrorMessage string      `json:"errorMessage"`
			Usage        *goai.Usage `json:"usage"`
		}
		_ = json.Unmarshal(result, &parsed)
		calls.mu.Lock()
		rec.DurationMs = float64(time.Since(rec.started).Microseconds()) / 1000
		switch parsed.StopReason {
		case "stop":
			rec.Status = "ok"
		case "aborted":
			rec.Status = "cancelled"
		default:
			rec.Status = "error"
		}
		if parsed.ErrorMessage != "" {
			rec.Error = truncatePreview(parsed.ErrorMessage, codemodeErrorPreviewChars)
		}
		if parsed.Usage != nil {
			rec.Cost = parsed.Usage.Cost.Total
		}
		calls.mu.Unlock()
		if parsed.Usage != nil && rt.AddUsage != nil {
			rt.AddUsage(*parsed.Usage)
		} else if parsed.Usage == nil && rt.UnknownUsage != nil {
			rt.UnknownUsage()
		}
		return result, nil
	}
	global := func(name string, fn func(context.Context, []jsValue) (json.RawMessage, error)) codemode.Global {
		return codemode.Global{Name: name, Spread: true, Execute: func(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
			return fn(ctx, scriptArgs(raw))
		}}
	}
	return []codemode.Global{
		global("models.getModelsOfType", func(_ context.Context, a []jsValue) (json.RawMessage, error) {
			typ, err := toModelType(argAtJS(a, 0))
			if err != nil {
				return nil, err
			}
			provider, err := toProvider(argAtJS(a, 1))
			if err != nil {
				return nil, err
			}
			return marshalJS(sortedModels(backend.modelsOfType(typ, provider)))
		}),
		global("models.getAvailableOfType", func(_ context.Context, a []jsValue) (json.RawMessage, error) {
			typ, err := toModelType(argAtJS(a, 0))
			if err != nil {
				return nil, err
			}
			provider, err := toProvider(argAtJS(a, 1))
			if err != nil {
				return nil, err
			}
			return marshalJS(sortedModels(backend.availableOfType(typ, provider)))
		}),
		global("models.getModelOfType", func(_ context.Context, a []jsValue) (json.RawMessage, error) {
			provider, okP := argAtJS(a, 1).str()
			id, okI := argAtJS(a, 2).str()
			if !okP || !okI {
				shown := make([]string, len(a))
				for i, v := range a {
					shown[i] = describeValue(v)
				}
				return nil, fmt.Errorf(`models.getModelOfType(type, provider, id) expects three strings, got (%s). The provider and the id are separate arguments, for example models.getModelOfType("classifier", "typesafe", "jev-latest").`, strings.Join(shown, ", "))
			}
			typ, err := toModelType(argAtJS(a, 0))
			if err != nil {
				return nil, err
			}
			if m := backend.modelOfType(typ, provider, id); m != nil {
				return marshalJS(m)
			}
			return nil, nil
		}),
		global("models.classify", func(ctx context.Context, a []jsValue) (json.RawMessage, error) {
			return runModelCall(ctx, "models.classify", "classifier", a, checkClassifierContext, backend.classify)
		}),
		global("models.generateImages", func(ctx context.Context, a []jsValue) (json.RawMessage, error) {
			result, err := runModelCall(ctx, "models.generateImages", "image", a, checkImagesContext, backend.generateImages)
			if err == nil {
				var parsed struct {
					Output []struct {
						Type string `json:"type"`
					} `json:"output"`
				}
				_ = json.Unmarshal(result, &parsed)
				calls.mu.Lock()
				for _, block := range parsed.Output {
					if block.Type == "image" {
						*generatedImages++
					}
				}
				calls.mu.Unlock()
			}
			return result, err
		}),
	}
}

// sortedModels keeps catalog output stable (provider, then id).
func sortedModels(models []map[string]any) []map[string]any {
	sort.SliceStable(models, func(i, j int) bool {
		pi, _ := models[i]["provider"].(string)
		pj, _ := models[j]["provider"].(string)
		if pi != pj {
			return pi < pj
		}
		ii, _ := models[i]["id"].(string)
		ij, _ := models[j]["id"].(string)
		return ii < ij
	})
	return models
}

package inference

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"

	goai "github.com/rcarmo/go-ai"
	_ "github.com/rcarmo/go-ai/inference/provider/anthropic"
	_ "github.com/rcarmo/go-ai/inference/provider/openai"
	_ "github.com/rcarmo/go-ai/inference/provider/openaicodex"
	_ "github.com/rcarmo/go-ai/inference/provider/openairesponses"
	"github.com/rcarmo/go-ai/oauth"

	"github.com/rcarmo/gi/internal/config"
)

var once sync.Once

type authEntry struct {
	Type    string `json:"type"`
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
	Expires int64  `json:"expires"`
	Token   string `json:"token"`
	// Key is Pi's stored API key ({"type": "api_key", "key": ...});
	// APIKey is gi's earlier field, still read.
	Key    string `json:"key"`
	APIKey string `json:"apiKey"`
	// Pi records the account's Copilot model list and enterprise domain;
	// listings are filtered to it like Pi's provider filterModels.
	AvailableModelIDs []string `json:"availableModelIds"`
	EnterpriseURL     string   `json:"enterpriseUrl"`
}

type ProviderOption struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Authenticated bool   `json:"authenticated"`
}

type ModelOption struct {
	ID            string `json:"id"`
	Provider      string `json:"provider"`
	Label         string `json:"label"`
	Name          string `json:"name,omitempty"`
	ContextWindow int    `json:"context_window,omitempty"`
	Reasoning     bool   `json:"reasoning,omitempty"`
	Authenticated bool   `json:"authenticated"`
	Enabled       bool   `json:"enabled,omitempty"`
}

func Init() {
	once.Do(func() {
		goai.RegisterBuiltinModels()
		registerCustomModels()
		registerPiModelsStore(PiModelsStorePath())
	})
}

// ResolveModelContextWindow returns the registered context window for the selected
// provider/model pair. The model argument may be either an ID or provider/id label.
func ResolveModelContextWindow(defaultProvider, model string) int {
	Init()
	provider, id := splitModelLabel(defaultProvider, model)
	if id == "" {
		return 0
	}
	if m := goai.GetModel(goai.Provider(provider), id); m != nil && m.ContextWindow > 0 {
		return m.ContextWindow
	}
	return 0
}

func registerCustomModels() {
	goai.RegisterModel(&goai.Model{
		ID:            "minimax-m2.5-free",
		Name:          "OpenCode Zen / MiniMax M2.5 Free",
		Api:           goai.ApiOpenAICompletions,
		Provider:      "opencode-zen",
		BaseURL:       "https://opencode.ai/zen/v1",
		Reasoning:     false,
		Input:         []string{"text"},
		ContextWindow: 128000,
		MaxTokens:     16384,
	})
}

// AuthFilePath is the credentials file: ~/.gi/agent/auth.json when it
// exists, else Pi's (PI_CODING_AGENT_DIR or ~/.pi/agent), which is shared
// with Pi and refreshed in place.
func AuthFilePath() string { return config.UserConfigFile("auth.json") }

func loadAuthEntries() (map[string]authEntry, error) {
	d, err := readCredentials()
	if err != nil {
		return nil, err
	}
	entries := map[string]authEntry{}
	for id, raw := range d.entries {
		var entry authEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			return nil, fmt.Errorf("invalid credential fields")
		}
		entries[id] = entry
	}
	return entries, nil
}

// AuthStatus describes a known OAuth/credential provider and whether the local
// auth.json currently holds credentials for it.
type AuthStatus struct {
	ID            string
	Name          string
	Authenticated bool
	Kind          string
}

// subscriptionProviders are the providers whose OAuth sign-in is backed by
// a subscription (pi-ai's isSubscription: Claude Pro/Max, Copilot, ChatGPT,
// Codex, Kimi Code, Meta Muse, SuperGrok/X Premium); other OAuth sign-ins
// are accounts.
var subscriptionProviders = map[string]bool{
	"anthropic": true, "github-copilot": true, "kimi-coding": true, "meta": true,
	"openai": true, "openai-codex": true, "xai": true,
}

// AuthTypeLabel is Pi's formatAuthSelectorProviderType: "API key",
// "subscription" for subscription-backed OAuth providers, else "account".
func (s AuthStatus) AuthTypeLabel() string {
	if s.Kind == "api-key" || s.Kind == "api_key" {
		return "API key"
	}
	if subscriptionProviders[s.ID] {
		return "subscription"
	}
	return "account"
}

// ListAuthStatus returns the registered OAuth providers plus any auth.json
// entries, marking which are currently authenticated. It never performs network
// calls.
func ListAuthStatus() []AuthStatus {
	Init()
	entries, _ := loadAuthEntries()
	seen := map[string]bool{}
	var out []AuthStatus
	for _, p := range oauth.ListProviders() {
		id := p.ID()
		seen[id] = true
		entry, ok := entries[id]
		kind := entry.Type
		if kind == "" && ok {
			kind = "oauth"
		}
		out = append(out, AuthStatus{ID: id, Name: p.Name(), Authenticated: ok, Kind: kind})
	}
	for id, entry := range entries {
		if seen[id] {
			continue
		}
		kind := entry.Type
		if kind == "" {
			if entry.APIKey != "" || entry.Key != "" {
				kind = "api-key"
			} else {
				kind = "credential"
			}
		}
		out = append(out, AuthStatus{ID: id, Name: providerName(id), Authenticated: true, Kind: kind})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// RemoveAuthEntry deletes a provider's credentials from auth.json. It returns
// (false, nil) when there was nothing to remove.
func RemoveAuthEntry(provider string) (bool, error) {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return false, fmt.Errorf("provider is required")
	}
	removed := false
	_, err := updateCredentials("", func(entries map[string]json.RawMessage) error {
		if _, ok := entries[provider]; ok {
			delete(entries, provider)
			removed = true
		}
		return nil
	})
	return removed, err
}

func authEntryToOAuthCredentials(entry authEntry) *oauth.Credentials {
	creds := &oauth.Credentials{
		Refresh: entry.Refresh,
		Access:  entry.Access,
		Expires: entry.Expires,
	}
	if entry.AvailableModelIDs != nil || entry.EnterpriseURL != "" {
		creds.Extra = map[string]interface{}{}
		if entry.AvailableModelIDs != nil {
			creds.Extra["availableModelIds"] = entry.AvailableModelIDs
		}
		if entry.EnterpriseURL != "" {
			creds.Extra["enterpriseUrl"] = entry.EnterpriseURL
		}
	}
	return creds
}

func providerName(id string) string {
	if p := oauth.GetProvider(id); p != nil {
		return p.Name()
	}
	parts := strings.FieldsFunc(id, func(r rune) bool { return r == '-' || r == '_' || r == '/' })
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	if len(parts) == 0 {
		return id
	}
	return strings.Join(parts, " ")
}

func splitModelLabel(defaultProvider, label string) (string, string) {
	label = strings.TrimSpace(label)
	if label == "" {
		return strings.TrimSpace(defaultProvider), ""
	}
	if slash := strings.Index(label, "/"); slash > 0 {
		return strings.TrimSpace(label[:slash]), strings.TrimSpace(label[slash+1:])
	}
	return strings.TrimSpace(defaultProvider), label
}

func modelLabel(provider, id string) string {
	provider = strings.TrimSpace(provider)
	id = strings.TrimSpace(id)
	if provider == "" {
		return id
	}
	if strings.HasPrefix(id, provider+"/") {
		return id
	}
	return provider + "/" + id
}

func ListRuntimeOptions(defaultProvider, defaultModel string, enabledModels []string) ([]ProviderOption, []ModelOption) {
	Init()

	authEntries, err := loadAuthEntries()
	if err != nil {
		authEntries = map[string]authEntry{}
	}
	authenticatedProviders := map[string]authEntry{}
	for id, entry := range authEntries {
		authenticatedProviders[strings.TrimSpace(id)] = entry
	}

	providerSet := map[string]bool{}
	enabledSet := map[string]bool{}
	modelOptions := map[string]ModelOption{}

	addProvider := func(id string) {
		id = strings.TrimSpace(id)
		if id != "" {
			providerSet[id] = true
		}
	}
	addModel := func(provider string, model *goai.Model, enabled bool) {
		provider = strings.TrimSpace(provider)
		if model == nil {
			return
		}
		id := strings.TrimSpace(model.ID)
		if id == "" {
			return
		}
		label := modelLabel(provider, id)
		current := modelOptions[label]
		current.ID = id
		current.Provider = provider
		current.Label = label
		if strings.TrimSpace(model.Name) != "" {
			current.Name = model.Name
		}
		if model.ContextWindow > 0 {
			current.ContextWindow = model.ContextWindow
		}
		current.Reasoning = current.Reasoning || model.Reasoning
		current.Authenticated = current.Authenticated || authenticatedProviders[provider].Type != "" || provider == "opencode-zen"
		current.Enabled = current.Enabled || enabled
		modelOptions[label] = current
		addProvider(provider)
	}
	addSyntheticModel := func(provider, id string, enabled bool) {
		provider = strings.TrimSpace(provider)
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}
		label := modelLabel(provider, id)
		current := modelOptions[label]
		current.ID = id
		current.Provider = provider
		current.Label = label
		current.Authenticated = current.Authenticated || authenticatedProviders[provider].Type != "" || provider == "opencode-zen"
		current.Enabled = current.Enabled || enabled
		modelOptions[label] = current
		addProvider(provider)
	}

	// GetModel clones an entire provider catalogue before finding one ID.
	// Snapshot each requested provider once per call; no cross-call cache can
	// conceal a registry change or an external credential rotation.
	catalogues := map[string][]*goai.Model{}
	findModel := func(provider, id string) *goai.Model {
		list, ok := catalogues[provider]
		if !ok {
			list = goai.ListModels(goai.Provider(provider))
			catalogues[provider] = list
		}
		for _, m := range list {
			if m.ID == id {
				return m
			}
		}
		return nil
	}
	addProvider(defaultProvider)
	for _, label := range enabledModels {
		provider, id := splitModelLabel(defaultProvider, label)
		full := modelLabel(provider, id)
		if full != "" {
			enabledSet[full] = true
		}
		if model := findModel(provider, id); model != nil {
			addModel(provider, model, true)
		} else {
			addSyntheticModel(provider, id, true)
		}
	}
	if provider, id := splitModelLabel(defaultProvider, defaultModel); id != "" {
		if model := findModel(provider, id); model != nil {
			addModel(provider, model, enabledSet[modelLabel(provider, id)])
		} else {
			addSyntheticModel(provider, id, enabledSet[modelLabel(provider, id)])
		}
	}

	for provider, entry := range authenticatedProviders {
		addProvider(provider)
		models, ok := catalogues[provider]
		if !ok {
			models = goai.ListModels(goai.Provider(provider))
			catalogues[provider] = models
		}
		if p := oauth.GetProvider(provider); p != nil {
			models = p.ModifyModels(models, authEntryToOAuthCredentials(entry))
		}
		for _, model := range models {
			addModel(provider, model, enabledSet[modelLabel(provider, model.ID)])
		}
	}

	providers := make([]ProviderOption, 0, len(providerSet))
	for id := range providerSet {
		providers = append(providers, ProviderOption{
			ID:            id,
			Name:          providerName(id),
			Authenticated: authenticatedProviders[id].Type != "" || id == "opencode-zen",
		})
	}
	sort.Slice(providers, func(i, j int) bool {
		if providers[i].Authenticated != providers[j].Authenticated {
			return providers[i].Authenticated && !providers[j].Authenticated
		}
		return strings.ToLower(providers[i].Name) < strings.ToLower(providers[j].Name)
	})

	models := make([]ModelOption, 0, len(modelOptions))
	for _, model := range modelOptions {
		models = append(models, model)
	}
	sort.Slice(models, func(i, j int) bool {
		if models[i].Enabled != models[j].Enabled {
			return models[i].Enabled && !models[j].Enabled
		}
		if models[i].Authenticated != models[j].Authenticated {
			return models[i].Authenticated && !models[j].Authenticated
		}
		return strings.ToLower(models[i].Label) < strings.ToLower(models[j].Label)
	})

	return providers, models
}

func loadAuth(provider string) (string, string, error) {
	if provider == "opencode-zen" {
		return "", "https://opencode.ai/zen/v1", nil
	}
	entries, err := loadAuthEntries()
	if err != nil {
		return "", "", err
	}
	entry, ok := entries[provider]
	if !ok {
		return "", "", fmt.Errorf("no auth entry for provider %q", provider)
	}

	if provider == "github-copilot" {
		return copilotAuth(entry)
	}

	if entry.Access != "" {
		return entry.Access, "", nil
	}
	if entry.Token != "" {
		return entry.Token, "", nil
	}
	if entry.Key != "" {
		return entry.Key, "", nil
	}
	if entry.APIKey != "" {
		return entry.APIKey, "", nil
	}
	return "", "", fmt.Errorf("no token/key in auth entry for %q", provider)
}

// StreamResult holds the full result from a streaming inference call.
type StreamResult struct {
	// OccurrenceID is assigned by Gi's turn owner, not the remote provider.
	OccurrenceID string `json:"-"`
	// RequestThinking is the turn owner's captured preparation, not a remote
	// provider assertion. Nil retains legacy admission-metadata fallback.
	RequestThinking *string `json:"-"`
	Message         *goai.Message
	Partial         *goai.Message   `json:"-"`
	Incomplete      bool            `json:"-"`
	TerminalReason  goai.StopReason `json:"-"`
	Usage           *goai.Usage
	Text            string
}

type StreamHooks struct {
	// Validated choice captured for this request; empty leaves the provider default unset.
	Thinking   string
	Transport  goai.Transport // Pi's configured transport; empty leaves go-ai's default
	OnPayload  func(payload any, model *goai.Model) (any, error)
	OnResponse func(status int, headers map[string]string, model *goai.Model)
	// OnProgress receives provider snapshots; an error stops this request.
	OnProgress func(*goai.Message) error
	// MaxTokens caps the response; CacheRetention "none" avoids cache
	// writes for one-off requests (compaction summaries, as Pi).
	MaxTokens      int
	CacheRetention goai.CacheRetention
}

// StreamWithTools streams a single LLM call (which may produce tool calls).
// It returns the complete assistant message including any tool-call content blocks.
// The broadcast callback receives SSE-shaped events for real-time UI updates.
func StreamWithTools(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*StreamResult, error) {
	return StreamWithToolsWithHooks(ctx, modelID, convCtx, broadcast, nil)
}

// StreamWithToolsWithHooks is the hook-aware variant of StreamWithTools.
func StreamWithToolsWithHooks(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any), hooks *StreamHooks) (*StreamResult, error) {
	Init()
	convCtx = contextWithObjectToolInputs(convCtx)

	provider, modelName := splitModelID(modelID)
	model := goai.GetModel(goai.Provider(provider), modelName)
	if model == nil {
		return nil, fmt.Errorf("model not found: %s/%s", provider, modelName)
	}
	if hooks != nil && hooks.Thinking != "" {
		if err := ValidateThinking(modelID, hooks.Thinking); err != nil {
			return nil, err
		}
		if provider == "opencode-zen" {
			return nil, fmt.Errorf("thinking configuration unavailable for this transport")
		}
	}
	if provider == "opencode-zen" {
		return streamOpenCodeZen(ctx, model, convCtx, broadcast, hooks)
	}

	apiKey, baseURLOverride, err := loadAuth(provider)
	if err != nil {
		return nil, fmt.Errorf("load auth for %s: %w", provider, err)
	}
	if baseURLOverride != "" {
		model.BaseURL = baseURLOverride
	}

	opts := &goai.StreamOptions{
		APIKey: apiKey,
	}
	if hooks != nil {
		opts.OnPayload = hooks.OnPayload
		opts.OnResponse = hooks.OnResponse
		opts.Transport = hooks.Transport
		if hooks.Thinking != "" {
			level := goai.ThinkingLevel(hooks.Thinking)
			opts.Reasoning = &level
		}
		if hooks.MaxTokens > 0 {
			maxTokens := hooks.MaxTokens
			opts.MaxTokens = &maxTokens
		}
		if hooks.CacheRetention != "" {
			opts.CacheRetention = hooks.CacheRetention
		}
	}
	if model.Api == goai.ApiOpenAICodexResponses {
		opts.OnPayload = codexPayloadHook(opts.OnPayload)
	}
	if provider == "github-copilot" {
		opts.Headers = goai.CopilotHeaders()
	}

	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var observe func(*goai.Message) error
	if hooks != nil && hooks.OnProgress != nil {
		observe = func(message *goai.Message) error {
			err := hooks.OnProgress(message)
			if err != nil {
				cancel()
			}
			return err
		}
	}
	return collectProviderStream(ctx, modelID, goai.Stream(streamCtx, model, convCtx, opts), broadcast, observe)
}

// Drain the producer through channel closure before returning ownership. This
// also makes duplicate/missing terminals observable instead of adopting the
// first DoneEvent and abandoning the producer.
func collectProviderStream(ctx context.Context, modelID string, events <-chan goai.Event, broadcast func(map[string]any), observers ...func(*goai.Message) error) (*StreamResult, error) {
	var fullText string
	var result *StreamResult
	var terminalCount int
	var terminalReason goai.StopReason
	var streamErr, progressErr error
	var lastPartial *goai.Message
	var fallback fallbackProgress
	for ev := range events {
		switch ev.(type) {
		case *goai.DoneEvent, *goai.ErrorEvent:
		default:
			fallbackMessage := fallback.observe(ev)
			partial := streamEventPartial(ev)
			if partial == nil {
				partial = fallbackMessage
			}
			if partial != nil {
				lastPartial = partial
				if progressErr == nil && len(observers) > 0 && observers[0] != nil {
					progressErr = observers[0](partial)
				}
			}
		}
		switch e := ev.(type) {
		case *goai.TextDeltaEvent:
			fullText += e.Delta
			if broadcast != nil {
				broadcast(map[string]any{"type": "text_delta", "delta": e.Delta})
			}
		case *goai.ThinkingDeltaEvent:
			if broadcast != nil {
				broadcast(map[string]any{"type": "thinking_delta", "delta": e.Delta})
			}
		case *goai.ToolCallStartEvent:
			if broadcast != nil && e.Partial != nil {
				// Extract tool name from partial message
				for _, b := range e.Partial.Content {
					if b.Type == "toolCall" && b.Name != "" {
						broadcast(map[string]any{"type": "tool_call_start", "name": b.Name, "index": e.ContentIndex})
						break
					}
				}
			}
		case *goai.ToolCallDeltaEvent:
			// accumulation happens inside go-ai; no need to broadcast deltas
		case *goai.ToolCallEndEvent:
			if broadcast != nil {
				broadcast(map[string]any{"type": "tool_call_end", "name": e.ToolCall.Name, "id": e.ToolCall.ID})
			}
		case *goai.DoneEvent:
			terminalCount++
			terminalReason = e.Reason
			if e.Message == nil {
				streamErr = fmt.Errorf("provider returned a nil terminal message")
				continue
			}
			if e.Reason != "" && e.Reason != e.Message.StopReason {
				streamErr = fmt.Errorf("provider terminal stop reason mismatch")
			}
			result = &StreamResult{Message: objectToolInputs(e.Message), Usage: e.Message.Usage, Text: fullText}
		case *goai.ErrorEvent:
			terminalCount++
			terminalReason = e.Reason
			streamErr = e.Err
			if streamErr == nil {
				streamErr = fmt.Errorf("provider terminal error: %s", e.Reason)
			}
			result = &StreamResult{Message: e.Error, Text: fullText}
			if e.Error != nil {
				result.Usage = e.Error.Usage
			}
			if broadcast != nil {
				broadcast(map[string]any{"type": "error", "error": streamErr.Error()})
			}
		}
	}

	if result == nil {
		result = &StreamResult{Text: fullText}
	}
	result.TerminalReason = terminalReason
	result.Partial = lastPartial
	result.Incomplete = terminalCount != 1 || result.Message == nil || ctx.Err() != nil || progressErr != nil || terminalReason == goai.StopReasonAborted || (result.Message != nil && result.Message.StopReason == goai.StopReasonAborted)
	if progressErr != nil {
		streamErr = progressErr
	}
	if terminalCount != 1 {
		streamErr = fmt.Errorf("provider stream has %d terminals, want 1", terminalCount)
	}
	if ctx.Err() != nil {
		streamErr = ctx.Err()
	}
	if streamErr == nil {
		streamErr = ValidateAssistantResult(result)
	}
	if streamErr != nil {
		return result, streamErr
	}
	if broadcast != nil {
		usageMap := map[string]any{}
		if usage := result.Usage; usage != nil {
			usageMap = map[string]any{"input": usage.Input, "output": usage.Output, "total": usage.TotalTokens, "cache_read": usage.CacheRead, "cache_write": usage.CacheWrite, "cost_input": usage.Cost.Input, "cost_output": usage.Cost.Output, "cost_total": usage.Cost.Total}
		}
		broadcast(map[string]any{"type": "done", "model": modelID, "usage": usageMap})
	}
	return result, nil
}

func responseHeadersMap(header http.Header) map[string]string {
	if len(header) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(header))
	for key, values := range header {
		if len(values) == 0 {
			out[key] = ""
			continue
		}
		out[key] = strings.Join(values, ", ")
	}
	return out
}

func splitModelID(id string) (string, string) {
	for i, c := range id {
		if c == '/' {
			return id[:i], id[i+1:]
		}
	}
	return "", id
}

func streamOpenCodeZen(ctx context.Context, model *goai.Model, convCtx *goai.Context, broadcast func(map[string]any), hooks *StreamHooks) (*StreamResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	messages, err := openAICompatMessages(convCtx)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"model":    model.ID,
		"stream":   true,
		"messages": messages,
	}
	if len(convCtx.Tools) > 0 {
		payload["tools"] = openAICompatTools(convCtx.Tools)
	}
	var payloadValue any = payload
	if hooks != nil && hooks.OnPayload != nil {
		replaced, err := hooks.OnPayload(payloadValue, model)
		if err != nil {
			return nil, err
		}
		if replaced != nil {
			payloadValue = replaced
		}
	}
	body, err := json.Marshal(payloadValue)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(model.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if hooks != nil && hooks.OnResponse != nil {
		hooks.OnResponse(resp.StatusCode, responseHeadersMap(resp.Header), model)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	var fullText, fullThinking string
	var terminalCount int
	var terminalStop goai.StopReason
	usage := &goai.Usage{}
	toolCalls := map[int]*goai.ToolCall{}
	argumentFragments := map[int]*strings.Builder{}
	var partial *goai.Message
	var progressErr error
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content   string `json:"content"`
					Reasoning string `json:"reasoning"`
					ToolCalls []struct {
						Index    int    `json:"index"`
						ID       string `json:"id"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Usage struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.Usage.TotalTokens > 0 {
			usage.Input = chunk.Usage.PromptTokens
			usage.Output = chunk.Usage.CompletionTokens
			usage.TotalTokens = chunk.Usage.TotalTokens
		}
		for _, choice := range chunk.Choices {
			if choice.FinishReason != "" {
				terminalCount++
				switch choice.FinishReason {
				case "stop":
					terminalStop = goai.StopReasonStop
				case "length":
					terminalStop = goai.StopReasonLength
				case "tool_calls":
					terminalStop = goai.StopReasonToolUse
				default:
					terminalStop = goai.StopReason(choice.FinishReason)
				}
			}
			if choice.Delta.Reasoning != "" {
				fullThinking += choice.Delta.Reasoning
				if broadcast != nil {
					broadcast(map[string]any{"type": "thinking_delta", "delta": choice.Delta.Reasoning})
				}
			}
			if choice.Delta.Content != "" {
				fullText += choice.Delta.Content
				if broadcast != nil {
					broadcast(map[string]any{"type": "text_delta", "delta": choice.Delta.Content})
				}
			}
			for _, tc := range choice.Delta.ToolCalls {
				call := toolCalls[tc.Index]
				if call == nil {
					call = &goai.ToolCall{Type: "toolCall"}
					toolCalls[tc.Index] = call
					if broadcast != nil && tc.Function.Name != "" {
						broadcast(map[string]any{"type": "tool_call_start", "name": tc.Function.Name, "index": tc.Index})
					}
				}
				if tc.ID != "" {
					call.ID = tc.ID
				}
				if tc.Function.Name != "" {
					call.Name = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					if call.Arguments == nil {
						call.Arguments = map[string]any{}
					}
					builder := argumentFragments[tc.Index]
					if builder == nil {
						builder = &strings.Builder{}
						argumentFragments[tc.Index] = builder
					}
					builder.WriteString(tc.Function.Arguments)
				}
			}
		}
		partial = compatProviderPartial(model, fullText, fullThinking, usage, toolCalls, argumentFragments)
		if hooks != nil && hooks.OnProgress != nil {
			if err := hooks.OnProgress(partial); err != nil {
				progressErr = err
				cancel()
				break
			}
		}
	}
	msg := &goai.Message{Role: goai.RoleAssistant, Model: model.ID, Provider: model.Provider, Api: model.Api, Usage: usage, StopReason: terminalStop}
	if fullThinking != "" {
		msg.Content = append(msg.Content, goai.ContentBlock{Type: "thinking", Thinking: fullThinking})
	}
	if fullText != "" {
		msg.Content = append(msg.Content, goai.ContentBlock{Type: "text", Text: fullText})
	}
	for _, idx := range sortedToolCallIndexes(toolCalls) {
		call := toolCalls[idx]
		if fragments := argumentFragments[idx]; fragments != nil {
			call.Arguments, call.ArgumentsError = completedCompatToolArguments(fragments.String())
		} else {
			call.Arguments, call.ArgumentsError = completedCompatToolArguments("")
		}
		msg.Content = append(msg.Content, goai.ContentBlock{Type: "toolCall", ID: call.ID, Name: call.Name, Arguments: call.Arguments, ArgumentsError: call.ArgumentsError})
		if broadcast != nil {
			broadcast(map[string]any{"type": "tool_call_end", "name": call.Name, "id": call.ID})
		}
	}
	result := &StreamResult{Message: msg, Usage: usage, Text: fullText, Partial: partial, TerminalReason: terminalStop, Incomplete: terminalCount != 1 || scanner.Err() != nil || ctx.Err() != nil || progressErr != nil}
	if progressErr != nil {
		return result, progressErr
	}
	if err := scanner.Err(); err != nil {
		return result, err
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if terminalCount != 1 {
		return result, fmt.Errorf("provider stream has %d terminals, want 1", terminalCount)
	}
	if err := ValidateAssistantResult(result); err != nil {
		return result, err
	}
	if broadcast != nil {
		broadcast(map[string]any{"type": "done", "model": string(model.Provider) + "/" + model.ID, "usage": map[string]any{"input": usage.Input, "output": usage.Output, "total": usage.TotalTokens}})
	}
	return result, nil
}

func openAICompatMessages(convCtx *goai.Context) ([]map[string]any, error) {
	messages := make([]map[string]any, 0, len(convCtx.Messages)+1)
	if strings.TrimSpace(convCtx.SystemPrompt) != "" {
		messages = append(messages, map[string]any{"role": "system", "content": convCtx.SystemPrompt})
	}
	for _, msg := range convCtx.Messages {
		role := strings.ToLower(string(msg.Role))
		content := goai.GetTextContent(&msg)
		var imageParts []any
		for _, block := range msg.Content {
			if block.Type == "image" {
				imageParts = append(imageParts, map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:" + block.MimeType + ";base64," + block.Data}})
			}
		}
		if role == "tool" || role == "toolresult" {
			entry := map[string]any{"role": "tool", "content": content}
			if msg.ToolCallID != "" {
				entry["tool_call_id"] = msg.ToolCallID
			}
			if len(imageParts) > 0 && strings.TrimSpace(content) == "" {
				entry["content"] = "(no tool output)"
			}
			messages = append(messages, entry)
			if len(imageParts) > 0 {
				parts := []any{map[string]any{"type": "text", "text": "Tool result image:"}}
				parts = append(parts, imageParts...)
				messages = append(messages, map[string]any{"role": "user", "content": parts})
			}
			continue
		}
		entry := map[string]any{"role": role, "content": content}
		if len(imageParts) > 0 {
			if role != "user" {
				return nil, fmt.Errorf("compatibility messages do not support images in role %q", role)
			}
			var parts []any
			for _, block := range msg.Content {
				switch block.Type {
				case "text":
					if block.Text != "" {
						parts = append(parts, map[string]any{"type": "text", "text": block.Text})
					}
				case "image":
					parts = append(parts, map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:" + block.MimeType + ";base64," + block.Data}})
				}
			}
			entry["content"] = parts
		}
		if role == "assistant" {
			var wireCalls []any
			for _, call := range goai.GetToolCalls(&msg) {
				args := call.Arguments
				if args == nil {
					args = map[string]any{}
				}
				arguments, err := json.Marshal(args)
				if err != nil {
					return nil, fmt.Errorf("encode tool arguments for %s: %w", call.Name, err)
				}
				wireCalls = append(wireCalls, map[string]any{"id": call.ID, "type": "function", "function": map[string]any{"name": call.Name, "arguments": string(arguments)}})
			}
			if len(wireCalls) > 0 {
				entry["tool_calls"] = wireCalls
				if content == "" {
					entry["content"] = nil
				}
			}
		}
		messages = append(messages, entry)
	}
	return messages, nil
}

func openAICompatTools(tools []goai.Tool) []map[string]any {
	out := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		out = append(out, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  json.RawMessage(tool.Parameters),
			},
		})
	}
	return out
}

func sortedToolCallIndexes(calls map[int]*goai.ToolCall) []int {
	idx := make([]int, 0, len(calls))
	for k := range calls {
		idx = append(idx, k)
	}
	for i := 0; i < len(idx); i++ {
		for j := i + 1; j < len(idx); j++ {
			if idx[j] < idx[i] {
				idx[i], idx[j] = idx[j], idx[i]
			}
		}
	}
	return idx
}

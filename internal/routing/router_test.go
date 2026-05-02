package routing

import (
	"strings"
	"testing"
)

func TestRouterSelectsLightModelForSimplePrompt(t *testing.T) {
	router := NewRouter(ModelRoutingConfig{Enabled: true, LightModel: "flash", Threshold: 0.35})
	model, usedLight, score := router.SelectModel("hello", nil, "heavy")
	if model != "flash" || !usedLight || score >= 0.35 {
		t.Fatalf("unexpected selection: model=%s light=%v score=%f", model, usedLight, score)
	}
}

func TestRouterSelectsPrimaryModelForCodePrompt(t *testing.T) {
	router := NewRouter(ModelRoutingConfig{Enabled: true, LightModel: "flash", Threshold: 0.35})
	prompt := "please fix this\n```go\nfmt.Println(\"hi\")\n```"
	model, usedLight, score := router.SelectModel(prompt, nil, "heavy")
	if model != "heavy" || usedLight || score < 0.35 {
		t.Fatalf("unexpected selection: model=%s light=%v score=%f", model, usedLight, score)
	}
}

func TestExtractFeaturesCountsRecentToolMessages(t *testing.T) {
	history := []HistoryMessage{{Payload: map[string]any{"kind": "tool"}}, {Payload: map[string]any{"source": "tool"}}, {Payload: map[string]any{}}}
	features := ExtractFeatures(strings.Repeat("x", 240), history)
	if features.RecentToolCalls != 2 || features.TokenEstimate == 0 {
		t.Fatalf("unexpected features: %#v", features)
	}
}

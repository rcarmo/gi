package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReadsWorkspacePiAndPiclawConfig(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".piclaw"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".piclaw", "config.json"), []byte(`{"assistant":{"assistantName":"Neo","assistantAvatar":"a.png"},"user":{"userName":"Rui","userAvatar":"u.png"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"defaultProvider":"ollama","defaultModel":"gemma4:latest","defaultThinkingLevel":"medium","enabledModels":["gemma4:latest"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if cfg.AssistantName != "Neo" || cfg.UserName != "Rui" || cfg.DefaultModel != "gemma4:latest" || cfg.DefaultProvider != "ollama" || cfg.DefaultThinkingLevel != "medium" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
	if len(cfg.EnabledModels) != 1 || cfg.EnabledModels[0] != "gemma4:latest" {
		t.Fatalf("unexpected enabled models: %#v", cfg.EnabledModels)
	}
	if cfg.SystemPrompt != "" {
		t.Fatalf("default prompt is built per turn, not preset: %q", cfg.SystemPrompt)
	}
}

func TestPersistModelSelectionUpdatesPiSettings(t *testing.T) {
	root := t.TempDir()
	if err := PersistModelSelection(root, "ollama", "gemma4:latest", "high", []string{"qwen3:latest"}); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if cfg.DefaultProvider != "ollama" || cfg.DefaultModel != "gemma4:latest" || cfg.DefaultThinkingLevel != "high" {
		t.Fatalf("unexpected persisted config: %#v", cfg)
	}
	if len(cfg.EnabledModels) != 2 || cfg.EnabledModels[1] != "gemma4:latest" {
		t.Fatalf("unexpected enabled models after persist: %#v", cfg.EnabledModels)
	}
}

func TestLoadFallsBackToGiDefaultsWhenNoPiSettingsExist(t *testing.T) {
	root := t.TempDir()
	cfg := Load(root)
	if cfg.DefaultProvider != "opencode-zen" || cfg.DefaultModel != "opencode-zen/minimax-m2.5-free" || cfg.DefaultThinkingLevel != "medium" {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
	if len(cfg.EnabledModels) != 1 || cfg.EnabledModels[0] != "opencode-zen/minimax-m2.5-free" {
		t.Fatalf("unexpected enabled-model defaults: %#v", cfg.EnabledModels)
	}
	if !cfg.InboundWork.Enabled || cfg.InboundWork.IntervalMS != 500 || cfg.InboundWork.BatchSize != 8 || cfg.InboundWork.WorkerID != "web-runtime" || cfg.InboundWork.LeaseTTLMS != 2000 {
		t.Fatalf("unexpected inbound-work defaults: %#v", cfg.InboundWork)
	}
}

func TestLoadPreservesExplicitInboundWorkDisable(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"inboundWork":{"enabled":false,"interval_ms":25,"batch_size":2,"worker_id":"test-worker","lease_ttl_ms":750}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if cfg.InboundWork.Enabled {
		t.Fatalf("expected inbound work dispatcher to remain disabled, got %#v", cfg.InboundWork)
	}
	if cfg.InboundWork.IntervalMS != 25 || cfg.InboundWork.BatchSize != 2 || cfg.InboundWork.WorkerID != "test-worker" || cfg.InboundWork.LeaseTTLMS != 750 {
		t.Fatalf("unexpected explicit inbound-work config: %#v", cfg.InboundWork)
	}
}

func TestPersistClipboardModeUpdatesPiSettings(t *testing.T) {
	root := t.TempDir()
	if err := PersistClipboardMode(root, "osc52"); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if cfg.TUIClipboardMode != "osc52" {
		t.Fatalf("unexpected clipboard mode: %#v", cfg)
	}
	if err := PersistClipboardMode(root, "bogus"); err != nil {
		t.Fatal(err)
	}
	if cfg := Load(root); cfg.TUIClipboardMode != "off" {
		t.Fatalf("unexpected normalized clipboard mode: %#v", cfg)
	}
}

func TestPersistScrollbackLimitUpdatesPiSettings(t *testing.T) {
	root := t.TempDir()
	if err := PersistScrollbackLimit(root, 250); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if cfg.ScrollbackLimit != 250 {
		t.Fatalf("unexpected scrollback limit: %#v", cfg)
	}
}

func TestLoadUsesCurrentWorkingDirectoryWhenWorkspaceRootEmpty(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"defaultProvider":"cwd-provider","defaultModel":"cwd-model","enabledModels":["cwd-model"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(cwd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	// os.Getwd may return the physical path when the project scratch root
	// is reached through the operator's /workspace mount alias.
	wantRoot, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cfg := Load("")
	if cfg.WorkspaceRoot != wantRoot {
		t.Fatalf("unexpected workspace root: %q, want cwd %q", cfg.WorkspaceRoot, wantRoot)
	}
	if cfg.DefaultProvider != "cwd-provider" || cfg.DefaultModel != "cwd-model" {
		t.Fatalf("unexpected cwd-loaded config: %#v", cfg)
	}
}

func TestLoadKeepsAgentsInstructionsForProjectContext(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("Project rule: keep APIs stable."), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	last := cfg.ContextFiles[len(cfg.ContextFiles)-1]
	if last.Content != "Project rule: keep APIs stable." || last.Path != filepath.Join(root, "AGENTS.md") {
		t.Fatalf("project instructions: %+v", cfg.ContextFiles)
	}
}

func TestLoadWorkspaceIndexSettingsAreStartupOnlyAndPreserved(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".pi/settings.json")
	raw := []byte(`{"defaultModel":"test","workspaceIndex":{"extraRoots":["docs"],"extraExtensions":["nim"],"optionalRoots":["notes",".pi/skills"]}}`)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if strings.Join(cfg.WorkspaceIndex.ExtraRoots, ",") != "docs" || strings.Join(cfg.WorkspaceIndex.ExtraExtensions, ",") != "nim" || len(cfg.WorkspaceIndex.OptionalRoots) != 2 {
		t.Fatal(cfg.WorkspaceIndex)
	}
	if err := PersistModelSelection(root, "test", "test", "low", []string{"test"}); err != nil {
		t.Fatal(err)
	}
	if next := Load(root); strings.Join(next.WorkspaceIndex.OptionalRoots, ",") != "notes,.pi/skills" {
		t.Fatal("model save lost index settings", next.WorkspaceIndex)
	}
	if err := os.WriteFile(path, []byte(`{"workspaceIndex":{"extraRoots":["other"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if cfg.WorkspaceIndex.ExtraRoots[0] != "docs" {
		t.Fatal("existing config mutated")
	}
	if next := Load(root); next.WorkspaceIndex.ExtraRoots[0] != "other" || len(next.WorkspaceIndex.OptionalRoots) != 0 {
		t.Fatal(next.WorkspaceIndex)
	}
	if cfg := Load(t.TempDir()); len(cfg.WorkspaceIndex.ExtraRoots) != 0 || len(cfg.WorkspaceIndex.OptionalRoots) != 0 {
		t.Fatal("non-strict defaults")
	}
}

func TestLoadPasskeyRelyingPartyConfiguration(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".pi"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".pi", "settings.json"), []byte(`{"passkeys":{"rp_id":"gi.example.com","origins":["https://gi.example.com"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := Load(dir)
	if cfg.Passkeys.RPID != "gi.example.com" || len(cfg.Passkeys.Origins) != 1 || cfg.Passkeys.Origins[0] != "https://gi.example.com" {
		t.Fatalf("passkeys %+v", cfg.Passkeys)
	}
	if empty := Load(t.TempDir()).Passkeys; empty.RPID != "" || len(empty.Origins) != 0 {
		t.Fatal("passkeys enabled by default")
	}
}

func TestLoadClipboardModePreservesUnsetAndExplicitPolicy(t *testing.T) {
	for _, tc := range []struct{ value, want string }{
		{"", ""}, {`"tuiClipboardMode":""`, ""},
		{`"tuiClipboardMode":"off"`, "off"},
		{`"tuiClipboardMode":" OSC52 "`, "osc52"},
		{`"tuiClipboardMode":"native"`, "native"},
		{`"tuiClipboardMode":"auto"`, "auto"},
		{`"tuiClipboardMode":"bogus"`, "off"},
	} {
		t.Run(tc.value, func(t *testing.T) {
			root := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, ".pi"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte("{"+tc.value+"}"), 0o644); err != nil {
				t.Fatal(err)
			}
			if got := Load(root).TUIClipboardMode; got != tc.want {
				t.Fatalf("mode = %q, want %q", got, tc.want)
			}
		})
	}
	if got := Load(t.TempDir()).TUIClipboardMode; got != "" {
		t.Fatalf("fresh workspace mode = %q", got)
	}
}

func TestLoadMergesGlobalPiSettingsUnderProject(t *testing.T) {
	agent := t.TempDir()
	t.Setenv("PI_CODING_AGENT_DIR", agent)
	if err := os.WriteFile(filepath.Join(agent, "settings.json"), []byte(`{"defaultProvider":"github-copilot","defaultModel":"gpt-5.4","defaultThinkingLevel":"high"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Load(t.TempDir())
	if cfg.DefaultProvider != "github-copilot" || cfg.DefaultModel != "gpt-5.4" || cfg.DefaultThinkingLevel != "high" {
		t.Fatalf("global settings not applied: %q %q %q", cfg.DefaultProvider, cfg.DefaultModel, cfg.DefaultThinkingLevel)
	}
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".pi"), 0o755)
	os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"defaultProvider":"test","defaultModel":"test-model"}`), 0o644)
	cfg = Load(root)
	if cfg.DefaultProvider != "test" || cfg.DefaultModel != "test-model" || cfg.DefaultThinkingLevel != "high" {
		t.Fatalf("project must override global: %q %q %q", cfg.DefaultProvider, cfg.DefaultModel, cfg.DefaultThinkingLevel)
	}
}

// Pi's tuiMode: project settings over global ones.
func TestLoadTUIModeSetting(t *testing.T) {
	agent := t.TempDir()
	t.Setenv("PI_CODING_AGENT_DIR", agent)
	if err := os.WriteFile(filepath.Join(agent, "settings.json"), []byte(`{"tuiMode":"regular"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if cfg := Load(t.TempDir()); cfg.TUIMode != "regular" {
		t.Fatalf("global tuiMode: %q", cfg.TUIMode)
	}
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".pi"), 0o755)
	os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"tuiMode":"fullscreen"}`), 0o644)
	if cfg := Load(root); cfg.TUIMode != "fullscreen" {
		t.Fatalf("project tuiMode: %q", cfg.TUIMode)
	}
}

func TestQuietStartupSetting(t *testing.T) {
	for raw, want := range map[string]string{"true": "true", `"header"`: "header", "false": "", "": "", `"x"`: ""} {
		if got := parseQuietStartup([]byte(raw)); got != want {
			t.Fatalf("%s: %q, want %q", raw, got, want)
		}
	}
}

// modelCatalogUrl is read from the user settings only: a project cannot
// choose where model definitions come from.
func TestLoadModelCatalogURLFromUserSettingsOnly(t *testing.T) {
	user, ws := t.TempDir(), t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", t.TempDir())
	t.Setenv("PI_CODING_AGENT_DIR", user)
	if err := os.MkdirAll(filepath.Join(ws, ".pi"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, ".pi", "settings.json"), []byte(`{"modelCatalogUrl":"https://project.example"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := Load(ws).ModelCatalogURL; got != "" {
		t.Fatalf("project setting used: %q", got)
	}
	if err := os.WriteFile(filepath.Join(user, "settings.json"), []byte(`{"modelCatalogUrl":" https://pi.dev "}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := Load(ws).ModelCatalogURL; got != "https://pi.dev" {
		t.Fatalf("user setting: %q", got)
	}
}

// Pi's /tree settings, treeFilterMode and branchSummary: project settings
// over global ones, field by field.
func TestLoadTreeSettings(t *testing.T) {
	agent := t.TempDir()
	t.Setenv("PI_CODING_AGENT_DIR", agent)
	if err := os.WriteFile(filepath.Join(agent, "settings.json"), []byte(`{"treeFilterMode":"user-only","branchSummary":{"reserveTokens":9000,"skipPrompt":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if cfg := Load(t.TempDir()); cfg.TreeFilterMode != "user-only" || cfg.BranchSummaryReserveTokens != 9000 || !cfg.BranchSummarySkipPrompt {
		t.Fatalf("global: %q %d %v", cfg.TreeFilterMode, cfg.BranchSummaryReserveTokens, cfg.BranchSummarySkipPrompt)
	}
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".pi"), 0o755)
	os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"treeFilterMode":"all","branchSummary":{"skipPrompt":false}}`), 0o644)
	if cfg := Load(root); cfg.TreeFilterMode != "all" || cfg.BranchSummaryReserveTokens != 9000 || cfg.BranchSummarySkipPrompt {
		t.Fatalf("project: %q %d %v", cfg.TreeFilterMode, cfg.BranchSummaryReserveTokens, cfg.BranchSummarySkipPrompt)
	}
}

func TestJavaScriptRuntimeProjectOverridesGlobal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(home, "agent"))
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(home, "agent"))
	if err := os.MkdirAll(filepath.Join(home, "agent"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "agent", "settings.json"), []byte(`{"javascriptRuntime":"quickjs"}`), 0600); err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	if cfg := Load(project); cfg.JavaScriptRuntime != "quickjs" {
		t.Fatal(cfg.JavaScriptRuntime)
	}
	if err := os.Mkdir(filepath.Join(project, ".gi"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".gi", "settings.json"), []byte(`{"javascriptRuntime":"goja"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if cfg := Load(project); cfg.JavaScriptRuntime != "goja" {
		t.Fatal(cfg.JavaScriptRuntime)
	}
}

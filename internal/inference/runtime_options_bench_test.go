package inference

import (
	"fmt"
	goai "github.com/rcarmo/go-ai"
	"sync/atomic"
	"testing"
)

func BenchmarkRuntimeOptions(b *testing.B) {
	dir := b.TempDir()
	b.Setenv("GI_CODING_AGENT_DIR", dir)
	b.Setenv("PI_CODING_AGENT_DIR", dir)
	Init()
	if err := saveAuthEntry("openai", map[string]any{"type": "api_key", "key": "fixture"}); err != nil {
		b.Fatal(err)
	}
	enabled := []string{"openai/gpt-4o", "openai/gpt-4o-mini", "openai/gpt-4.1", "openai/gpt-4.1-mini", "openai/gpt-5"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, m := ListRuntimeOptions("openai", "gpt-4o", enabled)
		if len(m) == 0 {
			b.Fatal("missing models")
		}
	}
}

var runtimeOptionsFixtureSeq atomic.Uint64

func TestRuntimeOptionsObserveRegistryAndCredentialChanges(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", dir)
	t.Setenv("PI_CODING_AGENT_DIR", dir)
	Init()
	provider := fmt.Sprintf("options-fixture-%d", runtimeOptionsFixtureSeq.Add(1))
	id := "model-one"
	if err := SaveAPIKeyLogin(provider, "fixture-key"); err != nil {
		t.Fatal(err)
	}
	goai.RegisterModel(&goai.Model{Provider: goai.Provider(provider), ID: id, Name: "First", ContextWindow: 1000})
	get := func() []ModelOption { _, m := ListRuntimeOptions(provider, id, []string{id, id}); return m }
	a := get()
	if len(a) != 1 || a[0].Name != "First" || !a[0].Authenticated {
		t.Fatal(a)
	}
	goai.RegisterModel(&goai.Model{Provider: goai.Provider(provider), ID: "model-two", Name: "Second", ContextWindow: 2000})
	a = get()
	if len(a) != 2 {
		t.Fatal("registry change concealed", a)
	}
	if _, err := RemoveAuthEntry(provider); err != nil {
		t.Fatal(err)
	}
	a = get()
	if len(a) != 1 || a[0].Authenticated {
		t.Fatal("credential removal concealed", a)
	}
}

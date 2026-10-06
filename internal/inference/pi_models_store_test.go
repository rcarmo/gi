package inference

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

// gi lists the same Copilot models as Pi: Pi's refreshed models-store adds
// models the compiled-in catalogue lacks, and the account's availableModelIds
// (recorded by Pi in auth.json) filter the listing.
var copilotStoreFixtureSeq atomic.Uint64

func TestListRuntimeOptionsMatchesPiCopilotCatalogue(t *testing.T) {
	Init()
	root := t.TempDir()
	t.Setenv("HOME", root)
	agent := filepath.Join(root, ".pi", "agent")
	if err := os.MkdirAll(agent, 0o755); err != nil {
		t.Fatal(err)
	}
	newID := fmt.Sprintf("gi-test-copilot-model-%d", copilotStoreFixtureSeq.Add(1))
	store := `{"github-copilot":{"models":[{"id":"` + newID + `","name":"Test Model 9","api":"openai-responses","provider":"github-copilot",
		"baseUrl":"https://api.individual.githubcopilot.com","reasoning":true,"input":["text"],"contextWindow":400000,"maxTokens":64000,
		"compat":{"supportsMidConvoSystemMessages":true}}]}}`
	if err := os.WriteFile(filepath.Join(agent, "models-store.json"), []byte(store), 0o600); err != nil {
		t.Fatal(err)
	}
	if added := registerPiModelsStore(PiModelsStorePath()); added != 1 {
		t.Fatalf("registered %d store models, want 1", added)
	}
	if again := registerPiModelsStore(PiModelsStorePath()); again != 0 {
		t.Fatalf("re-registered %d known models", again)
	}
	registered := goai.GetModel("github-copilot", newID)
	if registered == nil || registered.ContextWindow != 400000 || registered.ResponsesCompat == nil || registered.ResponsesCompat.SupportsMidConvoSystemMessages == nil || !*registered.ResponsesCompat.SupportsMidConvoSystemMessages {
		t.Fatalf("store model not decoded: %#v", registered)
	}

	auth := `{"github-copilot":{"type":"oauth","refresh":"r","access":"a","expires":9999999999999,
		"availableModelIds":["gpt-5-mini","` + newID + `","not-in-any-catalogue"]}}`
	if err := os.WriteFile(filepath.Join(agent, "auth.json"), []byte(auth), 0o600); err != nil {
		t.Fatal(err)
	}
	_, options := ListRuntimeOptions("github-copilot", "", nil)
	var copilot []string
	for _, o := range options {
		if o.Provider == "github-copilot" {
			copilot = append(copilot, o.ID)
		}
	}
	sort.Strings(copilot)
	if strings.Join(copilot, ",") != newID+",gpt-5-mini" {
		t.Fatalf("copilot models = %v, want only the account's available catalogue models", copilot)
	}
	for _, o := range options {
		if o.ID == newID && !UsableSessionModel(o) {
			t.Fatalf("store model listed but not selectable: %#v", o)
		}
	}
}

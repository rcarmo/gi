// Isolated terminal acceptance catalogue. No external provider is contacted.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tui"
	goai "github.com/rcarmo/go-ai"
)

func main() {
	db := flag.String("db", "", "isolated database")
	workspace := flag.String("workspace", "", "isolated workspace")
	mode := flag.String("tui-mode", "fullscreen", "terminal mode")
	flag.Parse()
	if *db == "" || *workspace == "" {
		log.Fatal("db and workspace required")
	}
	if err := os.Setenv("HOME", *workspace); err != nil {
		log.Fatal(err)
	}
	inference.Init()
	for _, entry := range []struct {
		id, name  string
		window    int
		reasoning bool
	}{
		{"picker-small", "Forest Small", 80, false},
		{"picker-pine", "Forest Pine", 32000, true},
		{"picker-piper", "Forest Piper", 32000, false},
		{"picker-oak1", "Forest Oak One", 200, false},
		{"picker-oak2", "Forest Oak Two", 200, false},
		{"picker-oak3", "Forest Oak Three", 200, false},
		{"picker-oak4", "Forest Oak Four", 200, false},
		{"picker-oak5", "Forest Oak Five", 200, false},
		{"picker-oak6", "Forest Oak Six", 200, false},
	} {
		goai.RegisterModel(&goai.Model{ID: entry.id, Name: entry.name, Provider: "opencode-zen", Api: goai.ApiOpenAICompletions, ContextWindow: entry.window, Reasoning: entry.reasoning})
	}
	s, err := store.Open(*db)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if _, err = s.GetSession(ctx, "picker-main"); err != nil {
		allocation := session.AllocateDefaultSession("agent", "gi", "default", "picker-main")
		if _, _, err = s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "picker-main", Title: "Picker", State: map[string]any{"model": "test-model", "provider": "test", "status": "idle"}, Allocation: allocation}); err != nil {
			log.Fatal(err)
		}
		if _, err = s.CreateTurn(ctx, "measurement-fixture", "picker-main", "terminal fit fixture", nil); err != nil {
			log.Fatal(err)
		}
		if err = s.AppendTurnEvent(ctx, "measurement-fixture", "picker-main", "context.measured", map[string]any{"tokens": 100, "input": 100, "model": "test/test-model"}); err != nil {
			log.Fatal(err)
		}
		if err = s.UpdateTurnStatus(ctx, "measurement-fixture", "completed"); err != nil {
			log.Fatal(err)
		}
		for i := 0; i < 12; i++ {
			if err = s.AddMessage(ctx, store.NowID("msg"), "picker-main", "assistant", fmt.Sprintf("Picker history %02d", i), map[string]any{"kind": "chat"}); err != nil {
				log.Fatal(err)
			}
		}
		if _, err = s.CreateSession(ctx, "picker-other", "Other", map[string]any{"model": "test-model", "provider": "test", "status": "idle"}); err != nil {
			log.Fatal(err)
		}
	}
	if err = s.Close(); err != nil {
		log.Fatal(err)
	}
	if err = tui.RunMode(*db, *workspace, "test-model", *mode); err != nil {
		log.Fatal(err)
	}
}

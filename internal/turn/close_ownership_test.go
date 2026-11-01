package turn

import (
	"context"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	goai "github.com/rcarmo/go-ai"
)

func TestCloseDoesNotAbortAnotherEngineSharedSessionHost(t *testing.T) {
	s := openChildJoinStore(t)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "shared", "shared", nil); err != nil {
		t.Fatal(err)
	}
	first, owner := New(s), New(s)
	defer first.Close()
	defer owner.Close()
	// A previous lookup registers the shared coordinator with this engine, but
	// does not transfer ownership of hosts subsequently launched by another one.
	first.runner("shared")
	started, cancelled := make(chan struct{}), make(chan struct{})
	withStreamWithToolsStub(t, func(runCtx context.Context, _ string, _ *goai.Context, _ func(map[string]any)) (*inference.StreamResult, error) {
		close(started)
		<-runCtx.Done()
		close(cancelled)
		return nil, runCtx.Err()
	})
	run, err := owner.SubmitPrompt(ctx, RunInput{SessionID: "shared", Prompt: "hold", Model: "mock-stream"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("owner did not start")
	}
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-cancelled:
		t.Fatal("closing a non-owner aborted the shared host")
	default:
	}
	row, err := s.GetTurn(ctx, run.TurnID)
	if err != nil || row.Status != "running" {
		t.Fatal("non-owner changed the turn", row, err)
	}
	if err = owner.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-cancelled:
	default:
		t.Fatal("owner close did not join its host")
	}
}

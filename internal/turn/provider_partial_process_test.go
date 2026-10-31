package turn

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestProviderPartialProcessOwner(t *testing.T) {
	path := os.Getenv("GI_PARTIAL_PROCESS_FIXTURE")
	if path == "" {
		t.Skip("subprocess helper")
	}
	ctx := context.Background()
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	e := New(s)
	streamWithToolsWithHooks = func(_ context.Context, _ string, _ *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
		if err := hooks.OnProgress(testProviderPartial()); err != nil {
			t.Fatal(err)
		}
		var requestID string
		if err := s.DB().QueryRow(`select json_extract(payload_json,'$.request_id') from messages where role='assistant'`).Scan(&requestID); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("PARTIAL_READY %s %s\n", e.LiveActiveTurn("A"), requestID)
		select {}
	}
	if _, err = e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "isolated process partial", Model: "provider/model"}); err != nil {
		t.Fatal(err)
	}
	select {}
}
func TestProviderPartialSurvivesKilledOwnerAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "process.db")
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "-test.run=^TestProviderPartialProcessOwner$", "-test.timeout=15s")
	command.Env = append(os.Environ(), "GI_PARTIAL_PROCESS_FIXTURE="+path, "HOME="+t.TempDir())
	command.Stderr = os.Stderr
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = command.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	t.Cleanup(func() {
		if !waited {
			_ = command.Process.Kill()
			_ = command.Wait()
		}
	})
	scanner := bufio.NewScanner(stdout)
	var fields []string
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "PARTIAL_READY ") {
			fields = strings.Fields(scanner.Text())
			break
		}
	}
	if len(fields) != 3 {
		t.Fatal("owner never committed a checkpoint", fields, scanner.Err())
	}
	if err = command.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err = command.Wait(); err == nil {
		t.Fatal("owner exited normally instead of being killed")
	}
	waited = true
	// Only after Wait establishes the isolated owner's death may the coordinator
	// convert its committed partial. No provider or turn is replayed here.
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	record, err := s.ProviderPartial(context.Background(), fields[1], "A", fields[2])
	if err != nil || record == nil || record.Message.Usage.Input != 17 || len(record.Message.Content) != 3 || record.Content != "streamed answer" {
		t.Fatal("checkpoint lost on process death", record, err)
	}
	sealed, err := s.SealProviderPartial(context.Background(), fields[1], "A", fields[2], "owner process terminated")
	if err != nil || !sealed.Applied || sealed.Accounting.Usage.Input != 17 || sealed.Accounting.Unknown != 1 {
		t.Fatal("partial counters or uncertainty lost", sealed, err)
	}
	snapshot, err := s.ContextSnapshot(context.Background(), "A")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range (&sessionRunner{store: s}).projectContextSnapshot(context.Background(), "A", snapshot) {
		if strings.Contains(goai.GetTextContent(&m), "streamed") || len(goai.GetToolCalls(&m)) != 0 {
			t.Fatal("killed owner's partial became context", m)
		}
	}
}

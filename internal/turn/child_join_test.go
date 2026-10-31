package turn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// Exercise the production WAL/immediate-transaction path. Shared-memory
// SQLite uses table locks and cannot provide this concurrent writer workload.
func openChildJoinStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "children.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}
func terminalEventSequence(t *testing.T, s *store.Store, id string) int64 {
	t.Helper()
	var sequence int64
	// Event seq is per-turn; the insertion identity orders commits across turns.
	if err := s.DB().QueryRowContext(context.Background(), `select id from turn_events where turn_id=? and event_type='turn.finished'`, id).Scan(&sequence); err != nil {
		t.Fatal("missing terminal receipt", id, err)
	}
	return sequence
}
func TestParentTerminalWaitsForRequiredChildHosts(t *testing.T) {
	for _, outcome := range []string{"completed", "failed", "cancelled", "cancel_during_join"} {
		t.Run(outcome, func(t *testing.T) {
			s := openChildJoinStore(t)
			e := New(s)
			ctx := context.Background()
			started := map[string]chan struct{}{}
			cancelled := map[string]chan struct{}{}
			for _, name := range []string{"parent", "noncritical", "critical", "late"} {
				started[name] = make(chan struct{})
				cancelled[name] = make(chan struct{})
				if _, err := s.CreateSession(ctx, name, name, nil); err != nil {
					t.Fatal(err)
				}
			}
			parentRelease, childRelease := make(chan struct{}), make(chan struct{})
			var parentOnce, childOnce sync.Once
			defer func() {
				parentOnce.Do(func() { close(parentRelease) })
				childOnce.Do(func() { close(childRelease) })
				_ = e.Close()
				_ = s.Close()
			}()
			withStreamWithToolsStub(t, func(runCtx context.Context, _ string, conv *goai.Context, _ func(map[string]any)) (*inference.StreamResult, error) {
				name := ""
				for _, message := range conv.Messages {
					if message.Role == goai.RoleUser && len(message.Content) > 0 {
						name = message.Content[0].Text
					}
				}
				signal, ok := started[name]
				if !ok {
					return nil, fmt.Errorf("unexpected inference request %q", name)
				}
				close(signal)
				if name == "parent" {
					select {
					case <-parentRelease:
					case <-runCtx.Done():
						return nil, runCtx.Err()
					}
					if outcome == "failed" {
						return nil, fmt.Errorf("provider failure")
					}
					return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "parent reply"}}}}, nil
				}
				<-runCtx.Done()
				close(cancelled[name])
				<-childRelease // deliberately ignore cancellation until released
				return nil, runCtx.Err()
			})
			parent, err := e.SubmitPrompt(ctx, RunInput{SessionID: "parent", Prompt: "parent", Model: "mock-stream"})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-started["parent"]:
			case <-time.After(time.Second):
				t.Fatal("parent did not start")
			}
			children := map[string]*SubmitResult{}
			for _, name := range []string{"noncritical", "critical"} {
				child, err := e.SubmitPrompt(ctx, RunInput{SessionID: name, Prompt: name, Model: "mock-stream", ParentTurnID: parent.TurnID, Metadata: map[string]any{"subturn_critical": name == "critical"}})
				if err != nil {
					t.Fatal(err)
				}
				children[name] = child
				select {
				case <-started[name]:
				case <-time.After(time.Second):
					t.Fatal("child did not start", name)
				}
			}
			if outcome == "cancelled" {
				if err := e.CancelTurn(ctx, "parent", parent.TurnID); err != nil {
					t.Fatal(err)
				}
			} else {
				parentOnce.Do(func() { close(parentRelease) })
			}
			select {
			case <-cancelled["noncritical"]:
			case <-time.After(time.Second):
				t.Fatal("required child not cancelled")
			}
			pending, err := s.GetTurn(ctx, parent.TurnID)
			if err != nil || isTerminalSubTurnStatus(pending.Status) || pending.Phase != "settling" || pending.FinishedAt != "" {
				t.Fatal("parent published terminal before child join", pending, err)
			}
			active, _, err := s.GetSessionActiveTurn(ctx, "parent")
			if err != nil || active != parent.TurnID {
				t.Fatal("parent claim released before child", active, err)
			}
			events, err := s.ListTurnEvents(ctx, parent.TurnID)
			if err != nil {
				t.Fatal(err)
			}
			var staged bool
			for _, event := range events {
				if event.Type == "turn.finished" {
					t.Fatal("parent terminal receipt preceded child")
				}
				staged = staged || event.Type == "turn.terminal_staged"
			}
			if !staged {
				t.Fatal("missing staged decision")
			}
			if _, err := e.SubmitPrompt(ctx, RunInput{SessionID: "late", Prompt: "late", Model: "mock-stream", ParentTurnID: parent.TurnID}); err == nil {
				t.Fatal("new child admitted after parent decision")
			}
			if queued, err := s.CountQueuedTurns(ctx, "late"); err != nil || queued != 0 {
				t.Fatal("rejected child stranded queued work", queued, err)
			}
			if _, _, err := s.GetSessionActiveTurn(ctx, "late"); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal("rejected child acquired claim", err)
			}
			if outcome == "cancel_during_join" {
				select {
				case <-cancelled["critical"]:
					t.Fatal("graceful decision cancelled critical child")
				default:
				}
				if err := e.CancelTurn(ctx, "parent", parent.TurnID); err != nil {
					t.Fatal(err)
				}
			}
			cancelAll := outcome == "cancelled" || outcome == "cancel_during_join"
			if cancelAll {
				select {
				case <-cancelled["critical"]:
				case <-time.After(time.Second):
					t.Fatal("changed parent decision did not cancel critical child")
				}
			} else {
				select {
				case <-cancelled["critical"]:
					t.Fatal("critical child lost independent lifetime")
				default:
				}
			}
			childOnce.Do(func() { close(childRelease) })
			expected := outcome
			if cancelAll {
				expected = "cancelled"
			}
			waitForCondition(t, time.Second, func() bool {
				turn, err := s.GetTurn(ctx, parent.TurnID)
				return err == nil && turn.Status == expected && turn.FinishedAt != ""
			}, "parent finishes after child settlement")
			parentSeq := terminalEventSequence(t, s, parent.TurnID)
			for name, child := range children {
				if name == "critical" && !cancelAll {
					continue
				}
				if childSeq := terminalEventSequence(t, s, child.TurnID); childSeq >= parentSeq {
					t.Fatal("child receipt followed parent", name, childSeq, parentSeq)
				}
				if _, _, err := s.GetSessionActiveTurn(ctx, name); !errors.Is(err, sql.ErrNoRows) {
					t.Fatal("parent passed child cleanup", name, err)
				}
			}
		})
	}
}
func TestFailedTerminalStageRetainsClaimWithoutLiveHost(t *testing.T) {
	s := openChildJoinStore(t)
	e := New(s)
	ctx := context.Background()
	defer s.Close()
	defer e.Close()
	if _, err := s.CreateSession(ctx, "parent", "parent", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`create trigger reject_stage before insert on turn_events when new.event_type='turn.terminal_staged' begin select raise(abort,'reject stage');end`); err != nil {
		t.Fatal(err)
	}
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "committed reply"}}}}, nil
	})
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "parent", Prompt: "parent", Model: "mock-stream"})
	if err != nil {
		t.Fatal(err)
	}
	joined := make(chan struct{})
	go func() { e.runs.Wait(); close(joined) }()
	select {
	case <-joined:
	case <-time.After(time.Second):
		t.Fatal("failed settlement host escaped join")
	}
	active, _, err := s.GetSessionActiveTurn(ctx, "parent")
	if err != nil || active != run.TurnID {
		t.Fatal("failed settlement released durable ownership", active, err)
	}
	turn, err := s.GetTurn(ctx, run.TurnID)
	if err != nil || turn.Status != "running" || turn.FinishedAt != "" {
		t.Fatal("failed stage claimed terminal", turn, err)
	}
	if terminal, err := s.HasTurnTerminalReceipt(ctx, run.TurnID); err != nil || terminal {
		t.Fatal("failed stage has terminal receipt", terminal, err)
	}
	if e.LiveActiveTurn("parent") != "" {
		t.Fatal("failed stage retained nonexistent live worker")
	}
}

func TestCancelCannotReopenTerminalDuringObserverJoin(t *testing.T) {
	s := openChildJoinStore(t)
	e := New(s)
	ctx := context.Background()
	release := make(chan struct{})
	observed := make(chan struct{})
	var once sync.Once
	defer func() { once.Do(func() { close(release) }); _ = e.Close(); _ = s.Close() }()
	if _, err := s.CreateSession(ctx, "parent", "parent", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := e.RegisterHook(HookTurnState, "terminal-observer", func(_ context.Context, req HookRequest) (HookResponse, error) {
		if req.TurnStatus == "completed" {
			close(observed)
			<-release
		}
		return HookResponse{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "reply"}}}}, nil
	})
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "parent", Prompt: "parent", Model: "mock-stream"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-observed:
	case <-time.After(time.Second):
		t.Fatal("terminal observer not entered")
	}
	if err := e.CancelTurn(ctx, "parent", run.TurnID); !errors.Is(err, store.ErrQueueConflict) {
		t.Fatal("late cancel reopened terminal", err)
	}
	turn, err := s.GetTurn(ctx, run.TurnID)
	if err != nil || turn.Status != "completed" || turn.FinishedAt == "" {
		t.Fatal("terminal mutated during observer join", turn, err)
	}
	once.Do(func() { close(release) })
}

func TestUnownedChildWithoutReceiptKeepsParentCheckpoint(t *testing.T) {
	s := openChildJoinStore(t)
	e := New(s)
	ctx := context.Background()
	release := make(chan struct{})
	var once sync.Once
	defer func() { once.Do(func() { close(release) }); _ = e.Close(); _ = s.Close() }()
	for _, id := range []string{"parent", "foreign"} {
		if _, err := s.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		<-release
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "reply"}}}}, nil
	})
	parent, err := e.SubmitPrompt(ctx, RunInput{SessionID: "parent", Prompt: "parent", Model: "mock-stream"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "foreign-child", "foreign", "running", "foreign", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSubTurn(ctx, parent.TurnID, "parent", "foreign-child", "foreign", "async", 1, nil); err != nil {
		t.Fatal(err)
	}
	token, err := store.NewActiveTurnClaimToken()
	if err != nil {
		t.Fatal(err)
	}
	if claimed, err := s.ClaimSessionActiveTurn(ctx, "foreign", "foreign-child", "other-process", token); err != nil || !claimed {
		t.Fatal(claimed, err)
	}
	once.Do(func() { close(release) })
	joined := make(chan struct{})
	go func() { e.runs.Wait(); close(joined) }()
	select {
	case <-joined:
	case <-time.After(time.Second):
		t.Fatal("unknown child left a hidden parent host")
	}
	parentRow, err := s.GetTurn(ctx, parent.TurnID)
	if err != nil || parentRow.Status != "running" || parentRow.Phase != "settling" || parentRow.FinishedAt != "" {
		t.Fatal("unknown child pronounced settled", parentRow, err)
	}
	if active, _, err := s.GetSessionActiveTurn(ctx, "parent"); err != nil || active != parent.TurnID {
		t.Fatal("parent ownership released", active, err)
	}
	child, err := s.GetTurn(ctx, "foreign-child")
	if err != nil || child.Status != "running" || child.FinishedAt != "" {
		t.Fatal("foreign child forcibly settled", child, err)
	}
	if active, _, err := s.GetSessionActiveTurn(ctx, "foreign"); err != nil || active != "foreign-child" {
		t.Fatal("foreign ownership released", active, err)
	}
}
func TestChildJoinRejectsOwnershipCycle(t *testing.T) {
	s := openChildJoinStore(t)
	e := New(s)
	defer s.Close()
	defer e.Close()
	ctx := context.Background()
	for _, id := range []string{"a", "b"} {
		if _, err := s.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := s.CreateTurnWithStatus(ctx, id, id, "running", id, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreateSubTurn(ctx, "a", "a", "b", "b", "async", 1, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSubTurn(ctx, "b", "b", "a", "a", "async", 2, nil); err != nil {
		t.Fatal(err)
	}
	if err := e.runner("a").joinRequiredChildren(ctx, "a", "cancelled", ""); err == nil {
		t.Fatal("ownership cycle admitted")
	}
}

func TestParentJoinsQueuedSameSessionChildWithoutLaunchingIt(t *testing.T) {
	s := openChildJoinStore(t)
	e := New(s)
	ctx := context.Background()
	release := make(chan struct{})
	started := make(chan struct{})
	var once sync.Once
	defer func() { once.Do(func() { close(release) }); _ = e.Close(); _ = s.Close() }()
	if _, err := s.CreateSession(ctx, "parent", "parent", nil); err != nil {
		t.Fatal(err)
	}
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case <-started:
			return nil, fmt.Errorf("unexpected queued child inference")
		default:
			close(started)
		}
		<-release
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "reply"}}}}, nil
	})
	parent, err := e.SubmitPrompt(ctx, RunInput{SessionID: "parent", Prompt: "parent", Model: "mock-stream"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("parent did not start")
	}
	child, err := e.SubmitPrompt(ctx, RunInput{SessionID: "parent", Prompt: "queued child", Intent: "queue", Model: "mock-stream", ParentTurnID: parent.TurnID})
	if err != nil {
		t.Fatal(err)
	}
	if !child.Queued {
		t.Fatal("same-session child not queued", child)
	}
	once.Do(func() { close(release) })
	waitForCondition(t, time.Second, func() bool {
		turn, err := s.GetTurn(ctx, parent.TurnID)
		return err == nil && turn.Status == "completed"
	}, "parent completes after queued child cancellation")
	receipt, err := s.HasTurnTerminalReceipt(ctx, child.TurnID)
	if err != nil || !receipt {
		t.Fatal("queued child lacks terminal receipt", receipt, err)
	}
	row, err := s.GetTurn(ctx, child.TurnID)
	if err != nil || row.Status != "cancelled" || row.StartedAt != "" {
		t.Fatal("queued child executed", row, err)
	}
	var childEnd, parentEnd int64
	if err := s.DB().QueryRow(`select id from turn_events where turn_id=? and event_type='turn.cancelled'`, child.TurnID).Scan(&childEnd); err != nil {
		t.Fatal(err)
	}
	parentEnd = terminalEventSequence(t, s, parent.TurnID)
	if childEnd >= parentEnd {
		t.Fatal("queued child settled after parent", childEnd, parentEnd)
	}
}

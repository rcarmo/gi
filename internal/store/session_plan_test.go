package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"github.com/rcarmo/gi/internal/plan"
)

func planText(v string) *string { return &v }
func TestSessionPlanAtomicMutationPersistenceAndIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plan.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, id := range []string{"a", "b"} {
		db.CreateSession(t.Context(), id, id, nil)
	}
	initial, err := db.SessionPlan(t.Context(), "a")
	if err != nil || initial.UpdatedAt != nil || initial.Markdown != plan.DefaultMarkdown {
		t.Fatal(initial, err)
	}
	written, err := db.MutateSessionPlan(t.Context(), "a", plan.Mutation{Action: "write", Markdown: planText("## Heading\n- [ ] alpha\n- [-] beta\n- [x] done")})
	if err != nil || written.UpdatedAt == nil || len(written.Plan) != 3 {
		t.Fatal(written, err)
	}
	other, err := db.SessionPlan(t.Context(), "b")
	if err != nil || other.Markdown != plan.DefaultMarkdown {
		t.Fatal("cross-session write", other, err)
	}
	_, err = db.MutateSessionPlan(t.Context(), "a", plan.Mutation{Action: "patch", Patches: []plan.Patch{{Operation: "update", Match: "beta", Status: planText("completed")}, {Operation: "remove", Match: "missing"}}})
	if !errors.Is(err, plan.ErrInvalid) {
		t.Fatal(err)
	}
	unchanged, _ := db.SessionPlan(t.Context(), "a")
	if unchanged.Markdown != written.Markdown || *unchanged.UpdatedAt != *written.UpdatedAt {
		t.Fatal("partial patch", unchanged)
	}
	db.Close()
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	saved, err := db.SessionPlan(t.Context(), "a")
	if err != nil || saved.Markdown != written.Markdown {
		t.Fatal(saved, err)
	}
	if _, err = db.SessionPlan(t.Context(), "missing"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
	if _, err = db.MutateSessionPlan(t.Context(), "missing", plan.Mutation{Action: "reset"}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
	reset, err := db.MutateSessionPlan(t.Context(), "a", plan.Mutation{Action: "reset"})
	if err != nil || reset.Markdown != plan.DefaultMarkdown {
		t.Fatal(reset, err)
	}
}
func TestSessionPlanWriteFailurePreservesSavedState(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "plan.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.CreateSession(t.Context(), "s", "s", nil)
	db.MutateSessionPlan(t.Context(), "s", plan.Mutation{Action: "write", Markdown: planText("- [ ] before")})
	if _, err = db.DB().Exec(`create trigger reject_plan before update on kv_store when new.namespace='session_plan' begin select raise(abort,'injected plan failure'); end`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.MutateSessionPlan(t.Context(), "s", plan.Mutation{Action: "write", Markdown: planText("- [x] after")}); err == nil {
		t.Fatal("expected write failure")
	}
	saved, err := db.SessionPlan(t.Context(), "s")
	if err != nil || saved.Markdown != "- [ ] before" {
		t.Fatal(saved, err)
	}
}

func TestSessionPlanConditionalRevisionConcurrentBrowserAndAgent(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "conditional.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.CreateSession(t.Context(), "s", "s", nil)
	loaded, err := db.SessionPlan(t.Context(), "s")
	if err != nil || loaded.Revision == "" {
		t.Fatal(loaded, err)
	}
	text := "- [ ] changed"
	if _, err := db.MutateSessionPlanConditional(t.Context(), "s", plan.Mutation{Action: "write", Markdown: &text}, ""); !errors.Is(err, ErrPlanRevisionRequired) {
		t.Fatal(err)
	}
	result := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := db.MutateSessionPlanConditional(t.Context(), "s", plan.Mutation{Action: "write", Markdown: &text}, loaded.Revision)
			result <- err
		}()
	}
	wg.Wait()
	close(result)
	success, conflict := 0, 0
	for err := range result {
		if err == nil {
			success++
		} else if errors.Is(err, ErrPlanRevisionConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatal(success, conflict)
	}
	current, _ := db.SessionPlan(t.Context(), "s")
	agent := "- [x] agent update"
	db.MutateSessionPlan(t.Context(), "s", plan.Mutation{Action: "write", Markdown: &agent})
	if latest, err := db.MutateSessionPlanConditional(t.Context(), "s", plan.Mutation{Action: "reset"}, current.Revision); !errors.Is(err, ErrPlanRevisionConflict) || latest.Markdown != agent {
		t.Fatal(latest, err)
	}
	latest, _ := db.SessionPlan(t.Context(), "s")
	reset, err := db.MutateSessionPlanConditional(t.Context(), "s", plan.Mutation{Action: "reset"}, latest.Revision)
	if err != nil || reset.Revision == latest.Revision || reset.Markdown != plan.DefaultMarkdown {
		t.Fatal(reset, err)
	}
}

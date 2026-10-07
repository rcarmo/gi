package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	gisession "github.com/rcarmo/gi/internal/session"
)

func TestCloneSessionKeepsAgentAndUniqueRoutingIdentity(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "copies.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateDefaultSession("neo", "web", "account-a", "source")
	source, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "source", Title: "Original", State: map[string]any{"model": "fixture", "loaded_tools": []string{"echo"}, "status": "running", "active_turn_id": "source-turn", "queue_count": 2, "archived_at": "yesterday", "pinned": true, treeParentKey: "old-parent", treeLabelsKey: map[string]any{"msg-0": "label"}}, Allocation: alloc})
	if err != nil {
		t.Fatal(err)
	}
	for i, role := range []string{"user", "assistant", "user"} {
		if err := s.AddMessage(ctx, fmt.Sprintf("msg-%d", i), source.ID, role, fmt.Sprintf("text-%d", i), map[string]any{"fixture": i}); err != nil {
			t.Fatal(err)
		}
	}
	copies := make(chan *Session, 8)
	errs := make(chan error, 8)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c, err := s.CloneSessionBefore(ctx, source.ID, fmt.Sprintf("copy-%d", i), "", "", "msg-2")
			if err != nil {
				errs <- err
				return
			}
			copies <- c
		}(i)
	}
	wg.Wait()
	close(copies)
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	originalIdentity, err := s.GetSessionIdentity(ctx, source.ID)
	if err != nil {
		t.Fatal(err)
	}
	aliases := map[string]string{}
	keys := map[string]bool{}
	opaqueKeys := map[string]bool{originalIdentity.OpaqueSessionKey: true}
	copyCount := 0
	for c := range copies {
		copyCount++
		if c.Scope.AgentID != source.Scope.AgentID || c.Scope.Channel != source.Scope.Channel || c.Scope.Account != source.Scope.Account || c.Title != source.Title || c.ParentSessionID != source.ID {
			t.Fatal(c)
		}
		if c.Scope.Values["chat"] == source.Scope.Values["chat"] {
			t.Fatal("copied source routing scope")
		}
		identity, err := s.RequireSessionIdentityRuntime(ctx, c.ID)
		if err != nil {
			t.Fatal(err)
		}
		if keys[identity.CanonicalScopeSignature] {
			t.Fatal("duplicate copy identity")
		}
		keys[identity.CanonicalScopeSignature] = true
		fullIdentity, err := s.GetSessionIdentity(ctx, c.ID)
		if err != nil {
			t.Fatal(err)
		}
		if fullIdentity.IsMainSession || fullIdentity.OpaqueSessionKey == "" || opaqueKeys[fullIdentity.OpaqueSessionKey] || identity.CanonicalScopeSignature == originalIdentity.CanonicalScopeSignature || len(c.Aliases) == 0 {
			t.Fatal("copy reused identity/main flag or has no aliases", fullIdentity)
		}
		opaqueKeys[fullIdentity.OpaqueSessionKey] = true
		resolved, err := s.ResolveSessionIDByKeyOrAlias(ctx, fullIdentity.OpaqueSessionKey)
		if err != nil || resolved != c.ID {
			t.Fatal("opaque key misrouted", resolved, err)
		}
		if c.State["model"] != source.State["model"] || !reflect.DeepEqual(c.State["loaded_tools"], source.State["loaded_tools"]) || c.State["status"] != "idle" || c.State["active_turn_id"] != nil || c.State["queue_count"] != float64(0) {
			t.Fatal("copy lost configuration or retained live work", c.State)
		}
		for _, key := range []string{"archived_at", "pinned", treeParentKey, treeLabelsKey} {
			if _, ok := c.State[key]; ok {
				t.Fatal("copy retained source state", key)
			}
		}
		for _, alias := range c.Aliases {
			if owner := aliases[alias]; owner != "" {
				t.Fatal("shared copy alias", owner, c.ID)
			}
			aliases[alias] = c.ID
			resolved, err := s.ResolveSessionIDByAlias(ctx, alias)
			if err != nil || resolved != c.ID {
				t.Fatal(alias, resolved, err)
			}
		}
		messages, err := s.ListMessages(ctx, c.ID)
		if err != nil || len(messages) != 2 || messages[0].Content != "text-0" || messages[1].Content != "text-1" {
			t.Fatal(messages, err)
		}
		for i, msg := range messages {
			if msg.ID == fmt.Sprintf("msg-%d", i) || msg.Payload["forked_from_message_id"] != fmt.Sprintf("msg-%d", i) || msg.Payload["fixture"] != float64(i) {
				t.Fatal("copy lost provenance or reused message identity", msg)
			}
		}
	}
	if copyCount != 8 {
		t.Fatal("missing concurrent copies", copyCount)
	}
	unchanged, err := s.GetSession(ctx, source.ID)
	if err != nil || !reflect.DeepEqual(unchanged.State, source.State) {
		t.Fatal("copy mutated source state", unchanged, err)
	}
	for _, alias := range source.Aliases {
		resolved, err := s.ResolveSessionIDByAlias(ctx, alias)
		if err != nil || resolved != source.ID {
			t.Fatal("source alias stolen", alias, resolved, err)
		}
	}
	main, err := s.ResolveMainSessionID(ctx, "neo", "web", "account-a")
	if err != nil || main != source.ID {
		t.Fatal("copy stole inbound main", main, err)
	}
	all, err := s.CloneSession(ctx, source.ID, "full", "", " ")
	if err != nil {
		t.Fatal(err)
	}
	messages, err := s.ListMessages(ctx, all.ID)
	if err != nil || len(messages) != 3 || all.Scope.AgentID != "neo" {
		t.Fatal(all, messages)
	}
	if _, err := s.CloneSessionBefore(ctx, source.ID, "invalid", "", "", "unknown"); err == nil {
		t.Fatal("unknown cutoff accepted")
	}
	if _, err := s.GetSession(ctx, "invalid"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("failed fork created session", err)
	}
	for _, tc := range []struct{ id, agent, channel, account string }{
		{"explicit-same", " Neo ", "web", "account-a"},
		{"explicit-peer", "peer7", "gi", "default"},
	} {
		copy, err := s.CloneSessionBefore(ctx, source.ID, tc.id, "Chosen title", tc.agent, "msg-0")
		if err != nil {
			t.Fatal(err)
		}
		if copy.Scope.Channel != tc.channel || copy.Scope.Account != tc.account || copy.Title != "Chosen title" {
			t.Fatal(copy)
		}
		messages, err := s.ListMessages(ctx, copy.ID)
		if err != nil || len(messages) != 0 {
			t.Fatal("first-message fork retained history", messages, err)
		}
	}
}

func TestCloneSessionRollsBackFailedMessageCopy(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "rollback.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := t.Context()
	source, err := s.CreateSession(ctx, "source", "Source", nil)
	if err != nil {
		t.Fatal(err)
	}
	for i, content := range []string{"first", "reject"} {
		if err := s.AddMessage(ctx, fmt.Sprintf("msg-%d", i), source.ID, "user", content, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.DB().ExecContext(ctx, `create trigger reject_copy before insert on messages when new.session_id = 'failed-copy' and new.content = 'reject' begin select raise(abort, 'fixture failure'); end;`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CloneSession(ctx, source.ID, "failed-copy", "", ""); err == nil {
		t.Fatal("failed message accepted")
	}
	if _, err := s.GetSession(ctx, "failed-copy"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("partial session survived", err)
	}
	for _, table := range []string{"messages", "session_identities", "session_identity_dimensions", "session_aliases"} {
		var count int
		if err := s.DB().QueryRowContext(ctx, "select count(*) from "+table+" where session_id = ?", "failed-copy").Scan(&count); err != nil || count != 0 {
			t.Fatal("partial copy survived", table, count, err)
		}
	}
	messages, err := s.ListMessages(ctx, source.ID)
	if err != nil || len(messages) != 2 {
		t.Fatal("source changed", messages, err)
	}
}

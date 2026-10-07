package tui

import (
	"context"
	"strings"
	"testing"
)

// Pi's /fork: pick an earlier user message; the new session holds only the
// history before it and the editor gets the message text back.
func TestForkFromEarlierUserMessage(t *testing.T) {
	c := sessionTestChat(t)
	ctx := context.Background()
	for _, m := range []struct{ id, role, text string }{
		{"m1", "user", "first question"}, {"m2", "assistant", "first answer"},
		{"m3", "user", "second question"}, {"m4", "assistant", "second answer"},
	} {
		if err := c.store.AddMessage(ctx, m.id, "A", m.role, m.text, map[string]any{}); err != nil {
			t.Fatal(err)
		}
	}
	c.handleCommand("/fork")
	if !c.modelMenuOpen || c.modelMenuKind != "fork" || len(c.modelMenuChoices) != 2 || c.modelMenuSelected != 1 {
		t.Fatalf("selector: open=%v kind=%q choices=%v selected=%d", c.modelMenuOpen, c.modelMenuKind, c.modelMenuChoices, c.modelMenuSelected)
	}
	rows := c.piForkSelectorRows(80)
	var screen []string
	for _, row := range rows {
		var line strings.Builder
		for _, span := range row {
			line.WriteString(span.Text)
		}
		screen = append(screen, line.String())
	}
	joined := strings.Join(screen, "\n")
	for _, want := range []string{"Fork from Message", "› second question", "Message 2 of 2", "  first question"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("selector missing %q:\n%s", want, joined)
		}
	}
	c.acceptForkSelection()
	if c.modelMenuOpen || c.sessionID == "A" {
		t.Fatalf("fork did not switch: open=%v session=%s", c.modelMenuOpen, c.sessionID)
	}
	original, err := c.store.GetSession(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	forked, err := c.store.GetSession(ctx, c.sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if forked.Scope.AgentID != original.Scope.AgentID || forked.Scope.Values["chat"] == original.Scope.Values["chat"] {
		t.Fatal("fork changed agent or shared routing identity", forked, original)
	}
	if c.input.Text() != "second question" {
		t.Fatalf("editor = %q, want the selected message", c.input.Text())
	}
	msgs, err := c.store.ListMessages(ctx, c.sessionID)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, m := range msgs {
		got = append(got, m.Content)
	}
	if strings.Join(got, "|") != "first question|first answer" {
		t.Fatalf("forked history = %v, want only messages before the selection", got)
	}
}

func TestForkWithoutUserMessages(t *testing.T) {
	c := sessionTestChat(t)
	c.handleCommand("/fork")
	if c.modelMenuOpen || !strings.Contains(strings.Join(c.transcript, "\n"), "No messages to fork from") {
		t.Fatalf("empty fork: open=%v transcript=%v", c.modelMenuOpen, c.transcript)
	}
}

func TestCloneDefaultKeepsAgentAndSpawnCreatesPeer(t *testing.T) {
	c := sessionTestChat(t)
	source := c.sessionID
	original, err := c.store.GetSession(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.store.AddMessage(context.Background(), "clone-message", source, "user", "kept", nil); err != nil {
		t.Fatal(err)
	}
	lines := c.cloneSessionLines([]string{"/clone"})
	if len(lines) == 0 || strings.HasPrefix(lines[0], "error:") {
		t.Fatal(lines)
	}
	cloned, err := c.store.GetSession(context.Background(), c.sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if cloned.Scope.AgentID != original.Scope.AgentID || cloned.ParentSessionID != source || cloned.Scope.Values["chat"] == original.Scope.Values["chat"] {
		t.Fatal(original, cloned)
	}
	// Agent references keep the existing deterministic session-list order;
	// exact IDs always let the user choose between same-agent copies.
	for _, id := range []string{source, cloned.ID} {
		resolved, err := c.resolveSessionRef(id)
		if err != nil || resolved.ID != id {
			t.Fatal("exact session reference misrouted", resolved, err)
		}
	}
	sessionIDs, agentIDs, err := c.loadSessionIdentityIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	expected, ok := findSessionIDByNormalizedAgentRef(sessionIDs, agentIDs, original.Scope.AgentID)
	resolved, err := c.resolveSessionRef("@" + original.Scope.AgentID)
	if !ok || err != nil || resolved.ID != expected {
		t.Fatal("agent reference changed selection order", resolved, err)
	}
	c.handleCommand("/spawn")
	peer, err := c.store.GetSession(context.Background(), c.sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if peer.Scope.AgentID == original.Scope.AgentID {
		t.Fatal("spawn no longer creates peer", peer)
	}
	messages, err := c.store.ListMessages(context.Background(), peer.ID)
	if err != nil || len(messages) != 2 || messages[0].Content != "kept" || messages[1].Payload["kind"] != "fork" {
		t.Fatal("spawn lost peer context clipping/provenance", messages, err)
	}
}

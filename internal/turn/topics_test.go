package turn

import (
	"context"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
	goai "github.com/rcarmo/go-ai"
)

func TestBroadcastPublishesNormalizedTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "turn.*", topics.SubscribeOptions{Buffer: 4})
	defer unsub()

	engine.broadcast("session_1", map[string]any{"type": "agent_status", "agent_id": "agent1", "status": "running"})

	select {
	case env := <-ch:
		if env.Topic != "turn.status" {
			t.Fatalf("unexpected topic: %#v", env)
		}
		if env.SessionID != "session_1" || env.AgentID != "agent1" {
			t.Fatalf("unexpected scoped envelope: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected topic event")
	}
}

func TestConnectivityBusBridgesIntoTopics(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	defer engine.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "connectivity.*", topics.SubscribeOptions{Buffer: 4})
	defer unsub()

	if err := engine.Connectivity().Emit(context.Background(), "route.http.demo", map[string]any{"ok": true, "session_id": "s1"}); err != nil {
		t.Fatalf("emit connectivity event: %v", err)
	}

	select {
	case env := <-ch:
		if env.Topic != "connectivity.route.http.demo" {
			t.Fatalf("unexpected connectivity topic: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected bridged connectivity topic")
	}
}

func TestEngineCloseStopsConnectivityTopicBridge(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "connectivity.*", topics.SubscribeOptions{Buffer: 4})
	defer unsub()
	if err := engine.Close(); err != nil {
		t.Fatalf("close engine: %v", err)
	}
	// Allow the connectivity bus subscription goroutine to observe cancellation.
	time.Sleep(20 * time.Millisecond)
	if err := engine.Connectivity().Emit(context.Background(), "route.http.demo", map[string]any{"ok": true, "session_id": "s1"}); err != nil {
		t.Fatalf("emit connectivity event after close: %v", err)
	}
	select {
	case env := <-ch:
		t.Fatalf("unexpected bridged connectivity topic after close: %#v", env)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestRecordExtensionPublishesExtensionTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "extension.*", topics.SubscribeOptions{Buffer: 4})
	defer unsub()

	engine.recordExtension(ExtensionInfo{Engine: "joker", Path: ".gi/extensions/demo.joke", Status: "loaded"})

	select {
	case env := <-ch:
		if env.Topic != "extension.loaded" {
			t.Fatalf("unexpected extension topic: %#v", env)
		}
		if got := env.Payload["path"]; got != ".gi/extensions/demo.joke" {
			t.Fatalf("unexpected payload: %#v", env.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("expected extension lifecycle topic")
	}
}

func TestSteeringBroadcastEventsMapToSessionSteeringTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "session.steering", topics.SubscribeOptions{Buffer: 16})
	defer unsub()

	types := []string{"steering_enqueued", "steering_dequeued", "steering_continue_staged", "steering_continued", "steering_injected"}
	for _, typ := range types {
		engine.broadcast("session_steering_topic", map[string]any{"type": typ, "turn_id": "turn_1"})
		select {
		case env := <-ch:
			if env.Topic != "session.steering" {
				t.Fatalf("unexpected topic for %s: %#v", typ, env)
			}
			if gotType, _ := env.Payload["type"].(string); gotType != typ {
				t.Fatalf("unexpected payload for %s: %#v", typ, env.Payload)
			}
		case <-time.After(time.Second):
			t.Fatalf("expected session.steering event for %s", typ)
		}
	}
}

func TestEmitHookPublishesRuntimeHookTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.hook", topics.SubscribeOptions{Buffer: 8})
	defer unsub()
	if _, err := engine.RegisterHook(HookToolCall, "test", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return HookResponse{Action: "modify"}, nil
	}); err != nil {
		t.Fatalf("register hook: %v", err)
	}
	if _, err := engine.emitHook(context.Background(), HookRequest{Name: HookToolCall, SessionID: "session_hook_topic", TurnID: "turn_hook_topic", AgentID: "agent"}); err != nil {
		t.Fatalf("emit hook: %v", err)
	}
	select {
	case env := <-ch:
		if env.Topic != "runtime.hook" {
			t.Fatalf("unexpected runtime hook topic: %#v", env)
		}
		if env.Payload["hook"] != HookToolCall || env.Payload["action"] != "modify" || env.Payload["source"] != "test" {
			t.Fatalf("unexpected runtime hook payload: %#v", env.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.hook topic event")
	}
}

func TestEmitHookPublishesRuntimeHookErrorTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.hook", topics.SubscribeOptions{Buffer: 8})
	defer unsub()
	if _, err := engine.RegisterHook(HookToolCall, "broken", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return HookResponse{}, context.DeadlineExceeded
	}); err != nil {
		t.Fatalf("register hook: %v", err)
	}
	engine.runtimeCfg.Hooks.OnError = "continue"
	if _, err := engine.emitHook(context.Background(), HookRequest{Name: HookToolCall, SessionID: "session_hook_topic", TurnID: "turn_hook_topic", AgentID: "agent"}); err != nil {
		t.Fatalf("emit hook with continue policy: %v", err)
	}
	select {
	case env := <-ch:
		if env.Topic != "runtime.hook" {
			t.Fatalf("unexpected runtime hook topic: %#v", env)
		}
		if env.Payload["hook"] != HookToolCall || env.Payload["source"] != "broken" || env.Payload["error"] == "" {
			t.Fatalf("unexpected runtime hook error payload: %#v", env.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.hook error topic event")
	}
}

func TestPublishRuntimeHookDecisionEventPublishesRuntimeHookTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.hook", topics.SubscribeOptions{Buffer: 8, SessionID: "session_hook_decision"})
	defer unsub()
	call := goai.ToolCall{ID: "call_hook_decision", Name: "grep"}
	engine.PublishRuntimeHookDecisionEvent("hook_deny", HookRequest{Name: HookApproveTool, SessionID: "session_hook_decision", TurnID: "turn_hook_decision", AgentID: "agent", Iteration: 2, ToolCall: &call}, map[string]any{"phase": "approve_tool", "reason": "tool not approved"})
	select {
	case env := <-ch:
		if env.Topic != "runtime.hook" || env.Payload["type"] != "hook_deny" || env.Payload["hook"] != HookApproveTool || env.Payload["tool"] != "grep" || env.Payload["tool_call_id"] != "call_hook_decision" {
			t.Fatalf("unexpected runtime hook decision payload: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.hook decision topic event")
	}
}

func TestPublishRuntimeTurnEventPublishesRuntimeTurnTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 8})
	defer unsub()

	engine.PublishRuntimeTurnEvent("turn_started", "session_turn_topic", "turn_topic_1", "agent_topic", "running", "setup", map[string]any{"reason": "setup"})

	select {
	case env := <-ch:
		if env.Topic != "runtime.turn" {
			t.Fatalf("unexpected runtime.turn topic: %#v", env)
		}
		if env.SessionID != "session_turn_topic" || env.AgentID != "agent_topic" {
			t.Fatalf("unexpected runtime.turn scope: %#v", env)
		}
		if env.Payload["type"] != "turn_started" || env.Payload["turn_id"] != "turn_topic_1" || env.Payload["status"] != "running" || env.Payload["phase"] != "setup" {
			t.Fatalf("unexpected runtime.turn payload: %#v", env.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.turn topic event")
	}
}

func TestPublishRuntimeSessionEventPublishesRuntimeSessionTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 8})
	defer unsub()

	engine.PublishRuntimeSessionEvent("session_idle", "session_state_topic", "agent_topic", "idle", map[string]any{"reason": "turn_completed", "turn_id": "turn_topic_1"})

	select {
	case env := <-ch:
		if env.Topic != "runtime.session" {
			t.Fatalf("unexpected runtime.session topic: %#v", env)
		}
		if env.SessionID != "session_state_topic" || env.AgentID != "agent_topic" {
			t.Fatalf("unexpected runtime.session scope: %#v", env)
		}
		if env.Payload["type"] != "session_idle" || env.Payload["status"] != "idle" || env.Payload["turn_id"] != "turn_topic_1" {
			t.Fatalf("unexpected runtime.session payload: %#v", env.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.session topic event")
	}
}

func TestPublishRuntimeToolEventPublishesRuntimeToolTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.tool", topics.SubscribeOptions{Buffer: 8})
	defer unsub()

	engine.PublishRuntimeToolEvent("tool_finished", "session_tool_topic", "turn_tool_topic", "agent_tool", "grep", "call_1", 3, nil, map[string]any{"output_length": 42})

	select {
	case env := <-ch:
		if env.Topic != "runtime.tool" {
			t.Fatalf("unexpected runtime.tool topic: %#v", env)
		}
		if env.SessionID != "session_tool_topic" || env.AgentID != "agent_tool" {
			t.Fatalf("unexpected runtime.tool scope: %#v", env)
		}
		if env.Payload["type"] != "tool_finished" || env.Payload["tool"] != "grep" || env.Payload["tool_call_id"] != "call_1" || env.Payload["turn_id"] != "turn_tool_topic" || env.Payload["iteration"] != 3 || env.Payload["output_length"] != 42 {
			t.Fatalf("unexpected runtime.tool payload: %#v", env.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.tool topic event")
	}
}

func TestEmitTurnStateHookPublishesRuntimeTurnStateTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	runner := &sessionRunner{store: s, engine: engine}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 8, SessionID: "session_turn_state"})
	defer unsub()

	runner.emitTurnStateHook(context.Background(), "session_turn_state", "turn_state_1", "agent_state", "model", "running", "waiting_on_tools", map[string]any{"reason": "tool_execution", "tool": "grep"})

	select {
	case env := <-ch:
		if env.Payload["type"] != "turn_state" || env.Payload["status"] != "running" || env.Payload["phase"] != "waiting_on_tools" || env.Payload["tool"] != "grep" {
			t.Fatalf("unexpected runtime.turn state payload: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.turn state topic event")
	}
}

func TestEmitSessionStateHookPublishesRuntimeSessionStateTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	runner := &sessionRunner{store: s, engine: engine}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 8, SessionID: "session_state_hook"})
	defer unsub()

	runner.emitSessionStateHook(context.Background(), "session_state_hook", "agent_state", "model", "running", map[string]any{"reason": "setup", "active_turn_id": "turn_state_1"})

	select {
	case env := <-ch:
		if env.Payload["type"] != "session_state" || env.Payload["status"] != "running" || env.Payload["active_turn_id"] != "turn_state_1" {
			t.Fatalf("unexpected runtime.session state payload: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.session state topic event")
	}
}

func TestHookOnlyStateEmittersDoNotPublishGenericRuntimeTopics(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	runner := &sessionRunner{store: s, engine: engine}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	turnCh, unsubTurn := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 4, SessionID: "session_hook_only"})
	defer unsubTurn()
	sessionCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 4, SessionID: "session_hook_only"})
	defer unsubSession()

	runner.emitTurnStateHookOnly(context.Background(), "session_hook_only", "turn_hook_only", "agent", "model", "completed", "completed", map[string]any{"reason": "completed"})
	runner.emitSessionStateHookOnly(context.Background(), "session_hook_only", "agent", "model", "idle", map[string]any{"reason": "turn_completed"})

	select {
	case env := <-turnCh:
		t.Fatalf("unexpected runtime.turn topic from hook-only emitter: %#v", env)
	case <-time.After(100 * time.Millisecond):
	}
	select {
	case env := <-sessionCh:
		t.Fatalf("unexpected runtime.session topic from hook-only emitter: %#v", env)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestFinishTurnStoresTerminalSystemMessageAsSystemRole(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	runner := &sessionRunner{store: s, engine: engine}
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_finish_system_role", "Test", map[string]any{"status": "running"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_finish_system_role", "session_finish_system_role", "running", "running", map[string]any{}); err != nil {
		t.Fatalf("create turn: %v", err)
	}

	runner.finishTurn(s, "turn_finish_system_role", "session_finish_system_role", "agent", "model", "failed", "Inference error: boom", "provider_error")

	msgs, err := s.ListMessages(ctx, "session_finish_system_role")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	var found bool
	for _, msg := range msgs {
		if msg.Content != "Inference error: boom" {
			continue
		}
		found = true
		if msg.Role != "system" {
			t.Fatalf("expected terminal status message role=system, got %#v", msg)
		}
		if msg.Payload["source"] != "system" {
			t.Fatalf("expected terminal status payload source=system, got %#v", msg.Payload)
		}
	}
	if !found {
		t.Fatal("expected stored terminal status message")
	}
}

func TestFinishTurnCompletedPublishesCompletedRuntimeTopics(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	runner := &sessionRunner{store: s, engine: engine}
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_finish_completed", "Test", map[string]any{"status": "running"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_finish_completed", "session_finish_completed", "running", "running", map[string]any{}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	subCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	turnCh, unsubTurn := engine.Topics().Subscribe(subCtx, "runtime.turn", topics.SubscribeOptions{Buffer: 8, SessionID: "session_finish_completed"})
	defer unsubTurn()
	sessionCh, unsubSession := engine.Topics().Subscribe(subCtx, "runtime.session", topics.SubscribeOptions{Buffer: 8, SessionID: "session_finish_completed"})
	defer unsubSession()

	runner.finishTurn(s, "turn_finish_completed", "session_finish_completed", "agent", "model", "completed", "Reached maximum iteration limit (1). The task may be incomplete.", "")

	select {
	case env := <-turnCh:
		if env.Payload["type"] != "turn_completed" || env.Payload["status"] != "completed" {
			t.Fatalf("unexpected runtime.turn payload for completed finishTurn: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.turn completed event")
	}
	select {
	case env := <-sessionCh:
		if env.Payload["type"] != "session_idle" || env.Payload["reason"] != "turn_completed" || env.Payload["turn_status"] != "completed" {
			t.Fatalf("unexpected runtime.session payload for completed finishTurn: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.session idle event")
	}
}

func TestPublishRuntimeRoutingEventUsesExpectedSessionScope(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	decisionCh, unsubDecision := engine.Topics().Subscribe(ctx, "runtime.routing", topics.SubscribeOptions{Buffer: 8, SessionID: "session_route_source"})
	defer unsubDecision()
	incomingCh, unsubIncoming := engine.Topics().Subscribe(ctx, "runtime.routing", topics.SubscribeOptions{Buffer: 8, SessionID: "session_route_target"})
	defer unsubIncoming()
	decision := store.RouteEvent{
		ID:             42,
		TurnID:         "turn_route_topic",
		SourceSession:  "session_route_source",
		TargetSession:  "session_route_target",
		SourceAgentID:  "agent_source",
		TargetAgentID:  "agent_target",
		Mode:           "prompt",
		MatchedBy:      "mention",
		RoutingPolicy:  "mention",
		RequestedAgent: "agent_target",
		Metadata:       map[string]any{"created_session": true},
		CreatedAt:      time.Now().UTC().Format(time.RFC3339Nano),
	}

	engine.PublishRuntimeRoutingEvent("routing_decision", decision)
	select {
	case env := <-decisionCh:
		if env.SessionID != "session_route_source" || env.Payload["type"] != "routing_decision" || env.Payload["route_event_id"] != int64(42) {
			t.Fatalf("unexpected routing decision topic: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected routing_decision topic event")
	}

	engine.PublishRuntimeRoutingEvent("routing_incoming", decision)
	select {
	case env := <-incomingCh:
		if env.SessionID != "session_route_target" || env.Payload["type"] != "routing_incoming" || env.Payload["target_agent_id"] != "agent_target" {
			t.Fatalf("unexpected routing incoming topic: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected routing_incoming topic event")
	}
}

func TestSubTurnBroadcastEventsMapToTurnSubTurnTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "turn.subturn", topics.SubscribeOptions{Buffer: 8})
	defer unsub()

	for _, typ := range []string{"subturn_created", "subturn_status", "subturn_result_ready", "subturn_result_delivered", "subturn_orphaned", "subturn_cancel_requested"} {
		engine.broadcast("session_subturn_topic", map[string]any{"type": typ, "parent_turn_id": "turn_parent", "child_turn_id": "turn_child"})
		select {
		case env := <-ch:
			if env.Topic != "turn.subturn" {
				t.Fatalf("unexpected subturn topic for %s: %#v", typ, env)
			}
			if gotType, _ := env.Payload["type"].(string); gotType != typ {
				t.Fatalf("unexpected payload type for %s: %#v", typ, env.Payload)
			}
		case <-time.After(time.Second):
			t.Fatalf("expected turn.subturn event for %s", typ)
		}
	}
}

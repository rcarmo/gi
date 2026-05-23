package scripting

import (
	"context"
	"testing"
)

func TestExecuteEmbeddedJoker(t *testing.T) {
	bridge := NewBridge("test-session", BridgeFuncs{})
	out, err := ExecuteEmbeddedJoker(context.Background(), `(+ 40 2)`, bridge)
	if err != nil {
		t.Fatalf("ExecuteEmbeddedJoker returned error: %v", err)
	}
	if out != "42" {
		t.Fatalf("expected 42, got %q", out)
	}
}

func TestJokerRunnerName(t *testing.T) {
	r := NewJokerRunner()
	if got := r.Name(); got != "joker" {
		t.Fatalf("unexpected runner name: %q", got)
	}
}

func TestBuildBridgeStateIncludesSessionID(t *testing.T) {
	bridge := NewBridge("abc123", BridgeFuncs{})
	state := buildBridgeState(context.Background(), bridge)
	if got := state["session-id"]; got != "abc123" {
		t.Fatalf("expected session-id abc123, got %#v", got)
	}
}

func TestBuildBridgeStateIncludesSessionMessages(t *testing.T) {
	bridge := NewBridge("abc123", BridgeFuncs{
		ListMessages: func(ctx context.Context, limit int) ([]map[string]any, error) {
			return []map[string]any{
				{"id": "m1", "role": "user", "content": "first"},
			}, nil
		},
	})
	state := buildBridgeState(context.Background(), bridge)
	msgs, ok := state["messages"].([]map[string]any)
	if !ok {
		t.Fatalf("expected messages in bridge state, got %#v", state["messages"])
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

func TestExecuteEmbeddedJokerSupportsSessionStateAndMetadata(t *testing.T) {
	state := map[string]any{"initialized": true}
	bridge := NewBridge("joker-session", BridgeFuncs{
		GetSessionState: func(ctx context.Context) (map[string]any, error) {
			return state, nil
		},
		SetSessionState: func(ctx context.Context, patch map[string]any) error {
			for k, v := range patch {
				state[k] = v
			}
			return nil
		},
		GetSessionInfo: func(ctx context.Context) (map[string]any, error) {
			return map[string]any{
				"session": map[string]any{"title": "demo"},
			}, nil
		},
		GetConfig: func(ctx context.Context) (map[string]any, error) {
			return map[string]any{"default_model": "joker-model"}, nil
		},
		ListTurns: func(ctx context.Context, limit int) ([]map[string]any, error) {
			return []map[string]any{{"id": "t1"}, {"id": "t2"}}, nil
		},
		ListMessages: func(ctx context.Context, limit int) ([]map[string]any, error) {
			msgs := []map[string]any{{"id": "m1"}, {"id": "m2"}, {"id": "m3"}}
			if limit > 0 && limit < len(msgs) {
				return msgs[:limit], nil
			}
			return msgs, nil
		},
	})

	out, err := ExecuteEmbeddedJoker(context.Background(), `
		(do
			(gi-set-session-state! {:markdown_ready true})
			(str "title=" (:title (:session (gi-get-session-info))) " / "
				(:default_model (gi-get-runtime-config)) " / turns=" (count (gi-list-turns)) " / messages=" (count (gi-list-messages)) "/lim=" (count (gi-list-messages {:limit 2}))))`, bridge)
	if err != nil {
		t.Fatalf("joker execute returned error: %v", err)
	}
	if out != "title=demo / joker-model / turns=2 / messages=3/lim=2" {
		t.Fatalf("unexpected output: %q", out)
	}
	if state["markdown_ready"] != true {
		t.Fatalf("expected markdown_ready=true, got %#v", state["markdown_ready"])
	}
}

func TestExecuteEmbeddedJokerSupportsListMessages(t *testing.T) {
	bridge := NewBridge("joker-session", BridgeFuncs{
		ListMessages: func(ctx context.Context, limit int) ([]map[string]any, error) {
			msgs := []map[string]any{
				{"id": "m1", "content": "hi"},
				{"id": "m2", "content": "there"},
				{"id": "m3", "content": "everyone"},
			}
			if limit > 0 && limit < len(msgs) {
				return msgs[:limit], nil
			}
			return msgs, nil
		},
	})

	out, err := ExecuteEmbeddedJoker(context.Background(), `(str (count (gi-list-messages)) ":" (count (gi-list-messages {:limit 2})))`, bridge)
	if err != nil {
		t.Fatalf("joker execute returned error: %v", err)
	}
	if out != "3:2" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestExecuteEmbeddedJokerSupportsTopicPublish(t *testing.T) {
	var published map[string]any
	bridge := NewBridge("joker-session", BridgeFuncs{
		PublishTopic: func(ctx context.Context, envelope map[string]any) error {
			published = envelope
			return nil
		},
	})
	out, err := ExecuteEmbeddedJoker(context.Background(), `
		(do
			(gi-topic-publish {:topic "runtime.test" :payload {:ok true} :type "notice" :source "script"})
			"done")
	`, bridge)
	if err != nil {
		t.Fatalf("joker execute returned error: %v", err)
	}
	if out != "done" {
		t.Fatalf("unexpected output: %q", out)
	}
	if published["topic"] != "runtime.test" {
		t.Fatalf("unexpected published topic envelope: %#v", published)
	}
	payload, _ := published["payload"].(map[string]any)
	if payload == nil || payload["ok"] != true {
		t.Fatalf("unexpected published topic payload: %#v", published)
	}
}

func TestExecuteEmbeddedJokerSupportsTopicSubscribeReadAndUnsubscribe(t *testing.T) {
	var unsubscribed bool
	bridge := NewBridge("joker-session", BridgeFuncs{
		SubscribeTopic: func(ctx context.Context, pattern string, opts TopicSubscribeOptions) (string, error) {
			if pattern != "runtime.*" {
				t.Fatalf("unexpected topic pattern: %q", pattern)
			}
			if opts.SessionID != "joker-session" {
				t.Fatalf("unexpected topic subscribe opts: %#v", opts)
			}
			return "sub-1", nil
		},
		ReadTopicSubscription: func(ctx context.Context, id string, limit int) ([]map[string]any, error) {
			if id != "sub-1" || limit != 5 {
				t.Fatalf("unexpected topic read request: id=%q limit=%d", id, limit)
			}
			return []map[string]any{{"topic": "runtime.test", "sequence": 7, "payload": map[string]any{"ok": true}}}, nil
		},
		UnsubscribeTopic: func(ctx context.Context, id string) error {
			if id != "sub-1" {
				t.Fatalf("unexpected topic unsubscribe id: %q", id)
			}
			unsubscribed = true
			return nil
		},
	})
	out, err := ExecuteEmbeddedJoker(context.Background(), `
		(do
			(def sub (gi-topic-subscribe "runtime.*" {:session_id "joker-session"}))
			(def events (gi-topic-read sub 5))
			(def first-event (first events))
			(gi-topic-unsubscribe sub)
			(str (get first-event "topic") ":" (get (get first-event "payload") "ok") ":" (get first-event "sequence")))
	`, bridge)
	if err != nil {
		t.Fatalf("joker execute returned error: %v", err)
	}
	if out != "runtime.test:true:7" {
		t.Fatalf("unexpected output: %q", out)
	}
	if !unsubscribed {
		t.Fatal("expected topic subscription to be unsubscribed")
	}
}

func TestExecuteEmbeddedJokerSupportsEventHooksAndHTTPRequest(t *testing.T) {
	var hookSeen bool
	var emitted bool
	var request HTTPCallSpec
	bridge := NewBridge("joker-session", BridgeFuncs{
		RegisterEventHook: func(ctx context.Context, spec EventHookSpec) error {
			hookSeen = true
			if spec.Name != "evt" {
				t.Fatalf("unexpected hook name: %q", spec.Name)
			}
			return nil
		},
		EmitEvent: func(ctx context.Context, name string, payload map[string]any) error {
			emitted = emitted || name == "evt"
			return nil
		},
		ClearEventHooks: func(ctx context.Context) error {
			if !hookSeen {
				t.Fatalf("clearEventHooks called before register")
			}
			return nil
		},
		DoHTTPRequest: func(ctx context.Context, req HTTPCallSpec) (HTTPResponse, error) {
			request = req
			return HTTPResponse{
				StatusCode: 201,
				Status:     "201 Created",
				Headers:    map[string][]string{"x-test": {"1"}},
				Body:       "pong",
				URL:        req.URL,
			}, nil
		},
	})

	script := `
		(do
			(gi-register-event-hook {:name "evt" :source "script" :filter {:kind "test"}})
			(gi-emit-event "evt" {:source "script" :payload {:n 1}})
			(gi-clear-event-hooks)
			(let [r (gi-http-request {"method" "POST" "url" "https://example.invalid/ok" "headers" {"x-test" ["1"]} "body" "ping" "timeout_ms" 250})]
				(str (:status_code r) ":" (:body r))))
	`
	out, err := ExecuteEmbeddedJoker(context.Background(), script, bridge)
	if err != nil {
		t.Fatalf("joker execute returned error: %v", err)
	}
	if out != "201:pong" {
		t.Fatalf("unexpected output: %q", out)
	}
	if !hookSeen {
		t.Fatal("expected event hook to be registered")
	}
	if !emitted {
		t.Fatal("expected event to emit")
	}
	if request.Method != "POST" || request.URL != "https://example.invalid/ok" {
		t.Fatalf("unexpected request spec: %#v", request)
	}
}

package scripting

import (
	"context"
	"testing"
)

func TestGojaRunnerName(t *testing.T) {
	if got := NewGojaRunner().Name(); got != "js" {
		t.Fatalf("unexpected runner name: %q", got)
	}
}

func TestGojaRunnerExecuteSupportsSessionStateAndMetadata(t *testing.T) {
	state := map[string]any{"initialized": true}
	bridge := NewBridge("goja-session", BridgeFuncs{
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
			return map[string]any{"default_model": "goja-model"}, nil
		},
		ListTurns: func(ctx context.Context, limit int) ([]map[string]any, error) {
			return []map[string]any{{"id": "t1"}, {"id": "t2"}}, nil
		},
		ListMessages: func(ctx context.Context, limit int) ([]map[string]any, error) {
			messages := []map[string]any{{"id": "m1"}, {"id": "m2"}, {"id": "m3"}}
			if limit > 0 && limit < len(messages) {
				return messages[:limit], nil
			}
			return messages, nil
		},
	})

	out, err := NewGojaRunner().Execute(context.Background(), `
		gi.setSessionState({markdown_ready: true});
		gi.getSessionInfo().session.title + "/" + gi.getRuntimeConfig().default_model + "/turns=" + gi.listTurns().length + "/messages=" + gi.listMessages().length + "/lim=" + gi.listMessages(2).length;
	`, bridge)
	if err != nil {
		t.Fatalf("goja execute returned error: %v", err)
	}
	if out != "demo/goja-model/turns=2/messages=3/lim=2" {
		t.Fatalf("unexpected output: %q", out)
	}
	if state["markdown_ready"] != true {
		t.Fatalf("expected markdown_ready=true, got %#v", state["markdown_ready"])
	}
}

func TestGojaRunnerExecuteSupportsTopicPublish(t *testing.T) {
	var published map[string]any
	bridge := NewBridge("goja-session", BridgeFuncs{
		PublishTopic: func(ctx context.Context, envelope map[string]any) error {
			published = envelope
			return nil
		},
	})
	out, err := NewGojaRunner().Execute(context.Background(), `
		gi.topics.publish({topic: "runtime.test", payload: {ok: true}, type: "notice", source: "script"});
		"done";
	`, bridge)
	if err != nil {
		t.Fatalf("goja execute returned error: %v", err)
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

func TestGojaRunnerExecuteSupportsTopicSubscribeReadAndUnsubscribe(t *testing.T) {
	var unsubscribed bool
	bridge := NewBridge("goja-session", BridgeFuncs{
		SubscribeTopic: func(ctx context.Context, pattern string, opts TopicSubscribeOptions) (string, error) {
			if pattern != "runtime.*" {
				t.Fatalf("unexpected topic pattern: %q", pattern)
			}
			if opts.SessionID != "goja-session" {
				t.Fatalf("unexpected topic subscribe opts: %#v", opts)
			}
			return "sub-1", nil
		},
		ReadTopicSubscription: func(ctx context.Context, id string, limit int) ([]map[string]any, error) {
			if id != "sub-1" || limit != 5 {
				t.Fatalf("unexpected topic read request: id=%q limit=%d", id, limit)
			}
			return []map[string]any{{"topic": "runtime.test", "payload": map[string]any{"ok": true}}}, nil
		},
		UnsubscribeTopic: func(ctx context.Context, id string) error {
			if id != "sub-1" {
				t.Fatalf("unexpected topic unsubscribe id: %q", id)
			}
			unsubscribed = true
			return nil
		},
	})
	out, err := NewGojaRunner().Execute(context.Background(), `
		var sub = gi.topics.subscribe("runtime.*", {session_id: "goja-session"});
		var events = gi.topics.read(sub, 5);
		gi.topics.unsubscribe(sub);
		events[0].topic + ":" + events[0].payload.ok;
	`, bridge)
	if err != nil {
		t.Fatalf("goja execute returned error: %v", err)
	}
	if out != "runtime.test:true" {
		t.Fatalf("unexpected output: %q", out)
	}
	if !unsubscribed {
		t.Fatal("expected topic subscription to be unsubscribed")
	}
}

func TestGojaRunnerExecuteSupportsEventHooksAndHTTPRequest(t *testing.T) {
	var hookSeen bool
	var emitted bool
	var request HTTPCallSpec
	bridge := NewBridge("goja-session", BridgeFuncs{
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
		gi.registerEventHook({name: "evt", source: "script", filter: {kind: "test"}});
		gi.emitEvent("evt", {source: "script", payload: {n: 1}});
		gi.clearEventHooks();
		var r = gi.http.request({
			method: "POST",
			url: "https://example.invalid/ok",
			headers: {"x-test": ["1"]},
			body: "ping",
			timeout_ms: 250
		});
		r.body;
	`
	out, err := NewGojaRunner().Execute(context.Background(), script, bridge)
	if err != nil {
		t.Fatalf("goja execute returned error: %v", err)
	}
	if out != "pong" {
		t.Fatalf("unexpected output: %q", out)
	}
	if !hookSeen {
		t.Fatal("expected event hook to be registered")
	}
	if !emitted {
		t.Fatal("expected event to emit")
	}
	if request.Method != "POST" || request.URL != "https://example.invalid/ok" || len(request.Headers["x-test"]) != 1 {
		t.Fatalf("unexpected request spec: %#v", request)
	}
}

package scripting

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestQuickJSBridgeRuntimeAndIsolation(t *testing.T) {
	state := map[string]any{"model": "fixture"}
	var mu sync.Mutex
	b := NewBridge("quick-session", BridgeFuncs{GetSessionState: func(context.Context) (map[string]any, error) { mu.Lock(); defer mu.Unlock(); return state, nil }, SetSessionState: func(ctx context.Context, p map[string]any) error {
		mu.Lock()
		defer mu.Unlock()
		for k, v := range p {
			state[k] = v
		}
		return nil
	}, ListMessages: func(ctx context.Context, n int) ([]map[string]any, error) {
		return []map[string]any{{"content": "fixture"}}, nil
	}, Log: func(context.Context, string, string) {}})
	r := NewQuickJSRunner()
	out, err := r.Execute(context.Background(), `(async()=>{await gi.setSessionState({ready:true});const s=await gi.getSessionState();const m=await gi.listMessages(1);globalThis.privateProbe=99;return gi.sessionId+"/"+s.model+"/"+m[0].content;})()`, b)
	if err != nil || out != "quick-session/fixture/fixture" {
		t.Fatal(out, err)
	}
	if state["ready"] != true {
		t.Fatal(state)
	}
	out, err = r.Execute(context.Background(), `typeof privateProbe`, b)
	if err != nil || out != "undefined" {
		t.Fatal("isolate leaked state", out, err)
	}
	out, err = r.Execute(context.Background(), `console.log("from quickjs");40+2`, b)
	if err != nil || out != "from quickjs" {
		t.Fatal(out, err)
	}
	out, err = r.Execute(context.Background(), `40+2`, b)
	if err != nil || out != "42" {
		t.Fatal(out, err)
	}
}

func TestQuickJSCancellationAndNoAmbientCapabilities(t *testing.T) {
	r := NewQuickJSRunner()
	b := NewBridge("q", BridgeFuncs{})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, err := r.Execute(ctx, `while(true){}`, b); err == nil || ctx.Err() == nil {
		t.Fatal(err)
	}
	if out, err := r.Execute(context.Background(), `typeof process+"/"+typeof require+"/"+typeof fetch`, b); err != nil || out != "undefined/undefined/undefined" {
		t.Fatal(out, err)
	}
	if _, err := r.Execute(context.Background(), `throw new Error("quick-fixture")`, b); err == nil || !strings.Contains(err.Error(), "quick-fixture") {
		t.Fatal(err)
	}
	if out, err := r.Execute(context.Background(), `6*7`, b); err != nil || out != "42" {
		t.Fatal(out, err)
	}
}

func TestQuickJSHostCallCancellationIsJoined(t *testing.T) {
	// Warm module compilation before testing cancellation of a started host call.
	if _, err := NewQuickJSRunner().Execute(context.Background(), `42`, NewBridge("warm", BridgeFuncs{})); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	done := make(chan struct{})
	bridge := NewBridge("cancel", BridgeFuncs{ReadFile: func(ctx context.Context, path string) (string, error) {
		close(started)
		<-ctx.Done()
		close(done)
		return "", ctx.Err()
	}})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() {
		select {
		case <-started:
			cancel()
		case <-ctx.Done():
		}
	}()
	if _, err := NewQuickJSRunner().Execute(ctx, `(async()=>await gi.readFile("wait"))()`, bridge); err == nil {
		t.Fatal("cancelled host call succeeded")
	}
	select {
	case <-done:
	default:
		t.Fatal("host call survived execution")
	}
}

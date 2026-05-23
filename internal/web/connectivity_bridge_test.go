package web

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	"github.com/rcarmo/gi/internal/topics"
	"github.com/rcarmo/gi/internal/turn"
)

func TestNormalizeSessionRouteFilterDoesNotMutateCallerMap(t *testing.T) {
	original := map[string]any{"name": "demo"}
	normalized, err := normalizeSessionRouteFilter("session_a", original)
	if err != nil {
		t.Fatalf("normalize route filter: %v", err)
	}
	if got, _ := normalized["session_id"].(string); got != "session_a" {
		t.Fatalf("expected normalized session id, got %#v", normalized)
	}
	if _, ok := original["session_id"]; ok {
		t.Fatalf("expected original filter not to be mutated, got %#v", original)
	}
}

func TestNormalizeSessionRouteFilterRejectsCrossSessionOverride(t *testing.T) {
	_, err := normalizeSessionRouteFilter("session_a", map[string]any{"session_id": "session_b"})
	if err == nil || !strings.Contains(err.Error(), "does not match current session") {
		t.Fatalf("expected cross-session filter rejection, got %v", err)
	}
}

func TestWithSessionIDDoesNotMutateCallerPayload(t *testing.T) {
	original := map[string]any{"ok": true}
	normalized := withSessionID(original, "session_a")
	if got, _ := normalized["session_id"].(string); got != "session_a" {
		t.Fatalf("expected normalized payload session id, got %#v", normalized)
	}
	if _, ok := original["session_id"]; ok {
		t.Fatalf("expected original payload not to be mutated, got %#v", original)
	}
}

func TestScriptConnectivityBridgeScopesListAndUnregisterToSession(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	engine := turn.New(s)
	defer engine.Close()
	srv := New(s, engine, config.RuntimeConfig{WorkspaceRoot: t.TempDir(), DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "low"})

	sessionA, err := s.CreateSession(context.Background(), store.NowID("session"), "A", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session A: %v", err)
	}
	sessionB, err := s.CreateSession(context.Background(), store.NowID("session"), "B", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session B: %v", err)
	}

	crossRegister := srv.scriptTool.Execute(context.Background(), tools.ScriptInput{SessionID: sessionA.ID, Script: `var r = gi.connect.registerRoute({name:"route-x", transport:"event", session_id:"` + sessionB.ID + `"}); r.id;`})
	if crossRegister.Error == "" || !strings.Contains(crossRegister.Error, "does not match current session") {
		t.Fatalf("expected cross-session register rejection, got result=%q err=%q", crossRegister.Result, crossRegister.Error)
	}

	outA := srv.scriptTool.Execute(context.Background(), tools.ScriptInput{SessionID: sessionA.ID, Script: `var r = gi.connect.registerRoute({name:"route-a", transport:"event"}); r.id;`})
	if outA.Error != "" {
		t.Fatalf("register route A: %v", outA.Error)
	}
	routeA := strings.TrimSpace(outA.Result)
	if routeA == "" {
		t.Fatal("expected route id for session A")
	}

	outB := srv.scriptTool.Execute(context.Background(), tools.ScriptInput{SessionID: sessionB.ID, Script: `var r = gi.connect.registerRoute({name:"route-b", transport:"event"}); r.id;`})
	if outB.Error != "" {
		t.Fatalf("register route B: %v", outB.Error)
	}
	routeB := strings.TrimSpace(outB.Result)
	if routeB == "" {
		t.Fatal("expected route id for session B")
	}

	listA := srv.scriptTool.Execute(context.Background(), tools.ScriptInput{SessionID: sessionA.ID, Script: `JSON.stringify(gi.connect.listRoutes())`})
	if listA.Error != "" {
		t.Fatalf("list routes A: %v", listA.Error)
	}
	if !strings.Contains(listA.Result, routeA) {
		t.Fatalf("expected session A to see own route, got %q", listA.Result)
	}
	if strings.Contains(listA.Result, routeB) {
		t.Fatalf("expected session A list to exclude session B route, got %q", listA.Result)
	}

	crossList := srv.scriptTool.Execute(context.Background(), tools.ScriptInput{SessionID: sessionA.ID, Script: `JSON.stringify(gi.connect.listRoutes({session_id:"` + sessionB.ID + `"}))`})
	if crossList.Error == "" || !strings.Contains(crossList.Error, "does not match current session") {
		t.Fatalf("expected cross-session list rejection, got result=%q err=%q", crossList.Result, crossList.Error)
	}

	crossUnregister := srv.scriptTool.Execute(context.Background(), tools.ScriptInput{SessionID: sessionA.ID, Script: `gi.connect.unregisterRoute("` + routeB + `"); "ok";`})
	if crossUnregister.Error == "" || !strings.Contains(crossUnregister.Error, "does not belong to session") {
		t.Fatalf("expected cross-session unregister rejection, got result=%q err=%q", crossUnregister.Result, crossUnregister.Error)
	}
	if _, ok := engine.Connectivity().Get(routeB); !ok {
		t.Fatal("expected session B route to remain after rejected cross-session unregister")
	}

	topicCh, unsubscribe := engine.Topics().Subscribe(context.Background(), "runtime.test", topics.SubscribeOptions{SessionID: sessionA.ID, Buffer: 1})
	defer unsubscribe()

	publishOK := srv.scriptTool.Execute(context.Background(), tools.ScriptInput{SessionID: sessionA.ID, Script: `gi.topics.publish({topic:"runtime.test", payload:{ok:true}}); "ok";`})
	if publishOK.Error != "" {
		t.Fatalf("publish topic A: %v", publishOK.Error)
	}
	select {
	case env := <-topicCh:
		if env.SessionID != sessionA.ID {
			t.Fatalf("expected published topic session %q, got %q", sessionA.ID, env.SessionID)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for published topic")
	}

	crossPublish := srv.scriptTool.Execute(context.Background(), tools.ScriptInput{SessionID: sessionA.ID, Script: `gi.topics.publish({topic:"runtime.test", session_id:"` + sessionB.ID + `", payload:{ok:true}}); "ok";`})
	if crossPublish.Error == "" || !strings.Contains(crossPublish.Error, "does not match current session") {
		t.Fatalf("expected cross-session topic publish rejection, got result=%q err=%q", crossPublish.Result, crossPublish.Error)
	}

	crossSubscribe := srv.scriptTool.Execute(context.Background(), tools.ScriptInput{SessionID: sessionA.ID, Script: `gi.topics.subscribe("runtime.*", {session_id:"` + sessionB.ID + `"});`})
	if crossSubscribe.Error == "" || !strings.Contains(crossSubscribe.Error, "does not match current session") {
		t.Fatalf("expected cross-session topic subscribe rejection, got result=%q err=%q", crossSubscribe.Result, crossSubscribe.Error)
	}

	jokerCrossPublish := srv.scriptTool.Execute(context.Background(), tools.ScriptInput{SessionID: sessionA.ID, Engine: "joker", Script: `(do (gi-topic-publish {:topic "runtime.test" :session_id "` + sessionB.ID + `" :payload {:ok true}}) "ok")`})
	if jokerCrossPublish.Error == "" || !strings.Contains(jokerCrossPublish.Error, "does not match current session") {
		t.Fatalf("expected joker cross-session topic publish rejection, got result=%q err=%q", jokerCrossPublish.Result, jokerCrossPublish.Error)
	}

	jokerCrossSubscribe := srv.scriptTool.Execute(context.Background(), tools.ScriptInput{SessionID: sessionA.ID, Engine: "joker", Script: `(gi-topic-subscribe "runtime.*" {:session_id "` + sessionB.ID + `"})`})
	if jokerCrossSubscribe.Error == "" || !strings.Contains(jokerCrossSubscribe.Error, "does not match current session") {
		t.Fatalf("expected joker cross-session topic subscribe rejection, got result=%q err=%q", jokerCrossSubscribe.Result, jokerCrossSubscribe.Error)
	}
}

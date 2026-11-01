package scripting

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"testing"

	core "github.com/rcarmo/go-joker/v42/core"
	coretypes "github.com/rcarmo/go-joker/v42/core/types"
)

// Compare native keywordization with the previous interpreted reference. Keep
// the original parser on both sides, including its numeric representation.
func TestEmbeddedJokerNativeBridgeKeywordization(t *testing.T) {
	wide := make(map[string]any)
	for i := range 128 {
		wide[fmt.Sprintf("key-%d", i)] = map[string]any{"nested/key": []any{i, nil, true, "value"}}
	}
	wideJSON, err := json.Marshal(wide)
	if err != nil {
		t.Fatal(err)
	}
	for _, payload := range []string{
		`{}`,
		`{"nested":[{"x/y":{"":[],"/":{},":prefix":null,"λ/β":true}},false,"literal/string",[1,2,3]]}`,
		`{"numbers":[0,-0,1,-1,1.5,1e3,9007199254740993,9223372036854775808,1e-300]}`,
		`{"quoted":"\"\\\n\u0000<script>","session-state":{"counter":7},"session-id":"fixture"}`,
		string(wideJSON),
	} {
		script := fmt.Sprintf(`(letfn [(signature [x]
 (cond (map? x) [(str (type x)) (into {} (map (fn [[k v]] [k (signature v)]) x))]
       (vector? x) [(str (type x)) (mapv signature x)]
       :else [(str (type x)) x]))]
 (let [old (walk/keywordize-keys (json/read-string %q))
       new (user/__gi-keywordize-bridge (json/read-string %q))]
   (= (signature old) (signature new))))`, payload, payload)
		result, err := ExecuteEmbeddedJoker(context.Background(), script, NewBridge("fixture", BridgeFuncs{}))
		if err != nil || result != "true" {
			diagnostic, _ := ExecuteEmbeddedJoker(context.Background(), strings.Replace(script, "(= (signature old) (signature new))", "[(signature old) (signature new)]", 1), NewBridge("fixture", BridgeFuncs{}))
			t.Fatalf("native bridge conversion differs from reference: payloadBytes=%d types=%s result=%q err=%v", len(payload), diagnostic, result, err)
		}
	}
}

func TestEmbeddedJokerNamespaceRefreshMatchesNS(t *testing.T) {
	if _, err := ExecuteEmbeddedJoker(context.Background(), "true", NewBridge("fixture", BridgeFuncs{})); err != nil {
		t.Fatal(err)
	}
	embeddedJokerMu.Lock()
	defer embeddedJokerMu.Unlock()
	env := core.GLOBAL_ENV
	user := env.EnsureSymbolIsNamespace(coretypes.MakeSymbol(core.STRINGS.Intern, "user"))
	enter := user.Resolve("__gi-enter-namespace").Value.(core.Proc)
	for _, name := range []string{"user", "gi.state", "gi.topics", "gi.runtime"} {
		env.SetCurrentNamespace(user)
		// Independently execute the original plain ns form, then compare every
		// referred and local Var, including private/local bindings it retains.
		if err := runEmbeddedJokerScript(fmt.Sprintf("(ns %s)", name)); err != nil {
			t.Fatal(err)
		}
		ns := env.CurrentNamespace()
		want := maps.Clone(ns.Mappings())
		delete(ns.Mappings(), core.STRINGS.Intern("map"))
		enter.Fn([]coretypes.Object{coretypes.MakeString(name)})
		if env.CurrentNamespace() != ns || !maps.Equal(want, ns.Mappings()) {
			t.Fatalf("native namespace refresh differs from ns for %s", name)
		}
	}
	env.SetCurrentNamespace(user)
}

func TestEmbeddedJokerNamespaceLoadedLibSnapshot(t *testing.T) {
	script := `(let [snapshot (disj joker.core/*loaded-libs* 'gi.state)]
 (joker.core/var-set #'joker.core/*loaded-libs* snapshot)
 (user/__gi-enter-namespace "gi.state")
 (and (not (contains? snapshot 'gi.state))
      (contains? joker.core/*loaded-libs* 'gi.state)))`
	result, err := ExecuteEmbeddedJoker(context.Background(), script, NewBridge("fixture", BridgeFuncs{}))
	if err != nil || result != "true" {
		t.Fatalf("loaded-lib registration changed an old persistent snapshot: result=%q err=%v", result, err)
	}
}

func TestEmbeddedJokerNamespaceRejectsInvalidEntry(t *testing.T) {
	bridge := NewBridge("fixture", BridgeFuncs{})
	for _, script := range []string{
		`(user/__gi-enter-namespace "joker.core")`,
		`(user/__gi-enter-namespace)`,
		`(user/__gi-enter-namespace "user" "extra")`,
		`(user/__gi-enter-namespace 7)`,
	} {
		if _, err := ExecuteEmbeddedJoker(context.Background(), script, bridge); err == nil {
			t.Fatalf("invalid namespace entry accepted: %s", script)
		}
	}
	if result, err := ExecuteEmbeddedJoker(context.Background(), "(+ 20 22)", bridge); err != nil || result != "42" {
		t.Fatalf("namespace entry error poisoned the next execution: result=%q err=%v", result, err)
	}
}

func TestEmbeddedJokerNativeBridgeFreshState(t *testing.T) {
	counter := 0
	bridge := NewBridge("fixture", BridgeFuncs{
		GetSessionState: func(context.Context) (map[string]any, error) {
			return map[string]any{"counter": counter}, nil
		},
	})
	for i := range 3 {
		counter = i + 7
		bridge.SessionID = fmt.Sprintf("fixture-%d", i)
		result, err := ExecuteEmbeddedJoker(context.Background(), `(str (:session-id *gi-bridge*) ":" (:counter (gi.state/session-state)))`, bridge)
		want := fmt.Sprintf("fixture-%d:%d", i, counter)
		if err != nil || result != want {
			t.Fatalf("stale bridge state: result=%q want=%q err=%v", result, want, err)
		}
	}
}

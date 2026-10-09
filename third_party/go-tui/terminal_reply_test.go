//go:build !windows

package tui

import "testing"

func TestTerminalReplyParsing(t *testing.T) {
	for _, seq := range []string{"\x1b]7501;?\x07", "\x1b]7501;?version=2\x1b\\", "\x1b[?62;22c"} {
		events, rest := parseInputWithRemainder([]byte("a" + seq + "b"))
		if len(rest) != 0 || len(events) != 3 {
			t.Fatalf("events %#v rest %q", events, rest)
		}
		if events[1].(TerminalReplyEvent).Sequence != seq || events[0].(KeyEvent).Rune != 'a' || events[2].(KeyEvent).Rune != 'b' {
			t.Fatal(events)
		}
		for split := 2; split < len(seq); split++ {
			events, rest = parseInputWithRemainder([]byte(seq[:split]))
			if len(events) != 0 || string(rest) != seq[:split] {
				t.Fatalf("split %d: %#v %q", split, events, rest)
			}
			events, rest = parseInputWithRemainder(append(rest, []byte(seq[split:]+"x")...))
			if len(rest) != 0 || len(events) != 2 || events[0].(TerminalReplyEvent).Sequence != seq || events[1].(KeyEvent).Rune != 'x' {
				t.Fatalf("joined %#v %q", events, rest)
			}
		}
		events, rest = parseInputWithRemainder([]byte(pasteStart + seq + pasteEnd))
		if len(rest) != 0 || len(events) != 1 || events[0].(PasteEvent).Text != seq {
			t.Fatal("paste interpreted", events)
		}
	}
	a := &App{}
	var got string
	a.SetTerminalReplyHandler(func(e TerminalReplyEvent) { got = e.Sequence })
	if !a.Dispatch(TerminalReplyEvent{Sequence: "reply"}) || got != "reply" {
		t.Fatal("reply dispatch")
	}
}

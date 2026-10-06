package tui

import (
	"github.com/rcarmo/gi/internal/config"
	"reflect"
	"testing"
)

func BenchmarkTranscriptProjection(b *testing.B) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Gi"}, transcript: windowTestTranscript(), transcriptExpanded: map[string]bool{}}
	c.transcriptRowsAtWidth(100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if len(c.transcriptRowsAtWidth(100)) == 0 {
			b.Fatal("no rows")
		}
	}
}

func TestTranscriptProjectionInvalidationAndBounds(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Gi"}, transcript: windowTestTranscript(), transcriptExpanded: map[string]bool{}}
	check := func(width int) {
		t.Helper()
		got := c.transcriptRowsAtWidth(width)
		fresh := &chatTUI{cfg: c.cfg, transcript: append([]string(nil), c.transcript...), transcriptExpanded: map[string]bool{}, selectedTranscriptBlock: c.selectedTranscriptBlock, outputWidth: c.outputWidth, regularMode: c.regularMode}
		for k, v := range c.transcriptExpanded {
			fresh.transcriptExpanded[k] = v
		}
		want := fresh.transcriptRowsAtWidth(width)
		if !reflect.DeepEqual(got, want) {
			t.Fatal("cached projection differs from fresh render")
		}
	}
	check(100)
	first := c.projectionMemo
	if first == nil {
		t.Fatal("projection not retained")
	}
	check(100)
	if c.projectionMemo != first {
		t.Fatal("unchanged projection rebuilt")
	}
	check(60)
	if c.projectionMemo == first {
		t.Fatal("width not invalidated")
	}
	c.transcript = append([]string(nil), c.transcript...)
	c.transcript[3] = "Gi: streamed replacement text"
	check(60)
	c.selectedTranscriptBlock = "tool"
	check(60)
	for _, b := range c.transcriptBlocks() {
		if b.Key != "" {
			c.transcriptExpanded[b.Key] = true
		}
	}
	check(60)
	oldGen := piThemeGeneration
	piThemeGeneration++
	defer func() { piThemeGeneration = oldGen }()
	check(60)
	c.cfg.AssistantName = "Renamed"
	check(60)
	// Even a projection with short text may own width-sized style/cell arrays.
	c.transcript = []string{"you: bounded"}
	c.projectionMemo = nil
	c.transcriptRowsAtWidth(transcriptProjectionMaxCells + 1)
	if c.projectionMemo != nil {
		t.Fatal("oversized projection retained")
	}
}

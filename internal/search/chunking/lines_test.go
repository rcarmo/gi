package chunking

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestLinesDeterministicLosslessUTF8AndSourceLocations(t *testing.T) {
	inputs := []string{"", "one", "one\r\ntwo\n", strings.Repeat("界é👋", 2500), strings.Repeat("line\n", 300), strings.Repeat("x", LineChunkBytes+5) + "\nend", strings.Repeat("\n", LineChunkLines+1)}
	for _, text := range inputs {
		t.Run(string(rune(len(text)+65)), func(t *testing.T) {
			result, err := (Lines{}).Chunk("test", []byte(text))
			if err != nil {
				t.Fatal(err)
			}
			again, err := (Lines{}).Chunk("other", []byte(text))
			if err != nil || !reflect.DeepEqual(result, again) {
				t.Fatal("not deterministic", err)
			}
			pos, line := 0, 1
			var joined strings.Builder
			for n, c := range result {
				if c.ChunkIndex != n || c.StartByte != pos || c.EndByte-c.StartByte != len(c.Content) || c.StartLine != line || !utf8.ValidString(c.Content) {
					t.Fatal(c)
				}
				count := strings.Count(c.Content, "\n")
				if c.EndLine != line+count || count > LineChunkLines || len(c.Content) > LineChunkBytes {
					t.Fatal(c)
				}
				if c.Content != text[c.StartByte:c.EndByte] {
					t.Fatal("source differs")
				}
				pos = c.EndByte
				line = c.EndLine
				joined.WriteString(c.Content)
			}
			if joined.String() != text || len(result) == 0 {
				t.Fatal("not lossless")
			}
		})
	}
	for _, raw := range [][]byte{{255}, {0}, []byte("a\x00b")} {
		if _, err := (Lines{}).Chunk("bad", raw); err == nil {
			t.Fatal("invalid accepted")
		}
	}
}

func FuzzLinesRoundTrip(f *testing.F) {
	for _, s := range []string{"", "héllo\n世界\n", "\r\n", strings.Repeat("a", 8193)} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 1<<20 {
			t.Skip()
		}
		chunks, err := (Lines{}).Chunk("fuzz", raw)
		if !utf8.Valid(raw) || bytes.IndexByte(raw, 0) >= 0 {
			if err == nil {
				t.Fatal("accepted invalid")
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		var out strings.Builder
		end := 0
		for n, c := range chunks {
			if c.ChunkIndex != n || c.StartByte != end || c.EndByte < c.StartByte || c.EndByte > len(raw) || len(c.Content) > LineChunkBytes || !utf8.ValidString(c.Content) {
				t.Fatal("invalid bounds")
			}
			end = c.EndByte
			out.WriteString(c.Content)
		}
		if out.String() != string(raw) || end != len(raw) {
			t.Fatal("roundtrip")
		}
	})
}

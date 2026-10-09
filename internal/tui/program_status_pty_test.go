//go:build linux || darwin

package tui

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
	"golang.org/x/sys/unix"
)

// Exercise the real reader, render loop and cleanup with a terminal emulator
// replying on a PTY. No live credentials, provider, tmux or network are used.
func TestProgramStatusPTY(t *testing.T) {
	if os.Getenv("GI_STATUS_PTY_CHILD") == "1" {
		programStatusPTYChild(t)
		return
	}
	for _, mode := range []string{"fullscreen", "regular"} {
		for _, support := range []string{"detected", "absent", "late", "forced", "disabled"} {
			t.Run(mode+"/"+support, func(t *testing.T) {
				override := ""
				if support == "forced" {
					override = "1"
				}
				if support == "disabled" {
					override = "0"
				}
				cmd := exec.Command(os.Args[0], "-test.run=^TestProgramStatusPTY$", "-test.timeout=15s")
				cmd.Env = append(os.Environ(), "GI_STATUS_PTY_CHILD=1", "GI_STATUS_MODE="+mode, "GI_STATUS_DIR="+t.TempDir(), "PI_PROGRAM_STATUS="+override, "TERM=xterm-256color")
				master, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 24, Cols: 90})
				if err != nil {
					t.Fatal(err)
				}
				defer master.Close()
				if err := unix.SetNonblock(int(master.Fd()), true); err != nil {
					t.Fatal(err)
				}
				defer func() {
					if cmd.ProcessState == nil {
						_ = cmd.Process.Kill()
						_ = cmd.Wait()
					}
				}()
				var out bytes.Buffer
				queried, typed := false, false
				deadline := time.Now().Add(10 * time.Second)
				buf := make([]byte, 16384)
				for time.Now().Before(deadline) {
					n, err := unix.Read(int(master.Fd()), buf)
					if n > 0 {
						out.Write(buf[:n])
					}
					if err != nil && err != unix.EAGAIN && err != unix.EINTR {
						break
					}
					if !queried && bytes.Contains(out.Bytes(), []byte(programStatusQuery)) {
						queried = true
						switch support {
						case "detected":
							// Separate writes/read cycles: replies must never become editor text.
							_, _ = master.Write([]byte("\x1b]7501;?version=2\x1b"))
							time.Sleep(30 * time.Millisecond)
							_, _ = master.Write([]byte("\\\x1b[?62;22c"))
						case "late":
							_, _ = master.Write([]byte("\x1b[?62;22c" + programStatusQuery))
						default:
							_, _ = master.Write([]byte("\x1b[?62;22c"))
						}
					}
					if !typed && (queried || support == "forced" || support == "disabled") {
						_, _ = master.Write([]byte("draft"))
						typed = true
					}
					if bytes.Contains(out.Bytes(), []byte("RESULT draft=")) {
						break
					}
					time.Sleep(5 * time.Millisecond)
				}
				if err := cmd.Wait(); err != nil {
					t.Fatalf("child %v: %q", err, out.String())
				}
				text := out.String()
				if !strings.Contains(text, "RESULT draft=draft") {
					t.Fatalf("terminal reply corrupted draft: %q", text)
				}
				reports := strings.Contains(text, "\x1b]7501;state=")
				wantReports := support == "forced" || support == "detected"
				if reports != wantReports {
					t.Fatalf("reports=%v want=%v: %q", reports, wantReports, text)
				}
				if queried != (override == "") {
					t.Fatalf("query override: %v", queried)
				}
				if reports {
					for _, state := range []string{"idle", "working", "blocked", "done", "error", "clear"} {
						if !strings.Contains(text, "\x1b]7501;state="+state) {
							t.Errorf("missing %s", state)
						}
					}
				}
			})
		}
	}
}

func programStatusPTYChild(t *testing.T) {
	dir := os.Getenv("GI_STATUS_DIR")
	s, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.CreateSession(context.Background(), "status", "Named session", nil); err != nil {
		t.Fatal(err)
	}
	cfg := config.Load(dir)
	c := &chatTUI{store: s, sessionID: "status", cfg: cfg, transcriptRef: gotui.NewRef(), regularMode: os.Getenv("GI_STATUS_MODE") == "regular"}
	opts := []gotui.AppOption{gotui.WithLegacyKeyboard()}
	if c.regularMode {
		opts = append(opts, gotui.WithInlineHeight(5), gotui.WithPostRenderHook(c.flushRegularTranscript))
	}
	app, err := gotui.NewApp(opts...)
	if err != nil {
		t.Fatal(err)
	}
	c.app = app
	c.programStatus = newProgramStatusTerminal(os.Stdout, os.Getenv("PI_PROGRAM_STATUS"))
	app.SetTerminalReplyHandler(func(e gotui.TerminalReplyEvent) { c.programStatus.reply(e.Sequence) })
	c.programStatus.start()
	cleanup := c.Init()
	defer cleanup()
	app.SetRootComponent(c)
	go func() {
		steps := []func(){
			func() { c.running = true },
			func() { c.openSelect("Choice", []string{"a"}, nil, nil) },
			func() { c.modelMenuOpen = false; c.running = false },
			func() { c.handleEvent(map[string]any{"type": "error", "error": "Failure\nprivate detail"}) },
			func() { c.running = true },
			func() {
				c.handleTopicEvent(topics.Envelope{Topic: "runtime.turn", SessionID: "status", Payload: map[string]any{"type": "turn_terminal", "status": "cancelled"}})
			},
			func() { app.Stop() },
		}
		for _, step := range steps {
			time.Sleep(140 * time.Millisecond)
			app.QueueUpdate(func() { step(); app.MarkDirty() })
		}
	}()
	if err := app.Run(); err != nil {
		t.Fatal(err)
	}
	c.programStatus.stop()
	fmt.Printf("RESULT draft=%s\n", c.input.Text())
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	tui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func main() {
	dbPath := flag.String("db", ".gi-run/gi.db", "SQLite database path")
	workspace := flag.String("workspace", "/workspace", "Workspace root")
	model := flag.String("model", "", "Override default model")
	flag.Parse()

	cfg := config.Load(*workspace)
	if *model != "" {
		cfg.DefaultModel = *model
	}

	s, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	engine := turn.NewWithSystemPrompt(s, cfg.SystemPrompt)

	sessions, _ := s.ListSessions(context.Background())
	var sessionID string
	if len(sessions) > 0 {
		sessionID = sessions[0].ID
	} else {
		sess, err := s.CreateSession(context.Background(), store.NowID("session"), "default", map[string]any{"status": "idle"})
		if err != nil {
			log.Fatalf("create session: %v", err)
		}
		sessionID = sess.ID
	}

	chat := &chatTUI{
		store:     s,
		engine:    engine,
		sessionID: sessionID,
		cfg:       cfg,
	}

	app, err := tui.NewApp(
		tui.WithInlineHeight(4),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}
	chat.app = app

	msgs, _ := s.ListMessages(context.Background(), sessionID)
	for _, m := range msgs {
		prefix := "you"
		if m.Role == "assistant" {
			prefix = cfg.AssistantName
		} else if m.Role == "system" {
			prefix = "sys"
		}
		app.PrintAboveln("%s: %s", prefix, truncate(m.Content, 200))
	}

	app.SetRootComponent(chat)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type chatTUI struct {
	app       *tui.App
	store     *store.Store
	engine    *turn.Engine
	sessionID string
	cfg       config.RuntimeConfig
	history   []string
	histIdx   int
	running   bool
	status    string
	draft     string
	eventCh   chan map[string]any
	input     *tui.Input
}

func (c *chatTUI) Init() func() {
	c.eventCh = c.engine.Subscribe(c.sessionID)
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.histIdx = -1
	c.history = []string{}

	c.input = tui.NewInput(
		tui.WithInputPlaceholder("Send a message…"),
		tui.WithInputAutoFocus(true),
		tui.WithInputOnSubmit(c.onSubmit),
	)

	return func() {
		c.engine.Unsubscribe(c.sessionID, c.eventCh)
	}
}

func (c *chatTUI) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.NewChannelWatcher(c.eventCh, c.handleEvent),
	}
}

func (c *chatTUI) handleEvent(ev map[string]any) {
	evType, _ := ev["type"].(string)
	switch evType {
	case "agent_draft_delta":
		delta, _ := ev["delta"].(string)
		c.draft += delta
		c.status = fmt.Sprintf("⏳ %s…", truncate(c.draft, 60))
	case "new_post":
		data, _ := ev["data"].(map[string]any)
		if data != nil {
			text, _ := data["content"].(string)
			if text != "" {
				c.app.PrintAboveln("%s: %s", c.cfg.AssistantName, text)
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
				c.running = false
				c.draft = ""
			}
		}
	case "agent_status":
		title, _ := ev["title"].(string)
		if title != "" {
			c.status = title
		} else {
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
		}
	}
}

func (c *chatTUI) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnStop(tui.KeyCtrlC, func(ke tui.KeyEvent) { c.app.Stop() }),
		tui.OnStop(tui.KeyCtrlD, func(ke tui.KeyEvent) {
			if c.input.Text() == "" {
				c.app.Stop()
			}
		}),
		tui.OnStop(tui.KeyUp, func(ke tui.KeyEvent) {
			if c.input.Text() != "" || len(c.history) == 0 {
				return
			}
			if c.histIdx < 0 {
				c.histIdx = len(c.history) - 1
			} else if c.histIdx > 0 {
				c.histIdx--
			}
			c.input.SetText(c.history[c.histIdx])
		}),
		tui.OnStop(tui.KeyDown, func(ke tui.KeyEvent) {
			if c.histIdx < 0 {
				return
			}
			c.histIdx++
			if c.histIdx >= len(c.history) {
				c.histIdx = -1
				c.input.SetText("")
			} else {
				c.input.SetText(c.history[c.histIdx])
			}
		}),
	}
}

func (c *chatTUI) onSubmit(text string) {
	if text == "" || c.running {
		return
	}
	c.history = append(c.history, text)
	c.histIdx = -1
	c.input.SetText("")
	c.running = true
	c.status = "sending…"
	c.draft = ""

	c.app.PrintAboveln("you: %s", text)

	go func() {
		_, err := c.engine.SubmitPrompt(context.Background(), turn.RunInput{
			SessionID: c.sessionID,
			Prompt:    text,
			Intent:    "prompt",
			Model:     c.cfg.DefaultModel,
		})
		if err != nil {
			c.app.PrintAboveln("error: %v", err)
			c.running = false
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
		}
	}()
}

func (c *chatTUI) Render(app *tui.App) *tui.Element {
	root := tui.New(
		tui.WithDirection(tui.Column),
		tui.WithWidthPercent(100),
		tui.WithHeightPercent(100),
		tui.WithGap(0),
	)

	statusEl := tui.New(
		tui.WithWidthPercent(100),
		tui.WithHeight(1),
		tui.WithTextStyle(tui.NewStyle().Dim()),
	)
	statusEl.SetText(c.status)
	root.AddChild(statusEl)

	sep := tui.New(
		tui.WithWidthPercent(100),
		tui.WithHeight(1),
		tui.WithTextStyle(tui.NewStyle().Dim()),
	)
	sep.SetText("─────────────────────────────────────────────────────────────────────────────────")
	root.AddChild(sep)

	if c.input != nil {
		inputEl := app.Mount(c, 0, func() tui.Component { return c.input })
		root.AddChild(inputEl)
	}

	return root
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

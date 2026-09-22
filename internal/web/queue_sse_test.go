package web

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
	"github.com/rcarmo/gi/internal/turn"
)

func TestLegacySSEQueueInvalidationIsReadyAndSessionScoped(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	engine := turn.New(s)
	defer engine.Close()
	server := New(s, engine, config.RuntimeConfig{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest("GET", "/sse/stream?chat_jid=gi:A", nil).WithContext(ctx)
	// Use real streaming HTTP so connected is an observable readiness boundary.
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleSSEStream))
	defer httpServer.Close()
	clientReq, _ := http.NewRequestWithContext(ctx, "GET", httpServer.URL+req.URL.String(), nil)
	response, err := httpServer.Client().Do(clientReq)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	events := make(chan string, 16)
	go func() {
		defer close(events)
		scanner := bufio.NewScanner(response.Body)
		var frame strings.Builder
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				events <- frame.String()
				frame.Reset()
			} else {
				fmt.Fprintln(&frame, line)
			}
		}
	}()
	next := func() string {
		t.Helper()
		select {
		case event := <-events:
			return event
		case <-time.After(2 * time.Second):
			t.Fatal("SSE timeout")
			return ""
		}
	}
	if event := next(); !strings.Contains(event, "event: connected") {
		t.Fatal(event)
	}
	engine.Topics().Publish(topics.Envelope{Topic: "session.queue", SessionID: "B", Payload: map[string]any{"type": "queue_changed"}})
	engine.Topics().Publish(topics.Envelope{Topic: "session.queue", SessionID: "A", Payload: map[string]any{"type": "queue_changed"}})
	if event := next(); !strings.Contains(event, "event: queue_changed") || !strings.Contains(event, `"chat_jid":"gi:A"`) {
		t.Fatal(event)
	}
	engine.PublishRuntimeTurnEvent("turn_submitted", "A", "queued", "agent", "queued", "queue", nil)
	if event := next(); !strings.Contains(event, "event: queue_changed") {
		t.Fatal(event)
	}
	cancel()
}

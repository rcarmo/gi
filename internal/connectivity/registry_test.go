package connectivity

import (
	"context"
	"testing"
	"time"
)

func TestRegistryRegisterDeliverAndList(t *testing.T) {
	reg := NewRegistry()
	info, err := reg.Register(context.Background(), RouteSpec{Name: "hook", Transport: "http", Match: map[string]any{"topic": "http.hook"}}, func(ctx context.Context, event EventEnvelope) (RouteResponse, error) {
		if event.Topic != "http.hook" {
			t.Fatalf("unexpected topic: %s", event.Topic)
		}
		return RouteResponse{Status: 202, Body: "accepted"}, nil
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if info.ID == "" || info.Lifetime != "session" || info.Mode != "respond" {
		t.Fatalf("unexpected info: %#v", info)
	}

	routes, err := reg.List(context.Background(), map[string]any{"transport": "http"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(routes) != 1 || routes[0].ID != info.ID {
		t.Fatalf("unexpected routes: %#v", routes)
	}

	resp, err := reg.Deliver(context.Background(), info.ID, EventEnvelope{Payload: map[string]any{"ok": true}})
	if err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if resp.Status != 202 || resp.Body != "accepted" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestEventBusSubscribePattern(t *testing.T) {
	bus := NewEventBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := bus.Subscribe(ctx, "mqtt.home.*", 1)
	defer unsub()
	if err := bus.Emit(ctx, EventEnvelope{Topic: "mqtt.home.power", Payload: map[string]any{"w": 42}}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	select {
	case ev := <-ch:
		if ev.Topic != "mqtt.home.power" {
			t.Fatalf("unexpected event: %#v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

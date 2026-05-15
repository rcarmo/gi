package connectivity

import (
	"context"
	"testing"
	"time"
)

func TestEventBusUnsubscribeIsIdempotent(t *testing.T) {
	bus := NewEventBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, unsubscribe := bus.Subscribe(ctx, "*", 1)
	unsubscribe()
	unsubscribe()
}

func TestEventBusEmitDoesNotPanicDuringConcurrentUnsubscribe(t *testing.T) {
	bus := NewEventBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, unsubscribe := bus.Subscribe(ctx, "runtime.*", 1)
	panicCh := make(chan any, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if p := recover(); p != nil {
				panicCh <- p
			}
		}()
		for i := 0; i < 1000; i++ {
			_ = bus.Emit(context.Background(), EventEnvelope{Topic: "runtime.test", Timestamp: time.Now().UTC()})
		}
	}()
	unsubscribe()
	<-done
	select {
	case p := <-panicCh:
		t.Fatalf("emit panicked during concurrent unsubscribe: %v", p)
	default:
	}
}

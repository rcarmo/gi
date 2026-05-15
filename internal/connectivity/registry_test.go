package connectivity

import (
	"context"
	"testing"
)

func TestRegistryDefaultsRouteLifetimeToProcess(t *testing.T) {
	r := NewRegistry()
	info, err := r.Register(context.Background(), RouteSpec{Name: "demo", Transport: "http"}, nil)
	if err != nil {
		t.Fatalf("register route: %v", err)
	}
	if info.Lifetime != "process" {
		t.Fatalf("expected default lifetime process, got %q", info.Lifetime)
	}
}

func TestRegistryUnregisterIsIdempotentWhenMissing(t *testing.T) {
	r := NewRegistry()
	if err := r.Unregister(context.Background(), "route_missing"); err != nil {
		t.Fatalf("unregister missing route should be idempotent: %v", err)
	}
}

func TestRegistryUnregisterRemovesExistingRoute(t *testing.T) {
	r := NewRegistry()
	info, err := r.Register(context.Background(), RouteSpec{Name: "demo", Transport: "http"}, nil)
	if err != nil {
		t.Fatalf("register route: %v", err)
	}
	if err := r.Unregister(context.Background(), info.ID); err != nil {
		t.Fatalf("unregister route: %v", err)
	}
	if _, ok := r.Get(info.ID); ok {
		t.Fatal("expected route to be removed after unregister")
	}
	if err := r.Unregister(context.Background(), info.ID); err != nil {
		t.Fatalf("second unregister should be idempotent: %v", err)
	}
}

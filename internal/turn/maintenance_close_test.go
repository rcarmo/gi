package turn

import (
	"context"
	"testing"
	"time"
)

func TestEngineCloseJoinsMaintenance(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	e := &Engine{bgCtx: ctx, bgCancel: cancel}
	e.maintenance.Add(1)
	stopped := make(chan struct{})
	release := make(chan struct{})
	go func() { <-ctx.Done(); close(stopped); <-release; e.maintenance.Done() }()
	done := make(chan struct{})
	go func() { e.Close(); close(done) }()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("maintenance not cancelled")
	}
	select {
	case <-done:
		t.Fatal("Close returned while maintenance still writes")
	default:
	}
	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Close did not join maintenance")
	}
}

package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestToolSettlementAdmissionBoundary(t *testing.T) {
	if _, err := admitToolSettlement(nil); err == nil {
		t.Fatal("nil admission accepted")
	}
	parent, cancel := context.WithCancel(context.WithValue(context.Background(), "settlement-test", "value"))
	admitted, err := admitToolSettlement(parent)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	if admitted.Err() != nil || admitted.Done() != nil {
		t.Fatal("admitted settlement cancelled")
	}
	if admitted.Value("settlement-test") != "value" {
		t.Fatal("request context values lost")
	}
	if _, err := admitToolSettlement(parent); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled admission: %v", err)
	}
	expired, stop := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer stop()
	if _, err := admitToolSettlement(expired); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expired admission: %v", err)
	}
}

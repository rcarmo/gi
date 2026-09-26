package store

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestWebSendReceiptUpgradeReopenAndBoundedDuplicate(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "receipts.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.DB().Exec(`drop table web_send_receipts`); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]any{"turn_id": "target", "session_id": "B", "source_session_id": "A", "routed": true}
	if err = s.RecordWebSendReceipt(ctx, "A", "token", result); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	raw, ok, err := s.GetWebSendReceipt(ctx, "A", "token")
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	want, _ := json.Marshal(result)
	if string(raw) != string(want) {
		t.Fatal(string(raw))
	}
	if _, ok, err = s.GetWebSendReceipt(ctx, "B", "token"); err != nil || ok {
		t.Fatal("foreign", ok, err)
	}
	if err = s.RecordWebSendReceipt(ctx, "A", "token", result); err != nil {
		t.Fatal(err)
	}
	if _, ok, err = s.GetWebSendReceipt(ctx, "A", "token"); err != nil || ok {
		t.Fatal("duplicate", ok, err)
	}
	var detail string
	rows, err := s.DB().Query(`explain query plan select result_json from web_send_receipts where source_session_id='A' and client_request_id='token' order by id limit 2`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() {
		var a, b, c int
		if err = rows.Scan(&a, &b, &c, &detail); err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(detail, "idx_web_send_receipts_request") {
		t.Fatal("not indexed", detail)
	}
}

func TestWebSendReceiptConcurrentDuplicateIsUnknown(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "receipts.db")
	a, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if _, err = a.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	b, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	start := make(chan struct{})
	out := make(chan error, 2)
	var wg sync.WaitGroup
	for _, s := range []*Store{a, b} {
		wg.Add(1)
		go func(s *Store) {
			defer wg.Done()
			<-start
			out <- s.RecordWebSendReceipt(ctx, "A", "same", map[string]any{"turn_id": "turn", "session_id": "A"})
		}(s)
	}
	close(start)
	wg.Wait()
	close(out)
	for err := range out {
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, confirmed, err := a.GetWebSendReceipt(ctx, "A", "same"); err != nil || confirmed {
		t.Fatal("racing writes inferred one admission", confirmed, err)
	}
}

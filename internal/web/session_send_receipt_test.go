package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func sendReceipt(t *testing.T, mux http.Handler, session, query string, code int) map[string]any {
	t.Helper()
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/api/sessions/"+session+"/send-receipt?"+query, nil))
	if rr.Code != code {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}
func TestWebSendReceiptLocalRoutedSteeredAndDuplicate(t *testing.T) {
	for _, mode := range []string{"local", "directed", "target", "steered"} {
		t.Run(mode, func(t *testing.T) {
			s, st := newTestWebServer(t, t.TempDir())
			ctx := context.Background()
			mux := s.Handler()
			if _, err := st.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			if _, err := st.CreateSession(ctx, "B", "B", nil); err != nil {
				t.Fatal(err)
			}
			if _, err := st.CreateTurnWithStatus(ctx, "active", "A", "running", "active", nil); err != nil {
				t.Fatal(err)
			}
			if ok, err := st.ClaimSessionActiveTurn(ctx, "A", "active", "fixture", "token"); err != nil || !ok {
				t.Fatal(ok, err)
			}
			body := `{"prompt":"receipt proof","model":"test-model","intent":"queue","client_request_id":"request"}`
			if mode == "directed" {
				body = `{"prompt":"@peer receipt proof","model":"test-model","intent":"queue","client_request_id":"request"}`
			}
			if mode == "target" {
				body = `{"prompt":"receipt proof","target_agent_id":"peer","model":"test-model","intent":"queue","client_request_id":"request"}`
			}
			if mode == "steered" {
				body = `{"prompt":"receipt proof","model":"test-model","intent":"steer","client_request_id":"request"}`
			}
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, httptest.NewRequest("POST", "/api/sessions/A/prompt", strings.NewReader(body)))
			if rr.Code != 202 {
				t.Fatal(rr.Code, rr.Body.String())
			}
			var admitted map[string]any
			if err := json.Unmarshal(rr.Body.Bytes(), &admitted); err != nil {
				t.Fatal(err)
			}
			receipt := sendReceipt(t, mux, "A", "client_request_id=request", 200)
			if receipt["confirmed"] != true {
				t.Fatal(receipt)
			}
			expected, _ := json.Marshal(admitted)
			actual, _ := json.Marshal(receipt["result"])
			if string(expected) != string(actual) {
				t.Fatal("receipt not exact returned result", receipt, admitted)
			}
			if mode == "directed" || mode == "target" {
				if admitted["session_id"] == "A" || admitted["source_session_id"] != "A" {
					t.Fatal(admitted)
				}
			}
			if mode == "steered" && admitted["turn_id"] != "active" {
				t.Fatal(admitted)
			}
			if sendReceipt(t, mux, "B", "client_request_id=request", 200)["confirmed"] != false {
				t.Fatal("foreign receipt leaked")
			}
			if err := st.RecordWebSendReceipt(ctx, "A", "request", admitted); err != nil {
				t.Fatal(err)
			}
			if sendReceipt(t, mux, "A", "client_request_id=request", 200)["confirmed"] != false {
				t.Fatal("duplicate picked arbitrarily")
			}
		})
	}
}
func TestWebSendReceiptStrictQueriesAndFailedWrite(t *testing.T) {
	s, st := newTestWebServer(t, t.TempDir())
	ctx := context.Background()
	mux := s.Handler()
	st.CreateSession(ctx, "A", "A", nil)
	for _, q := range []string{"", "client_request_id=", "client_request_id=a&client_request_id=b", "client_request_id=a&session=B", "client_request_id=%00", "client_request_id=" + strings.Repeat("x", 129)} {
		sendReceipt(t, mux, "A", q, 400)
	}
	sendReceipt(t, mux, "missing", "client_request_id=a", 404)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("POST", "/api/sessions/A/send-receipt?client_request_id=a", nil))
	if rr.Code != 405 {
		t.Fatal(rr.Code)
	}
	if _, err := st.CreateTurnWithStatus(ctx, "active", "A", "running", "active", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := st.ClaimSessionActiveTurn(ctx, "A", "active", "fixture", "token"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if _, err := st.DB().Exec(`create trigger reject_receipt before insert on web_send_receipts begin select raise(abort,'no receipt'); end`); err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("POST", "/api/sessions/A/prompt", strings.NewReader(`{"prompt":"accepted","intent":"queue","client_request_id":"lost","model":"test-model"}`)))
	if rr.Code != 202 {
		t.Fatal("receipt failure rejected accepted prompt", rr.Code, rr.Body.String())
	}
	if sendReceipt(t, mux, "A", "client_request_id=lost", 200)["confirmed"] != false {
		t.Fatal("fabricated confirmation")
	}
	turns, err := st.ListTurns(ctx, "A")
	if err != nil || len(turns) != 2 {
		t.Fatal(turns, err)
	}
	// Validate before explicit-target dispatch too, not only local admission.
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("POST", "/api/sessions/A/prompt", strings.NewReader(`{"prompt":"invalid","target_agent_id":"peer","client_request_id":"`+strings.Repeat("x", 129)+`"}`)))
	if rr.Code != 400 {
		t.Fatal(rr.Code)
	}
}

func TestWebSendReceiptDeletedTargetAndUnsupportedSelectors(t *testing.T) {
	s, st := newTestWebServer(t, t.TempDir())
	ctx := context.Background()
	mux := s.Handler()
	if _, err := st.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateTurnWithStatus(ctx, "target", "A", "completed", "private text", nil); err != nil {
		t.Fatal(err)
	}
	if err := st.RecordWebSendReceipt(ctx, "A", "token", map[string]any{"turn_id": "target", "session_id": "A", "status": "completed"}); err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/api/sessions/A/send-receipt?client_request_id=token", nil))
	if rr.Code != 200 || rr.Header().Get("Cache-Control") != "no-store" || strings.Contains(rr.Body.String(), "private text") {
		t.Fatal(rr.Code, rr.Body.String())
	}
	if err := st.DeleteTurn(ctx, "target"); err != nil {
		t.Fatal(err)
	}
	if sendReceipt(t, mux, "A", "client_request_id=token", 200)["confirmed"] != false {
		t.Fatal("deleted target confirmed")
	}
	for _, q := range []string{"client_request_id=x&limit=100", "client_request_id=x&target_session_id=B", "client_request_id=%ZZ"} {
		sendReceipt(t, mux, "A", q, 400)
	}
}

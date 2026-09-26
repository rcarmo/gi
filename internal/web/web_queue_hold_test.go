package web

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebQueueHoldResumeHTTPContract(t *testing.T) {
	s, st := newTestWebServer(t, t.TempDir())
	ctx := context.Background()
	st.CreateSession(ctx, "A", "A", nil)
	for _, body := range []string{`{}`, `{"stop_turn_id":""}`, `{"stop_turn_id":"x","extra":true}`, `{"stop_turn_id":"x"}{}`} {
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, httptest.NewRequest("POST", "/api/sessions/A/resume-queue", strings.NewReader(body)))
		if rr.Code != 400 {
			t.Fatal(rr.Code, body)
		}
	}
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/api/sessions/A/resume-queue", nil))
	if rr.Code != 405 {
		t.Fatal(rr.Code)
	}
	if _, err := st.DB().Exec(`insert into web_queue_holds values('A','stopped',datetime('now'))`); err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/api/sessions/A/activity", nil))
	var activity map[string]any
	json.Unmarshal(rr.Body.Bytes(), &activity)
	if activity["queue_hold_turn_id"] != "stopped" || activity["status"] != "idle" {
		t.Fatal(activity)
	}
	for _, id := range []string{"stale", "stopped", "stopped"} {
		rr = httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, httptest.NewRequest("POST", "/api/sessions/A/resume-queue", strings.NewReader(`{"stop_turn_id":"`+id+`"}`)))
		if id == "stale" {
			if rr.Code != 409 {
				t.Fatal(rr.Code)
			}
		} else {
			if rr.Code != 200 {
				t.Fatal(rr.Code, rr.Body.String())
			}
			break
		}
	}
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest("POST", "/api/sessions/A/resume-queue", strings.NewReader(`{"stop_turn_id":"stopped"}`)))
	if rr.Code != 409 {
		t.Fatal(rr.Code)
	}
}

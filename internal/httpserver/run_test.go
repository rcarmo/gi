package httpserver

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func listener(t *testing.T, handler http.Handler) (Listener, string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	srv := &http.Server{Handler: handler}
	return Listener{Server: srv, Serve: func() error { return srv.Serve(ln) }, Label: "test"}, "http://" + ln.Addr().String()
}
func await(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out")
	}
}
func TestRunJoinsGracefulHandlersAndAllListeners(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	entered, release, stopped := make(chan struct{}), make(chan struct{}), make(chan struct{})
	primary, url := listener(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-release; io.WriteString(w, "drained") }))
	secondary, _ := listener(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	var served atomic.Int32
	for _, l := range []*Listener{&primary, &secondary} {
		serve := l.Serve
		l.Serve = func() error { defer served.Add(1); return serve() }
	}
	done := make(chan error, 1)
	go func() { done <- Run(ctx, func() { cancel(); close(stopped) }, time.Second, primary, secondary) }()
	requestDone := make(chan error, 1)
	go func() {
		res, err := http.Get(url)
		if err == nil {
			defer res.Body.Close()
			_, err = io.ReadAll(res.Body)
		}
		requestDone <- err
	}()
	await(t, entered)
	cancel()
	await(t, stopped)
	select {
	case err := <-done:
		t.Fatal("returned before handler drained", err)
	default:
	}
	close(release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("drain stuck")
	}
	if err := <-requestDone; err != nil {
		t.Fatal(err)
	}
	if served.Load() != 2 {
		t.Fatal("secondary listener abandoned", served.Load())
	}
}

func TestRunTimeoutCancelsButStillJoinsHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	entered, cancelled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	l, url := listener(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
		close(cancelled)
		<-release
	}))
	done := make(chan error, 1)
	go func() { done <- Run(ctx, cancel, 20*time.Millisecond, l) }()
	requestDone := make(chan struct{})
	go func() {
		defer close(requestDone)
		res, err := http.Get(url)
		if err == nil {
			res.Body.Close()
		}
	}()
	await(t, entered)
	cancel()
	await(t, cancelled)
	select {
	case err := <-done:
		t.Fatal("timeout abandoned handler", err)
	default:
	}
	close(release)
	if err := <-done; !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	await(t, requestDone)
}

func TestRunUnexpectedListenerFailureStopsAndJoinsOthers(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	l, _ := listener(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	sentinel := errors.New("bind failed")
	failing := Listener{Server: &http.Server{}, Serve: func() error { return sentinel }, Label: "secondary"}
	if err := Run(ctx, cancel, time.Second, l, failing); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	if ctx.Err() == nil {
		t.Fatal("application not stopped")
	}
}

package web

import (
	"context"
	"net"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
)

func TestVNCInitialGreetingTimeoutAndCloseDoNotLeakWorkers(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		c, err := listener.Accept()
		if err == nil {
			accepted <- c
		}
	}()
	host, port, _ := net.SplitHostPort(listener.Addr().String())
	n, _ := strconv.Atoi(port)
	srv := New(nil, nil, config.RuntimeConfig{WorkspaceRoot: t.TempDir(), VNCTargets: []config.VNCTarget{{ID: "silent", Host: host, Port: n}}})
	srv.vnc.greetingTimeout = 30 * time.Millisecond
	server := httptest.NewServer(srv.Handler())
	defer server.Close()
	defer srv.CloseVNC()
	c := vncDial(t, server, "silent", "")
	vncRead(t, c)
	upstream := <-accepted
	defer upstream.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _, err = c.Read(ctx)
	if err == nil || ctx.Err() != nil {
		t.Fatal("silent upstream did not time out", err)
	}
	done := make(chan struct{})
	go func() { srv.CloseVNC(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("shutdown blocked on reader")
	}
}

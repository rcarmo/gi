// Package httpserver joins HTTP listeners and handlers before their owners
// release application resources. Shutdown's deadline limits graceful draining,
// not the wait for a handler that ignores cancellation.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Listener struct {
	Server *http.Server
	Serve  func() error
	Label  string
}

// Run serves all listeners until cancellation or the first listener exit. Stop
// cancels application work on either path. All handlers, shutdown calls and
// serve goroutines are joined before return, including secondary ACME listeners.
// Listeners must not already be serving; their handlers are wrapped for joining.
func Run(ctx context.Context, stop context.CancelFunc, timeout time.Duration, listeners ...Listener) error {
	if len(listeners) == 0 {
		return fmt.Errorf("HTTP runner requires a listener")
	}
	var mu sync.Mutex
	var handlers sync.WaitGroup
	stopping := false
	for _, listener := range listeners {
		handler := listener.Server.Handler
		if handler == nil {
			handler = http.DefaultServeMux
		}
		listener.Server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			if stopping {
				mu.Unlock()
				http.Error(w, "Server stopping", http.StatusServiceUnavailable)
				return
			}
			handlers.Add(1)
			mu.Unlock()
			defer handlers.Done()
			handler.ServeHTTP(w, r)
		})
	}
	results := make(chan error, len(listeners))
	for _, listener := range listeners {
		go func() {
			err := listener.Serve()
			if errors.Is(err, http.ErrServerClosed) {
				err = nil
			}
			if err != nil {
				err = fmt.Errorf("listen %s: %w", listener.Label, err)
			}
			results <- err
		}()
	}
	var result error
	remaining := len(listeners)
	select {
	case <-ctx.Done():
	case result = <-results:
		remaining--
	}
	mu.Lock()
	stopping = true
	mu.Unlock()
	stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var shutdowns sync.WaitGroup
	shutdownErrors := make(chan error, len(listeners))
	for _, listener := range listeners {
		shutdowns.Add(1)
		go func() {
			defer shutdowns.Done()
			if err := listener.Server.Shutdown(shutdownCtx); err != nil {
				shutdownErrors <- fmt.Errorf("shutdown %s: %w", listener.Label, err)
				_ = listener.Server.Close() // cancel remaining network request contexts
			}
		}()
	}
	shutdowns.Wait()
	close(shutdownErrors)
	for err := range shutdownErrors {
		result = errors.Join(result, err)
	}
	for i := 0; i < remaining; i++ {
		result = errors.Join(result, <-results)
	}
	handlers.Wait()
	return result
}

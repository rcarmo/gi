// Package lockdir holds proper-lockfile's lock: a "<file>.lock" directory,
// refreshed while held and taken over when stale. Pi locks its shared files
// (auth.json, mcp-auth.json, models-store.json) this way, so gi and Pi do
// not interleave writes.
package lockdir

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const retry = 100 * time.Millisecond

// With runs fn holding lockPath, waiting up to wait for another holder; a
// lock not refreshed for stale was abandoned by a killed process.
func With(lockPath string, stale, wait time.Duration, fn func() error) error {
	return WithContext(context.Background(), lockPath, stale, wait, fn)
}

// WithContext is With with cancellable acquisition; fn owns the acquired lock.
func WithContext(ctx context.Context, lockPath string, stale, wait time.Duration, fn func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o700); err != nil {
		return err
	}
	deadline := time.Now().Add(wait)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := os.Mkdir(lockPath, 0o700)
		if err == nil {
			break
		}
		if !os.IsExist(err) {
			return err
		}
		if info, statErr := os.Stat(lockPath); statErr == nil && time.Since(info.ModTime()) > stale {
			_ = os.Remove(lockPath)
			continue
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("lock %s is held", lockPath)
		}
		timer := time.NewTimer(retry)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	done := make(chan struct{})
	go func() { // keep the lock fresh, like proper-lockfile's update
		ticker := time.NewTicker(stale / 2)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case now := <-ticker.C:
				_ = os.Chtimes(lockPath, now, now)
			}
		}
	}()
	defer func() {
		close(done)
		_ = os.Remove(lockPath)
	}()
	return fn()
}

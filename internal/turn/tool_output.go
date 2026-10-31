package turn

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

// Each reporter owns its coalescing timer. Closing joins any in-flight write;
// terminal replacement bypasses pacing, even for an empty final result.
func (r *sessionRunner) toolOutputReporter(ctx context.Context, turnID, sessionID, callID, occurrence string) (func(string, bool) error, func()) {
	ctx = context.WithoutCancel(ctx)
	var mu sync.Mutex
	var workers sync.WaitGroup
	var timer *time.Timer
	var generation uint64
	var next time.Time
	var failure error
	var previous, pending string
	var stopped, dirty bool
	write := func(text string, final bool) {
		if failure != nil || (!final && text == previous) {
			return
		}
		preview := store.ToolOutputPreview(text)
		if preview == nil {
			preview = map[string]any{"output_preview": "", "output_total_lines": 0, "output_preview_lines": 0, "output_truncated": false}
		}
		preview["tool_call_id"], preview["occurrence_id"], preview["final"] = callID, occurrence, final
		failure = r.store.AppendTurnEffectEvent(ctx, turnID, sessionID, "tool.output", preview)
		if failure != nil {
			return
		}
		previous = text
		// Pace committed writes by size, with the native 100ms minimum.
		bytes := len(preview["output_preview"].(string))
		delay := time.Duration(bytes) * time.Second / (100 * 1024)
		if delay < 100*time.Millisecond {
			delay = 100 * time.Millisecond
		}
		next = time.Now().Add(delay)
		r.engine.broadcast(sessionID, map[string]any{"type": "tool_activity_changed", "chat_jid": "gi:" + sessionID, "turn_id": turnID})
	}
	stopTimer := func() {
		generation++
		if timer != nil {
			if timer.Stop() {
				workers.Done()
			}
			timer = nil
		}
	}
	report := func(text string, final bool) error {
		mu.Lock()
		if stopped {
			err := failure
			mu.Unlock()
			return err
		}
		pending, dirty = text, true
		if final {
			stopped = true
			stopTimer()
			write(pending, true)
			pending, previous = "", ""
			err := failure
			mu.Unlock()
			workers.Wait()
			return err
		}
		if failure == nil {
			if !time.Now().Before(next) {
				stopTimer()
				write(pending, false)
				dirty = false
			} else if timer == nil {
				workers.Add(1)
				current := generation
				timer = time.AfterFunc(time.Until(next), func() {
					defer workers.Done()
					mu.Lock()
					defer mu.Unlock()
					if current != generation {
						return
					}
					timer = nil
					if !stopped && dirty {
						write(pending, false)
						dirty = false
					}
				})
			}
		}
		err := failure
		mu.Unlock()
		return err
	}
	close := func() {
		mu.Lock()
		if !stopped {
			stopped = true
			stopTimer()
			if dirty {
				write(pending, true)
			}
		}
		pending, previous = "", ""
		mu.Unlock()
		workers.Wait()
	}
	return report, close
}

// firstLine is the first line of s, for log lines about multi-line errors.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i] + " …"
	}
	return s
}

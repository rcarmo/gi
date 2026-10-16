package turn

import "context"

// ResetSession admits an in-place history reset under the same runner lock used
// by submissions; busy requests remain durable until a safe execution boundary.
func (e *Engine) ResetSession(ctx context.Context, sessionID, text string) (bool, error) {
	r := e.runner(sessionID)
	r.mu.Lock()
	defer r.mu.Unlock()
	return e.store.RequestContextReset(ctx, sessionID, text)
}

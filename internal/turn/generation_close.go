package turn

import (
	"context"
	"errors"

	"github.com/rcarmo/gi/internal/store"
)

// The owned provider may flush its final bounded snapshot after cancellation.
// Keep this narrow capability in the recorder, not in the provider/hook context:
// it cannot admit requests, settle a response or authorise tool execution. The
// controller does not publish the checkpoint until the recorder and host join.
func (r *sessionRunner) providerPartialPersistence(ctx context.Context, s *store.Store, sessionID, turnID, occurrenceID string) func(string, map[string]any) error {
	if r.sharedSessionCoord == nil {
		return nil
	}
	r.mu.Lock()
	owner := r.current
	r.mu.Unlock()
	if owner == nil || owner.turnID != turnID {
		return nil
	}
	writer := func() (context.Context, bool) {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.current == owner && owner.closingProgressToken != "" {
			return store.WithTurnWriter(ctx, sessionID, turnID, owner.closingProgressToken), true
		}
		return ctx, false
	}
	return func(content string, payload map[string]any) error {
		current, closing := writer()
		err := s.SaveProviderPartial(current, turnID, sessionID, occurrenceID, content, payload)
		if !closing && errors.Is(err, store.ErrQueueConflict) {
			// Reservation may linearize between reading the old token and writing.
			// Retry only that ownership transition, once; never borrow a newer host.
			if reserved, ok := writer(); ok {
				return s.SaveProviderPartial(reserved, turnID, sessionID, occurrenceID, content, payload)
			}
		}
		return err
	}
}

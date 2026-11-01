package turn

import (
	"context"
	"errors"

	"github.com/rcarmo/gi/internal/store"
)

// checkpointProgressPersistence binds bounded recorders to this physical host.
// The closing incarnation is never exposed to tools or used for settlements.
// A replacement host on the same logical turn cannot inherit these callbacks.
func (r *sessionRunner) checkpointProgressPersistence(ctx context.Context, sessionID, turnID string) func(func(context.Context) error) error {
	ctx = context.WithoutCancel(ctx)
	if r.sharedSessionCoord == nil {
		return func(persist func(context.Context) error) error { return persist(ctx) }
	}
	r.mu.Lock()
	owner := r.current
	r.mu.Unlock()
	closingContext := func() (context.Context, bool) {
		r.mu.Lock()
		defer r.mu.Unlock()
		if owner != nil && r.current == owner && owner.turnID == turnID && owner.closingProgressToken != "" {
			return store.WithTurnWriter(ctx, sessionID, turnID, owner.closingProgressToken), true
		}
		return ctx, false
	}
	return func(persist func(context.Context) error) error {
		writeCtx, closing := closingContext()
		err := persist(writeCtx)
		// Reservation can win between reading the token and the DB write.
		if errors.Is(err, store.ErrQueueConflict) && !closing {
			if next, available := closingContext(); available {
				err = persist(next)
			}
		}
		if err != nil {
			r.mu.Lock()
			if owner != nil && r.current == owner && owner.turnID == turnID && owner.closingProgressToken != "" {
				owner.closingProgressErr = errors.Join(owner.closingProgressErr, err)
			}
			r.mu.Unlock()
		}
		return err
	}
}

func (r *sessionRunner) providerPartialPersistence(ctx context.Context, s *store.Store, sessionID, turnID, requestID string) func(string, map[string]any) error {
	if r.sharedSessionCoord == nil {
		return nil
	}
	r.mu.Lock()
	owner := r.current
	r.mu.Unlock()
	if owner == nil || owner.turnID != turnID {
		return nil
	}
	persist := r.checkpointProgressPersistence(ctx, sessionID, turnID)
	return func(content string, payload map[string]any) error {
		return persist(func(writeCtx context.Context) error {
			return s.SaveProviderPartial(writeCtx, turnID, sessionID, requestID, content, payload)
		})
	}
}

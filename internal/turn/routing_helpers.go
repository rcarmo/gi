package turn

import (
	"context"

	"github.com/rcarmo/gi/internal/routing/audit"
)

func (e *Engine) recordRouteDecision(ctx context.Context, sourceSessionID, turnID string, metadata map[string]any) error {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	return audit.RecordDecision(opCtx, e.store, sourceSessionID, turnID, metadata, audit.Options{
		PublishRuntimeRoutingEvent: e.PublishRuntimeRoutingEvent,
		Broadcast:                  e.broadcast,
	})
}

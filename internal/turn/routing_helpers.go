package turn

import (
	"context"

	"github.com/rcarmo/gi/internal/turn/routeaudit"
)

func (e *Engine) recordRouteDecision(ctx context.Context, sourceSessionID, turnID string, metadata map[string]any) error {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	return routeaudit.RecordDecision(opCtx, e.store, sourceSessionID, turnID, metadata, routeaudit.Options{
		PublishRuntimeRoutingEvent: e.PublishRuntimeRoutingEvent,
		Broadcast:                  e.broadcast,
	})
}

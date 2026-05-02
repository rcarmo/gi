package indexer

import "context"

// Rebuilder captures the full rebuild surface for search indexes.
type Rebuilder interface {
	Rebuild(ctx context.Context) error
}

func RebuildAll(ctx context.Context, svc Rebuilder) error {
	return svc.Rebuild(ctx)
}

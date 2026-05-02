package indexer

import "context"

// PathReindexer captures the minimal reindex surface used by the walker layer.
type PathReindexer interface {
	ReindexPath(ctx context.Context, path string) error
}

func ReindexPaths(ctx context.Context, svc PathReindexer, paths []string) error {
	for _, path := range paths {
		if err := svc.ReindexPath(ctx, path); err != nil {
			return err
		}
	}
	return nil
}

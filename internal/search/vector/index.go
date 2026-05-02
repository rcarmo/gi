package vector

import "context"

type VectorRecord struct {
	ChunkID   int64
	Embedding []float32
}

type Hit struct {
	ChunkID int64
	Score   float64
}

// Index abstracts semantic vector lookup.
type Index interface {
	Upsert(ctx context.Context, rows []VectorRecord) error
	DeleteByChunkIDs(ctx context.Context, ids []int64) error
	Search(ctx context.Context, embedding []float32, k int) ([]Hit, error)
}

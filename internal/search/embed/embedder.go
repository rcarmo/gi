package embed

import "context"

// Embedder turns text into a dense vector.
type Embedder interface {
	Name() string
	Dimension() int
	Embed(ctx context.Context, text string) ([]float32, error)
}

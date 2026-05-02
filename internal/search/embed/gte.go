package embed

import (
	"context"
	"errors"
)

var ErrGTEUnavailable = errors.New("embed/gte: implementation scaffold only")

// GTEEmbedder is the planned gte-go backed embedder.
type GTEEmbedder struct{}

func NewGTEEmbedder() *GTEEmbedder { return &GTEEmbedder{} }

func (e *GTEEmbedder) Name() string   { return "gte-small" }
func (e *GTEEmbedder) Dimension() int { return 384 }
func (e *GTEEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	return nil, ErrGTEUnavailable
}

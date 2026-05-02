package search

import (
	"context"
	"errors"
)

var ErrNotImplemented = errors.New("search: not implemented")

// Dependencies groups the main pluggable search backends.
type Dependencies struct {
	FTS           FTSSearcher
	Vec           VectorSearcher
	Store         IndexStore
	QueryEmbedder QueryEmbedder
}

// FTSSearcher returns lexical hits.
type FTSSearcher interface {
	SearchFTS(ctx context.Context, q SearchQuery) ([]SearchHit, error)
}

// VectorSearcher returns semantic hits.
type VectorSearcher interface {
	SearchVector(ctx context.Context, q SearchQuery, embedding []float32) ([]SearchHit, error)
}

// QueryEmbedder embeds query text for vector lookup.
type QueryEmbedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// IndexStore owns reindex/rebuild persistence work.
type IndexStore interface {
	ReindexPath(ctx context.Context, path string) error
	Rebuild(ctx context.Context) error
}

// HybridService is the default Service implementation.
type HybridService struct {
	deps Dependencies
}

func NewHybridService(deps Dependencies) *HybridService {
	return &HybridService{deps: deps}
}

func (s *HybridService) Search(ctx context.Context, q SearchQuery) ([]SearchHit, error) {
	var ftsHits, vecHits []SearchHit
	if q.UseFTS && s.deps.FTS != nil {
		hits, err := s.deps.FTS.SearchFTS(ctx, q)
		if err != nil {
			return nil, err
		}
		ftsHits = hits
	}
	if q.UseVector && s.deps.Vec != nil && s.deps.QueryEmbedder != nil {
		emb, err := s.deps.QueryEmbedder.Embed(ctx, q.Text)
		if err != nil {
			return nil, err
		}
		hits, err := s.deps.Vec.SearchVector(ctx, q, emb)
		if err != nil {
			return nil, err
		}
		vecHits = hits
	}
	if !q.UseFTS && !q.UseVector {
		return nil, ErrNotImplemented
	}
	return MergeAndRank(ftsHits, vecHits, q.Limit), nil
}

func (s *HybridService) ReindexPath(ctx context.Context, path string) error {
	if s.deps.Store == nil {
		return ErrNotImplemented
	}
	return s.deps.Store.ReindexPath(ctx, path)
}

func (s *HybridService) Rebuild(ctx context.Context) error {
	if s.deps.Store == nil {
		return ErrNotImplemented
	}
	return s.deps.Store.Rebuild(ctx)
}

package search

import "context"

// DocumentRecord describes an indexed workspace file.
type DocumentRecord struct {
	ID          int64
	Path        string
	Kind        string
	Language    string
	SizeBytes   int64
	MTimeNS     int64
	ContentHash string
	ChunkCount  int
	IndexState  string
	LastError   string
	IndexedAtMS int64
}

// ChunkRecord describes one searchable chunk for a document.
type ChunkRecord struct {
	ID               int64
	DocumentID       int64
	ChunkIndex       int
	StartByte        int
	EndByte          int
	StartLine        int
	EndLine          int
	TokenEstimate    int
	Heading          string
	Content          string
	EmbeddingVersion string
}

// VectorRecord binds a chunk id to an embedding.
type VectorRecord struct {
	ChunkID    int64
	Embedding  []float32
	Dimensions int
}

// SearchQuery captures a hybrid search request.
type SearchQuery struct {
	Text       string
	Limit      int
	ScopePaths []string
	Language   string
	UseFTS     bool
	UseVector  bool
}

// SearchHit is the merged output of FTS + vector retrieval.
type SearchHit struct {
	ChunkID    int64
	DocumentID int64
	Path       string
	Language   string
	Heading    string
	Content    string
	StartLine  int
	EndLine    int
	FTSScore   float64
	VecScore   float64
	FinalScore float64
}

// Service is the high-level hybrid indexing/query surface.
type Service interface {
	Search(ctx context.Context, q SearchQuery) ([]SearchHit, error)
	ReindexPath(ctx context.Context, path string) error
	Rebuild(ctx context.Context) error
}

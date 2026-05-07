# Workspace search: vec + FTS

This document describes the proposed internal layout for gi's hybrid workspace search subsystem.

## Runtime virtual search namespace

In addition to this package-level architecture, gi now exposes a read-only virtual search surface via `fts://...` locators.

See:
- `docs/internal/search/fts-namespace.md`
- `docs/internal/vfs/chat-projection.md`

This enables model-friendly retrieval paths and source linking without requiring SQL queries in prompts/tool calls.

## Current implementation status

Implemented now:

- ADR and internal design docs
- `internal/search/` package scaffold
- schema definitions for workspace search tables
- query classification and hybrid rank helper scaffolding
- chunking/embed/vector/indexer interfaces

Still pending:

- wiring the search schema into the main DB initialization path
- real `gte-go` embedding implementation
- real `sqlite-vec` backend implementation
- real chunk persistence / FTS updates / hybrid query execution against the database

## Goals

The subsystem should provide:

- exact/token-aware search over workspace contents
- semantic search over workspace chunks
- incremental indexing with SQLite as the source of truth
- stable internal APIs that can be used by web, TUI, tools, and the agent loop

## Core components

- **chunker** — splits files into stable searchable chunks
- **embedder** — turns text into dense vectors
- **index store** — persists documents/chunks/metadata in SQLite
- **fts index** — lexical retrieval via FTS5
- **vec index** — semantic retrieval via `sqlite-vec`
- **hybrid query engine** — runs both retrieval paths and reranks results

## Recommended package layout

```text
internal/search/
  types.go            # shared query/hit/document/chunk types
  service.go          # high-level indexing/search orchestration
  rank.go             # merge + rerank logic
  query.go            # query heuristics and search entrypoints
  search_test.go

internal/search/chunking/
  chunker.go          # Chunker interface
  markdown.go         # heading-aware markdown/doc chunking
  code.go             # code/window chunking
  text.go             # generic text chunking
  classify.go         # file-kind/language helpers
  chunker_test.go

internal/search/embed/
  embedder.go         # Embedder interface
  gte.go              # gte-go implementation
  gte_test.go

internal/search/store/
  schema.go           # schema helpers / migrations for search tables
  documents.go        # workspace_documents operations
  chunks.go           # workspace_chunks operations
  fts.go              # FTS maintenance helpers
  meta.go             # version metadata
  store_test.go

internal/search/vector/
  index.go            # VectorIndex interface
  sqlitevec.go        # sqlite-vec implementation
  sqlitevec_test.go

internal/search/indexer/
  walker.go           # filesystem walk/filtering
  reindex.go          # incremental reindexing pipeline
  rebuild.go          # full rebuild helpers
  indexer_test.go
```

## Primary interfaces

### Chunker

```go
type Chunk struct {
    ChunkIndex    int
    StartByte     int
    EndByte       int
    StartLine     int
    EndLine       int
    TokenEstimate int
    Heading       string
    Content       string
}

type Chunker interface {
    Version() string
    Chunk(path string, data []byte) ([]Chunk, error)
}
```

### Embedder

```go
type Embedder interface {
    Name() string
    Dimension() int
    Embed(ctx context.Context, text string) ([]float32, error)
}
```

### Vector index

```go
type VectorHit struct {
    ChunkID int64
    Score   float64
}

type VectorIndex interface {
    Upsert(ctx context.Context, rows []VectorRecord) error
    DeleteByChunkIDs(ctx context.Context, ids []int64) error
    Search(ctx context.Context, embedding []float32, k int) ([]VectorHit, error)
}
```

### Search service

```go
type SearchQuery struct {
    Text       string
    Limit      int
    ScopePaths []string
    Language   string
    UseFTS     bool
    UseVector  bool
}

type SearchHit struct {
    ChunkID    int64
    DocumentID int64
    Path       string
    Heading    string
    Content    string
    StartLine  int
    EndLine    int
    FTSScore   float64
    VecScore   float64
    FinalScore float64
}

type Service interface {
    Search(ctx context.Context, q SearchQuery) ([]SearchHit, error)
    ReindexPath(ctx context.Context, path string) error
    Rebuild(ctx context.Context) error
}
```

## Proposed schema

### `workspace_documents`

```sql
CREATE TABLE workspace_documents (
  id INTEGER PRIMARY KEY,
  path TEXT NOT NULL UNIQUE,
  kind TEXT NOT NULL,
  language TEXT NOT NULL DEFAULT '',
  size_bytes INTEGER NOT NULL,
  mtime_ns INTEGER NOT NULL,
  content_hash TEXT NOT NULL,
  chunk_count INTEGER NOT NULL DEFAULT 0,
  index_state TEXT NOT NULL DEFAULT 'ready',
  last_error TEXT NOT NULL DEFAULT '',
  indexed_at_ms INTEGER NOT NULL
);
```

### `workspace_chunks`

```sql
CREATE TABLE workspace_chunks (
  id INTEGER PRIMARY KEY,
  document_id INTEGER NOT NULL REFERENCES workspace_documents(id) ON DELETE CASCADE,
  chunk_index INTEGER NOT NULL,
  start_byte INTEGER NOT NULL,
  end_byte INTEGER NOT NULL,
  start_line INTEGER NOT NULL DEFAULT 0,
  end_line INTEGER NOT NULL DEFAULT 0,
  token_estimate INTEGER NOT NULL DEFAULT 0,
  heading TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL,
  embedding_version TEXT NOT NULL,
  UNIQUE(document_id, chunk_index)
);
```

### `workspace_chunks_fts`

```sql
CREATE VIRTUAL TABLE workspace_chunks_fts USING fts5(
  content,
  heading,
  path,
  language,
  tokenize = 'unicode61'
);
```

### `workspace_index_meta`

```sql
CREATE TABLE workspace_index_meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
```

### `workspace_chunks_vec`

Exact DDL depends on `sqlite-vec`, but the join key should be `chunk_id` and the default embedding dimension is `384`.

## Query flow

1. classify the query as lexical-heavy, semantic-heavy, or mixed
2. run FTS retrieval
3. run vec retrieval
4. merge hits by `chunk_id`
5. normalize scores and compute a final rank
6. optionally collapse repeated hits from the same file

### Suggested initial rank formula

```text
final_score =
  0.55 * vec_norm +
  0.35 * fts_norm +
  0.05 * heading_boost +
  0.05 * path_boost
```

## Chunking defaults

### Markdown/docs/notes
- heading-aware split
- roughly 400–800 words
- 80–120 word overlap

### Code
- try light symbol-aware splitting first
- fallback to ~120 line windows
- 20–30 line overlap

### Generic text
- paragraph/window based chunks

## Incremental indexing policy

Reindex a file when any of these change:

- file path added/removed
- file mtime/hash changed
- embedding version changed
- chunker version changed

Support both:

- explicit rebuild command
- targeted path reindex

## Initial migration plan

1. add search tables
2. add index metadata/version rows
3. add a manual rebuild command
4. add hybrid query API
5. add background refresh later

## Notes

For v1, prefer boring and inspectable behavior over an overly clever indexer. Debuggability is more important than maximizing recall immediately.

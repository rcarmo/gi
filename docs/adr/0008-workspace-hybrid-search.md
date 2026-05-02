# ADR 0008: Workspace Hybrid Search with Vec + FTS

- Status: Draft
- Date: 2026-04-28

## Decision

gi will add a **hybrid workspace search subsystem** built from:

- **SQLite** as the source of truth for indexed documents and chunks
- **SQLite FTS5** for lexical search
- **`sqlite-vec`** for vector similarity search
- **`rcarmo/gte-go`** for local text embeddings

The system will index workspace files into chunks, store chunk metadata in SQLite, store chunk text in FTS5, and store embeddings in a vector index keyed by chunk id.

Search queries will run against both FTS and vec, then merge and rerank results.

## Why

We want both:

- **exact/token-aware retrieval** for symbols, filenames, identifiers, literals, and grep-like queries
- **semantic retrieval** for fuzzy or paraphrased questions like “where do we handle tool retries”

FTS alone is not enough for semantic lookup. Vec alone is not enough for precise symbol/path matching. Hybrid search gives the best first production shape while keeping gi entirely local and Go-native.

## Embeddings

The default embedding model is **GTE-small** via `github.com/rcarmo/gte-go`.

Rationale:

- pure Go implementation
- local/in-process operation
- no Python or model-server sidecar
- small 384-dimensional embeddings suitable for local indexing
- predictable deployment and latency characteristics

The initial embedding contract is:

- model id: `gte-small`
- dimension: `384`
- normalized embeddings
- cosine similarity as the default semantic metric

## Storage model

The canonical search-index state will live in the main gi SQLite database.

The search subsystem will maintain:

- a file-level table for indexed workspace documents
- a chunk-level table for searchable segments
- an FTS5 table mirroring searchable chunk text
- a vec table keyed by chunk id
- metadata describing embedding/chunker/index versions

This keeps indexing state transactional and inspectable alongside the rest of gi runtime state.

## Initial query model

Each user query runs in three stages:

1. **FTS retrieval** over chunk text, headings, path, and language
2. **vector retrieval** over chunk embeddings
3. **merge + rerank** into a final hit list

Initial ranking will be a weighted hybrid score rather than a complex learned ranker.

## Chunking policy

The initial chunking design favors deterministic and debuggable rules over aggressive structure inference.

- markdown/docs/notes: heading-aware chunks with size caps
- code: symbol-ish splitting when easy, otherwise line windows with overlap
- generic text: paragraph/window chunks

Chunk metadata must retain enough information to reopen the file at the relevant location.

## Operational policy

The subsystem will support:

- incremental reindexing by file hash/mtime
- full rebuild when embedding or chunker versions change
- selective indexing of supported workspace file types
- conservative size/binary filtering

## Rejected alternatives

### FTS only

Rejected because it does not provide semantic recall for natural-language queries.

### Vec only

Rejected because it performs poorly for exact symbol/path/token lookup and weakens inspectability.

### External vector database first

Rejected for v1 because it complicates deployment and undermines the local SQLite-centric architecture.

### Custom ANN engine first

Rejected because there are already viable Go-native components and we do not yet have evidence that a custom engine is required.

## Preferred implementation path

### Phase 1

- add schema for documents/chunks/fts/meta
- add chunking + embedding pipeline
- add `sqlite-vec` integration
- add a hybrid search API inside gi

### Phase 2

- incremental/background reindexing
- query heuristics for lexical-vs-semantic weighting
- result collapsing and improved ranking

### Phase 3

- branch/session-aware search scopes
- richer code-aware chunking
- optional future fallback/alternate vector backends

## Consequences

Positive:

- stronger retrieval for both agents and humans
- all-Go, local-first deployment path
- preserves SQLite as the primary runtime store
- can evolve ranking independently from storage

Negative:

- more schema and indexing complexity than plain FTS
- chunking quality materially affects recall
- vector indexing introduces rebuild/versioning concerns

## Follow-up

This ADR should be implemented together with:

- a concrete schema migration
- an `internal/search` package layout
- internal docs for search/indexing behavior under `docs/internal/`

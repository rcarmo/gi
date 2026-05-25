# Shared media ingestion contract

Status: design contract for the first shared ingestion slice. TUI image paste and provider multimodal projection depend on this contract and must not introduce a separate payload shape.

## Goals

- Use one durable media reference model for web, TUI, API/direct ingress, steering, turns, and provider request projection.
- Keep SQLite/WAL as the canonical runtime truth; JSON fields may carry compatibility projections but are not the source of ownership or lifecycle policy.
- Preserve existing text-only API contracts. Media support is additive and optional.
- Keep transcript rendering deterministic and terminal-friendly: media references are metadata plus plain text placeholders, never raw binary in transcripts.

## Current inventory

Existing media-adjacent surfaces before this contract:

- `media` table (`internal/store/schema.go`) stores session-owned blobs in SQLite with filename, content type, metadata JSON, original size, compressed size, compression flag, content bytes, and timestamps.
- Store APIs (`internal/store/media.go`) expose `CreateMedia`, `GetMedia`, and `GetMediaContent`; these already compress/decompress blobs and attach arbitrary metadata.
- `steering_queue.media_json` stores a JSON array of strings and `SteeringMessage.Media` exposes it as `[]string`.
- Turn steering paths preserve `media` from metadata into queued steering and later into user-message payloads, but only as opaque strings today.
- User-visible messages currently have no first-class relational message/media join. `messages.payload_json.media` may carry the compatibility list when steering is injected.
- `/api/sessions/{id}/prompt` is text-only today. It must remain valid with no `media` field.
- Provider payload construction treats media only as a placeholder note today (`[media attachments included]`) rather than loading binary parts.

## Media reference model

The canonical external/reference shape is a JSON object:

```json
{
  "id": "media:123",
  "media_id": 123,
  "session_id": "session_...",
  "filename": "screenshot.png",
  "content_type": "image/png",
  "size": 12345,
  "sha256": "hex-encoded-content-hash",
  "source": "web|tui|api|direct|extension|tool",
  "created_at": "2026-05-25T13:00:00Z"
}
```

Rules:

- `media_id` is the durable SQLite `media.id` integer.
- `id` is a stable string form for payloads and UI code: `media:<media_id>`.
- `session_id`, `filename`, `content_type`, `size`, `sha256`, `source`, and `created_at` are descriptive fields and may be omitted by callers when submitting references, but are filled by store/API responses.
- Compatibility lists may contain strings (`"media:123"` or legacy opaque values) or objects. New code should emit objects and accept both forms while this feature is rolling forward.
- A reference is session-scoped. Submitting `media:123` to another session is rejected unless a future explicit clone/share API creates a new session-owned media row.

## Storage policy

The `media` table remains the durable blob store for the first slice.

- Binary content lives in SQLite as compressed-or-raw `content` bytes, matching the existing store object compression helper.
- `metadata_json` carries derived metadata that does not need separate indexes yet: `size`, `sha256`, `source`, `detected_content_type`, optional image dimensions, and UI hints.
- The initial implementation does not deduplicate rows across sessions. It computes and records `sha256` so a later cleanup/dedup slice can safely identify duplicates without changing the reference shape.
- Cleanup is session-owned: deleting a session deletes its media via the existing foreign key cascade. Standalone pruning of unreferenced media is deferred until message/media joins exist.

## Limits and validation

Initial limits should be enforced consistently at all ingestion entrypoints:

- maximum single media size: 10 MiB by default;
- maximum media references per submitted prompt/steering message: 8;
- accepted types for provider-bound media: `image/png`, `image/jpeg`, `image/webp`, and `image/gif` initially;
- unknown or unsupported binary media may still be stored for transcript/reference purposes but must not be projected to providers as image parts.

MIME/content-type handling:

- trust explicit `content_type` only as a hint;
- detect content type from bytes with the standard Go sniffer or a stricter image decoder when projecting to providers;
- store both caller-provided and detected values when they differ;
- default to `application/octet-stream` when unknown.

## Submission payloads

Text-only submissions are unchanged.

New optional request field for web/API/direct/TUI submit paths:

```json
{
  "prompt": "describe this",
  "media": ["media:123"]
}
```

or, equivalently:

```json
{
  "prompt": "describe this",
  "media": [{"id":"media:123"}]
}
```

Accepted field name is `media`; `attachments` may be accepted later as a UI alias but must normalize to `media` before entering `turn.RunInput.Metadata` or steering queues.

For active-turn steering, `media` follows the existing steering queue lane and is also copied into `messages.payload_json.media` for compatibility.

For new turns, the turn metadata should carry a normalized `media` array until a relational turn/media join exists. The first implementation must also persist the user message with the same normalized compatibility payload.

## Message, turn, and tool representation

Until relational joins are added, normalized media references appear in JSON metadata in these locations:

- `turns.metadata_json.media` for the submitted turn;
- `steering_queue.media_json` for active-turn steering;
- `messages.payload_json.media` for transcript/history compatibility;
- topic payloads as counts by default (`media_count`) and full references only on ingestion/registration events where useful.

Tools should receive media references, not raw bytes, unless an explicit tool contract asks the store for content. Tool calls and tool results must not embed binary data in transcript-safe JSON.

## Provider projection

Provider projection is a separate step from ingestion.

- The turn engine loads referenced media only when constructing a provider request that supports media/image parts.
- Provider-specific request builders map normalized refs to the provider's image-part format.
- Unsupported providers keep the current safe textual placeholder behavior and should include a clear system/user-visible note only when media cannot be consumed.
- Projection must validate session ownership, content type, and size again before reading bytes.

## Topics and SSE

Media ingestion publishes durable-friendly runtime notices without broadcasting raw content:

- `runtime.media` / `media.created`: session id, media id/ref, filename, content type, size, sha256, source.
- `runtime.media` / `media.rejected`: session id, source, reason, optional filename/content type/size.

Existing `session.steering` and turn/session lifecycle topics continue to carry `media_count` rather than raw references unless a UI specifically needs refs.

## TUI boundary

TUI image paste work starts only after the minimal store/API primitives exist.

- Direct paste support, if terminal/library support exists, creates media first and then submits refs through the same `media` field.
- If paste events are not available, a TUI fallback command such as `/attach <path>` should ingest the file and produce the same refs.
- Ordinary text paste remains unchanged.

## Implementation order

1. Add shared normalization helpers and tests for string/object media refs.
2. Add content hash and detected MIME metadata to `CreateMedia` without changing existing callers.
3. Add a small authenticated media upload/list/get API under session scope. **Implemented:** `GET /api/sessions/{session_id}/media`, `POST /api/sessions/{session_id}/media`, and `GET /api/sessions/{session_id}/media/{media_id}`.
4. Add optional `media` to prompt/direct submit paths and preserve it in turn/message metadata.
5. Add provider-safe projection tests before enabling binary provider requests.

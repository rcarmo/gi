# Provider prompt caching

Gi does not store a local copy of provider KV caches. The provider controls cache
availability, minimum prefix length, lifetime, billing and eviction. A cache miss
is not an inference failure, and a stable key is not a cache-hit guarantee.

## Request identity and prefix

Normal agent generations pass the durable **session ID**, not a turn, request,
iteration or timestamp ID, through `inference.StreamHooks.SessionID` to
`go-ai.StreamOptions.SessionID`. It stays the same across tool iterations, turns,
transport retries and reopening that session. Another session gets another key.
Native adapters decide which cache-key fields and affinity headers their API
supports; Gi does not inject arbitrary OpenAI fields into other APIs.

Gi's legacy OpenCode Zen adapter sends `x-opencode-session` itself. It also reads
`usage.prompt_tokens_details.cached_tokens`: those tokens are subtracted from
`prompt_tokens` before storing uncached input, so input and cache reads do not
overlap. Native go-ai adapters already use this convention. The terminal cache
hit percentage is `cache_read / (input + cache_read + cache_write)` for the latest
measured generation, not a local cache counter or the entire conversation size.

Normal requests leave retention at the provider/go-ai default (usually short).
Anthropic already receives system, tool and message cache breakpoints by default.
Gi's one-off compaction/side requests explicitly use `CacheRetentionNone`; they
must not accidentally inherit the main conversation's cache key. Long retention
is not forced: it may cost more and some providers do not support it.

The system prompt is built from stable named sections. Updates retain their
recorded historical position where the model supports mid-conversation system
messages; incompatible models fold updates into the leading prompt. Tool
registration/declaration order is stable, and tools loaded through search are
appended and recorded per session. History replays stored provider messages and
tool content rather than reconstructing assistant text from UI rendering.
Compaction, actual tool/schema changes, model/provider changes, context edits and
hook-supplied prefix replacements can legitimately invalidate cached prefixes.
Do not suppress correctness updates or retain stale tools just to improve CH%.

## November 4 audit and repair

Two concrete defects were found:

1. Normal generations never supplied a session ID to go-ai. OpenAI/Codex cache
   keys and several providers' affinity mechanisms were therefore omitted.
2. The fallback OpenCode stream discarded reported cache reads, so displayed
   hit rates could underreport caching even when the service reused input.

Both paths are repaired. Local wire tests prove stable keys for the same
conversation, distinct keys for different conversations, omission when caching
is disabled, unchanged tools and append-only message prefixes, deterministic
Anthropic payloads with cache breakpoints, and OpenCode affinity/non-overlapping
usage counters. An engine regression verifies identity and stable prompt/tool/
history prefixes across separate durable turns. Full inference, turn and
scripting package suites passed with Go 1.27.3 and go-joker v42.12.6.

These are offline contract tests, **not measured live cache-hit improvements**.
No live credentials or transcripts were read to perform this audit. After a
restart with the rebuilt binary, compare provider-reported cache reads on
successive long requests in the same session, within the provider TTL. The first
request after a prefix/model change may be cold. Very short prompts, TTL expiry,
provider routing/eviction and extensions rewriting earlier context remain
possible causes of low hit rates.

## Verification

```sh
make test TEST_PKGS=./internal/inference TEST_RUN=TestPromptCache
make test TEST_PKGS=./internal/turn TEST_RUN=TestPromptCache
```

Use the project Makefile scratch overrides for tmpfs caches; no fixture makes a
paid remote request. Do not log request content or credentials to diagnose cache
identity. Request IDs remain per-occurrence audit identifiers, not cache keys.

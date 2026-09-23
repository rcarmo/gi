# ADR-0047: Index settings and conservative optional roots

## Status

Accepted — 2026-09-23. Startup settings can extend indexed roots/extensions and explicitly allow initially missing roots. No background scheduling, automatic freshness or terminal index UI is enabled. Coverage stays 45/236 Classic and 2/42 shared, with 191/40 unmapped.

## Startup configuration

Configure `.pi/settings.json`:

```json
{
  "workspaceIndex": {
    "extraRoots": ["docs"],
    "extraExtensions": ["nim"],
    "optionalRoots": ["notes", ".pi/skills"]
  }
}
```

`config.Load` captures these settings at startup; `/api/runtime/config` exposes them as `workspace_index` with the same camel-case inner keys. There is no hot reload, new environment override or settings editor. Existing model-setting persistence preserves the object. Missing settings retain the previous strict defaults: notes and skills are required, with no extra roots/extensions.

Extra roots widen only `all`; extra extensions augment each scope's supported extensions. Notes/skills keep their own root ownership. Optional roots must exactly match distinct resolved roots in the all-scope configuration. Unknown, traversal, root-dot and covered subroot declarations fail resolution. A parent scope cannot simultaneously treat one of its descendants as an optional separately scanned root.

`ConfiguredScopeConfig` validates the global policy and applies optional flags only to the selected scope's roots. It includes sorted optional roots in the fingerprint. With no optional roots the old fingerprint bytes remain identical. New configurations do not relabel old committed rows: status becomes stale and incompatible query results are hidden until explicit refresh. Index status includes `optional_roots`; `required_roots` means all selected roots are required when true. It does not mean no roots are required when false.

Malformed settings JSON follows the existing loader's fallback behaviour; semantically invalid index roots fail the index API with 400. This slice does not redesign global configuration error reporting.

## Absence and disappearance

The scanner accepts ENOENT only at an explicitly optional root or one of its missing ancestors. Permission errors, non-directories, excluded ancestors and symlinks still fail. It does not create directories. Missing ancestors are not cached as visited, and each missing root is rechecked after scanning. Appearance or an invalid replacement during the scan fails the snapshot.

A complete snapshot lists `MissingRoots` explicitly. The store validates that every entry is a unique configured optional root and rejects documents attributed to an absent root. Inside the fenced commit transaction, it checks prior memberships for every missing root. If a root owns any committed document in that scope, the refresh fails and retains all prior content, count, generation and last-success timestamp. A never-populated or previously emptied optional root can be absent without blocking other roots.

This distinction prevents a missing mount/root from silently deleting indexed content. To remove old files deliberately, keep the directory present, remove files and complete a refresh, or deliberately change the scope configuration. Restoring the missing root and explicitly retrying publishes a new generation. Existing required-root behaviour is unchanged.

## Evidence

Native tests cover settings load/startup lifetime/model-save preservation, default strictness, invalid option policy, stable/default and changed fingerprints, selected-scope optional roots and immutable configuration. Scanner/store tests exercise initial absence, appearance during scanning, symlink ancestor rejection, contradictory snapshots, transactional disappearance retention and successful retry. Endpoint tests construct a new server from modified settings and verify stale/config-isolated queries before repopulation.

`make test-ux-index-config` seeds a real startup settings file in an isolated server and runs six browser projects. It indexes `.nim` content under `docs`, accepts initially absent optional notes/skills, populates notes, moves notes away to cause a native retained-index error, verifies persisted status/results and drafts after reload, restores the root and retries. It creates no chat message and fabricates no index response.

Two Gi-only derived scenarios (016/017) record startup policy and safe disappearance. There are now **17 proposal scenarios** outside the unchanged frozen corpus; the parser check grants no automatic runtime parity credit. Workspace-005 remains unmapped.

**498 browser executions pass:** 330 main, 54 reconnect, 66 compaction, 18 Steer/admission, 12 context-fit, 12 meter and 6 index-configuration. **76/76 functional**, **32/32 helpers**, full Go/vet/build/hook and config/search-store/indexer/web race tests ×3 pass. The new settings suite and strict existing index suite both pass.

The meter suite twice reproduced the previously recorded capability-snapshot race: it sampled usage before claim cleanup, then the UI showed Compact availability. Its fixture now waits until the native compaction capability no longer reports `Session has active or queued work` before sampling. Exact label/tooltip/colour assertions are unchanged; the full six-project meter suite passed twice after this synchronisation repair. Display-idle or filtered queue state is not used as claim-release evidence.

Desktop/tablet/phone screenshots were attached from `/workspace/tmp/gi-ui-captures/september23/index-optional-root-*.png`. Logs are `/workspace/tmp/gi-index-config-*`. Supplied components/panes/helpers and frozen features are unchanged. No terminal code changed; PTY suites were not rerun or credited.

## Remaining integration

Add application-owned scheduling/invalidation and background refresh, then native query consumers and compact terminal status/actions. Terminal roots must use this same policy with temporary error/detail surfaces, retain editor/cursor/reader and add no idle rows. The new settings are loaded by runtime config but do not by themselves create terminal controls. Vector/semantic search remains unimplemented.

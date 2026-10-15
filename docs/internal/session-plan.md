# Session Plan backend

Gi stores each session's Plan Markdown in `kv_store` under namespace `session_plan`. The `plan` tool and HTTP API use the same transactional mutation code. The shared frontend sidebar comes from fixtures-vibes; browser saves now require the loaded revision.

## HTTP

Authenticated `GET /api/sessions/{session}/plan` returns:

```json
{
  "ok": true,
  "plan": {
    "chat_jid": "gi:session-id",
    "revision": "plan-v1-opaque-value",
    "markdown": "- [ ] Example",
    "updated_at": "2026-10-04T20:00:00Z",
    "explanation": null,
    "plan": [{"step": "Example", "status": "pending"}]
  }
}
```

`updated_at` is null for an unsaved plan. Its initial Markdown is Piclaw's five default checklist items: update the plan, clarify the objective, do the next step, verify the result and report progress.

`POST` to the same route accepts `{"markdown":"...","expected_revision":"..."}` or `{"action":"reset","expected_revision":"..."}` and returns the same envelope. Missing revisions return 428 `revision_required`; stale ones return 409 `revision_conflict` without changing the plan. Internal agent mutations advance the revision too. See [revision-safe writes](revision-safe-writes.md). Writes require HTTPS or a loopback URL, same-origin requests and JSON. Unsupported methods return 405; missing sessions return 404; invalid bodies or Markdown return 400. Unknown fields and trailing JSON values are rejected.

Markdown is limited to 256 KiB. CRLF becomes LF, checked markers become lowercase `[x]`, and at most one checklist item may use `[-]`. Headings and other prose survive Markdown write/edit operations. A failed mutation leaves both Markdown and its update timestamp unchanged.

## Tool

`plan` supports Piclaw's `read`, `write`, `edit`, `patch` and `update` actions:

- `read`: returns `Plan for gi:{session}:` followed by the Markdown, with structured plan details.
- `write`: replaces Markdown; requires a `markdown` string.
- `edit`: applies exact replacements, deletions, insertions, appends or prepends. Anchors match exactly once against the original Markdown. Overlapping edits fail atomically.
- `patch`: sequential checklist add/update/remove operations. Indices are 1-based; text matches identify exactly one item. Adds can use start/end, before or after. This operation rebuilds the checklist and its leading explanation.
- `update`: replaces the structured checklist and optional explanation. Each item has a non-empty step and a pending, in_progress or completed status.

An explicit `chat_jid` must name the tool's current session. The tool also accepts Piclaw's `get`/`set` aliases and infers update/patch when their arrays are supplied without an action. Reset is API-only.

Tool details contain the same plan object as the HTTP response. The tool does not complete a model turn: the model receives its ordinary result and can continue.

## Events and model context

Every committed mutation emits a session-scoped `extension_ui_status` event:

```json
{
  "type": "extension_ui_status",
  "key": "plan.changes",
  "addon": "plan-sidebar",
  "chat_jid": "gi:session-id",
  "updated_at": "2026-10-04T20:00:00Z",
  "source": "tool",
  "action": "patch"
}
```

Source is `tool` or `api`; action identifies the mutation. Reads and failed mutations emit no change event. The sidebar should refresh the owning session while preserving dirty local text, as in Piclaw's add-on.

A saved non-empty plan adds stable Plan-tool instructions to the system prompt and supplies the current Markdown in a separate message on the outgoing provider-request copy. Persisted conversation and compaction input stay unchanged. The Markdown is excluded from the system prompt. Default unsaved and saved empty plans add no context.

## Submit to model

The UI captures the session, saves its current Markdown to that session's Plan API, then submits it through the same session's prompt API with normal `client_request_id` and intent handling. The composer draft stays untouched. There is no separate Plan-submit endpoint.

## Verification

`make test-session-plan` covers Markdown normalisation, exact edits and overlap rejection, sequential patches, transactional failure, reopen, session isolation, tool results/details, API authentication and validation, change events and model-context separation. The full plan/store/tools/turn/web packages pass. Independent Plan/widget acceptance at fixtures4259e82 passes 66/66 across all six Chromium/WebKit projects, including Markdown/progress rendering, edit/save, dirty-state preservation, captured-session submission and session-scoped tool behaviour. The profile claims `@cap-plan-sidebar`; eight Plan skips are removed. See [acceptance and profiles](../implementation/maintenance/classic-ui-acceptance.md).

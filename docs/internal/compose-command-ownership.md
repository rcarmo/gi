# Composer slash-command ownership

Composer autocomplete now uses Gi's advertised native commands. It no longer
calls Piclaw's `/agent/commands` endpoint or falls back to the bundled catalogue
of unsupported commands. The guarded build adapter leaves supplied component
files unchanged.

## Catalogue

`getAgentCommands()` reads `/api/quick-actions`, the same native capability source
used by Quick Actions: `/model`, `/compact` and startup-loaded workspace skills.
The composer validates command names/descriptions, preserves server order and
removes duplicate names. Empty catalogues are valid. Loading, malformed or failed
responses expose no fallback suggestions.

Each load clears previous matches. Responses from retired effects are ignored.
Successful reads refresh suggestions from the textarea's current value only
when Settings and search mode do not own input. A separate fixed status message
survives typing after catalogue failure: reopen the chat or reload to retry.
No catalogue request submits text or changes runtime settings.

## Keyboard contract

- Typing `/` in the composer opens its autocomplete, not Quick Actions.
- Tab accepts the highlighted command, preserving arguments and placing the
  cursor at the end, without sending.
- Enter on a bare fragment accepts and submits the highlighted command once.
  `/m` resolves to native `/model`, which is handled without an inference turn.
- With arguments present, Enter submits the literal composer value. Tab remains
  the explicit completion action; `/m argument` is not silently rewritten.
- Escape dismisses suggestions and retains the draft. Shift+Enter inserts a
  newline and closes the single-line completion list without submitting.
- Already-consumed events, modern composition events, legacy keyCode229 and
  repeated Enter/Tab are declined. Ordinary arrow-key repetition is unchanged.
- Idle-timeline typing belongs to Quick Actions. Settings controls, search mode
  and session/model pickers retain their own keyboard handling.

Shared16 versus Classic command-prefill semantics and the disputed Classic008
skill replacement mapping are unchanged. This slice does not adjudicate those
conflicts or add full frozen-scenario credit. Modifier-command semantics and
physical IME behaviour need separate acceptance.

## Evidence

The original composer requested an unserved endpoint and exposed56 bundled
commands in the empty native fixture. After fixing that path,18 synthetic
key-boundary cases reproduced unwanted submission for repeated Enter, consumed
Enter and keyCode229 across six browser projects. Modern `isComposing` was
already safe. Native Enter positive controls accompany the rejection tests;
synthetic composition flags do not establish physical keyboard/IME behaviour.

`make test-ux-slash` runs12 tests across Chromium/WebKit at three viewport sizes:
72 executions. It uses fresh disposable stores and local deterministic inference,
with native prompt/session/command paths. Tests cover completion/submission,
exact prompt identity, catalogue unavailable/malformed/delayed replies, Settings
and search races, and Quick Actions/picker exclusion. The dedicated browser CI
job runs this suite after startup and picker geometry.

Local regression evidence:174 Quick Actions/model cases,18 skill cases,42 startup
cases,24 geometry cases and112 functional cases (five existing skips). Go/vet,
hook and112 helper tests/3,013assertions passed. Review found stale match-reset
and modal-completion gaps; both were fixed and re-reviewed. A current search-mode
ref and delayed-search test address the remaining stale-closure concern.

No persistent terminal control, status row or shortcut is added. Equivalent
terminal command handling must retain pi-tui's existing editor/event ownership
and use transient completions; browser popup styling is not a TUI contract.

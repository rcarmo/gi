# Queue Return replaces the draft

Installed Piclaw 3.2.4 replaces editor text and references when returning a queued item. Gi's previous recovery merge prepended the queued content and retained the older draft, which did not match that contract.

Gi now replaces the current draft with the returned text, references and recovered queued attachments. Existing unsent content is not concatenated. The full canonical reference IDs and native attachment ownership checks are unchanged.

## Async and persistence safeguards

Gi recovers attachment bytes from its native media API, so Return can wait on a fetch. The action captures the current draft before that fetch. If text, references or files change while recovery is pending, it leaves both the current draft and queue item untouched and asks for an explicit retry. Retrying authorises replacement of the then-current draft.

After replacement is published, newer typing survives delayed persistence/deletion responses. A durable queue-ID recovery marker makes retry persist the existing replacement plus newer edits rather than replacing a second time. IndexedDB revision CAS and cross-tab conflict handling remain unchanged. Persistence must complete before DELETE; failure leaves the queue recoverable. A server-side consumed-item conflict retains the local replacement and reports uncertainty rather than silently sending it.

This retains native durability safeguards; it does not claim that Piclaw's asynchronous backend/attachment behaviour is identical in every failure case. Existing previously merged recovery records are not rewritten on upgrade.

## Evidence

- `make test-piclaw-queue-return`: six shipped-Piclaw browser cases pass across Chromium/WebKit and phone/tablet/desktop. The oracle confirms replacement before removal, restored text/file/message refs, focus and caret position. No Gi helper is used by the oracle.
- `make test-ux-queue-return`: 24 native browser cases pass, including compact-reference regression, edit-during-fetch rejection, storage failure, transport retry/reload, session switch, consumed item and typing after replacement.
- `make check`: Go/vet/build/hooks and 144 functional tests pass; 11 skipped.
- `make ux-parity-inventory`: 207 support tests / 7,770 assertions pass; no new mappings.
- Focused delegate review timed out. No independent review approval is claimed.

Protected source and historical contracts are unchanged. No deployment, restart or live-chat writes. Logs: `/workspace/tmp/gi-return-{final-check,final-browser}.log`; installed oracle artifacts: `test-results/ux-oracle/queue-return/`.

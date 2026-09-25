# Passkey criterion ledger

The [JSON ledger](passkey-criteria.json) links all 26 additive scenarios to
individual assertions or named gaps. It records 169 scenario steps, four shared
Background steps and 40 example rows. Gherkin expansion produces 56 cases.
These counts describe the source contract; the latest browser suite has 189
executions and uses a different test grouping.

Initial source-evidence baseline: `e7d9b0450db9dc2c103dc06797dd5387ab9f81d7`.
Later tested changes update individual dispositions and anchors in the ledger.
Source hash:
`bd48cab9126778763cab3ddfd8ee04a89bd8c29da62e3bc4188cf24dc3a55b82`.
The [frozen feature](../../tests/ux/features/additions/piclaw-2026-09-24/piclaw-single-user-passkey-settings.feature)
is unchanged. All 26 full-scenario mappings remain unawarded. Classic/shared
mapping counts remain 101 of236 Classic scenario IDs (256 expanded cases) and
30 of42 shared cases, including the separately disputed Classic008 mapping.

## Reading a row

Each criterion stores its exact source line, keyword and text. Example rows keep
the exact column values. An evidence key resolves to a repository path, unique
test-declaration anchor and fixture scope. The fixture itself is labelled
separately. No line number into a moving test file is
required. Every scenario inherits the four Background requirements and the
ledger's common limits.

| State | Meaning |
|---|---|
| `bounded` | A direct assertion exists within the named fixture scope. |
| `substituted` | The test uses a specified alternative to the requested precondition or interaction. |
| `partial` | Some assertions exist, but a named part of the criterion is missing. |
| `gap` | No direct evidence for this criterion. |
| `manual` | Device, native prompt or assistive-technology execution is needed. |
| `unsupported` | The required runtime surface is not implemented. |

These states are source-review dispositions. They do not represent a new test
run, and `bounded` does not mean a complete scenario passes. Every
`formalMapping` field is false. The ledger is deliberately separate from
`mappedIds` and `sharedMappedIds` in the existing parity catalogue.

Common substitutions apply even to rows labelled bounded: current ceremonies run
on HTTP localhost rather than the pinned HTTPS RP, UI evidence is Classic only,
and CDP virtual authenticators do not establish physical or synced devices.
A separate Classic first-owner journey now enables TOTP in Settings from an
unenrolled instance, registers the first virtual passkey and signs in with each
factor. Its setup recovery tests also run in WebKit. This adds prerequisite
evidence to002 without a full mapping. The production RP choice remains open.

## Findings that determine the next work

- 010 cancels through the application, not a native prompt Cancel control. The
  failed attempt also requires Refresh before Add becomes available.
- 011 has synthetic blur ownership checks, without successful completion after
  real native focus transfer.
- 012 now has a Settings duplicate journey with the real exclusion list and a
  crafted client finish accepted by verification then refused by the duplicate
  guard. Server credentials are not seeded. The crafted none-attestation proof
  remains a substitution for physical authenticator behaviour.
- 014 now has both pane-return and Settings-reopen evidence: native finish is
  committed but its response lost, then an automatic held inventory GET gates
  reconciliation. Second return and reload issue no additional auth POST. This
  closes the earlier Refresh-only gap within the common fixture limits.
- 015 now has actual insecure non-localhost origin evidence: the login gate
  blocks even a copied valid cookie before Settings opens. The frozen journey
  asks for unavailable Add inside Settings, so this stronger entry boundary
  remains an explicit difference. Capability overrides are not old-browser proof.
- 017 now checks the exact list/register/rename/remove/account-query refusals
  with a nonempty native inventory, immutable auth state and an owner-list
  positive control. Family-shared mode is unsupported; that example stays open.
- 018's revocation test holds the completed credential before finish delivery;
  it does not revoke while a physical prompt is open.
- 020 has native/Settings decision and explanation coverage for its six supported
  examples. Pending TOTP addition to an established owner is unsupported; the
  defensive stored-disabled-secret test cannot substitute for that workflow.
- 021 can reject the concurrent loser at the non-blocking writer lock. A later
  deliberate retry proves the factor check, not the required direct refusal.
- 023 has no legacy passkey-delete command with Settings guidance.
- 024 now preserves sessions through removal, then exercises explicit logout
  and native expiry independently. Expiry crosses a two-second fixture deadline;
  it does not wait twelve hours. Logout is cookie-owner-only and leaves the other
  browser signed in even when the selected session's proof key was removed.

The [scenario review](passkey-scenario-review.md) is the compact index. The
criterion ledger supplies the exact missing step or example behind a partial
classification.

## Validation

Run `make test-passkey-criteria` for this ledger or `make ux-parity-inventory`
for all support checks. The existing CI passkey-browser target now requires this
ledger check before running its browser matrix. The test checks the pinned hash,
Gherkin parse, all Background/scenario steps and example values, unique IDs,
allowed dispositions, evidence references and distinct file/anchor pairs. Each
anchor must occur once and appear in a test declaration or the declared fixture.
It also
checks that Classic/shared mappings remain unchanged. It cannot judge whether a
test assertion semantically satisfies the referenced criterion; that requires
review of the named test.

## Terminal adaptation

Passkey ceremonies and credential management remain in the authenticated browser.
The terminal must not emulate WebAuthn, collect enrolment secrets or acquire
browser-owner authority from its local process credentials. No persistent status
row, pane or prompt is added. Existing transient help can direct users to Settings
when an applicable command exists; the unsupported legacy delete workflow still
needs an explicit design decision. This ledger adds no terminal implementation
or terminal acceptance credit.

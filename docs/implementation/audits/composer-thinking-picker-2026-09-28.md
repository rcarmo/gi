## Composer thinking selection

Gi's composer model picker can now change thinking levels through the same validated session API as Settings. Previously it displayed a disabled, read-only level even when the native session supported changes.

The build-time adapter in `scripts/patch-model-thinking.mjs` leaves supplied components unchanged. It accepts only advertised levels (or an empty string for provider default), uses the current model and thinking token, and shares the model-mutation lock. A rejected change refreshes authoritative state without replaying the mutation. An unmounted session cannot receive the response; the keyed composer and parent selection guards preserve the destination draft and state.

The picker stays open. A pending request disables selection, and native select keys bypass model-row navigation. Focus restoration is conditional: it only restores a previously focused thinking control when the browser has dropped focus to the body, the control is enabled, and Settings does not own a modal. Newer composer or Settings focus is preserved.

## Installed reference

`make test-piclaw-picker-thinking` passed six isolated journeys against the installed Piclaw 3.2.4 bundle: Chromium and WebKit at phone, tablet and desktop sizes. Each sent one `/thinking high` command, disabled the select while pending, kept the draft and picker open, and returned focus to the picker trigger on Escape. The probe records installed bundle provenance in `test-results/ux-oracle/picker-thinking/evidence.json`.

Piclaw lost thinking-control focus after the disabled/pending transition in all six reference cases (`thinkingFocusedAfter: false`). Gi restores that focus under the ownership conditions above. This is a deliberate accessibility difference, not an exact focus match. Piclaw commands are mocked in this oracle; no live session or provider configuration is changed.

## Validation

| Check | Result |
| --- | --- |
| Installed picker oracle | 6/6 |
| Complete native picker journey suite | 30/30 |
| Existing model-panel suite | 42/42 |
| Existing Settings thinking suite, separate browser groups | Chromium 21/21, WebKit 21/21 |
| Native SessionThinking race checks, repeated three times | Store, inference, turn and web passed with the Settings gates |
| `make test-model-panel-helpers` | 10 tests, 114 assertions |
| `make check ux-parity-inventory` | 144 functional passed, 11 skipped; 209 support tests, 7,786 assertions; Go tests/vet and build checks passed |

Native journeys use disposable databases and the local reasoning provider fixture. They cover explicit and default levels, disabled pending state, rejection recovery, retained drafts, durable reload, selected thinking reaching the next provider request, newer focus ownership, delayed origin responses after a session switch, and stale-token conflict without replay. The support test checks arrow/Enter/Home/End/Page keys cannot activate model rows from the native select.

## Retained failures and limits

Early test runs exposed lost focus after a disabled select, an incorrect test route for turn status, an incorrect empty-turn-array assumption, a phone-only close-button locator used on larger screens, and a stale-token test expecting 400 instead of the native 409 conflict. Those failures preceded the passing runs. The focus defect was fixed; the other cases corrected tests to use existing native contracts.

Incomplete runs hit outer command limits at 190 seconds for the picker and 275/280 seconds for combined or full Settings runs. Their logs are retained. The model-panel suite completed in the combined run; the Settings browser groups later completed separately. No Playwright timeout was increased, and no SSE or browser-error exception was added. Independent review attempts timed out; there is no independent approval result.

The evidence archive retains the available earlier logs, installed reference output and final native report. Earlier overwritten browser artifacts cannot be recovered from the logs alone. This slice does not test physical devices, screen readers, real provider mutations or pixel equality. It adds no parity mapping: the inventory remains 241 scenarios/262 expanded cases, Classic 97 mapped and 144 unmapped, Shared 30 mapped and 12 unmapped. The strict installed WebKit reload failure and the remaining whole-web audit are still open. Nothing was deployed or restarted.

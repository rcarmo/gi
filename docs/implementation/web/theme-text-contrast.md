# Pinned theme text contrast

Gi now computes primary and secondary text colours using the pinned Classic palette rule. The default dark secondary colour changes from `#71767b` to `#82868b`; the model label now has the same computed colour as the reference.

## Reference and guarded adaptation

Pinned Classic `0afe5366ced9`, asset `15958f2c3dc9`, bundle SHA256 `6dcd31d89b14a7c99066f469d07ead18e35a48d701ae11b599b94e4603778c5d7b8`, contains these functions:

- `Rk` computes primary text against the primary, secondary and hover backgrounds, then secondary text against the same surfaces, using the adjusted primary as its target.
- `gc` selects the first 0–100% RGB blend meeting 4.5:1 against every supplied background. Channels are rounded at each 1% step; an impossible match returns the target.
- `Ck`, `xk` and `Mg` provide rounded RGB mixing, WCAG luminance with the 0.04045 linearisation threshold, and contrast ratio. Invalid colours leave the input unchanged.

`gi-theme-text-contrast.ts` ports only these primary/secondary calculations for the hex/RGB colours emitted by Gi's existing palettes and tint renderer. `patch-theme-text-contrast.mjs` adds exact build anchors after `patch-accent-contrast.mjs`. The supplied `web/src/ui/theme.ts` bytes remain unchanged. Accent foreground, syntax/terminal colours, borders, theme choices, persistence and metadata colour remain on their existing paths.

Default/no-tint rendering previously cleared variables and bypassed the palette renderer. It still clears the existing variables, then installs the two corrected text colours. Preset/tinted rendering corrects the same two entries in the existing variable map. System colour changes and Reset follow those same paths. No DOM observer, extra control, font/raster flag or idle TUI chrome is added.

The contrast threshold applies to the computed palette colours against those three surfaces. Opacity and other compositing can change final rendered contrast; this is not blanket accessibility or screen-reader acceptance.

## Tests and pixel evidence

- Seven compose/contrast helper tests pass. New coverage pins `#82868b`, light-mode stability, the preceding insufficient blend, generated RGB tints, malformed/fallback inputs, absent hover, monochrome target and guarded adapter composition.
- `make test-ux-compose-surface`: 36 passes across Chromium/WebKit and three viewports. New checks cover system light/dark transitions, default colour, Monokai/Solarized, purple/amber tints, Reset, focus, draft and all-three-surface contrast.
- Settings/shell: 216 passes. Stock functional: 107 passes, 11 existing fixture-dependent skips. Go/vet/hooks and 15 pixel helpers pass.
- Initial tests exposed the default/no-tint bypass. Two early runs were interrupted after incorrect Appearance control labels timed out; the completed run uses the actual Save/Custom tint contract. The Settings suite also had one old menu locator; it now uses the shipped Models listbox. No tests removed or timeouts increased.
- The requested read-only delegate could not select an approved executable model. Reference analysis was performed locally; no independent review result is claimed.

Full capture run `run-1790392290634-774386` completed all 72 captures without errors. All 18 full-frame cross-host comparisons still fail; 15/36 repeat pairs are unstable. Explicit saved-run comparison confirms failure. Compose-region differences:

| Viewport | Light | Dark |
|---|---:|---:|
| Phone | 238 | 226 |
| Tablet | 237 | 230 |
| Desktop | 298 | 239 |

Both desktop/dark model hints compute `rgb(130, 134, 139)`. The preceding full run measured 242–1,025 compose pixels. Raw frames, hashes, requests and failures are retained. No masks, tolerances, image resizing, altered raster flags or repeat-selection rules apply. Full-frame equality remains mandatory; capability differences and unstable rasterisation still block acceptance.

## Deployment

Exact `e0ed68a` passed whole rerun CI36215674941 and was rebuilt from detached checkout `/workspace/tmp/gi-theme-context-e0ed68a`. It is deployed on8090 as PID839609, PGID/SID839508. The earlier CI36214891606 exceeded its15-minute compose-job limit during voice tests after surface/model/session tests passed; required voice coverage was split separately in `d2caf16`, without removing targets or raising timeouts. That CI-only change and TUI search `efe1880` are excluded from the deployed source.

Six live Chromium/WebKit × viewport probes pass system light/dark text colours, disabled context reason, existing model listbox/Tab order, Settings handoff, resize, session metadata and draft retention. Writes/errors are empty; no permissions requested or compaction/session/auth mutations. DB62sessions/51turns/146messages, integrity/FKs clean, auth absent before/after. SQL excluding runtime leases is unchanged SHA256 `300ce6794c6fe9b8fb93c826fcfecf093676a6cf369f4be4e11dd624a829653d`.

Frozen feature mappings and broader TUI/physical-device/Visual acceptance are unchanged. The exact pixel failures above remain open.

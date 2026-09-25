# Compose contrast and remaining controls

The active-session pill now uses the pinned Classic foreground contrast rule and vertical padding. The six compose-region pixel comparisons improved to 930–1,735 changed pixels, but none passes exact equality.

## Colour and padding

The older supplied theme chose white text whenever accent luminance was at most 0.4. On the default blue accent, that gave poorer contrast than black. Pinned Classic `0afe5366ced9` compares black and white WCAG contrast and chooses the greater ratio. `gi-accent-contrast.ts` implements that rule; `patch-accent-contrast.mjs` replaces one exact function body at build time. The supplied theme file remains byte-identical. The rule applies to the shared `--accent-contrast-text` variable, including accent-filled controls outside the composer.

`gi-compose-surface.css` sets the session pill's padding to `0 8px`, matching the reference. Bounds, session identity, keyboard behaviour and draft ownership are unchanged. No idle TUI rows were added.

## Remaining control differences

Matched desktop DOM/style inspection found identical footer, model label, usage line and context-ring bounds. Piclaw renders six footer controls while Gi renders four: voice input and browser notifications are absent in Gi. Their missing width shifts Search and adjacent buttons by 56px in this fixture. No disabled placeholders or hidden reference controls were added to improve pixels. Native speech playback is separate from voice-input capture; neither it nor an unimplemented notification service establishes those missing capabilities.

The reference and Gi session-pill text used the same font but different foreground colours before this change. Other repeated differences cluster around antialiased edges. Diagnostic font-raster experiments were not adopted; screenshots continue to use the existing renderer flags, exact RGBA, no masks and no tolerance.

## Verification

- Four compose/contrast helper tests pass with 4,126 assertions. A 4,096-colour RGB sample stays at or above 4.5:1 against the chosen black/white foreground. Anchor drift and repeat patching reject the build.
- `make test-ux-compose-surface`: 30 Chromium/WebKit × viewport cases pass. The 12 new cases cover default light/dark, dark/pale custom tint, reset, padding, contrast, retained draft and focus.
- Settings/shell suite: 216 passes. Go, vet, hooks and 14 pixel helpers pass. Stock functional suite: 107 passes, 11 fixture-dependent skips.
- The initial browser run used a nonexistent Save-button label and was interrupted. Correcting it to the real Appearance controls preceded the completed run; no product behaviour was weakened.

Final pixel run: `run-1790375609940-255818`. All 72 captures completed without fixture/page errors; all 18 cross-host full frames differ, and 16/36 repeats are unstable.

| Compose region | Changed pixels, light / dark |
|---|---|
| Phone | 930 / 1,441 |
| Tablet | 952 / 1,502 |
| Desktop | 933 / 1,735 |

The earlier surface run measured 1,514–2,438 changed compose pixels. Counts are diagnostics while repeat instability persists. Full pixel, physical-device and Visual acceptance remain open. Frozen feature files and mapping counts are unchanged; CI must pass before deployment.

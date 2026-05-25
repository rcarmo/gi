# TUI clipboard and media parity

Status: clipboard text copy has an opt-in implementation; media/image paste remains deferred.

## Current behavior

Gi keeps `/copy` safe by default: without flags, it locates the last non-empty assistant message and prints it back into the transcript with a clear `copy:` prefix. It does **not** write to the OS clipboard and does **not** emit terminal escape sequences unless the user opts in.

This preserves the tmux/script-friendly baseline and keeps transcript storage deterministic.

## `/copy` targets

Supported forms:

```text
/copy
/copy --fallback
/copy --osc52
/copy --native
/copy --auto
/copy --mode <off|osc52|native|auto> [--persist]
```

Modes:

- `off` / `--fallback`: transcript fallback only.
- `osc52` / `--osc52`: write an OSC 52 sequence directly to the terminal output path, then print a plain transcript confirmation.
- `native` / `--native`: use a detected native helper, or fall back with an error note if none is available.
- `auto` / `--auto`: try native helper first, then fall back to transcript output.

`--persist` stores the selected mode in `.pi/settings.json` as `tuiClipboardMode`. The safe default is `off`.

## OSC 52 support

OSC 52 copies text to a terminal clipboard by writing an escape sequence like:

```text
ESC ] 52 ; c ; <base64 payload> BEL
```

Gi's implementation is intentionally opt-in and has these guardrails:

- escape sequences are written to the terminal writer, not stored in transcript lines;
- transcript output contains only a plain success/failure message;
- payload size is capped to avoid terminal/tmux limits;
- default `/copy` remains transcript-only.

Terminal caveats still apply:

- many terminals gate or disable OSC 52 clipboard writes;
- tmux may require passthrough configuration or its own clipboard integration;
- SSH sessions may copy to the local terminal clipboard depending on passthrough.

## Native clipboard helper support

Native helper support is opt-in and dependency-light. Gi detects helpers already present on the system rather than installing anything.

Helper order:

- macOS: `pbcopy`;
- Linux Wayland: `wl-copy`;
- Linux X11: `xclip -selection clipboard`, then `xsel --clipboard --input`;
- Windows/WSL fallback: `clip.exe`.

Native execution uses a short timeout and returns bounded transcript errors. Tests cover helper selection without requiring a desktop clipboard in CI.

## Media/image ingestion boundary

Pi-style image paste is not just a keybinding. Gi still needs a media ingestion contract first:

- where pasted media is stored;
- how media is referenced from messages/turns;
- how tools and providers receive image parts;
- how TUI, web, and API submissions share the same payload shape;
- what limits and cleanup policies apply.

Until that contract exists, image paste remains out of scope for the TUI.

## Current support summary

- `/copy`: supported as transcript fallback by default.
- OSC 52 copy: supported opt-in via `/copy --osc52` or persisted `tuiClipboardMode=osc52`.
- Native clipboard helpers: supported opt-in via `/copy --native` or `--auto`.
- Ordinary text paste: unchanged terminal-rune behavior.
- Bracketed paste: reassessed, deferred pending parser/editor support.
- Image/media paste: contract defined at a boundary level, implementation deferred pending shared media ingestion.

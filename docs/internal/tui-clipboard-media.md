# TUI clipboard and media parity

Status: Phase F working note for Pi-like TUI fit/gap work.

## Current behavior

Gi currently implements `/copy` as a terminal-safe fallback: it locates the last non-empty assistant message and prints it back into the transcript with a clear `copy:` prefix. It does **not** write to the OS clipboard and does **not** emit terminal escape sequences.

This is intentional for the current tmux/script-friendly TUI baseline.

## OSC 52 investigation

OSC 52 can copy text to a terminal clipboard by writing an escape sequence like:

```text
ESC ] 52 ; c ; <base64 payload> BEL
```

It is attractive because it avoids per-platform clipboard helper dependencies, but it is not safe as a normal transcript line:

- many terminals gate or disable OSC 52 clipboard writes;
- tmux requires passthrough configuration or its own clipboard integration;
- remote SSH sessions can copy to an unexpected terminal clipboard depending on passthrough;
- raw escape sequences in transcript buffers would make captures and tests less deterministic;
- some scrollback/logging paths may persist the escape sequence instead of executing it.

## Decision for current slice

Do **not** emit OSC 52 from `/copy` by default.

If OSC 52 is added later, it should be opt-in and handled outside transcript content, for example:

- `/copy --osc52` or a setting such as `tuiClipboard=osc52`;
- write the escape sequence directly to the terminal output path rather than storing it in the transcript;
- print a plain transcript confirmation after the attempt;
- cap payload size to avoid terminal/tmux limits;
- document terminal/tmux requirements.

## Native clipboard helper investigation

Native helper support is possible, but the dependency/OS surface is wider than the current fallback:

- macOS: `pbcopy`;
- Linux desktop: `wl-copy`, `xclip`, or `xsel` depending on Wayland/X11/session availability;
- Windows/WSL: `clip.exe`, PowerShell `Set-Clipboard`, or WSL interop quirks.

This should remain deferred unless a concrete user need justifies adding helper detection and tests.

## Media/image ingestion boundary

Pi-style image paste is not just a keybinding. Gi needs a media ingestion contract first:

- where pasted media is stored;
- how media is referenced from messages/turns;
- how tools and providers receive image parts;
- how TUI, web, and API submissions share the same payload shape;
- what limits and cleanup policies apply.

Until that contract exists, image paste remains out of scope for the TUI.

## Current support summary

- `/copy`: supported as transcript fallback.
- OSC 52 copy: investigated, deferred by default.
- Native clipboard helpers: investigated, deferred.
- Ordinary text paste: unchanged terminal-rune behavior.
- Bracketed paste: deferred pending parser/editor support.
- Image/media paste: deferred pending media ingestion contract.

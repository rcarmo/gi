# TUI clipboard and media parity

Status: clipboard text copy has an opt-in implementation; clipboard image paste is supported via `/paste-image`; `/attach <path> [prompt]` is the terminal-safe file media fallback.

## Clipboard image paste

`/paste-image [prompt]` (alias `/paste`) reads an image from the system clipboard, stores it in the session media store (`source=tui-paste`), and either reports the `media:<id>` reference or, with a prompt, submits the prompt with the image attached via the shared media ingestion contract.

The clipboard image reader is platform-aware and best-effort:

- Linux Wayland: `wl-paste --type image/png`;
- Linux X11: `xclip -selection clipboard -t image/png -o`;
- macOS: `pngpaste -`;
- Windows: PowerShell `System.Windows.Forms.Clipboard.GetImage()` to PNG.

If no helper is available, `/paste-image` returns a clear error. The reader is injectable for tests. Images larger than 10 MiB are rejected.

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

## TUI media fallback

Direct terminal image paste is still not available in the current `go-tui` keyboard parser: key handling is rune/key-event oriented and does not expose clipboard image payloads or bracketed-paste metadata that would safely distinguish image data from ordinary text paste.

Gi therefore provides an explicit fallback command:

```text
/attach <path> [prompt]
```

Behavior:

- reads a local path relative to `workspace_root` when the path is not absolute;
- stores the file through the shared session media store with `source=tui`;
- enforces the same 10 MiB single-file limit as the web/API media endpoint;
- with no prompt, prints the created `media:<id>` reference without submitting a turn;
- with a prompt, submits the prompt through the same media metadata path used by web/API submissions.

Ordinary text paste remains unchanged and is still handled as terminal text input.

## Media/image ingestion boundary

Pi-style image paste is not just a keybinding. Gi still needs a media ingestion contract first:

- where pasted media is stored;
- how media is referenced from messages/turns;
- how tools and providers receive image parts;
- how TUI, web, and API submissions share the same payload shape;
- what limits and cleanup policies apply.

The shared contract is documented in [`media-ingestion-contract.md`](media-ingestion-contract.md) and its store/API primitives are implemented. `/attach` and `/paste-image` both project media through that same contract.

## Current support summary

- `/copy`: supported as transcript fallback by default.
- OSC 52 copy: supported opt-in via `/copy --osc52` or persisted `tuiClipboardMode=osc52`.
- Native clipboard helpers: supported opt-in via `/copy --native` or `--auto`.
- Ordinary text paste: unchanged terminal-rune behavior.
- Bracketed paste: reassessed, deferred pending parser/editor support.
- Command-driven clipboard image paste: supported via `/paste-image [prompt]` (alias `/paste`).
- Raw Ctrl-V image payloads and inline terminal image protocols: deferred pending terminal/parser support.
- TUI file fallback: supported via `/attach <path> [prompt]` and the shared media ingestion contract.

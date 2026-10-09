# TUI file-tool syntax highlighting

Issue [#13](https://github.com/rcarmo/gi/issues/13). Applies to fullscreen and
regular terminal output; no web renderer or model-visible tool contract changes.

## Presentation

- `read`: when a file path is available, successful output is hidden while
  collapsed. Ctrl-O or the existing block toggle expands even a one-line read.
  Errors/skipped reasons remain visible; errors are not syntax highlighted.
- `write`: display the submitted `content`, not the success acknowledgment.
  Collapsed calls show the first ten source lines, expanded calls show all
  available content. Failed writes retain both the submitted preview and error.
- Infer language only from the explicit Pi 0.99.2 extension/name mapping. Unknown
  paths remain literal plain text. Use Chroma's existing tokenizer and active
  Pi `syntax*` theme colours, not language auto-detection.
- Tokenize the complete available source before limiting the preview, preserving
  multiline lexer state. Tabs display as three spaces, carriage returns are
  removed, and terminal controls are stripped. Literal Markdown-looking content
  stays literal. Source lines wrap by go-tui clusters inside the one-cell left
  and right band padding, without a second rich-text wrap or synthetic NBSPs.
- Existing per-tool `hidden`/`compact` extension modes remain respected. Unknown
  historical records without file arguments retain the old plain preview.

## Source and history

Live runtime events retain `path`/`file_path` and optional write `content` in
presentation-only block metadata, separate from the abbreviated header. End
notifications lacking arguments retain the start's metadata.

On reopening, the TUI pairs stored assistant `tool_calls` with their results by
**turn ID plus call ID**, so reused IDs in different turns cannot cross-associate.
No filesystem read is performed to recover old content. Available read lines
are no longer cut at 200 characters during historical projection. Stored source,
arguments, results and model context remain unchanged; underlying read-tool
truncation limits still apply.

Selection and search index the displayed rich-text cells, not lexer markers or
hidden content. Search/selection reproject file blocks at the requested viewport
width. Prewrapped code retains row-local search (cross-wrap code matching is not
claimed). Chroma and highlight.js can classify individual tokens differently;
this is not a claim of byte-identical Pi rendering.

## Acceptance

- `make test-tui-tool-syntax`: live metadata, successful read folding, write
  previews, multiline tokens, errors, unknown extensions, extension modes,
  turn-scoped history recovery, long lines, narrow/resize rendering and literal
  search/selection text.
- `make test-tui-tool-syntax-pty`: local deterministic HTTP provider drives the
  production turn engine and real read/write tools. Six isolated tmux PTYs cover
  fullscreen/regular at widths 60/100/140, ANSI keyword colour, expanding reads,
  resizing through width 38 and reopening after fixture files are changed.
  Artifacts: `test-results/tui-tool-syntax/`. No live credentials or services.

These terminal tests establish ANSI/cell behavior in tmux, not physical Ghostty
rendering or resolution of the separate Unicode-scroll report.

Verification on the ChromeOS development host: full `make test`, `make vet`,
`make build-web`, `make bun-checks`, and the six PTY scenarios pass. Browser
acceptance is skipped under this host's repository guidance; no web code changed.
A separate on-disk Go cache was used to avoid concurrent cache-clearing builds.

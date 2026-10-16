# Unreleased

Candidate changes for the next patch. These notes do not announce a release.

- Upgrade inference to go-ai 1.3.0 and the build toolchain to Go 1.27.2.
- Reject invalid tool arguments before execution and retain paired error results for replay.
- Preserve browser drafts through uncertain HTTP delivery and session changes.
- Render codemode nested calls, script source, errors and saved output references.
- Add VNC browser acceptance for configured read-only and interactive targets.
- Repair multiline TUI navigation, picker cancellation and external queue redraws.
- Acknowledge stored appearance-command results without restoring handled commands as unsent drafts; retain drafts when persistence fails.
- Repair native model-picker focus and touch dismissal, session prefix typeahead, hidden-file subtree refresh, clipboard failure feedback and clean external Markdown-preview updates.
- Reserve mobile menu space beside editor tabs and align browser acceptance with native controls and revision-conditional saves.
- Extract same-package TUI command/tool-text, shell-turn, queue-launch/cancellation and session-collection boundaries with matching regression tests.

Acceptance is still in progress. See docs/internal/tui-pi-1.3-alignment.md for
measured results and remaining limitations. No deployment is implied.

# Terminal program status

Gi reports program status through OSC 7501 in fullscreen and regular TUI modes. The wire format follows [pi-tui at `503c605528f9`](https://github.com/badlogic/pi-mono/blob/503c605528f9af993c0e37ede468cf884fb0ff5b/packages/tui/src/program-status.ts) and the [Program Status Protocol](https://www.superlogical.com/rex/docs/build/program-status).

## Detection and reports

Startup sends `OSC 7501 ; ? ST`, followed by a primary device-attributes query. A matching reply, terminated by BEL or ST, enables reporting. Optional future reply fields are accepted. DA1 closes the detection window; later replies cannot enable reporting. Silent terminals receive only the query. `PI_PROGRAM_STATUS=1` forces reports and `PI_PROGRAM_STATUS=0` disables both detection and reports, matching pi-tui's override.

Reports use `app=gi`. Identical consecutive reports are suppressed. A supported terminal receives `state=clear` on normal TUI exit.

| Activity | State and metadata |
| --- | --- |
| Initial or switched session | `idle` |
| Active agent or branch summary | `working`, with the session name if named |
| Active compaction | `working`, `Compacting context` |
| Login dialog | `blocked`, `kind=auth`, dialog title |
| Select/editor dialog or editor ask | `blocked`, `kind=question`, dialog title/prompt label |
| Completed turn/manual compaction | `done` |
| Aborted/cancelled turn or manual compaction | `idle` |
| Failure | `error`, first line of the error |

Settings, model and history browsing do not report `blocked`. Dialog closure restores the underlying activity or outcome. Failed individual tools do not by themselves make an agent run fail. Starting another run resets its outcome, so a successful retry can report `done`.

Chat prompts, assistant output, dialog answers and authentication tokens are not report metadata. Explicit dialog labels and session names can be visible to the terminal. Control-character runs in metadata become spaces. Messages are base64-encoded after trimming and UTF-8-safe truncation to 2,048 decoded bytes; invalid app names are omitted.

## Input and verification

The vendored go-tui reader delivers OSC and DA replies as terminal events on the UI loop, rather than editor keys. It buffers fragmented replies and preserves bracketed-paste bodies. Unterminated reply buffering is limited to 8 KiB. A lone Escape remains a key; as with other escape input, splitting immediately after the initial Escape is ambiguous. Windows console input does not implement automatic reply detection; the force/disable overrides remain available.

Verification commands:

```sh
make test TEST_PKGS=./internal/tui TEST_RUN=TestProgramStatus
make test-program-status-reader
make test TEST_PKGS=./internal/tui
make build
make vet
```

`TestProgramStatusPTY` runs ten real-PTY cases: both rendering modes with detected, absent, late, forced and disabled support. It checks lifecycle states, exit clear and preservation of an unsent draft. The tests use isolated stores and no live providers or credentials.

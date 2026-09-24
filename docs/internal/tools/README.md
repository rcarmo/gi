# Built-in tools

Each tool page should document:
- tool name
- purpose
- input schema
- output shape
- path semantics (`workspace` vs `vfs://` where relevant)
- side effects
- failure modes
- examples

Initial priority pages:
- `read.md`
- `write.md`
- `edit.md`
- `script.md`
- `messages.md`
- `schedule.md`
- `exit.md`

## Shell runtime portability

The shell-backed prompt runtime requires `sh` on `PATH` on every platform. The Windows build does not supply a shell or translate commands to PowerShell.

Unix starts the shell in its own process group and kills that group on cancellation. Windows attempts process-tree termination through `%SystemRoot%/System32/taskkill.exe` with a two-second timeout, then falls back to `Process.Kill`; other platforms kill the direct process. Cancellation closes local output readers so a detached descendant retaining a pipe cannot block return indefinitely. Captured output may be partial on cancellation. Normal completion drains stdout/stderr before `Wait` to avoid truncation. The turn engine cancels the context; the runtime is the sole owner of process termination.

`make test-shell-runtime` runs shell drain/cancellation and turn-cancellation regressions with the race detector three times. `make check-cross-build` builds the five CI targets into a disposable directory: Linux and macOS on amd64/arm64, plus Windows amd64. These checks fixed the repeated Windows `Setpgid`/`syscall.Kill` compilation failure; the cross-build is not a Windows runtime test.

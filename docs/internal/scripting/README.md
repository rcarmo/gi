# Scripting

This section documents gi scripting runtimes and bridge semantics.

Primary execution engines in v1 are **Goja (`js`)** and **Joker (`joker`)**.

It should cover:
- supported engines
- how engine selection works
- bridge globals and host functions
- file access semantics for workspace paths and `vfs://`
- state/session access
- output/logging behavior
- examples

See also:
- `joker.md`
- `bridge.md`
- `contract.md`

# Hooks

This section documents agent/runtime hook points.

- [Lifecycle contract](lifecycle.md) defines the current canonical hook taxonomy, shared JSON-safe request/response DTO, action semantics, process-hook protocol, timeout/failure policy, audit rows, and topic publication behavior.

For each hook, keep documenting:
- when it runs
- ordering guarantees
- inputs
- allowed outputs/mutations
- retry/failure semantics
- examples

Hook docs are expected to evolve alongside implementation, but this directory should describe shipped runtime behavior rather than aspirational phases.

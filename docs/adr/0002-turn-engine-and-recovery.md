# ADR 0002: Turn Engine and Recovery

- Status: Draft
- Date: 2026-04-22

## Context

The main failure mode to eliminate is unreliable turn/session handling, especially mid-turn context compaction and retries.

## Decision

Gi turn execution will be based on:
- an **append-only event log**
- **minimal in-memory turn state** reconstructed from persisted events/checkpoints
- **serialized turns per session**
- **concurrent turns across sessions**

Checkpoints are required at:
- tool call boundaries
- provider retry boundaries
- context compaction boundaries

Compaction and retries stay **inside the same turn**.

Gi also needs centralized runtime budget controls for turn handling, including:
- maximum tool calls per turn
- other per-turn and per-session safety/resource limits
- a single code-level configuration structure so limits are defined consistently and applied in one place

A turn is considered complete whenever control returns to the user with consistent persisted state, including:
- successful completion
- partial output with surfaced failure
- cancelled state with preserved outputs

## Recovery policy

- provider failures: retry with exponential backoff; preserve partial output if final failure occurs; surface retry progress to the user
- tool failures: tool-metadata-driven policy; stateful tools handled carefully; retries/failures visible to both model and user
- crash during turn: preserve and return; no auto-resume after crash
- clean restart: auto-resume unfinished turns and show last progress status

## UI requirements

Mid-turn compaction, retries, cancellation, and recovery must always emit visible UI progress/status hints using Piclaw conventions.

# Hook lifecycle contract

## Status
Planned.

## Purpose
This page will define the stable hook points gi exposes to scripts and/or Go extensions.

## Expected hook families
Based on current ADRs and checklists, the initial hook families are expected to include:
- before provider call
- after provider response
- compaction boundaries
- tool start / tool end
- turn start / turn end
- schedule/task lifecycle
- error handling

## Required documentation fields per hook
When a hook is implemented, document:
- hook name
- when it runs
- ordering guarantees
- input payload shape
- whether it may mutate state
- retry / failure behavior
- whether it may emit messages/events
- example usage

## Notes
This page is intentionally a bootstrap contract so hook implementation work has a documentation home immediately.

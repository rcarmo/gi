# Implementation notes

Feature implementation details, plans, investigations and verification results live here. They are repository documentation and are not embedded in Gi's `vfs://reference/...` tree.

## Browse

- [Terminal features](terminal/README.md), including [OSC 7501 program status](terminal/tui-program-status.md)
- [Web features](web/README.md)
- [Plans and assessments](plans/README.md)
- [Maintenance and verification](maintenance/README.md)
- [UX audits and investigations](audits/README.md)

## Where to write

- Put feature implementation notes, source comparisons and test findings in the relevant folder above.
- Keep agent-facing API, tool, scripting, hook and VFS contracts in [the internal reference](../internal/README.md). Update those contracts when behaviour changes.
- Record architecture decisions in [ADRs](../adr/) and track work in the [implementation checklist](../checklists/implementation.md).
- Use the [feature matrix](../feature-parity.md) for current release and acceptance status. Dated notes retain their original scope; their counts and gaps describe the named revision.

The September re-audit JSON/CSV datasets retain historical source paths. Use the accompanying Markdown navigation for current document locations.

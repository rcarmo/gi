# Skills and packages

This section documents:
- skill structure
- frontmatter expectations
- associated scripts/assets
- package/import conventions
- managed VFS storage rules
- discovery and execution semantics

## Current TUI/runtime behavior

Skills are discovered from:

- `.gi/skills/*/SKILL.md`
- `.pi/skills/*/SKILL.md`

`/skills [query]` lists discovered skills with:

- name and description;
- `/skill:<name> [args]` invocation hint;
- source path;
- metadata warnings.

`/skill:<name> [args]` loads the matching `SKILL.md` into the transcript and echoes optional args for clarity.

## Metadata validation

Gi accepts legacy/minimal skills but warns when Agent Skills-style metadata is incomplete:

- missing `Name` field — falls back to the skill directory name;
- missing `Description` field — falls back to the first prose/header line when available;
- empty `Name` field;
- empty `Description` field.

Warnings are visible in `/skills` output so authors can fix packages without breaking existing local skills.

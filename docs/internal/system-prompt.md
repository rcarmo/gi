# System prompt (#32)

gi builds its system prompt as Pi does (`core/system-prompt.js`): named,
ordered sections, each replaceable on its own.

| Section | Content |
|---|---|
| `preamble` | gi's: `You are <assistant>, an expert coding assistant operating inside gi, a coding agent harness. You help <user> by …` (no tags) |
| `tools` | `- <name>: <snippet>` for each declared tool with a snippet, then Pi's line about other custom tools |
| `rules` | Pi's `buildRules`: the shell rule (`Use shell for file operations like ls, rg, find` when no grep/find/ls tool), the declared tools' guidelines, gi's rules, then Pi's `Be concise in your responses` and `Show file paths clearly when working with files`; trimmed and deduplicated |
| `docs` | gi's: where gi's reference lives (`vfs://reference/...`) |
| `project_context` | context files (agent directory, then root to workspace; [config-files.md](config-files.md)), in Pi's `<project_instructions path="…">` form |
| `skills` | Pi's `formatSkillsForPrompt` (`<available_skills>` with name, description, location) |
| `cwd` | the workspace root |
| `mcp_servers` | MCP servers reached through codemode or `tool_search` ([mcp.md](mcp.md)) |

Every section but the preamble is wrapped in `<name>…</name>`; sections are
joined by blank lines. `internal/prompt` is the builder;
`TestBuildSectionsMatchesPi` compares every section except the preamble and
docs (gi's wording) with Pi's own output (`scripts/golden-system-prompt.mjs`).

A configured `SystemPrompt` (for example the acceptance fixture) is Pi's
`customPrompt`: it replaces the preamble, and the tools, rules and docs
sections are left out.

## Tool contributions

A tool contributes a snippet and guidelines while it is declared
(`RegisteredTool.PromptSnippet` / `PromptGuidelines`; built-in tools fall back
to the table in `internal/turn/prompt_sections.go`):

- `read`, `write`, `edit`, `tool_search`: Pi's text;
- `codemode`: Pi's snippet and guideline (`vendor/codemode-texts.json`);
- gi's tools (`shell`, `rtk`, `tools`, `skills`, `messages`, `script`,
  `compact`, `peering`): gi's text. MCP tools have none.

Because the tools section follows the declared tools, switching codemode on
or loading tools changes the prompt from the next turn, as in Pi.

Gi's script contribution explicitly requires the `gi` JavaScript namespace
and captured `console.log` output. Common rules distinguish VFS paths from
OS paths and require recorded turn/inference errors when diagnosing a failed
turn; a successful tool retry alone is not evidence of a runtime repair.
The docs section links the [scoped-models guide](scoped-models.md) so model
shortlists are explained in gi's context rather than another framework's.

## Changes mid-conversation

The sections when a session starts lead the conversation. Each turn the
engine builds the sections again; a change is recorded once (Pi's
`diffSystemPromptSections`) and sent as a system message before the prompt
whose turn introduced it, with pi-ai's text:

```
Updated system prompt section "tools":

<tools>
…
</tools>
```

(`Removed system prompt section "<name>".` for a removed section). Updates
keep their position on later turns, so earlier messages stay cached. The
initial sections and the updates are kept in the session state
(`system_prompt_sections`); stored messages are unchanged.

- Models whose compat does not declare `supportsMidConvoSystemMessages` get
  the updates folded into the leading prompt instead (pi-ai's
  `collapseSystemMessages`: a replaced section keeps its place, a new one
  goes last).
- If compaction removes an update's anchor, the update moves into the leading
  prompt.
- A hook that replaces the system prompt (`before_agent_start`, `context`,
  `before_provider_request`) replaces it whole; updates are then not added.
- Sessions recorded before this prompt start with the new sections as their
  initial ones.

## Not ported

- Pi's `SYSTEM.md` / `APPEND_SYSTEM.md` prompt files and extension-defined
  sections.

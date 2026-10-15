# Configuration file lookup (#26)

gi reads Pi's configuration files unchanged and lets gi's own directories
override them (`internal/config/config_paths.go`).

| Level | Lookup order |
|---|---|
| user | `$GI_CODING_AGENT_DIR` or `~/.gi/agent`, then `$PI_CODING_AGENT_DIR` or `~/.pi/agent` (Pi's `getAgentDir`) |
| project | `<workspace>/.gi`, then `<workspace>/.pi` |

For a file, the first existing location wins; files are never merged across
the two. A file is written where it was read; when neither location has it,
it is created in Pi's location, so it stays shared with Pi.

| File | Level | Where |
|---|---|---|
| `settings.json` | project, then user (Pi's merge of global under project) | `config.Load`, `settings_file.go` (reads and writes the project file that exists) |
| `auth.json` | user | `inference.AuthFilePath`, credential store (writes the file read, holding Pi's `auth.json.lock`; refreshes Copilot tokens in place, like Pi) |
| `models-store.json` | user | `inference.PiModelsStorePath`; refreshed with `modelCatalogUrl` set (below), holding Pi's `models-store.json.lock` |
| `mcp.json` | user and project | `internal/mcp` |
| `mcp-auth.json` | user | MCP OAuth credentials, shared with Pi |

Directories of items are scanned `.gi` before `.pi`, the first item of a
name winning: `skills/`, `tools/` (script tool manifests), `extensions/`.
The workspace index includes `.gi/skills` as a skills root only where it
exists, so other workspaces keep their index fingerprint.

## Provider transport

Pi's `transport` setting is read from project settings over user settings and passed to go-ai for normal turns, side prompts, compaction and branch summaries. Values are `auto` (default), `sse`, `websocket` and `websocket-cached`. The legacy `websockets` boolean maps to `websocket` or `sse` when the enum is absent. Reading either form leaves the file unchanged.

AgentsInTheCloud's managed Pi settings use `sse`. Gi honours that preference instead of making a WebSocket attempt first. Codex WebSocket connections now explicitly allow incoming messages up to 4 MiB in go-ai, including auto and cached mode, instead of inheriting coder/websocket's 32 KiB default. Larger messages are rejected to bound memory use.

## VNC target environment

VNC startup configuration uses `GI_WEB_VNC_TARGETS` (JSON target array) and `GI_WEB_VNC_ALLOW_DIRECT` (only literal `true` enables direct host:port access). Defaults disable VNC. These are operator environment settings, not model-editable Pi preferences. See [VNC protocol and security limits](web-vnc.md).

## Skills (#36)

`internal/skills` ports Pi's `loadSkills`: user skills first
(`<gi agent dir>/skills`, then `<Pi agent dir>/skills`), then project skills:
`.gi/skills`, ancestor `.agents-in-the-cloud/skills` and `.agents/skills`
(closest directory first), then the workspace's `.pi/skills`. The first skill
of a name wins, and a file reached twice through symlinks loads once. This
includes host skills when Gi starts inside a nested repository. Within a directory, Pi's rules:

- a directory containing `SKILL.md` is one skill (not searched further);
- otherwise `.md` files directly in the skills root are skills, and
  subdirectories are searched for `SKILL.md`;
- dot entries and `node_modules` are skipped; `.gitignore`, `.ignore` and
  `.fdignore` exclude paths below them;
- the YAML front matter must have a `description` (otherwise the file is not
  a skill); `name` defaults to the directory; names are checked against the
  Agent Skills spec as warnings; `disable-model-invocation: true` keeps a
  skill out of the system prompt (it can still be invoked with
  `/skill:name`).

The prompt tells the model to load skills with `read`, so user skill
directories (and the directories of discovered user skills) are readable
outside the workspace; writes there are refused (`tools.SetReadOnlyRoots`).

## Context files (#36)

`config.LoadContextFiles` ports Pi's `loadProjectContextFiles`: the agent
directory's file (gi's, else Pi's), then one file per directory from the
filesystem root down to the workspace, each the first of `AGENTS.override.md`,
`AGENTS.md`, `AGENTS.MD`, `CLAUDE.md`, `CLAUDE.MD`. In a linked git worktree
nested in its main repository, the worktree's file shadows the main
repository's. AgentsInTheCloud's `.agents-in-the-cloud/AGENTS.md` is loaded
after the normal context for each ancestor directory. Canonical-path
comparison avoids duplicating linked instructions. They become the prompt's
`project_context` entries.

`.piclaw/config.json` (assistant and user names and avatars) is Piclaw's
file, not Pi's; gi reads and writes it only there.

## Copilot credentials (`internal/inference/copilot_auth.go`)

As in Pi (pi-ai `resolveRefreshCredential` and `refreshGitHubCopilotToken`):
the Copilot token in `access` is used until `expires`, with the API base
URL taken from the token's `proxy-ep`. Then, holding `auth.json.lock`
(proper-lockfile's directory lock, as Pi's `FileAuthStorageBackend`), the
entry is read again, since another gi or Pi may have refreshed it, and, if
still expired, refreshed through go-ai's port. A refresh also fetches
`{copilot-base}/models` and stores the picker-enabled, tool-capable models
the policy allows as `availableModelIds` (Pi's picker and policy rules,
with Pi's policy fallback for the individual endpoint); they filter the
Copilot model listing. A failed model fetch fails the refresh, as in Pi.

At startup (`cmd/gi`, in the background, skipped with `PI_OFFLINE` like
Pi's offline mode), expired Copilot credentials are refreshed, as Pi's
startup model refresh does.

## Model catalogue (`internal/inference/model_catalog.go`)

A port of Pi's `withRemoteCatalog` and `FileModelsStore`. For each go-ai
provider with credentials (Pi resolves the credential first; radius is
excluded), `GET <modelCatalogUrl>/api/models/providers/<id>?types=chat,image,classifier`
at most every four hours. A cached body is revalidated with
`If-None-Match`; 304 only moves `checkedAt`; 404/501 store an empty
catalogue with `lastModified: 0`; other failures keep the body and its
ETag. Requests are retried as Pi's `fetchWithRetry` (two immediate retries
on 408/425/429/5xx and network errors, four seconds per attempt; 15 seconds
overall at startup). Entries are written one provider at a time under
`models-store.json.lock`, keeping the others. Models the store adds are
registered; go-ai's definitions are never replaced.

Difference: Pi always uses `https://pi.dev`. gi refreshes only when the
user settings set `modelCatalogUrl` (project settings cannot), so offline
and test instances never call out; otherwise it reads the store Pi
refreshes.

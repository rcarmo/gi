# Gi spec chat transcript

Source: `/workspace/.piclaw/store/messages.db`
Chat: `web:default:branch:e032ce143fdc`
Format: verbatim message export from the SQLite store.

## Rui Carmo @ 2026-04-22T10:37:41.229Z

```text
Ok, new project. gi, a coding agent based on go-ai and lessons learned from pi, piclaw and vibes. Create the project folder and Ask me what are my pain points so we can start on a spec
```

## Neo @ 2026-04-22T10:38:06.029Z

```text
Created:
- `/workspace/projects/gi`

To start the spec, tell me your pain points.

Use this if you want a structured dump:

## Pain points
1. **What frustrates you most** with current coding agents?
2. **Where pi/piclaw/vibes fall short** for your workflow
3. **What go-ai already solves** that you want to keep
4. **What you want gi to feel like**
   - CLI?
   - TUI?
   - library?
   - daemon?
   - web UI?
5. **Core jobs gi must do**
   - code editing
   - repo audit
   - planning
   - long-running tasks
   - infra ops
   - browser automation
   - document generation
   - multi-agent workflows
6. **Non-goals**
   - what gi should explicitly not do
7. **Reliability pain points**
   - retries
   - context drift
   - broken tool calls
   - partial edits
   - poor observability
   - flaky model/provider behavior
8. **UX pain points**
   - verbosity
   - interruption handling
   - progress reporting
   - approval flow
   - diff review
   - memory
9. **Architecture pain points**
   - provider abstraction
   - tool system
   - prompts
   - session model
   - persistence
   - testing
10. **What would make you switch**
   - what would make gi clearly better than what you use now

If you want, I can also turn your answers into a first-pass spec with sections like:
- vision
- user stories
- architecture
- tool/runtime model
- memory/session model
- provider strategy
- testing strategy
- MVP vs later phases
```

## Rui Carmo @ 2026-04-22T10:38:12.293Z

```text
Ask me one question at a time
```

## Neo @ 2026-04-22T10:38:17.654Z

```text
What frustrates you most about current coding agents?
```

## Rui Carmo @ 2026-04-22T10:38:55.085Z

```text
Pi keeps changing upstream and I want a unified TUI/CLI/web ui
```

## Neo @ 2026-04-22T10:39:04.688Z

```text
Got it: **upstream churn** and lack of a **single coherent TUI/CLI/web UI**.

Next question:

**What do pi/piclaw/vibes each get wrong for you today?**
```

## Rui Carmo @ 2026-04-22T10:39:29.991Z

```text
The turn handling mechanism is buggy and requires too much maintenance
```

## Neo @ 2026-04-22T10:39:37.015Z

```text
Got it: **turn handling is fragile and high-maintenance**.

Next question:

**If gi had to fix just one thing about turn handling, what should it do differently?**
```

## Rui Carmo @ 2026-04-22T10:41:14.758Z

```text
It should be stupidly reliable and auto recover from provider failures, tool failures, context compaction, session sizes and rotation, etc. it should return control to the user solidly after each turn
```

## Neo @ 2026-04-22T10:41:30.126Z

```text
Got it: **turn execution must be boringly reliable**, self-healing around provider/tool/context/session failures, and always **hand control back cleanly after every turn**.

Next question:

**When gi returns control after a turn, what exact state guarantees do you want?**
```

## Rui Carmo @ 2026-04-22T10:41:40.685Z

```text
What do you mean?
```

## Neo @ 2026-04-22T10:41:47.115Z

```text
I mean: after any turn finishes, fails, or partially succeeds, **what must be true from the user’s point of view?**

For example:
- the UI is responsive again
- no half-running hidden task is left behind
- partial tool output is preserved
- the conversation state is consistent
- retries are either done or clearly reported
- the agent is ready for the next prompt
- pending approvals are explicit
- background tasks are detached and tracked

So the question is:

**After a turn ends, what guarantees should gi give you every time?**
```

## Rui Carmo @ 2026-04-22T10:42:04.184Z

```text
All the above
```

## Neo @ 2026-04-22T10:42:11.472Z

```text
Got it: **all of those should be guaranteed**.

Next question:

**What kind of unified interface do you want first: TUI-first, CLI-first, or web-first?**
```

## Rui Carmo @ 2026-04-22T10:42:54.615Z

```text
All of them. Non negotiable. Some UX parity (except CLI, which should be used for batching and maintenance)
```

## Neo @ 2026-04-22T10:43:03.107Z

```text
Got it: **TUI + web with UX parity**, and **CLI for batch/maintenance**. Non-negotiable.

Next question:

**Which UI should be the canonical one internally — the source of truth that the others map onto?**
```

## Rui Carmo @ 2026-04-22T10:43:18.539Z

```text
The web ui.
```

## Neo @ 2026-04-22T10:43:25.675Z

```text
Got it: **web UI is the canonical interface**, and TUI/CLI should map onto it.

Next question:

**What are the must-have capabilities for gi v1?**
```

## Rui Carmo @ 2026-04-22T10:43:36.906Z

```text
In what category
```

## Neo @ 2026-04-22T10:43:46.224Z

```text
I mean **core product capabilities** — the things gi must be able to do in v1, regardless of UI.

For example:
- chat / agent turns
- read/write/edit files
- shell commands
- diff review
- tool calling
- long-running/background tasks
- planning / task tracking
- memory / notes
- web browsing
- browser automation
- multi-repo support
- session recovery
- provider failover
- approvals / permissions
- attachments / artifacts

So the question is:

**What are the must-have capabilities for gi v1?**
```

## Rui Carmo @ 2026-04-22T10:44:23.089Z

```text
Let’s do that in smaller chunks. Core features, core tools, etc
```

## Neo @ 2026-04-22T10:44:32.653Z

```text
Good.

First chunk:

**What are the core features gi must have in v1?**  
Not tools yet — product behaviors. For example:
- reliable turns
- session recovery
- background tasks
- memory
- approvals
- artifacts
- planning
- multi-chat/session management
```

## Rui Carmo @ 2026-04-22T10:45:24.247Z

```text
No need for planning. No need for approvals either. Need scheduling. Skills.
```

## Neo @ 2026-04-22T10:45:34.447Z

```text
Got it.

So far for **core features**:
- **reliable turns**
- **session recovery**
- **scheduling**
- **skills**
- **no planning**
- **no approvals**

Next question:

**Do you want persistent memory in v1?**  
If yes, what kind:
- conversation history only
- durable notes/knowledge
- project memory
- user preferences
- all of the above
```

## Rui Carmo @ 2026-04-22T10:46:16.733Z

```text
All the above. Modeled on piclaw’s system, with a more developed memory system coming later
```

## Neo @ 2026-04-22T10:46:26.416Z

```text
Got it: **all of the above for v1**, modeled on **piclaw’s memory system**, with a more advanced memory layer later.

Next question:

**Do you want background/long-running tasks in v1, or should every interaction stay strictly turn-bounded except for scheduling?**
```

## Rui Carmo @ 2026-04-22T10:46:47.146Z

```text
I need a way to track background tasks yes
```

## Neo @ 2026-04-22T10:46:51.195Z

```text
Got it: **background tasks are required in v1**, and they must be **trackable**.

Next question:

**What core tools must gi have in v1?**
```

## Rui Carmo @ 2026-04-22T10:48:54.323Z

```text
shell, read, write, edit, messages (message list search, extraction, attachments) schedule, exit, script. MCP and ACP in 1.1 for additional tools and subagents
```

## Neo @ 2026-04-22T10:49:04.246Z

```text
Got it.

## Core tools for v1
- `shell`
- `read`
- `write`
- `edit`
- `messages`
  - list/search
  - extraction
  - attachments
- `schedule`
- `exit`
- `script`

## Deferred to 1.1
- `MCP`
- `ACP`
- additional tools
- subagents

Next question:

**What should “script” mean in gi?**  
For example:
- run named project scripts
- run reusable agent skills
- run arbitrary local code
- packaged workflows
- all of the above
```

## Rui Carmo @ 2026-04-22T10:49:57.420Z

```text
Use the joker Clojure interpreter to run scripts that can introspect gi and go-ai state and extend agent behavior
```

## Neo @ 2026-04-22T10:50:01.766Z

```text
Got it: **`script` means running Joker/Clojure scripts** that can:
- introspect **gi** state
- introspect **go-ai** state
- extend agent behavior

Next question:

**How much power should scripts have in v1?**  
Should they be able to:
- call all core tools
- read/write agent/session state
- register skills/hooks
- mutate memory
- spawn background tasks
- access provider internals

Or should some of that be restricted?
```

## Rui Carmo @ 2026-04-22T10:50:14.125Z

```text
Everything
```

## Neo @ 2026-04-22T10:50:23.282Z

```text
Got it: **scripts get full power in v1**.

Next question:

**What should a skill be in gi?**  
For v1, should a skill be:
- just a prompt/instruction bundle
- a prompt plus scripts
- a packaged workflow with metadata
- all of the above
```

## Rui Carmo @ 2026-04-22T10:50:48.495Z

```text
A prompt/SKILL.md with frontmatter and associated scripts
```

## Neo @ 2026-04-22T10:50:53.155Z

```text
Got it: **a skill is `SKILL.md` + frontmatter + associated scripts**.

Next question:

**How should skills be discovered and loaded in v1?**  
For example:
- auto-discovered from workspace folders
- installed from packages
- both
- with per-project overrides or inheritance
```

## Rui Carmo @ 2026-04-22T10:51:38.617Z

```text
Both. We will define the packaging approach later, but it should be simple to extract a specific skill from a GitHub URL and download it
```

## Neo @ 2026-04-22T10:51:43.385Z

```text
Got it: **both local discovery and packaged skills**, and packaging should later support **simple extraction/download of a specific skill from a GitHub URL**.

Next question:

**How should the workspace be modeled in v1?**  
Should gi assume:
- one current workspace root
- multiple project roots under one workspace
- arbitrary folders per session
- all of the above
```

## Rui Carmo @ 2026-04-22T10:52:10.987Z

```text
One workspace, like piclaw
```

## Neo @ 2026-04-22T10:52:16.076Z

```text
Got it: **one workspace root**, Piclaw-style.

Next question:

**Should gi be single-user only in v1, or do you want multi-user/multi-tenant support from the start?**
```

## Rui Carmo @ 2026-04-22T10:53:49.247Z

```text
I want three kinds of identities: users (the local user), agents and remote agents (1.1). I do need multi-session (and simultaneous contexts) from the start, sharing the same SQLite database to track state, chat, attachments, keychain, etc.
```

## Neo @ 2026-04-22T10:54:01.695Z

```text
Got it.

## Identity/session model so far
- **identities**
  - local **users**
  - **agents**
  - **remote agents** in 1.1
- **from v1**
  - **multi-session**
  - **simultaneous contexts**
  - shared **SQLite** state store for:
    - chat
    - attachments
    - keychain
    - other agent/session state

Next question:

**How do you want sessions to behave in v1?**  
Should a session be:
- tied to one chat/thread
- tied to one workspace task
- resumable across UIs
- branchable/forkable
- all of the above, or something else?
```

## Rui Carmo @ 2026-04-22T10:55:16.386Z

```text
I dont want them tied to a task. The rest is needed, but the branching strategy can be much simpler than pi.
```

## Neo @ 2026-04-22T10:55:20.366Z

```text
Got it:

- **not tied to tasks**
- **tied to chat/thread**
- **resumable across UIs**
- **branchable/forkable**
- but with a **simpler branching model than pi**

Next question:

**What should the simplified branching model be?**  
For example:
- linear chat with optional snapshots
- one parent + named branches
- lightweight forks from any message
- separate sessions linked by ancestry
- something else
```

## Rui Carmo @ 2026-04-22T10:55:59.853Z

```text
Lightweight forks. A fork can then be tied to a new UI session/agent that inherits previous context
```

## Neo @ 2026-04-22T10:56:03.658Z

```text
Got it: **lightweight forks**, where a fork can become a new UI session/agent inheriting prior context.

Next question:

**What should happen to background tasks when a session is forked?**
```

## Rui Carmo @ 2026-04-22T10:56:22.320Z

```text
They stick around and all forks know about them until they die
```

## Neo @ 2026-04-22T10:56:25.587Z

```text
Got it: **background tasks are shared across forks and remain visible until they finish or are stopped**.

Next question:

**Should background tasks belong to the workspace globally, or to a session/fork lineage with global visibility?**
```

## Rui Carmo @ 2026-04-22T10:56:32.551Z

```text
Workspacr
```

## Neo @ 2026-04-22T10:56:36.607Z

```text
Got it: **background tasks are workspace-global**.

Next question:

**What provider model do you want in v1?**  
For example:
- OpenAI-compatible first
- multiple first-class providers
- go-ai providers reused directly
- custom provider plugin layer over go-ai
```

## Rui Carmo @ 2026-04-22T10:57:14.196Z

```text
Go-ai with compatibility with pi settings files for all providers. Full settings compatibility.
```

## Neo @ 2026-04-22T10:57:18.320Z

```text
Got it: **go-ai as the provider layer**, with **full compatibility with pi settings files** across providers.

Next question:

**Do you want gi to preserve pi’s model/provider naming and config semantics exactly, or only be able to import/read them compatibly?**
```

## Rui Carmo @ 2026-04-22T10:57:31.067Z

```text
Preserve
```

## Neo @ 2026-04-22T10:57:34.326Z

```text
Got it: **preserve pi naming and config semantics exactly**.

Next question:

**What should the keychain/secrets model be in v1?**  
For example:
- Piclaw-compatible keychain
- SQLite-backed only
- OS keychain integration
- env vars plus keychain
- something else
```

## Rui Carmo @ 2026-04-22T10:58:12.720Z

```text
Piclaw, with the current env injection features for all tools, scripts, skills, etc. Direct port
```

## Neo @ 2026-04-22T10:58:17.879Z

```text
Got it: **direct Piclaw keychain port**, including **env injection for tools, scripts, skills, etc.**

Next question:

**What should the artifact/attachment model be in v1?**  
Should gi support:
- uploaded files
- generated files
- inline image/media rendering
- downloadable artifacts
- attachment references inside chats
- all of the above
```

## Rui Carmo @ 2026-04-22T10:59:21.355Z

```text
All the above, like piclaw, including indexing. We will need to support the same viewing UX in the web and try to render images in advanced terminals like ghostty
```

## Neo @ 2026-04-22T10:59:25.398Z

```text
Got it: **full Piclaw-style attachments/artifacts**, including:
- uploads
- generated files
- inline rendering
- downloads
- chat references
- indexing
- web viewing parity
- image rendering in capable terminals like Ghostty

Next question:

**What should scheduling cover in v1?**  
Should it support:
- one-off tasks
- recurring tasks
- agent prompts
- shell/script jobs
- notifications/messages back into chat
- all of the above
```

## Rui Carmo @ 2026-04-22T10:59:37.684Z

```text
All of the above.
```

## Neo @ 2026-04-22T10:59:40.687Z

```text
Got it: **all of the above** for scheduling.

Next question:

**What should the TUI be able to do in v1 that the web UI can do too?**  
I’m asking to pin down the required parity surface.
```

## Rui Carmo @ 2026-04-22T11:01:00.345Z

```text
we should look into what go-tui supports, but it should be minimalistic. Good scrollback, minimal formatting, expandable text field, equivalent progress and status notifications to what piclaw currrently has in its compose box
```

## Neo @ 2026-04-22T11:01:06.009Z

```text
Got it: **minimalist TUI**, with:
- good scrollback
- minimal formatting
- expandable text field
- progress/status notifications equivalent to Piclaw’s compose box

Next question:

**What is the minimum CLI surface for v1?**  
For example:
- send a prompt to a session
- create/list/resume sessions
- run batch jobs non-interactively
- inspect messages/tasks/artifacts
- maintenance/admin commands
```

## Rui Carmo @ 2026-04-22T11:02:03.801Z

```text
All the above. It should manage the keychain, run scripts directly, take a prompt, prune state, search, etc.
```

## Neo @ 2026-04-22T11:02:08.117Z

```text
Got it: **full operational CLI** in v1, including:
- prompt submission
- create/list/resume sessions
- batch jobs
- inspect messages/tasks/artifacts
- maintenance/admin
- keychain management
- direct script execution
- state pruning
- search

Next question:

**What should search cover in v1?**  
For example:
- chat/messages
- attachments
- indexed file contents
- memory/notes
- tasks
- scripts/skills
- all of the above
```

## Rui Carmo @ 2026-04-22T11:02:32.868Z

```text
All the above within reason for attachments. SQLite FTS
```

## Neo @ 2026-04-22T11:02:36.671Z

```text
Got it: **SQLite FTS-backed search** across:
- chats/messages
- indexed file contents
- memory/notes
- tasks
- scripts/skills
- attachments where reasonable

Next question:

**What should “within reason for attachments” mean in v1?**  
For example:
- filename + metadata only
- OCR/text extraction when possible
- document text indexing
- image captions/thumbnails only
- something else
```

## Rui Carmo @ 2026-04-22T11:03:31.943Z

```text
File name, metadata, text indexing
```

## Neo @ 2026-04-22T11:03:36.688Z

```text
Got it: **attachments in v1 should support filename, metadata, and text indexing**.

Next question:

**What should the web UI minimally include in v1 besides chat itself?**  
For example:
- session list
- fork history
- task list
- artifact browser
- search UI
- keychain/settings UI
- model/provider picker
```

## Rui Carmo @ 2026-04-22T11:04:28.720Z

```text
Identical to piclaw, with the same UI extensions (but plain JS)
```

## Neo @ 2026-04-22T11:04:32.801Z

```text
Got it: **web UI v1 should be functionally identical to Piclaw**, including the same UI extensions, but implemented in **plain JS**.

Next question:

**Do you want gi to keep Piclaw’s message/content model exactly, or should we simplify it in v1?**
```

## Rui Carmo @ 2026-04-22T11:04:43.568Z

```text
What do you mean
```

## Neo @ 2026-04-22T11:04:47.238Z

```text
I mean the internal structure of chat/output items.

For example, Piclaw has a fairly rich message model with things like:
- user/agent/system messages
- structured content blocks
- attachments
- cards/widgets
- tool outputs
- status/progress updates
- message metadata

So the question is:

**Should gi preserve that same rich structured message model, or do you want a simpler internal model and just render it similarly in the UI?**
```

## Rui Carmo @ 2026-04-22T11:04:59.786Z

```text
Same
```

## Neo @ 2026-04-22T11:05:04.278Z

```text
Got it: **same structured message/content model as Piclaw**.

Next question:

**Should gi v1 preserve Piclaw’s Adaptive Cards and dashboard widgets exactly, or just the underlying message/attachment primitives first?**
```

## Rui Carmo @ 2026-04-22T11:06:39.865Z

```text
No adaptive cards. Simpler confirmation dialogs and tables drop-downs/checkboxes and a submit button. I will want the interactive widget and inline charting
```

## Neo @ 2026-04-22T11:06:43.730Z

```text
Got it:

- **no Adaptive Cards**
- use **simpler interactive UI primitives**:
  - confirmation dialogs
  - tables
  - dropdowns
  - checkboxes
  - submit button
- keep **interactive widgets**
- keep **inline charting**

Next question:

**Should those interactive UI elements exist only in the web UI in v1, or do you want TUI equivalents from the start too?**
```

## Rui Carmo @ 2026-04-22T11:07:12.410Z

```text
TUI only for forms and image previews in advanced terminals
```

## Neo @ 2026-04-22T11:07:18.073Z

```text
Got it: **web gets the full interactive UI**, and **TUI only needs forms plus image previews in capable terminals**.

Next question:

**What should image preview support look like in v1?**  
For example:
- just local attachments
- generated images too
- inline terminal preview when supported, fallback to path/metadata otherwise
- zoom/open externally
- all of the above
```

## Rui Carmo @ 2026-04-22T11:08:36.877Z

```text
All the above
```

## Neo @ 2026-04-22T11:08:40.539Z

```text
Got it: **all of the above** for image previews.

Next question:

**What should gi store in SQLite in v1 besides chat/messages/attachments/keychain?**  
For example:
- sessions
- forks/ancestry
- scheduled jobs
- background tasks
- memory indices
- search index
- tool/script run logs
- provider usage/cost stats
```

## Rui Carmo @ 2026-04-22T11:13:34.420Z

```text
All the above
```

## Neo @ 2026-04-22T11:13:40.226Z

```text
Got it: **all of the above** in SQLite for v1.

Next question:

**What should gi log and expose for observability in v1?**  
For example:
- provider requests/retries/failures
- tool calls and durations
- script runs
- background task lifecycle
- scheduling events
- session/fork events
- token/cost usage
- compaction/rotation/recovery events
- all of the above, or a filtered subset
```

## Rui Carmo @ 2026-04-22T11:14:46.089Z

```text
All the above. We can use SQLite JSON storage for the vast majority of it
```

## Neo @ 2026-04-22T11:14:52.100Z

```text
Got it: **all of the above for observability**, with **SQLite JSON storage for most event data**.

Next question:

**What should the session rotation model be in v1?**  
For example:
- automatic rollover when context gets too large
- new session with linked ancestry
- same visible chat, different underlying turn/session segments
- manual + automatic rotation
- something else
```

## Rui Carmo @ 2026-04-22T11:17:40.363Z

```text
Automatic rollover based on timeline length per session
```

## Neo @ 2026-04-22T11:17:44.542Z

```text
Got it: **automatic session rollover based on timeline length per session**.

Next question:

**When a session rolls over, should the new session inherit full prior context, a compacted summary/context pack, or both?**
```

## Rui Carmo @ 2026-04-22T11:20:27.391Z

```text
Compacted context and the last 10 messages or so (setting variable), rebuilt from database data
```

## Neo @ 2026-04-22T11:20:35.133Z

```text
Got it: **automatic rollover builds a new session from DB data using compacted context plus the last N messages**, with **N configurable**.

Next question:

**What should trigger context compaction in v1?**  
For example:
- token budget thresholds
- provider/model context window
- timeline length
- explicit recovery after failure
- all of the above
```

## Rui Carmo @ 2026-04-22T11:54:04.816Z

```text
All the above except timeline length. That should not be a direct factor since context won’t cover the entire timeline thanks to summaries
```

## Neo @ 2026-04-22T11:54:15.096Z

```text
Got it: **compaction triggers on token/context limits and recovery paths, not timeline length**.

Next question:

**Who generates the summaries/compacted context in v1?**  
For example:
- the same active model
- a dedicated cheaper summarizer model
- configurable per provider
- rule-based/scriptable pipeline with model fallback
```

## Rui Carmo @ 2026-04-22T11:54:53.514Z

```text
Whatever pi does plus custom compaction hooks we will define
```

## Neo @ 2026-04-22T11:54:57.317Z

```text
Got it: **start with Pi-style summarization/compaction behavior**, and allow **custom compaction hooks** later.

Next question:

**What should happen when compaction fails in v1?**  
For example:
- retry with the same model
- fall back to a cheaper/smaller model
- fall back to rule-based truncation
- ask the user
- abort the turn but preserve recovery state
```

## Rui Carmo @ 2026-04-22T11:56:01.414Z

```text
We need to have all the options here, but I don’t know how to deal with a separate model yet. We will need to make it a setting
```

## Neo @ 2026-04-22T11:56:09.028Z

```text
Got it: **compaction failure handling should be policy-driven and configurable**, with support for:
- retry
- fallback strategies
- abort with preserved recovery state

And the **separate compaction model should be a setting**, even if we don’t lock that design yet.

Next question:

**What should happen when a provider call fails mid-turn in v1?**  
For example:
- automatic retry on same provider/model
- provider failover
- model fallback within same provider
- preserve partial output
- convert to recoverable session error
- always return control immediately with status
```

## Rui Carmo @ 2026-04-22T11:57:38.754Z

```text
Automatic retry with exponential backoff (go-ai will help but needs setting) preserving partial output if everything fails but surfacing information to the user of retry progress and strategies
```

## Neo @ 2026-04-22T11:57:45.789Z

```text
Got it: **automatic retry with exponential backoff**, **preserve partial output on ultimate failure**, and **show retry progress/strategy to the user during the turn**.

Next question:

**What should happen when a tool call fails mid-turn in v1?**  
For example:
- retry automatically
- mark tool result as error and let the model continue
- abort the turn
- let scripts/skills define recovery
- all of the above via policy
```

## Rui Carmo @ 2026-04-22T11:58:52.062Z

```text
Depends on the tool. Stateful tools need to be handled carefully, but in general retries surfaced to the model and user, so that scripts and skills can handle it
```

## Neo @ 2026-04-22T11:59:02.492Z

```text
Got it: **tool failure policy is tool-specific**, especially for stateful tools, but in general:
- retries should be visible to the **user**
- retries/failures should be visible to the **model**
- scripts/skills should be able to handle recovery

Next question:

**How should tools declare whether they are safe to retry in v1?**  
For example:
- simple metadata flag (`retryable`, `stateful`, etc.)
- richer policy in tool definition
- script-defined policy
- all of the above
```

## Rui Carmo @ 2026-04-22T11:59:27.331Z

```text
Metadata. We can reuse the piclaw metadata approach for discovery, hints, etc
```

## Neo @ 2026-04-22T11:59:31.528Z

```text
Got it: **tool retry/safety behavior should come from metadata**, reusing the **Piclaw-style metadata/discovery/hints approach**.

Next question:

**What should a tool definition minimally contain in v1?**  
For example:
- name
- description
- input schema
- output shape
- retry/safety metadata
- timeout
- UI hints
- auth/keychain requirements
- something else
```

## Rui Carmo @ 2026-04-22T12:00:05.370Z

```text
The above, as used in piclaw. Plus hooks for progress reporting when called.
```

## Neo @ 2026-04-22T12:00:16.214Z

```text
Got it: **Piclaw-style tool definitions**, plus **progress-reporting hooks during execution**.

Next question:

**How should progress reporting work in v1?**  
For example:
- simple status text updates
- percent progress
- step/substep events
- streaming logs
- structured progress events with all of the above
```

## Rui Carmo @ 2026-04-22T12:00:34.364Z

```text
All the above. Piclaw model
```

## Neo @ 2026-04-22T12:00:42.239Z

```text
Got it: **Piclaw-style structured progress events**, including:
- status text
- percentage
- step/substep
- streaming logs

Next question:

**What should the permissions/sandbox model be in v1, given you don’t want approval flows?**  
For example:
- full local trust by default
- configurable per-tool restrictions
- workspace sandbox boundaries
- read/write path allowlists
- network policy settings
```

## Rui Carmo @ 2026-04-22T12:01:30.452Z

```text
Full local trust for v1. Sandboxing is an OS concern. Some tools might have boundaries in 1.1
```

## Neo @ 2026-04-22T12:01:35.483Z

```text
Got it: **full local trust in v1**, with sandboxing treated as an OS concern; tool boundaries can come later in 1.1.

Next question:

**What should gi’s configuration model look like in v1?**  
For example:
- Piclaw-compatible settings files
- workspace config + user config
- DB-backed runtime settings
- env overrides
- all of the above
```

## Rui Carmo @ 2026-04-22T12:02:33.309Z

```text
Pi/piclaw settings files plus occasional metadata and minimal critical env overrides (keychain master key, workspace details, etc)
```

## Neo @ 2026-04-22T12:02:39.482Z

```text
Got it: **Pi/Piclaw-compatible settings files first**, plus **small critical env overrides** for bootstrap/runtime essentials.

Next question:

**What language/runtime constraints do you want for gi itself?**  
For example:
- pure Go as much as possible
- plain JS for web UI
- Joker for scripting
- no Node/Bun dependency in core
- something else
```

## Rui Carmo @ 2026-04-22T12:04:00.605Z

```text
The four you mention are critical. Tinygo for creating WASM components, bun as scripting tool for vendoring and bundling, plain JS to simplify writing code
```

## Neo @ 2026-04-22T12:04:07.078Z

```text
Got it. Critical stack so far:

- **Go** for core runtime
- **plain JS** for web UI
- **Joker** for scripting
- **no Node/Bun dependency in core**
- **TinyGo** for WASM components
- **Bun only for vendoring/bundling/build-time scripting**

Next question:

**What should gi’s extension model be in v1?**  
For example:
- built-in tools + skills only
- load local extensions from the workspace
- installable extension packages
- WASM components for UI/tooling
- something else
```

## Rui Carmo @ 2026-04-22T12:04:41.742Z

```text
There are both UX and tooling scopes there. Separate them
```

## Neo @ 2026-04-22T12:04:46.779Z

```text
Agreed.

Let’s separate them.

**Next question: what should the _tooling extension model_ be in v1?**  
Ignore UI/UX extensions for now. Just tools/skills/runtime behavior.
```

## Rui Carmo @ 2026-04-22T12:06:34.763Z

```text
Built in tools plus scripts plus MCP/ACP. Nothing else. We will be smart about our MCP adapters. QuickJS is an option for 1.1 after we define the internal scripting bridge (filesystem, internal state, sockets, web requests)
```

## Neo @ 2026-04-22T12:06:40.376Z

```text
Got it: **tooling extensions in v1 = built-in tools + Joker scripts**, with **MCP/ACP in 1.1**, and possible **QuickJS in 1.1** after the scripting bridge is defined.

Next question:

**What should the _UI/UX extension model_ be in v1?**  
Ignore tools now. Just web/TUI extensibility.
```

## Rui Carmo @ 2026-04-22T12:08:35.143Z

```text
Web and TUI will define core interfaces for status, drafts, intents, input, streaming, slash commands that will be shared, and web will have additional SSE handling for live widgets, codemirror, and the pane management that we have in piclaw for editors, etc.
```

## Neo @ 2026-04-22T12:08:42.713Z

```text
Got it: **shared core UI interfaces** for web and TUI, with **web-specific extras** for SSE widgets, CodeMirror, and pane management.

Next question:

**What should slash commands cover in v1?**  
For example:
- session management
- model/provider switching
- tool/script invocation
- search
- scheduling
- maintenance/admin
- all of the above
```

## Rui Carmo @ 2026-04-22T12:08:50.552Z

```text
All the above
```

## Neo @ 2026-04-22T12:08:55.046Z

```text
Got it: **slash commands should cover all of the above**.

Next question:

**Do you want gi v1 to support editing/opening files in the web UI like Piclaw does, or is chat + panes enough initially?**
```

## Rui Carmo @ 2026-04-22T12:09:24.254Z

```text
I need the workspace browser and the e editor panes, and they should be openable by the agent
```

## Neo @ 2026-04-22T12:09:28.188Z

```text
Got it: **workspace browser + editor panes are required in v1**, and the **agent must be able to open them**.

Next question:

**What editing capabilities must those panes support in v1?**  
For example:
- view only
- plain text editing
- syntax highlighting
- diff view
- multiple tabs
- split panes
- agent cursor/focus control
- save/revert
- all of the above
```

## Rui Carmo @ 2026-04-22T12:09:47.997Z

```text
All the above, starting with the piclaw feature set
```

## Neo @ 2026-04-22T12:09:52.116Z

```text
Got it: **start from Piclaw’s editor/browser feature set**, including all of the above.

Next question:

**What should gi’s auth/login model be in v1?**  
For example:
- local-only, no login
- local user identity with web session auth
- passwordless local auth
- multi-user login later only
```

## Rui Carmo @ 2026-04-22T12:10:44.215Z

```text
Same as piclaw: no multi user, but with the option for Totp and webauthn
```

## Neo @ 2026-04-22T12:10:49.236Z

```text
Got it: **same as Piclaw** for auth in v1:
- single-user
- optional **TOTP**
- optional **WebAuthn**

Next question:

**What should gi’s startup/deployment model be in v1?**  
For example:
- local binary + workspace
- long-running daemon/service
- auto-start web server
- terminal-first command that can also host the web UI
- something else
```

## Rui Carmo @ 2026-04-22T12:12:04.696Z

```text
Single binary maintained by supervisor/systemd in web mode, cli and tui share database but runas separate processes. I do not expect a lot of contention that SQLite WAL can’t cope with
```

## Neo @ 2026-04-22T12:12:08.854Z

```text
Got it: **single binary**, with:
- **web mode** run under `supervisor`/`systemd`
- **CLI** and **TUI** as separate processes
- all sharing the same **SQLite WAL** database

Next question:

**What should gi’s test strategy be in v1?**  
For example:
- unit tests
- fake-server provider tests
- DB integration tests
- golden tests for UI/message rendering
- end-to-end CLI/TUI/web smoke tests
- all of the above
```

## Rui Carmo @ 2026-04-22T12:12:28.518Z

```text
All of the above plus vetting and fuzzing
```

## Neo @ 2026-04-22T12:12:31.862Z

```text
Got it: **all of the above**, plus **vetting** and **fuzzing**.

Next question:

**What is the MVP boundary for gi v1?**  
What must exist before you’d consider it usable enough to start migrating from Piclaw?
```

## Rui Carmo @ 2026-04-22T12:12:43.006Z

```text
what we described so far
```

## Neo @ 2026-04-22T12:12:49.192Z

```text
Got it: **the MVP is exactly the surface we’ve described so far**—not a reduced subset.

Next question:

**What should gi be called internally and externally?**  
Do you want:
- product name: **gi**
- module/package/binary names aligned with that
- or a different internal naming scheme?
```

## Rui Carmo @ 2026-04-22T12:12:59.765Z

```text
Gi
```

## Neo @ 2026-04-22T12:13:03.979Z

```text
Got it: **Gi**.

Next question:

**What should the binary/module naming be in v1?**  
For example:
- binary: `gi`
- Go module: `github.com/rcarmo/gi`
- internal package prefix: `gi/...`

Or do you want something else?
```

## Rui Carmo @ 2026-04-22T12:13:35.997Z

```text
That’s good. rcarmo/gi already exists
```

## Neo @ 2026-04-22T12:13:43.310Z

```text
Got it:
- binary: `gi`
- module: `github.com/rcarmo/gi`
- package prefix: `gi/...`

Next question:

**Do you want Gi v1 to keep Pi/Piclaw’s kanban/work item system, or leave project/task tracking out entirely?**
```

## Rui Carmo @ 2026-04-22T12:14:57.678Z

```text
It should be left out and decided upon as a skill. The kanban YAML format piclaw edits should be enough for other uses, but we will use GitHub and the gh tool for development
```

## Neo @ 2026-04-22T12:15:03.920Z

```text
Got it: **no built-in kanban/task system**. Treat it as a **skill**, and keep the existing **kanban YAML format** available if needed. For development, use **GitHub + `gh`**.

Next question:

**Do you want Gi v1 to support Git/GitHub operations as built-in tools, or should that also live in scripts/skills initially?**
```

## Rui Carmo @ 2026-04-22T12:15:29.835Z

```text
no built in tools for that. Gh and skills
```

## Neo @ 2026-04-22T12:15:34.224Z

```text
Got it: **no built-in Git/GitHub tooling** — use **`gh` + skills**.

Next question:

**Do you want Gi v1 to include built-in web search/browser automation, or should that also be deferred to scripts/skills?**
```

## Rui Carmo @ 2026-04-22T12:16:17.803Z

```text
Scripts/skills, but we will have some sort of socket/http tool for scripts to use once we define that
```

## Neo @ 2026-04-22T12:16:22.229Z

```text
Got it: **web search/browser automation via scripts/skills**, with some future **socket/HTTP capability for scripts** once that bridge is defined.

Next question:

**What is the first real workflow Gi must handle end-to-end better than Piclaw?**
```

## Rui Carmo @ 2026-04-22T12:16:32.615Z

```text
Developing itself
```

## Neo @ 2026-04-22T12:16:39.042Z

```text
Got it: **Gi’s first proving workflow is self-hosted development of Gi itself**.

Next question:

**What is the single most important thing Gi must do better than Piclaw when developing itself?**
```

## Rui Carmo @ 2026-04-22T12:17:21.076Z

```text
Reliability and ease of maintenance. Top spec dependencies besides go-tui and go-ai
```

## Neo @ 2026-04-22T12:17:25.497Z

```text
Got it: **reliability**, **ease of maintenance**, and **minimal top-level dependencies** beyond `go-tui` and `go-ai`.

Next question:

**What parts of Piclaw are the biggest maintenance burden that Gi should avoid or redesign first?**
```

## Rui Carmo @ 2026-04-22T12:18:03.997Z

```text
Turn management and failure retries. Piclaw is very unreliable in session handling in general
```

## Neo @ 2026-04-22T12:18:11.587Z

```text
Got it: **turn management, failure retries, and session handling are the first things Gi should redesign**.

Next question:

**What is the worst session-handling failure mode you want Gi to eliminate first?**
```

## Rui Carmo @ 2026-04-22T12:19:19.666Z

```text
Context compaction mid turn and the associated retries, which should be automatic
```

## Neo @ 2026-04-22T12:19:26.667Z

```text
Got it: **automatic handling of mid-turn context compaction and its retries** is the first failure mode to eliminate.

Next question:

**When mid-turn compaction happens, should Gi hide that from the user unless it fails, or always show a visible status/progress event for it?**
```

## Rui Carmo @ 2026-04-22T12:19:39.353Z

```text
Always show a UI hint.
```

## Neo @ 2026-04-22T12:19:42.831Z

```text
Got it: **always show a UI hint when mid-turn compaction happens**.

Next question:

**Should Gi treat compaction as part of the same turn, or should it create an explicit internal sub-turn/recovery step with its own logged events?**
```

## Rui Carmo @ 2026-04-22T12:20:04.453Z

```text
Part of the same turn
```

## Neo @ 2026-04-22T12:20:08.275Z

```text
Got it: **compaction stays inside the same turn**, but with visible logged/progress events.

Next question:

**What should a turn look like internally in v1?**  
For example:
- one request → many recoverable internal phases
- explicit state machine
- append-only event log with resumable checkpoints
- something else
```

## Rui Carmo @ 2026-04-22T12:20:21.932Z

```text
Simplest possible thing
```

## Neo @ 2026-04-22T12:20:30.054Z

```text
Understood: **simplest possible turn model**.

Next question:

**If you had to choose, should that simplest turn model be based on an explicit state machine, or just an append-only event log with a small amount of turn state?**
```

## Rui Carmo @ 2026-04-22T12:20:45.910Z

```text
Event log.
```

## Neo @ 2026-04-22T12:20:50.057Z

```text
Got it: **append-only event log with minimal turn state**.

Next question:

**What should be the minimum checkpoint/resume boundary inside a turn?**  
For example:
- after every provider chunk
- after every tool call
- after every persisted event batch
- only between major phases
```

## Rui Carmo @ 2026-04-22T12:21:12.061Z

```text
Tool call. Persistence is implied
```

## Neo @ 2026-04-22T12:21:17.989Z

```text
Got it: **tool-call boundaries are the minimum resume checkpoints**, with persistence implied.

Next question:

**Besides tool calls, should Gi also checkpoint at provider retry/compaction boundaries, or only log those as events within the same resumable segment?**
```

## Rui Carmo @ 2026-04-22T12:21:49.906Z

```text
Checkpoint, but not fail the turn
```

## Neo @ 2026-04-22T12:21:54.052Z

```text
Got it: **provider retry and compaction boundaries should also create checkpoints**, but **must not fail the turn by themselves**.

Next question:

**What should count as a completed turn in v1?**  
For example:
- provider finished successfully
- provider failed but partial output + recovery info were persisted
- tool failure was surfaced and control returned cleanly
- all of the above, if control is back with consistent state
```

## Rui Carmo @ 2026-04-22T12:23:31.910Z

```text
All of the above
```

## Neo @ 2026-04-22T12:23:36.418Z

```text
Got it: **a turn is complete whenever control returns cleanly with consistent state**, whether it ended in success, partial success, or surfaced failure.

Next question:

**What should Gi show the user during a turn, in the minimum useful progress view?**  
For example:
- current phase
- retries
- compaction/recovery
- tool currently running
- streamed output
- checkpoint/resume status
- all of the above or a smaller subset
```

## Rui Carmo @ 2026-04-22T12:24:10.878Z

```text
All the above using piclaw conventions. Most can translate to TUI
```

## Neo @ 2026-04-22T12:24:14.674Z

```text
Got it: **full turn progress visibility using Piclaw conventions**, with most of it mapped into TUI too.

Next question:

**What should happen if Gi itself crashes during a turn?**  
For example:
- automatically resume the turn from the last checkpoint
- mark it recoverable and ask the user
- preserve partial output and return control
- all of the above via policy
```

## Rui Carmo @ 2026-04-22T12:24:33.281Z

```text
Preserve and return
```

## Neo @ 2026-04-22T12:24:37.031Z

```text
Got it: **on Gi crash during a turn, preserve state and return control**.

Next question:

**Should Gi attempt automatic turn recovery after a crash on next startup, or only show the preserved failed/recoverable state and let the user decide?**
```

## Rui Carmo @ 2026-04-22T12:25:05.512Z

```text
No auto recovery upon crash. Auto recovery if restarted cleanly
```

## Neo @ 2026-04-22T12:25:11.273Z

```text
Got it: **no auto-recovery after a crash**, but **auto-recovery is allowed after a clean restart**.

Next question:

**What should count as a clean restart in v1?**  
For example:
- intentional `/exit` or shutdown command
- service restart after upgrade
- normal process termination with state flushed
- all of the above
```

## Rui Carmo @ 2026-04-22T12:25:22.056Z

```text
All the above
```

## Neo @ 2026-04-22T12:25:26.028Z

```text
Got it: **all of the above count as a clean restart**.

Next question:

**What should Gi do on startup if it finds unfinished turns from a clean restart?**  
For example:
- resume them automatically in place
- mark them as resuming and replay progress in the UI
- only resume the active session
- resume none until the user opens the session
```

## Rui Carmo @ 2026-04-22T12:25:39.969Z

```text
Resume them automatically
```

## Neo @ 2026-04-22T12:25:45.371Z

```text
Got it: **unfinished turns from a clean restart should resume automatically**.

Next question:

**When an unfinished turn resumes automatically, should Gi replay the prior progress/status events in the UI, or only show new events from the resume point onward?**
```

## Rui Carmo @ 2026-04-22T12:26:08.921Z

```text
We should see the last progress status
```

## Neo @ 2026-04-22T12:26:12.707Z

```text
Got it: on auto-resume, **show the last progress status**, then continue from there.

Next question:

**Should Gi allow more than one turn to run at the same time across different sessions in v1, or should turn execution be globally serialized?**
```

## Rui Carmo @ 2026-04-22T12:26:28.069Z

```text
More than one
```

## Neo @ 2026-04-22T12:26:31.789Z

```text
Got it: **multiple turns can run concurrently across sessions**.

Next question:

**Should Gi allow more than one turn at the same time _within the same session_, or serialize turns per session?**
```

## Rui Carmo @ 2026-04-22T12:26:38.974Z

```text
Serialize
```

## Neo @ 2026-04-22T12:26:42.847Z

```text
Got it: **turns run concurrently across sessions, but are serialized within each session**.

Next question:

**If the user submits a new prompt to a session while a turn is still running, what should Gi do?**  
For example:
- queue it
- reject it with status
- offer interrupt/cancel
- fork automatically
- something else
```

## Rui Carmo @ 2026-04-22T12:27:08.469Z

```text
Queue, using piclaw’s queue/reorder/steer approach
```

## Neo @ 2026-04-22T12:27:12.740Z

```text
Got it: **queue new prompts within a session**, using **Piclaw’s queue/reorder/steer model**.

Next question:

**Should queued prompts be editable/reorderable/cancellable while waiting in v1, or just reorderable like Piclaw?**
```

## Rui Carmo @ 2026-04-22T12:27:37.181Z

```text
Reorderable/cancellable
```

## Neo @ 2026-04-22T12:27:47.649Z

```text
Got it: **queued prompts should be reorderable and cancellable** in v1.

Next question:

**Should queued prompts also be editable before they run, or should editing require cancel-and-resubmit?**
```

## Rui Carmo @ 2026-04-22T12:27:58.184Z

```text
Cancel/resubmit
```

## Neo @ 2026-04-22T12:28:02.397Z

```text
Got it: **queued prompts are reorderable/cancellable, but editing is cancel-and-resubmit**.

Next question:

**Should Gi support draft prompts per session in v1, shared across web and TUI?**
```

## Rui Carmo @ 2026-04-22T12:28:31.896Z

```text
No. Local state only, history navigation with cursor keys per session
```

## Neo @ 2026-04-22T12:28:35.480Z

```text
Got it: **drafts stay local to each UI**, with **per-session input history navigation** via cursor keys.

Next question:

**Should slash-command history and prompt history be shared within a session, or kept separate?**
```

## Rui Carmo @ 2026-04-22T12:28:47.646Z

```text
Separate between sessions
```

## Neo @ 2026-04-22T12:28:52.839Z

```text
Got it: **history should be session-local**, not shared across sessions.

Next question:

**Should command history and prompt history be separate within a session, or merged into one history stream?**
```

## Rui Carmo @ 2026-04-22T12:29:32.125Z

```text
Separate, but tools can search the entire history for all sessions (like messages in piclaw and the scripting bridge)
```

## Neo @ 2026-04-22T12:29:36.362Z

```text
Got it: **prompt history and slash-command history are separate within each session**, while tools/scripts can still search the full cross-session history.

Next question:

**Should Gi have a built-in notion of intents in v1, like Piclaw’s compose-box intent/status model, or should that remain just UI metadata over normal prompts?**
```

## Rui Carmo @ 2026-04-22T12:30:14.617Z

```text
Exactly like piclaw. Maybe some minor unification for intents and some in flight notifications like compaction
```

## Neo @ 2026-04-22T12:30:22.456Z

```text
Got it: **keep Piclaw’s intent model essentially as-is**, with maybe some light unification around intents and in-flight notifications like compaction.

Next question:

**What intents must exist in v1?**  
For example:
- normal prompt
- steer
- queue
- reorder
- interrupt/cancel
- maintenance/admin
- something else
```

## Rui Carmo @ 2026-04-22T12:30:35.314Z

```text
Those are good
```

## Neo @ 2026-04-22T12:30:40.079Z

```text
Got it: v1 intents include:
- **normal prompt**
- **steer**
- **queue**
- **reorder**
- **interrupt/cancel**
- **maintenance/admin**

Next question:

**Should interrupts/cancels try to stop the currently running provider/tool immediately, or just mark the turn as cancelled and let the runtime unwind safely?**
```

## Rui Carmo @ 2026-04-22T12:30:49.083Z

```text
Try immediately
```

## Neo @ 2026-04-22T12:30:58.005Z

```text
Got it: **interrupt/cancel should try to stop the active provider/tool immediately**.

Next question:

**If immediate cancellation fails, should Gi keep trying in the background, or should it mark the turn as cancelling and return control while cleanup continues?**
```

## Rui Carmo @ 2026-04-22T12:31:08.482Z

```text
Cancelling
```

## Neo @ 2026-04-22T12:31:16.854Z

```text
Got it: if immediate stop fails, Gi should mark the turn as **cancelling** and return control while cleanup continues.

Next question:

**What should the user be allowed to do in a session while a previous turn is still in `cancelling` state?**
```

## Rui Carmo @ 2026-04-22T12:31:49.544Z

```text
Queue more prompts and change settings, issue more commands - like in piclaw
```

## Neo @ 2026-04-22T12:31:53.792Z

```text
Got it: while a turn is **cancelling**, the user can still:
- queue prompts
- change settings
- issue commands

Next question:

**Should a `cancelling` turn still block execution of queued prompts in that session, or can the next queued prompt start before cancellation cleanup fully finishes?**
```

## Rui Carmo @ 2026-04-22T12:32:37.220Z

```text
Blocks. But I expect cancellation to be definitive and effective, killing subprecesses and preserving outputs
```

## Neo @ 2026-04-22T12:32:43.429Z

```text
Got it: **`cancelling` still blocks the session queue**, and cancellation must be **definitive/effective**—kill subprocesses and preserve outputs.

Next question:

**What should Gi do if a subprocess/tool ignores cancellation and refuses to die?**
```

## Rui Carmo @ 2026-04-22T12:32:59.223Z

```text
That should never happen
```

## Neo @ 2026-04-22T12:33:05.234Z

```text
Got it: **design assumption for v1 is that cancellation is enforceable**—subprocesses/tools must die when cancelled.

Next question:

**What should Gi do with partial tool/provider output when a turn is cancelled?**  
For example:
- always preserve and show it
- preserve but collapse it
- discard tool logs but keep model output
- configurable per tool/provider
```

## Rui Carmo @ 2026-04-22T12:33:34.151Z

```text
Preserve and show with a clip annotation oof provenance
```

## Neo @ 2026-04-22T12:33:40.992Z

```text
Got it: **preserve and show partial output**, with a clear **clipped/cancelled provenance annotation**.

Next question:

**What should Gi use as the canonical artifact storage layout in v1?**  
For example:
- Piclaw-compatible workspace paths
- DB-managed blobs plus exported files
- filesystem-first with DB metadata
- something else
```

## Rui Carmo @ 2026-04-22T12:35:52.251Z

```text
Workspace paths and db managed blobs for images and generated outputs the user explicitly asks for, attached by the internal tool. Referenced files should be rendered in the web as clickable pills to open the editor/workspace, and there will be fts metadata for some items like piclaw. When you write the implementation checklist that should prompt you to investigate further
```

## Neo @ 2026-04-22T12:36:00.085Z

```text
Got it:

- **filesystem/workspace paths** as the main artifact layout
- **DB-managed blobs** for images and explicit generated outputs attached by the internal tool
- referenced files rendered in web UI as **clickable pills** opening editor/workspace
- some artifact metadata indexed in **FTS**, Piclaw-style
- implementation checklist should explicitly flag this area for deeper investigation

Next question:

**What should Gi’s internal “messages” capability cover in v1?**  
For example:
- list/search messages
- grep/extract patterns
- fetch context windows
- post agent messages
- delete/prune low-value messages
- attach/export/read attachments
- diff timeline changes
- all of the above
```

## Rui Carmo @ 2026-04-22T12:36:19.184Z

```text
All of the above
```

## Neo @ 2026-04-22T12:36:26.389Z

```text
Got it: **all of the above** for `messages` in v1.

Next question:

**Should `messages` be a built-in internal tool only, or also exposed as a CLI/TUI/web user-facing capability directly?**
```

## Rui Carmo @ 2026-04-22T12:36:35.090Z

```text
both
```

## Neo @ 2026-04-22T12:36:44.070Z

```text
Got it: **`messages` should be both an internal tool and a direct user-facing capability**.

Next question:

**What should the `exit` capability do in v1?**  
For example:
- end the current turn only
- stop the current UI client
- request clean shutdown/restart of the Gi web service
- all of the above, depending on context
```

## Rui Carmo @ 2026-04-22T12:37:12.166Z

```text
Stop and exit the current process cleanly
```

## Neo @ 2026-04-22T12:37:16.008Z

```text
Got it: **`exit` should stop and exit the current process cleanly**.

Next question:

**What should the `schedule` capability be allowed to run in v1?**  
For example:
- prompts into a session
- shell commands
- Joker scripts
- slash commands
- all of the above
```

## Rui Carmo @ 2026-04-22T12:38:18.524Z

```text
All of the above but slash commands. Tied to an explicit session
```

## Neo @ 2026-04-22T12:38:22.567Z

```text
Got it: `schedule` can run:
- prompts into a session
- shell commands
- Joker scripts

And:
- **not slash commands**
- always **tied to an explicit session**

Next question:

**Should scheduled jobs run under the target session’s identity/context, or as a separate scheduler identity that posts back into the session?**
```

## Rui Carmo @ 2026-04-22T12:38:45.365Z

```text
Separate temporary session that can post back
```

## Neo @ 2026-04-22T12:38:51.450Z

```text
Got it: **scheduled jobs run in a separate temporary session** that can **post back into the target session**.

Next question:

**What should happen if a scheduled job fails repeatedly in v1?**  
For example:
- retry with backoff
- disable the schedule after N failures
- keep running forever but post failures
- make it policy/config driven
```

## Rui Carmo @ 2026-04-22T12:39:06.225Z

```text
1 and 2
```

## Neo @ 2026-04-22T12:39:13.213Z

```text
Got it: **retry with backoff**, then **disable after N failures**.

Next question:

**Should schedules be visible and manageable from all three surfaces in v1 — web, TUI, and CLI?**
```

## Rui Carmo @ 2026-04-22T12:39:27.369Z

```text
Yes, via commands
```

## Neo @ 2026-04-22T12:39:31.463Z

```text
Got it: **schedules should be visible/manageable from web, TUI, and CLI via commands**.

Next question:

**What should the `shell` tool guarantee in v1?**  
For example:
- streaming stdout/stderr
- detached/background process support
- full process-group cancellation
- cwd/env control
- timeout support
- exit code capture
- all of the above
```

## Rui Carmo @ 2026-04-22T12:39:45.500Z

```text
All the above
```

## Neo @ 2026-04-22T12:39:49.790Z

```text
Got it: **all of the above** for `shell`.

Next question:

**What should the `read`/`write`/`edit` file tools guarantee in v1?**  
For example:
- exact targeted edits
- safe overwrite semantics
- directory auto-create
- binary/image support for `read`
- truncation/offset support for large files
- clear diff/error reporting
- all of the above
```

## Rui Carmo @ 2026-04-22T12:40:00.371Z

```text
All the above
```

## Neo @ 2026-04-22T12:40:04.192Z

```text
Got it: **all of the above** for `read` / `write` / `edit`.

Next question:

**Should `edit` in v1 be strictly exact-match replacement like Piclaw’s current tool, or do you want fallback modes too?**  
For example:
- exact-match only
- exact-match plus patch/diff mode
- exact-match plus model-assisted retry mode
```

## Rui Carmo @ 2026-04-22T12:40:15.215Z

```text
All of the above
```

## Neo @ 2026-04-22T12:40:22.531Z

```text
Got it: **`edit` should support exact-match replacement, patch/diff mode, and model-assisted retry mode**.

Next question:

**What should determine which edit mode Gi uses first in v1?**  
For example:
- always exact-match first, then fallback
- tool metadata/policy
- model choice
- user/session setting
- all of the above
```

## Rui Carmo @ 2026-04-22T12:40:32.251Z

```text
Model choice
```

## Neo @ 2026-04-22T12:40:36.998Z

```text
Got it: **edit strategy should be chosen based on the model**.

Next question:

**What should the default model-based edit policy be in v1?**  
For example:
- strong models can use patch/model-assisted fallback
- weaker models are restricted to exact-match
- configurable per model in settings
```

## Rui Carmo @ 2026-04-22T12:40:57.828Z

```text
Patch
```

## Neo @ 2026-04-22T12:41:04.138Z

```text
Got it: default edit policy should be **patch-oriented**.

Next question:

**Should Gi still try exact-match edits first when a patch-capable model is in use, or go straight to patch mode by default?**
```

## Rui Carmo @ 2026-04-22T12:41:13.465Z

```text
The model decides
```

## Neo @ 2026-04-22T12:41:21.890Z

```text
Got it: **the model decides the edit mode**.

Next question:

**What metadata should a model carry in v1 besides Pi-compatible provider/model settings?**  
For example:
- edit strategy hints
- retry policy hints
- compaction suitability
- streaming/tool-calling support
- cost/speed class
- context window
- something else
```

## Rui Carmo @ 2026-04-22T12:41:39.975Z

```text
What pi has
```

## Neo @ 2026-04-22T12:41:45.720Z

```text
Got it: **model metadata should match what Pi already has**.

Next question:

**Should Gi v1 support per-session model switching, or should model choice be fixed when the session is created?**
```

## Rui Carmo @ 2026-04-22T12:44:28.406Z

```text
Per session
```

## Neo @ 2026-04-22T12:44:41.343Z

```text
Got it: **model choice is switchable per session**.

Next question:

**Should a model switch take effect immediately for the next queued turn only, or also try to affect the currently running turn when possible?**
```

## Rui Carmo @ 2026-04-22T12:45:01.382Z

```text
Next turn
```

## Neo @ 2026-04-22T12:45:05.359Z

```text
Got it: **model switches apply on the next turn**, not the current one.

Next question:

**Should sessions also support per-session provider settings overrides in v1, or only model selection?**
```

## Rui Carmo @ 2026-04-22T12:45:39.117Z

```text
Only model selection some/commands for some parameters
```

## Neo @ 2026-04-22T12:45:43.839Z

```text
Got it: **per-session model selection**, plus **some parameter changes via commands**, but not arbitrary per-session provider config overrides.

Next question:

**Which per-session parameters must be command-adjustable in v1?**  
For example:
- thinking level
- temperature
- max output tokens
- compaction policy
- retry policy
- streaming on/off
- verbosity
```

## Rui Carmo @ 2026-04-22T12:46:29.517Z

```text
Thinking level and verbose
```

## Neo @ 2026-04-22T12:46:34.458Z

```text
Got it: **per-session command-adjustable parameters are:**
- **thinking level**
- **verbose**

Next question:

**Should `verbose` affect only UI/log detail, or also influence the model prompt/agent behavior?**
```

## Rui Carmo @ 2026-04-22T12:46:52.832Z

```text
Only prompt agent behavior
```

## Neo @ 2026-04-22T12:47:04.797Z

```text
Got it: **`verbose` should affect prompt/agent behavior only**, not just UI logging.

Next question:

**Should “thinking level” map exactly to Pi’s current thinking model/semantics, or can Gi define a simpler compatible version in v1?**
```

## Rui Carmo @ 2026-04-22T12:47:21.781Z

```text
It has to map exactly to what model supports
```

## Neo @ 2026-04-22T12:47:26.115Z

```text
Got it: **thinking level must map exactly to what the selected model supports**.

Next question:

**What should happen if the user asks for a thinking level the current model does not support?**  
For example:
- reject with a clear message
- clamp to the nearest supported level
- switch model automatically
- keep requested value pending until model changes
```

## Rui Carmo @ 2026-04-22T12:47:47.882Z

```text
Reject
```

## Neo @ 2026-04-22T12:47:54.789Z

```text
Got it: **reject unsupported thinking levels with a clear message**.

Next question:

**What should the default session model be in v1?**  
For example:
- inherit Pi/Piclaw default selection logic
- a fixed configured default
- workspace-configured default with per-session override
```

## Rui Carmo @ 2026-04-22T12:48:12.449Z

```text
Inherit
```

## Neo @ 2026-04-22T12:48:17.107Z

```text
Got it: **inherit Pi/Piclaw default model selection logic**.

Next question:

**Should Gi v1 preserve Pi/Piclaw’s prompt template system too, or start with a smaller built-in system and only keep settings compatibility where necessary?**
```

## Rui Carmo @ 2026-04-22T12:48:52.738Z

```text
Preserve
```

## Neo @ 2026-04-22T12:48:56.435Z

```text
Got it: **preserve Pi/Piclaw’s prompt template system**.

Next question:

**Should Gi v1 preserve Pi/Piclaw’s hook system too, or simplify hooks initially?**
```

## Rui Carmo @ 2026-04-22T12:49:18.349Z

```text
What do you mean?
```

## Neo @ 2026-04-22T12:49:22.304Z

```text
I mean the runtime hooks around a turn and provider/tool execution.

For example, Piclaw/go-ai style hooks can include things like:
- before/after provider request
- payload transformation
- response inspection
- context compaction hooks
- tool lifecycle hooks
- session/turn lifecycle hooks

So the question is:

**Should Gi preserve that full hook surface, or start with a smaller hook system in v1?**
```

## Rui Carmo @ 2026-04-22T12:49:44.814Z

```text
Smart smaller, but add those to the roadmap
```

## Neo @ 2026-04-22T12:49:50.998Z

```text
Got it: **smaller, smarter hook surface in v1**, with the broader hook system on the roadmap.

Next question:

**Which hook points are absolutely required in v1?**  
For example:
- before provider call
- after provider response
- context compaction
- tool start/end
- turn start/end
- schedule/task lifecycle
```

## Rui Carmo @ 2026-04-22T12:50:08.094Z

```text
All the above plus error handling
```

## Neo @ 2026-04-22T12:50:15.253Z

```text
Got it: **v1 needs hooks for:**
- before provider call
- after provider response
- context compaction
- tool start/end
- turn start/end
- schedule/task lifecycle
- error handling

Next question:

**Should v1 hooks be implemented as Joker scripts, Go interfaces, or both?**
```

## Rui Carmo @ 2026-04-22T12:50:43.465Z

```text
Both plus roadmap to JavaScript with quickjs
```

## Neo @ 2026-04-22T12:50:50.089Z

```text
Got it: **v1 hooks should support both Go and Joker**, with **QuickJS on the roadmap**.

Next question:

**Which one should be the primary extension surface in v1: Joker hooks first, or Go hooks first?**
```

## Rui Carmo @ 2026-04-22T12:51:04.156Z

```text
Joker
```

## Neo @ 2026-04-22T12:51:10.903Z

```text
Got it: **Joker is the primary hook/extension surface in v1**, with Go hooks also supported.

Next question:

**How should Joker hooks be loaded in v1?**  
For example:
- from workspace config
- from skills
- from packaged installs
- all of the above
```

## Rui Carmo @ 2026-04-22T12:51:52.252Z

```text
All the above plus database. One of the goals is that you should be able to take all your environment with you by just copying the databaseand the binary.
```

## Neo @ 2026-04-22T12:51:57.062Z

```text
Got it: **Joker hooks can come from workspace config, skills, packaged installs, and the database**, and a major goal is **portable state via just the binary + database**.

Next question:

**What should still live on disk outside the database in v1?**  
For example:
- workspace files only
- settings files
- skills/scripts
- cached assets
- generated artifacts
- logs
- or should most of that also be mirrored/importable into the DB
```

## Rui Carmo @ 2026-04-22T12:52:23.820Z

```text
Let’s break that down one by one
```

## Neo @ 2026-04-22T12:52:28.586Z

```text
Good.

First one:

**Should settings/config live only on disk in v1, or also be importable/mirrored into the database?**
```

## Rui Carmo @ 2026-04-22T12:52:39.249Z

```text
Mirror
```

## Neo @ 2026-04-22T12:52:45.774Z

```text
Got it: **settings/config should be mirrored into the database**.

Next question:

**Should skills live only on disk in v1, or also be importable/mirrored into the database?**
```

## Rui Carmo @ 2026-04-22T12:53:10.031Z

```text
Mirror and the database should be query before the file system
```

## Neo @ 2026-04-22T12:53:15.288Z

```text
Got it: **skills should be mirrored into the database**, and **the database should be queried before the filesystem**.

Next question:

**Should scripts also be mirrored into the database, or stay filesystem-first in v1?**
```

## Rui Carmo @ 2026-04-22T12:53:34.084Z

```text
Also
```

## Neo @ 2026-04-22T12:53:43.446Z

```text
Got it: **scripts should also be mirrored into the database**.

Next question:

**Should prompt templates also be mirrored into the database, or remain settings-file only in v1?**
```

## Rui Carmo @ 2026-04-22T12:54:20.809Z

```text
Also. There should be an asset stable with all of that kind of thing embedded and full text search
```

## Neo @ 2026-04-22T12:54:25.715Z

```text
Got it: **prompt templates should also be mirrored into the database**, and there should be an **assets table** for this kind of embedded content with **full-text search**.

Next question:

**Should generated artifacts/files also be mirrored into the database, or only selected ones like images and explicit user-requested outputs?**
```

## Rui Carmo @ 2026-04-22T12:54:35.437Z

```text
Only selected ones
```

## Neo @ 2026-04-22T12:54:44.525Z

```text
Got it: **only selected generated artifacts should be mirrored into the database**.

Next question:

**Should logs/observability data live only in SQLite in v1, or also be written to filesystem log files?**
```

## Rui Carmo @ 2026-04-22T12:54:53.574Z

```text
Only counters
```

## Neo @ 2026-04-22T12:55:00.036Z

```text
Got it: **don’t treat filesystem logs as a first-class output**.

To clarify this one precisely:

**Do you want full observability events stored only in SQLite, with only lightweight counters/metrics elsewhere, and no normal log files?**
```

## Rui Carmo @ 2026-04-22T12:55:38.253Z

```text
I want counters inside the database and events, but not all of the data that is associated with log files.
```

## Neo @ 2026-04-22T12:55:45.326Z

```text
Got it: **store counters and structured events in SQLite**, but **not full traditional logfile payloads**.

Next question:

**Should cached assets/dependencies in v1 live only on disk, or also be mirrored into the database?**
```

## Rui Carmo @ 2026-04-22T12:56:12.733Z

```text
What cash assets dependencies
```

## Neo @ 2026-04-22T12:56:44.474Z

```text
I mean things like:
- vendored JS/CSS for the web UI
- downloaded skill/package contents
- model/provider metadata caches
- thumbnails/previews
- temporary bundled assets

So the question is:

**For those kinds of cached/generated support assets, should Gi keep them only on disk in v1, or also mirror some of them into the database?**
```

## Rui Carmo @ 2026-04-22T12:58:01.635Z

```text
Web should be packed into the binary, skills and packages should be imported, cashes, live elsewhere, thumbnails are dynamically generated and kept in a separate table for cleaning up. We will need a good image processing library.
```

## Neo @ 2026-04-22T12:58:08.357Z

```text
Got it:

- **web assets packed into the binary**
- **skills/packages imported into the DB**
- **caches live elsewhere**
- **thumbnails generated dynamically**
- **thumbnails stored in a separate cleanupable table**
- and we’ll need a **good image processing library**

Next question:

**Should keychain secrets themselves live in the SQLite database in v1, or should Gi keep Piclaw’s current split model if there is one?**
```

## Rui Carmo @ 2026-04-22T12:58:27.157Z

```text
They should live in the database encrypted in the same way
```

## Neo @ 2026-04-22T12:58:32.970Z

```text
Got it: **keychain secrets should live in the database**, encrypted the same way as in Piclaw.

Next question:

**Should the encryption key for those secrets come only from an env var/bootstrap secret, or do you also want interactive unlock methods in v1?**
```

## Rui Carmo @ 2026-04-22T12:59:27.100Z

```text
Both
```

## Neo @ 2026-04-22T12:59:33.833Z

```text
Got it: **both** for v1:
- env/bootstrap secret support
- interactive unlock support

Next question:

**When Gi starts and the keychain is locked, what should happen by default?**  
For example:
- web/TUI prompt for unlock
- CLI fails with a clear message unless explicitly interactive
- start in degraded mode until unlocked
- something else
```

## Rui Carmo @ 2026-04-22T13:00:12.570Z

```text
Degraded
```

## Neo @ 2026-04-22T13:00:20.108Z

```text
Got it: **start in degraded mode until unlocked**.

Next question:

**In degraded mode, what should still work in v1?**  
For example:
- browsing chats/history
- reading files
- local shell
- non-secret scripts
- search
- editing
- everything except secret-dependent operations
```

## Rui Carmo @ 2026-04-22T13:00:36.278Z

```text
Everything
```

## Neo @ 2026-04-22T13:00:40.241Z

```text
Got it: **everything should still work in degraded mode except operations that actually need secret material**.

Next question:

**What should happen when a turn or tool needs a locked secret in degraded mode?**  
For example:
- fail immediately with a clear unlock-needed error
- emit a UI prompt to unlock and leave the turn paused
- let scripts decide
- something else
```

## Rui Carmo @ 2026-04-22T13:00:51.830Z

```text
Admit the UI prompt
```

## Neo @ 2026-04-22T13:01:01.474Z

```text
Got it: **emit a UI prompt to unlock**, and leave the operation waiting on that.

Next question:

**Should unlocking the keychain resume the blocked turn/tool automatically, or should it return control and ask the user to retry?**
```

## Rui Carmo @ 2026-04-22T13:01:19.118Z

```text
Resume
```

## Neo @ 2026-04-22T13:01:28.078Z

```text
Got it: **unlocking should automatically resume the blocked turn/tool**.

Next question:

**Should the unlocked keychain stay unlocked for the whole process lifetime in v1, or should it re-lock after inactivity / on demand?**
```

## Rui Carmo @ 2026-04-22T13:01:36.346Z

```text
Stay
```

## Neo @ 2026-04-22T13:01:45.153Z

```text
Got it: **once unlocked, it stays unlocked for the process lifetime**.

Next question:

**Should each process unlock independently in v1, or should unlocking in the web service also make secrets available to CLI/TUI processes through some shared mechanism?**
```

## Rui Carmo @ 2026-04-22T13:01:56.373Z

```text
Each process
```

## Neo @ 2026-04-22T13:02:02.105Z

```text
Got it: **each process unlocks independently**.

Next question:

**What should the first spec artifact be for Gi?**  
For example:
- a single `README/spec.md`
- a full `docs/` tree
- a kanban/workitems breakdown
- an architecture decision record set
- something else
```

## Rui Carmo @ 2026-04-22T13:03:09.635Z

```text
An architecture decision record set with full checklist for implementation. That should be in a docs tree. And I want you to build all of the work items that you can using the same checklist.
```

## Neo @ 2026-04-22T13:03:25.407Z

```text
Got it: **first spec artifact = a `docs/` tree with ADRs plus a full implementation checklist**, and that checklist should also drive the work items we can execute.

Next question:

**Do you want one top-level ADR for the whole architecture plus smaller ADRs per subsystem, or only small focused ADRs from the start?**
```

## Rui Carmo @ 2026-04-22T13:03:52.438Z

```text
Top + smaller. You can dump the entire contents of this chat there for reference
```

## Neo @ 2026-04-22T13:04:01.636Z

```text
Got it: **one top-level ADR plus smaller subsystem ADRs**, and we can include the full chat contents there as reference.

Next question:

**Do you want the implementation checklist organized by subsystem, by milestone/phase, or both?**
```

## Rui Carmo @ 2026-04-22T13:04:40.568Z

```text
Checklists by sub system and groups of them per phase
```

## Neo @ 2026-04-22T13:04:46.414Z

```text
Got it: **checklists by subsystem**, grouped into **phases**.

Next question:

**What should Phase 1 focus on first: turn engine, database/state model, or the web UI shell?**
```

## Rui Carmo @ 2026-04-22T13:05:07.272Z

```text
All the above with minimal functionality
```

## Neo @ 2026-04-22T13:05:21.392Z

```text
Got it: **Phase 1 should cover turn engine, database/state model, and web UI shell together, with minimal functionality**.

Next question:

**What is the absolute minimum “vertical slice” that proves Gi is alive in Phase 1?**  
For example:
- start web UI
- create/open a session
- send one prompt
- run one shell tool call
- stream progress
- persist messages/events in SQLite
- cleanly return control after the turn
```

## Rui Carmo @ 2026-04-22T13:05:33.405Z

```text
That works for me
```

## Neo @ 2026-04-22T13:05:42.637Z

```text
Got it.

Next question:

**What should be the next vertical slice after that?**  
For example:
- session resume/fork
- context compaction + retry
- keychain unlock/resume
- scheduling
- editor/browser panes
- skills/scripts
```

## Rui Carmo @ 2026-04-22T13:05:58.363Z

```text
Keychain and context and skills
```

## Neo @ 2026-04-22T13:06:04.488Z

```text
Got it: next slice is **keychain + context + skills**.

Next question:

**Of those three, which one must be implemented first: keychain, context compaction/recovery, or skills?**
```

## Rui Carmo @ 2026-04-22T13:06:33.415Z

```text
Context. Ask me only things that you cannot figure out from what we’ve discussed so far and the design of piclaw
```

## Neo @ 2026-04-22T13:06:45.808Z

```text
Understood. I’ll stop broad elicitation and only ask about gaps I can’t derive from this chat plus Piclaw’s design.

First unresolved item:

**Should I work in `/workspace/projects/gi` as a fresh local repo, or should I clone/connect it to the existing `github.com/rcarmo/gi` repository first?**
```

## Rui Carmo @ 2026-04-22T13:06:54.860Z

```text
Both
```

## Neo @ 2026-04-22T13:07:12.784Z

```text
I can proceed locally, but I can’t confirm/access `github.com/rcarmo/gi` from here right now.

**One necessary question:**  
**Is `rcarmo/gi` already an existing remote repo I should pull from, or should I initialize locally in `/workspace/projects/gi` and connect/push to that remote later?**
```

## Rui Carmo @ 2026-04-22T13:08:01.375Z

```text
Init and push. Dump this chat to a transcript before you forget, then start drafting the docs and project structure
```

## Neo @ 2026-04-22T13:09:32.988Z

```text
Now I’m writing the transcript, top-level ADR, subsystem ADRs, and phased implementation checklist from this chat.
```

## Rui Carmo @ 2026-04-22T13:11:13.633Z

```text
I want a full transcript, not a summary. You do have a messages tool
```

package turn

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	goai "github.com/rcarmo/go-ai"

	"github.com/rcarmo/gi/internal/codemode"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/prompt"
	"github.com/rcarmo/gi/internal/store"
)

// The system prompt is Pi's structured prompt (internal/prompt): named
// sections built per turn from the declared tools, the project instructions,
// skills, cwd and extension sections (mcp_servers). As in Pi, the sections
// when a session starts lead the conversation; a later change is appended
// once, as a system message updating the changed sections before the prompt
// whose turn introduced it, so earlier messages stay cached. gi records the
// initial sections and every update in the session state. Models without
// mid-conversation system messages get the updates folded into the leading
// prompt (pi-ai's collapseSystemMessages).

const promptSectionsStateKey = "system_prompt_sections"

// promptContribution is a tool's prompt snippet and guidelines (Pi's
// promptSnippet / promptGuidelines).
type promptContribution struct {
	snippet    string
	guidelines []string
}

// builtinPromptContributions: Pi's own for the tools gi shares with Pi,
// gi's wording for gi's tools.
var builtinPromptContributions = map[string]promptContribution{
	"read":  {"Read file contents", []string{"Use read to examine files instead of cat or sed."}},
	"write": {"Create or overwrite files", []string{"Use write only for new files or complete rewrites."}},
	"edit": {"Make precise file edits with exact text replacement, including multiple disjoint edits in one call", []string{
		"Use edit for precise changes (edits[].oldText must match exactly)",
		"When changing multiple separate locations in one file, use one edit call with multiple entries in edits[] instead of multiple edit calls",
		"Each edits[].oldText is matched against the original file, not after earlier edits are applied. Do not emit overlapping or nested edits. Merge nearby changes into one edit.",
		"Keep edits[].oldText as small as possible while still being unique in the file. Do not pad with large unchanged regions.",
	}},
	"shell":          {"Execute shell commands (ls, grep, find, tests, package tooling, etc.)", nil},
	"rtk":            {"Run noisy commands (git, search, listings, tests) with compact output", []string{"Use rtk for noisy command output and shell when the raw output is needed"}},
	"tools":          {"List, inspect and activate tools (query first, then fetch one tool's full schema)", []string{"Activate only the extra tools you need with tools, and reset them afterwards"}},
	"skills":         {"List discovered skills and read a skill's SKILL.md", nil},
	"messages":       {"Search chat messages or read them by row id", nil},
	"script":         {"Run Goja JavaScript or Joker scripts through gi's script bridge", nil},
	"compact":        {"Inspect compaction thresholds and whether the session should compact", nil},
	"peering":        {"Inspect tsnet/Tailscale peer-discovery status (disabled unless configured)", nil},
	"tool_search":    {"Search for tools that are not loaded yet and load the matches", nil},
	codemodeToolName: {codemode.Texts.PromptSnippet, codemode.Texts.PromptGuidelines},
}

// giPromptGuidelines are gi's own rules, after the tools' guidelines.
var giPromptGuidelines = []string{
	"Read relevant files before editing them",
	"Workspace paths go through gi's path resolver; vfs://namespace/path reaches gi's managed files",
	"Keep secrets, tokens and other auth material out of chat output",
	"Avoid destructive operations unless the user asks for them or they are clearly scoped and reversible",
}

const giDocsSection = `gi documentation (read only when the user asks about gi itself: its tools, scripting, hooks, MCP, codemode or TUI):
- Reference: vfs://reference/README.md (read-only; read it with the read tool)
- Tools: vfs://reference/tools/, MCP: vfs://reference/mcp.md, codemode: vfs://reference/codemode.md and vfs://reference/codemode-scripts.md
- gi reads Pi's settings, models and credentials (.gi first, then .pi)`

// promptSections builds this turn's sections for the declared tools.
func (e *Engine) promptSections(declared []goai.Tool) []prompt.Section {
	return e.promptSectionsWithConfig(declared, e.runtimeCfg)
}

func (e *Engine) promptSectionsWithConfig(declared []goai.Tool, cfg config.RuntimeConfig) []prompt.Section {
	assistant := strings.TrimSpace(cfg.AssistantName)
	if assistant == "" {
		assistant = "Gi"
	}
	user := strings.TrimSpace(cfg.UserName)
	if user == "" {
		user = "the user"
	}
	o := prompt.Options{
		CustomPrompt:     e.systemPrompt,
		Preamble:         fmt.Sprintf("You are %s, an expert coding assistant operating inside gi, a coding agent harness. You help %s by reading files, executing commands, editing code, and writing new files.", assistant, user),
		Docs:             giDocsSection,
		ToolSnippets:     map[string]string{},
		ToolGuidelines:   map[string][]string{},
		PromptGuidelines: giPromptGuidelines,
		Cwd:              cfg.WorkspaceRoot,
	}
	for _, t := range declared {
		o.SelectedTools = append(o.SelectedTools, t.Name)
		c := builtinPromptContributions[t.Name]
		if reg, ok := e.tools.GetRegistered(t.Name); ok && (reg.PromptSnippet != "" || len(reg.PromptGuidelines) > 0) {
			c = promptContribution{reg.PromptSnippet, reg.PromptGuidelines}
		}
		if c.snippet != "" {
			o.ToolSnippets[t.Name] = c.snippet
		}
		if len(c.guidelines) > 0 {
			o.ToolGuidelines[t.Name] = c.guidelines
		}
	}
	for _, f := range cfg.ContextFiles {
		o.ContextFiles = append(o.ContextFiles, prompt.ContextFile{Path: f.Path, Content: f.Content})
	}
	for _, s := range cfg.Discovery.Skills {
		if !s.DisableModelInvocation { // Pi: explicit invocation only
			o.Skills = append(o.Skills, prompt.Skill{Name: s.Name, Description: s.Description, Location: s.Path})
		}
	}
	if text := e.currentMCPSectionText(); text != "" {
		o.Extra = append(o.Extra, prompt.Section{Name: "mcp_servers", Text: text})
	}
	sections, err := prompt.BuildSections(o)
	if err != nil {
		log.Printf("system prompt: %v", err)
	}
	return sections
}

// currentMCPSectionText is the mcp_servers section body (without tags), or "".
func (e *Engine) currentMCPSectionText() string {
	section := e.mcpServersSection()
	if section == "" {
		return ""
	}
	return section[len("<mcp_servers>\n") : len(section)-len("\n</mcp_servers>")]
}

type sectionUpdate struct {
	Before string         `json:"before"` // message ID the update precedes
	Patch  []prompt.Patch `json:"patch"`
}

type sectionRecord struct {
	Initial []prompt.Section `json:"initial"`
	Updates []sectionUpdate  `json:"updates,omitempty"`
}

func (r sectionRecord) current() []prompt.Section {
	s := r.Initial
	for _, u := range r.Updates {
		s = prompt.Apply(s, u.Patch)
	}
	return s
}

// sectionInsert is an update positioned in a turn's messages.
type sectionInsert struct {
	index  int          // insert before convCtx.Messages[index]
	anchor goai.Message // message expected at index (detects mid-turn compaction)
	patch  []prompt.Patch
}

// sectionPlan is the prompt layout of a session's running turn.
type sectionPlan struct {
	leading []prompt.Section // the leading prompt (initial, plus compacted-away updates)
	latest  []prompt.Section // every update applied
	inserts []sectionInsert
}

// planPromptSections records a change of this turn's sections and places
// every recorded update: before its anchor message, or folded into the
// leading prompt when compaction removed the anchor. messageIDs are the
// snapshot messages, in convCtx order after offset.
func (e *Engine) planPromptSections(ctx context.Context, sessionID string, current []prompt.Section, messageIDs []string, offset int, messages []goai.Message, turnIDs ...string) *sectionPlan {
	sess, err := e.store.GetSession(ctx, sessionID)
	if err != nil {
		return &sectionPlan{leading: current, latest: current}
	}
	var rec sectionRecord
	recorded := false
	if raw, ok := sess.State[promptSectionsStateKey]; ok {
		if b, err := json.Marshal(raw); err == nil && json.Unmarshal(b, &rec) == nil && len(rec.Initial) > 0 {
			recorded = true
		}
	}
	anchor := ""
	for i := len(messageIDs) - 1; i >= 0; i-- {
		if messageIDs[i] != "" {
			anchor = messageIDs[i]
			break
		}
	}
	changed := false
	if !recorded {
		rec, changed = sectionRecord{Initial: current}, true
	} else if patch := prompt.Diff(rec.current(), current); patch != nil && anchor != "" {
		rec.Updates, changed = append(rec.Updates, sectionUpdate{Before: anchor, Patch: patch}), true
	}
	if changed {
		var err error
		if len(turnIDs) > 0 {
			err = e.store.RecordTurnPromptSections(ctx, sessionID, turnIDs[0], rec)
		} else {
			err = e.store.TouchSessionState(ctx, sessionID, map[string]any{promptSectionsStateKey: rec})
		}
		if err != nil {
			log.Printf("system prompt: record sections: %v", err)
		}
	}
	plan := &sectionPlan{leading: rec.Initial, latest: rec.current()}
	index := map[string]int{}
	for i, id := range messageIDs {
		index[id] = i + offset
	}
	for _, u := range rec.Updates {
		if i, ok := index[u.Before]; ok && i < len(messages) {
			plan.inserts = append(plan.inserts, sectionInsert{index: i, anchor: messages[i], patch: u.Patch})
			continue
		}
		// The anchor was compacted away (compaction rewrites the prefix
		// anyway): this and every earlier update lead.
		for _, ins := range plan.inserts {
			plan.leading = prompt.Apply(plan.leading, ins.patch)
		}
		plan.leading, plan.inserts = prompt.Apply(plan.leading, u.Patch), nil
	}
	return plan
}

// systemPrompt is the leading prompt text.
func (p *sectionPlan) systemPrompt() string { return prompt.Render(p.leading) }

// renderSectionUpdate is pi-ai's renderSystemMessageUpdate for a patch.
func renderSectionUpdate(patch []prompt.Patch) string {
	parts := make([]string, 0, len(patch))
	for _, p := range patch {
		if p.Text == nil {
			parts = append(parts, `Removed system prompt section "`+p.Name+`".`)
		} else {
			parts = append(parts, `Updated system prompt section "`+p.Name+`":`+"\n\n"+*p.Text)
		}
	}
	return strings.Join(parts, "\n\n")
}

// apply lays the plan out for a request: updates as system messages when
// the model takes them mid-conversation, otherwise folded into the leading
// prompt. A prompt replaced by a hook is used as is, without updates; after
// a mid-turn compaction (anchors moved) the latest sections lead.
func (p *sectionPlan) apply(systemPrompt string, messages []goai.Message, midConversation bool) (string, []goai.Message) {
	if p == nil || systemPrompt != p.systemPrompt() {
		return systemPrompt, messages
	}
	for _, ins := range p.inserts {
		if ins.index >= len(messages) || !reflect.DeepEqual(messages[ins.index], ins.anchor) {
			return prompt.Render(p.latest), messages
		}
	}
	if len(p.inserts) == 0 {
		return systemPrompt, messages
	}
	if !midConversation {
		return prompt.Render(p.latest), messages
	}
	out := make([]goai.Message, 0, len(messages)+len(p.inserts))
	next := 0
	for i, m := range messages {
		for next < len(p.inserts) && p.inserts[next].index == i {
			out = append(out, goai.Message{Role: goai.RoleSystem, Content: []goai.ContentBlock{{Type: "text", Text: renderSectionUpdate(p.inserts[next].patch)}}})
			next++
		}
		out = append(out, m)
	}
	return systemPrompt, out
}

// snapshotMessageIDs lists snapshot message IDs and the convCtx offset (1
// when a compaction summary precedes them).
func snapshotMessageIDs(snapshot store.ContextSnapshot) ([]string, int) {
	ids := make([]string, len(snapshot.Messages))
	for i, m := range snapshot.Messages {
		ids[i] = m.ID
	}
	offset := 0
	if snapshot.Summary != "" {
		offset = 1
	}
	return ids, offset
}

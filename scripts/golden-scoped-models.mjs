// Golden data for gi's port of Pi's /scoped-models selector
// (internal/tui/scoped_models.go), from Pi's own
// ScopedModelsSelectorComponent: rows as plain text, and the enabled ids
// after each key.
//   bun scripts/golden-scoped-models.mjs [node_modules dir]
import { existsSync, writeFileSync } from "node:fs";
import { homedir } from "node:os";
import { requirePi150, goldenPath } from './pi-reference.mjs';

const candidates = [process.argv[2], `${homedir()}/.bun/install/global/node_modules`, "/workspace/.cache/pi-ref/node_modules"];
const root = candidates.find((dir) => dir && existsSync(`${dir}/@earendil-works/pi-coding-agent/dist/index.js`));
if (!root) throw new Error("pi-coding-agent not found");
requirePi150(root, ['@earendil-works/pi-tui', '@earendil-works/pi-coding-agent']);
const agent = `${root}/@earendil-works/pi-coding-agent/dist`;
const { initTheme } = await import(`${agent}/modes/interactive/theme/theme.js`);
const { ScopedModelsSelectorComponent } = await import(`${agent}/modes/interactive/components/scoped-models-selector.js`);
const { KeybindingsManager } = await import(`${agent}/core/keybindings.js`);
const { setKeybindings } = await import(`${root}/@earendil-works/pi-tui/dist/index.js`);
initTheme("dark");
setKeybindings(new KeybindingsManager());

const strip = (s) => s.replace(/\x1b\[[0-9;]*m/g, "").replace(/\x1b_[^\x07]*\x07/g, "");
const models = [
	["anthropic", "claude-opus", "Claude Opus"],
	["anthropic", "claude-sonnet", "Claude Sonnet"],
	["openai", "gpt-5", "GPT-5"],
	["openai", "gpt-5-mini", "GPT-5 mini"],
	["google", "gemini-pro", "Gemini Pro"],
	["google", "gemini-flash", "Gemini Flash"],
	["xai", "grok", "Grok"],
	["mistral", "large", "Mistral Large"],
	["mistral", "small", "Mistral Small"],
	["groq", "llama", "Llama"],
].map(([provider, id, name]) => ({ provider, id, name }));
const keys = {
	up: "\x1b[A", down: "\x1b[B", enter: "\r", esc: "\x1b", "ctrl+c": "\x03", "ctrl+a": "\x01", "ctrl+x": "\x18",
	"ctrl+p": "\x10", "ctrl+s": "\x13", "alt+up": "\x1b[1;3A", "alt+down": "\x1b[1;3B", backspace: "\x7f",
};
// A step is a key name or ["type", text].
const scenarios = {
	all_enabled: { enabled: null, steps: ["down", "down", "enter", "down", "enter", "up", "alt+down", "ctrl+s"] },
	scoped_reorder: { enabled: ["openai/gpt-5", "anthropic/claude-opus", "gone/old"], steps: ["alt+down", "alt+down", "down", "alt+up", "up", "up", "enter", "ctrl+a"] },
	provider_toggle: { enabled: ["openai/gpt-5"], steps: ["ctrl+p", "down", "down", "ctrl+p", "ctrl+p", "ctrl+x", "enter"] },
	search: { enabled: null, steps: [["type", "gem"], "ctrl+x", "down", "enter", "backspace", "backspace", "backspace", ["type", "zzz"], "ctrl+c", "up"] },
	scroll: { enabled: null, steps: ["up", "up", "up", "down"] },
	cancel: { enabled: null, steps: ["esc"] },
};
const out = {};
for (const width of [80, 50]) {
	for (const [name, scenario] of Object.entries(scenarios)) {
		let changed = null, persisted = null, cancelled = false;
		const selector = new ScopedModelsSelectorComponent({ allModels: models, enabledModelIds: scenario.enabled }, {
			onChange: (ids) => { changed = ids; },
			onPersist: (ids) => { persisted = ids; },
			onCancel: () => { cancelled = true; },
		});
		const states = [{ step: "open", rows: selector.render(width).map((line) => strip(line).trimEnd()), changed, persisted, cancelled }];
		for (const step of scenario.steps) {
			if (Array.isArray(step)) for (const ch of step[1]) selector.handleInput(ch);
			else selector.handleInput(keys[step]);
			states.push({ step: Array.isArray(step) ? step.join(":") : step, rows: selector.render(width).map((line) => strip(line).trimEnd()), changed, persisted, cancelled });
		}
		(out[name] ??= { enabled: scenario.enabled, steps: scenario.steps, widths: {} }).widths[width] = states;
	}
}
writeFileSync(goldenPath('pi-scoped-models.json'), JSON.stringify({ models, scenarios: out }, null, 1) + "\n");

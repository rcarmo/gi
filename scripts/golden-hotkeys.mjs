// Golden data for gi's /hotkeys (internal/tui/hotkeys.go): Pi's own
// handleHotkeysCommand markdown, with Pi's default keybindings on Linux,
// macOS, WSL and Windows.
//   bun scripts/golden-hotkeys.mjs [node_modules dir]
//   (runs itself once per variant)
import { spawnSync } from "node:child_process";
import { existsSync, writeFileSync } from "node:fs";
import { homedir } from "node:os";
import { requirePi110, goldenPath } from './pi-reference.mjs';

const variant = process.env.GOLDEN_HOTKEYS_VARIANT;
if (!variant) {
	for (const [name, env] of [["default", {}], ["darwin", {}], ["wsl", { WSL_DISTRO_NAME: "Ubuntu" }], ["windows", {}]]) {
		const run = spawnSync(process.execPath, [process.argv[1], ...process.argv.slice(2)], { stdio: "inherit", env: { ...process.env, WSL_DISTRO_NAME: "", WSL_INTEROP: "", ...env, GOLDEN_HOTKEYS_VARIANT: name } });
		if (run.status !== 0) process.exit(run.status ?? 1);
	}
	process.exit(0);
}
// Pi picks its keys and their names (Option on macOS) from the platform.
if (variant === "windows") Object.defineProperty(process, "platform", { value: "win32" });
if (variant === "darwin") Object.defineProperty(process, "platform", { value: "darwin" });

const candidates = [process.argv[2], `${homedir()}/.bun/install/global/node_modules`, "/workspace/.cache/pi-ref/node_modules"];
const root = candidates.find((dir) => dir && existsSync(`${dir}/@earendil-works/pi-coding-agent/dist/index.js`));
if (!root) throw new Error("pi-coding-agent not found");
requirePi110(root, ['@earendil-works/pi-tui', '@earendil-works/pi-coding-agent']);
const agent = `${root}/@earendil-works/pi-coding-agent/dist`;
const { initTheme } = await import(`${agent}/modes/interactive/theme/theme.js`);
const { InteractiveMode } = await import(`${agent}/modes/interactive/interactive-mode.js`);
const { KeybindingsManager } = await import(`${agent}/core/keybindings.js`);
const { setKeybindings } = await import(`${root}/@earendil-works/pi-tui/dist/index.js`);
initTheme("dark");
const keybindings = new KeybindingsManager();
setKeybindings(keybindings);

const children = [];
const fake = Object.create(InteractiveMode.prototype);
const stub = {
	keybindings,
	session: { extensionRunner: { getShortcuts: () => new Map() } },
	chatContainer: { addChild: (child) => children.push(child) },
	ui: { requestRender() {} },
	getMarkdownThemeWithSettings: () => undefined,
};
for (const [key, value] of Object.entries(stub)) Object.defineProperty(fake, key, { value, configurable: true });
InteractiveMode.prototype.handleHotkeysCommand.call(fake);
const markdown = children.find((child) => child.constructor.name === "Markdown");
writeFileSync(goldenPath(`pi-hotkeys${variant === "default" ? "" : `-${variant}`}.md`), `${markdown.text}\n`);

// Golden data for gi's port of pi-tui's editor keys (internal/tui/
// multiline_input.go): word moves, kill ring, yank-pop, undo and character
// jump, recorded from pi-tui's own Editor with Pi's default keybindings.
//   bun scripts/golden-editor-keys.mjs [node_modules dir]
import { existsSync, writeFileSync } from "node:fs";
import { homedir } from "node:os";
import { requirePi110, goldenPath } from './pi-reference.mjs';

const candidates = [process.argv[2], `${homedir()}/.bun/install/global/node_modules`, "/workspace/.cache/pi-ref/node_modules"];
const root = candidates.find((dir) => dir && existsSync(`${dir}/@earendil-works/pi-tui/dist/index.js`));
if (!root) throw new Error("pi-tui not found");
requirePi110(root, ['@earendil-works/pi-tui']);
const tui = await import(`${root}/@earendil-works/pi-tui/dist/index.js`);

const identity = (s) => s;
const theme = new Proxy({}, { get: (_, key) => (key === "selectList" ? new Proxy({}, { get: () => identity }) : identity) });
// Legacy terminal bytes for the key names the Go test replays.
const keys = {
	left: "\x1b[D", right: "\x1b[C", "ctrl+a": "\x01", "ctrl+e": "\x05", "ctrl+b": "\x02", "ctrl+f": "\x06",
	"alt+b": "\x1bb", "alt+f": "\x1bf", "alt+left": "\x1b[1;3D", "alt+right": "\x1b[1;3C", "ctrl+left": "\x1b[1;5D", "ctrl+right": "\x1b[1;5C",
	backspace: "\x7f", "ctrl+w": "\x17", "alt+backspace": "\x1b\x7f", "alt+d": "\x1bd", "ctrl+u": "\x15", "ctrl+k": "\x0b",
	"ctrl+y": "\x19", "alt+y": "\x1by", "ctrl+-": "\x1f", "ctrl+]": "\x1d", "ctrl+alt+]": "\x1b\x1d",
};
// A step is ["set", text] (setText; the cursor goes to the end), ["type",
// text] (one key per character), ["paste", text] (bracketed paste) or a
// key name.
const scenarios = {
	word_moves_punctuation: [["set", "foo.bar(baz)  qux_1 é"], "alt+b", "alt+b", "alt+b", "alt+b", "alt+b", "alt+b", "alt+b", "ctrl+left", "ctrl+left", "alt+f", "alt+f", "ctrl+right", "alt+right", "alt+f", "alt+f", "alt+f"],
	word_moves_across_lines: [["set", "one two\nthree"], "alt+b", "alt+b", "alt+b", "alt+f", "alt+f", "alt+left"],
	kills_accumulate: [["set", "one two three four"], "ctrl+w", "alt+backspace", "ctrl+y", "ctrl+-", "ctrl+-", "ctrl+-"],
	kill_ring_yank_pop: [["set", "alpha beta gamma"], "ctrl+w", "alt+b", "ctrl+w", "ctrl+e", "ctrl+y", "alt+y", "alt+y", "alt+y", "ctrl+y"],
	yank_pop_needs_yank: [["set", "a b c"], "ctrl+w", "left", "ctrl+w", "alt+y", "ctrl+y", "right", "alt+y"],
	forward_kills: [["set", "one two three"], "ctrl+a", "alt+d", "alt+d", "ctrl+e", "ctrl+y"],
	line_kills: [["set", "first line\nsecond line"], "left", "left", "left", "ctrl+u", "ctrl+u", "ctrl+u", "ctrl+y", "ctrl+a", "ctrl+k", "ctrl+k", "ctrl+k", "ctrl+y"],
	line_edge_word_kills: [["set", "ab\ncd"], "ctrl+a", "ctrl+w", "ctrl+e", "alt+d", "ctrl+-", "ctrl+-"],
	undo_coalesces_words: [["type", "hello big  world"], "ctrl+-", "ctrl+-", "ctrl+-", "ctrl+-", "ctrl+-", "ctrl+-", "ctrl+-"],
	undo_after_move: [["type", "ab"], "left", ["type", "cd"], "ctrl+-", "ctrl+-", "backspace", "ctrl+-"],
	jump: [["set", "a.b.c d.e"], "ctrl+a", "ctrl+]", ["type", "."], "ctrl+]", ["type", "."], "ctrl+]", ["type", "z"], "ctrl+alt+]", ["type", "a"], "ctrl+e", "ctrl+alt+]", ["type", "."], "ctrl+]", "ctrl+]", ["type", "x"], "ctrl+]", "left", ["type", "y"]],
	jump_across_lines: [["set", "abc\ndef"], "ctrl+alt+]", ["type", "b"], "ctrl+]", ["type", "e"]],
	paste_marker_atomic: [["type", "see "], ["paste", "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11"], ["type", " ok"], "alt+b", "alt+b", "alt+b", "alt+f", "alt+f", "ctrl+e", "ctrl+w", "ctrl+w"],
};

const cursorOffset = (editor) => {
	const { line, col } = editor.getCursor();
	const lines = editor.getLines();
	let offset = 0;
	for (let i = 0; i < line; i++) offset += [...lines[i]].length + 1;
	return offset + [...lines[line].slice(0, col)].length;
};

const out = {};
for (const [name, steps] of Object.entries(scenarios)) {
	const editor = new tui.Editor({ requestRender() {} }, theme);
	const states = [];
	for (const step of steps) {
		if (Array.isArray(step)) {
			const [kind, text] = step;
			if (kind === "set") editor.setText(text);
			else if (kind === "paste") editor.handleInput(`\x1b[200~${text}\x1b[201~`);
			else for (const ch of text) editor.handleInput(ch);
		} else {
			editor.handleInput(keys[step]);
		}
		states.push({ step: Array.isArray(step) ? step.join(":") : step, text: editor.getText(), cursor: cursorOffset(editor) });
	}
	out[name] = { steps, states };
}
writeFileSync(goldenPath('pi-editor-keys.json'), JSON.stringify(out, null, 1) + "\n");

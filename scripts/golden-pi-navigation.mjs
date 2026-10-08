// Extract navigation defaults from the published pi-tui package, not Gi code.
// bun scripts/golden-pi-navigation.mjs /absolute/path/to/pi-tui-1.1.0
import { resolve, join } from 'node:path';
const root = resolve(process.argv[2] || '');
const pkg = await Bun.file(join(root, 'package.json')).json();
if (pkg.name !== '@earendil-works/pi-tui' || pkg.version !== '1.1.0') {
  throw Error('Provide the published @earendil-works/pi-tui 1.1.0 directory');
}
const { TUI_KEYBINDINGS } = await import(join(root, 'dist/keybindings.js'));
const ids = ['tui.editor.cursorLineStart', 'tui.editor.cursorLineEnd', 'tui.altScreen.top', 'tui.altScreen.bottom'];
const bindings = Object.fromEntries(ids.map(id => [id, [TUI_KEYBINDINGS[id].defaultKeys].flat()]));
await Bun.write('internal/tui/testdata/pi-1.1.0-navigation.json', JSON.stringify(bindings, null, 2) + '\n');

// Event-ownership-only adaptation; never edit supplied component source bytes.
function replace(source, from, to) {
  if (source.split(from).length !== 2) throw new Error(`Popup key adapter anchor changed: ${from}`);
  return source.replace(from, to);
}
export function patchQuickActionKeys(source) {
  source = replace(source, "import { getAgentCommands, getQuickActionsSettings } from '../api.js';", "import { getAgentCommands, getQuickActionsSettings, isQuickActionsReady } from '../api.js';\nimport { blocksQuickActions, settingsOwnsKeyboard } from '../gi-quick-actions.js';");
  source = replace(source, '    if (!isPopupTypeaheadKey(event)) return false;', '    if (blocksQuickActions(event, isQuickActionsReady())) return false;\n    if (!isPopupTypeaheadKey(event)) return false;');
  source = replace(source, '        const onPointerDown = (event) => {\n            if (!open) return;', '        const onPointerDown = (event) => {\n            if (settingsOwnsKeyboard()) return;\n            if (!open) return;');
  return replace(source, '        const onKeyDown = (event) => {', '        const onKeyDown = (event) => {\n            if (settingsOwnsKeyboard()) return;');
}
export function patchComposePopupKeys(source) {
  source = replace(source, '    const handlePopupKeyboardEvent = useCallback((e) => {\n        if (searchMode', '    const handlePopupKeyboardEvent = useCallback((e) => {\n        if (settingsOwnsKeyboard()) return false;\n        if (searchMode');
  for (const ref of ['modelPopupRef', 'sessionPopupRef']) {
    source = replace(source, `        const onPointerDown = (event) => {\n            const popup = ${ref}.current;`, `        const onPointerDown = (event) => {\n            if (settingsOwnsKeyboard()) return;\n            const popup = ${ref}.current;`);
  }
  return "import { settingsOwnsKeyboard } from '../gi-quick-actions.js';\n" + source;
}

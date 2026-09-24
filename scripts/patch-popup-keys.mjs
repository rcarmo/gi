// Event ownership and dismissal adaptation; supplied source bytes stay intact.
function replace(source, from, to) {
  if (source.split(from).length !== 2) throw new Error(`Popup key adapter anchor changed: ${from}`);
  return source.replace(from, to);
}
export function patchQuickActionKeys(source) {
  source = replace(source, "import { getAgentCommands, getQuickActionsSettings } from '../api.js';", "import { getAgentCommands, getQuickActionsSettings, isQuickActionsReady } from '../api.js';\nimport { blocksQuickActions, settingsOwnsKeyboard } from '../gi-quick-actions.js';");
  source = replace(source, '    if (!isPopupTypeaheadKey(event)) return false;', '    if (blocksQuickActions(event, isQuickActionsReady())) return false;\n    if (!isPopupTypeaheadKey(event)) return false;');
  source = replace(source, "import { html, useCallback, useEffect, useMemo, useRef, useState }", "import { html, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState }");
  source = "import { bindQuickActionsFocus, quickActionsOpener } from '../gi-quick-actions-focus.js';\n" + source;
  source = replace(source, '    const rootRef = useRef(null);', '    const rootRef = useRef(null);\n    const openerRef = useRef(null);');
  source = replace(source, `    useEffect(() => {
        if (!open) return;
        requestAnimationFrame(() => inputRef.current?.focus?.());
    }, [open]);`, `    useLayoutEffect(() => {
        if (!open || !rootRef.current || !inputRef.current) return;
        return bindQuickActionsFocus(rootRef.current, inputRef.current, openerRef.current, () => {
            setOpen(false); setQuery('');
        });
    }, [open]);`);
  source = replace(source, "                setQuery(String(event.key || ''));", "                openerRef.current = quickActionsOpener();\n                setQuery(String(event.key || ''));");
  source = replace(source, `            if (event.key === 'Escape') {
                event.preventDefault();
                setOpen(false);
                setQuery('');
                return;
            }`, "            if (event.key === 'Escape') return; // Focus/dismissal binding owns Escape.");
  source = replace(source, `        const onPointerDown = (event) => {
            if (!open) return;
            if (rootRef.current?.contains(event.target)) return;
            setOpen(false);
            setQuery('');
        };`, '');
  source = replace(source, "        document.addEventListener('pointerdown', onPointerDown, true);", '');
  source = replace(source, "            document.removeEventListener('pointerdown', onPointerDown, true);", '');
  return replace(source, '        const onKeyDown = (event) => {', '        const onKeyDown = (event) => {\n            if (settingsOwnsKeyboard()) return;');
}
export function patchComposePopupKeys(source) {
  source = replace(source, '    const handlePopupKeyboardEvent = useCallback((e) => {\n        if (searchMode', '    const handlePopupKeyboardEvent = useCallback((e) => {\n        if (settingsOwnsKeyboard()) return false;\n        if (searchMode');
  for (const ref of ['modelPopupRef', 'sessionPopupRef']) {
    source = replace(source, `        const onPointerDown = (event) => {\n            const popup = ${ref}.current;`, `        const onPointerDown = (event) => {\n            if (settingsOwnsKeyboard()) return;\n            const popup = ${ref}.current;`);
  }
  return "import { settingsOwnsKeyboard } from '../gi-quick-actions.js';\n" + source;
}

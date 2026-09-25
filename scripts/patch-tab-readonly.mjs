// Keep the supplied tab-strip bytes intact. Only Gi's read-only host opts in:
// no standalone viewer routes, viewport-bounded menu, no shortcuts under Settings.
export function patchTabReadonly(source) {
    for (const [from, to] of [
        ['html, useCallback, useEffect, useMemo, useRef, useState', 'html, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState'],
        ['// Close context menu on outside click or Escape\n    useEffect(() => {', '// Close context menu on outside click or Escape\n    useLayoutEffect(() => {'],
        ['export function TabStrip({ tabs,', 'export function TabStrip({ readOnlyHost = false, hostVisible = true, tabs,'],
        ['const onKeyDown = (e) => {', 'const onKeyDown = (e) => {\n            if (readOnlyHost && (e.defaultPrevented || e.isComposing || e.keyCode === 229 || stripRef.current?.closest("[hidden]") || document.querySelector(\'.settings-dialog[aria-modal="true"]\'))) return;'],
        ['    const stripRef = useRef(null);', `    const stripRef = useRef(null);
    const contextRef = useRef(null);
    useLayoutEffect(() => {
        if (readOnlyHost && !hostVisible) setContextMenu(null);
    }, [readOnlyHost, hostVisible]);
    useLayoutEffect(() => {
        if (!readOnlyHost || !stripRef.current) return;
        return bindReadonlyTabKeys(stripRef.current, id => onActivate?.(id), (id, x, y) => setContextMenu({id, x, y}));
    }, [readOnlyHost, tabs, onActivate]);
    useLayoutEffect(() => {
        if (!readOnlyHost || !contextMenu || !contextRef.current || !stripRef.current) return;
        return bindReadonlyTabMenu(contextRef.current, stripRef.current, contextMenu.id);
    }, [readOnlyHost, contextMenu]);`],
        ['    const handleClosePointerDown = useCallback((e) => {', `    const handleClosePointerDown = useCallback((e) => {
        // WebKit touch needs its compatibility click. Stop propagation without
        // cancelling the touch pointer default; close-click still owns removal.
        if (readOnlyHost && e.pointerType === 'touch') { e.stopPropagation(); return; }`],
        ['                    role="tab"', '                    role="tab"\n                    tabIndex=${readOnlyHost ? (tab.id === activeId ? 0 : -1) : undefined}\n                    data-readonly-tab-id=${readOnlyHost ? tab.id : undefined}'],
        ['<div class="tab-context-menu" style=', '<div class="tab-context-menu" ref=${contextRef} style='],
        ['const standaloneUrl = getStandaloneTabUrl(contextMenu.id, {', 'if (readOnlyHost) return null;\n                    const standaloneUrl = getStandaloneTabUrl(contextMenu.id, {'],
        ["style=${{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }}", "style=${{ left: (readOnlyHost ? Math.max(4, Math.min(contextMenu.x, window.innerWidth - 148)) : contextMenu.x) + 'px', top: (readOnlyHost ? Math.max(4, Math.min(contextMenu.y, window.innerHeight - 208)) : contextMenu.y) + 'px' }}"],
    ]) {
        if (source.split(from).length !== 2) throw new Error(`Read-only tab adapter anchor changed: ${from}`);
        source = source.replace(from, to);
    }
    return "import { bindReadonlyTabKeys, bindReadonlyTabMenu } from '../gi-readonly-tab-focus.js';\n" + source;
}

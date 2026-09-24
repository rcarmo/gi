// Keep the supplied tab-strip bytes intact. Only Gi's read-only host opts in:
// no standalone viewer routes, viewport-bounded menu, no shortcuts under Settings.
export function patchTabReadonly(source) {
    for (const [from, to] of [
        ['html, useCallback, useEffect, useMemo, useRef, useState', 'html, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState'],
        ['// Close context menu on outside click or Escape\n    useEffect(() => {', '// Close context menu on outside click or Escape\n    useLayoutEffect(() => {'],
        ['export function TabStrip({ tabs,', 'export function TabStrip({ readOnlyHost = false, tabs,'],
        ['const onKeyDown = (e) => {', 'const onKeyDown = (e) => {\n            if (readOnlyHost && document.querySelector(\'.settings-dialog[aria-modal="true"]\')) return;'],
        ['const standaloneUrl = getStandaloneTabUrl(contextMenu.id, {', 'if (readOnlyHost) return null;\n                    const standaloneUrl = getStandaloneTabUrl(contextMenu.id, {'],
        ["style=${{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }}", "style=${{ left: (readOnlyHost ? Math.max(4, Math.min(contextMenu.x, window.innerWidth - 148)) : contextMenu.x) + 'px', top: (readOnlyHost ? Math.max(4, Math.min(contextMenu.y, window.innerHeight - 208)) : contextMenu.y) + 'px' }}"],
    ]) {
        if (source.split(from).length !== 2) throw new Error(`Read-only tab adapter anchor changed: ${from}`);
        source = source.replace(from, to);
    }
    return source;
}

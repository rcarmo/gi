// @ts-nocheck
/**
 * use-editor-state.ts — Tab orchestration hook for the editor pane.
 *
 * Manages tab strip state (open/close/activate/pin/dirty), SSE
 * workspace-update file reload, rename sync, and beforeunload.
 *
 * File I/O and dirty tracking live in the StandaloneEditorInstance
 * (panes/editor-extension.ts). This hook coordinates tabs only.
 */

import { useState, useCallback, useRef, useEffect } from '../vendor/preact-htm.js';
import { paneRegistry, tabStore } from '../panes/index.js';

function renamePaneOverrides(previous, oldPath, newPath, type) {
    if (!(previous instanceof Map) || previous.size === 0 || !oldPath || !newPath) return previous;
    let changed = false;
    const next = new Map();
    for (const [key, value] of previous.entries()) {
        let nextKey = key;
        if (type === 'dir') {
            if (key === oldPath) {
                nextKey = newPath;
                changed = true;
            } else if (key.startsWith(`${oldPath}/`)) {
                nextKey = `${newPath}${key.slice(oldPath.length)}`;
                changed = true;
            }
        } else if (key === oldPath) {
            nextKey = newPath;
            changed = true;
        }
        next.set(nextKey, value);
    }
    return changed ? next : previous;
}

function renamePathSet(previous, oldPath, newPath, type) {
    if (!(previous instanceof Set) || previous.size === 0 || !oldPath || !newPath) return previous;
    let changed = false;
    const next = new Set();
    for (const value of previous.values()) {
        let nextValue = value;
        if (type === 'dir') {
            if (value === oldPath) {
                nextValue = newPath;
                changed = true;
            } else if (value.startsWith(`${oldPath}/`)) {
                nextValue = `${newPath}${value.slice(oldPath.length)}`;
                changed = true;
            }
        } else if (value === oldPath) {
            nextValue = newPath;
            changed = true;
        }
        next.add(nextValue);
    }
    return changed ? next : previous;
}

/**
 * Custom hook that manages editor tab orchestration.
 *
 * @returns Tab state, handlers, and active tab info.
 */
export function useEditorState({ onTabClosed } = {}) {
    // Store callback in ref so close handlers never re-create when the caller changes identity
    const onTabClosedRef = useRef(onTabClosed);
    onTabClosedRef.current = onTabClosed;

    // ── Tab strip state (driven by tabStore) ────────────────────
    const [tabStripTabs, setTabStripTabs] = useState(() => tabStore.getTabs());
    const [tabStripActiveId, setTabStripActiveId] = useState(() => tabStore.getActiveId());
    const [editorOpen, setEditorOpen] = useState(() => tabStore.getTabs().length > 0);

    useEffect(() => {
        return tabStore.onChange((tabs, activeId) => {
            setTabStripTabs(tabs);
            setTabStripActiveId(activeId);
            setEditorOpen(tabs.length > 0);
        });
    }, []);

    // ── Markdown preview state ────────────────────────────────
    const [previewTabs, setPreviewTabs] = useState(() => new Set());
    const [diffTabs, setDiffTabs] = useState(() => new Set());
    const [tabPaneOverrides, setTabPaneOverrides] = useState(() => new Map());

    const handleTabTogglePreview = useCallback((id) => {
        setPreviewTabs((prev) => {
            const next = new Set(prev);
            if (next.has(id)) {
                next.delete(id);
            } else {
                next.add(id);
            }
            return next;
        });
    }, []);

    // Clean up preview state when tabs close — declared before tab
    // action callbacks so it can appear in their dependency arrays
    // without a temporal dead zone violation.
    const cleanupPreviewTab = useCallback((id) => {
        setPreviewTabs((prev) => {
            if (!prev.has(id)) return prev;
            const next = new Set(prev);
            next.delete(id);
            return next;
        });
    }, []);

    const cleanupDiffTab = useCallback((id) => {
        setDiffTabs((prev) => {
            if (!prev.has(id)) return prev;
            const next = new Set(prev);
            next.delete(id);
            return next;
        });
    }, []);

    const cleanupPaneOverride = useCallback((id) => {
        setTabPaneOverrides((prev) => {
            if (!prev.has(id)) return prev;
            const next = new Map(prev);
            next.delete(id);
            return next;
        });
    }, []);

    // ── Tab actions ─────────────────────────────────────────────

    /** Open a file in the editor. Creates a tab and sets it active. */
    const openEditor = useCallback((path, options = {}) => {
        if (!path) return;
        const paneOverrideId = typeof options?.paneOverrideId === 'string' && options.paneOverrideId.trim()
            ? options.paneOverrideId.trim()
            : null;
        // Verify there's a pane handler for this file type
        const context = { path, mode: 'edit' };
        try {
            const pane = paneOverrideId ? paneRegistry.get(paneOverrideId) : paneRegistry.resolve(context);
            if (!pane) {
                const fallback = paneRegistry.get('editor');
                if (!fallback) {
                    console.warn(`[openEditor] No pane handler for: ${path}`);
                    return;
                }
            }
        } catch (err) {
            console.warn(`[openEditor] paneRegistry.resolve() error for "${path}":`, err);
        }
        const label = typeof options?.label === 'string' && options.label.trim() ? options.label.trim() : undefined;
        const viewState = options?.viewState && typeof options.viewState === 'object' ? options.viewState : null;
        const diffMode = options?.diffMode === 'saved' ? 'saved' : null;
        tabStore.open(path, label);
        if (viewState) {
            tabStore.saveViewState(path, viewState);
        }
        if (paneOverrideId) {
            setTabPaneOverrides((prev) => {
                if (prev.get(path) === paneOverrideId) return prev;
                const next = new Map(prev);
                next.set(path, paneOverrideId);
                return next;
            });
        }
        if (diffMode === 'saved') {
            setDiffTabs((prev) => {
                if (prev.has(path)) return prev;
                const next = new Set(prev);
                next.add(path);
                return next;
            });
        }
    }, []);

    /** Close the active tab (with dirty confirmation). */
    const closeEditor = useCallback(() => {
        const activeId = tabStore.getActiveId();
        if (activeId) {
            const tab = tabStore.get(activeId);
            if (tab?.dirty) {
                const confirmed = window.confirm(`"${tab.label}" has unsaved changes. Close anyway?`);
                if (!confirmed) return;
            }
            tabStore.close(activeId);
            cleanupPreviewTab(activeId);
            cleanupDiffTab(activeId);
            cleanupPaneOverride(activeId);
            onTabClosedRef.current?.(activeId);
        }
    }, [cleanupDiffTab, cleanupPaneOverride, cleanupPreviewTab]);

    /** Close a specific tab (from tab strip). */
    const handleTabClose = useCallback((id) => {
        const tab = tabStore.get(id);
        if (tab?.dirty) {
            const confirmed = window.confirm(`"${tab.label}" has unsaved changes. Close anyway?`);
            if (!confirmed) return;
        }
        tabStore.close(id);
        cleanupPreviewTab(id);
        cleanupDiffTab(id);
        cleanupPaneOverride(id);
        onTabClosedRef.current?.(id);
    }, [cleanupDiffTab, cleanupPaneOverride, cleanupPreviewTab]);

    /** Activate a tab by id. */
    const handleTabActivate = useCallback((id) => {
        tabStore.activate(id);
    }, []);

    /** Close all other tabs. */
    const handleTabCloseOthers = useCallback((id) => {
        const others = tabStore.getTabs().filter(t => t.id !== id && !t.pinned);
        const dirtyCount = others.filter(t => t.dirty).length;
        if (dirtyCount > 0) {
            const confirmed = window.confirm(`${dirtyCount} unsaved tab${dirtyCount > 1 ? 's' : ''} will be closed. Continue?`);
            if (!confirmed) return;
        }
        const closedIds = others.map(t => t.id);
        tabStore.closeOthers(id);
        closedIds.forEach(cid => {
            cleanupPreviewTab(cid);
            cleanupDiffTab(cid);
            cleanupPaneOverride(cid);
            onTabClosedRef.current?.(cid);
        });
    }, [cleanupDiffTab, cleanupPaneOverride, cleanupPreviewTab]);

    /** Close all tabs. */
    const handleTabCloseAll = useCallback(() => {
        const tabs = tabStore.getTabs().filter(t => !t.pinned);
        const dirtyCount = tabs.filter(t => t.dirty).length;
        if (dirtyCount > 0) {
            const confirmed = window.confirm(`${dirtyCount} unsaved tab${dirtyCount > 1 ? 's' : ''} will be closed. Continue?`);
            if (!confirmed) return;
        }
        const closedIds = tabs.map(t => t.id);
        tabStore.closeAll();
        closedIds.forEach(cid => {
            cleanupPreviewTab(cid);
            cleanupDiffTab(cid);
            cleanupPaneOverride(cid);
            onTabClosedRef.current?.(cid);
        });
    }, [cleanupDiffTab, cleanupPaneOverride, cleanupPreviewTab]);

    /** Toggle pin on a tab. */
    const handleTabTogglePin = useCallback((id) => {
        tabStore.togglePin(id);
    }, []);

    /** Toggle Compare to Saved diff mode on a tab. */
    const handleTabToggleDiff = useCallback((id) => {
        if (!id) return;
        setDiffTabs((prev) => {
            const next = new Set(prev);
            if (next.has(id)) next.delete(id);
            else next.add(id);
            return next;
        });
        tabStore.activate(id);
    }, []);

    /** Replace a specialized editor tab with the generic source editor. */
    const handleTabEditSource = useCallback((id) => {
        if (!id) return;
        setTabPaneOverrides((prev) => {
            if (prev.get(id) === 'editor') return prev;
            const next = new Map(prev);
            next.set(id, 'editor');
            return next;
        });
        tabStore.activate(id);
    }, []);

    /** Reveal active tab in workspace explorer. */
    const revealInExplorer = useCallback(() => {
        const activeId = tabStore.getActiveId();
        if (activeId) {
            window.dispatchEvent(new CustomEvent('workspace-reveal-path', { detail: { path: activeId } }));
        }
    }, []);

    // ── SSE rename sync ─────────────────────────────────────────
    useEffect(() => {
        const handleFileRenamed = (e) => {
            const { oldPath, newPath, type } = e.detail || {};
            if (!oldPath || !newPath) return;
            if (type === 'dir') {
                for (const tab of tabStore.getTabs()) {
                    if (tab.path === oldPath || tab.path.startsWith(`${oldPath}/`)) {
                        const updatedPath = `${newPath}${tab.path.slice(oldPath.length)}`;
                        tabStore.rename(tab.id, updatedPath);
                    }
                }
            } else {
                tabStore.rename(oldPath, newPath);
            }
            setPreviewTabs((prev) => renamePathSet(prev, oldPath, newPath, type));
            setDiffTabs((prev) => renamePathSet(prev, oldPath, newPath, type));
            setTabPaneOverrides((prev) => renamePaneOverrides(prev, oldPath, newPath, type));
        };
        window.addEventListener('workspace-file-renamed', handleFileRenamed);
        return () => window.removeEventListener('workspace-file-renamed', handleFileRenamed);
    }, []);

    // ── Warn on close with unsaved changes ──────────────────────
    useEffect(() => {
        const handleBeforeUnload = (e) => {
            if (tabStore.hasUnsaved()) {
                e.preventDefault();
                e.returnValue = '';
            }
        };
        window.addEventListener('beforeunload', handleBeforeUnload);
        return () => window.removeEventListener('beforeunload', handleBeforeUnload);
    }, []);

    return {
        // State
        editorOpen,
        tabStripTabs,
        tabStripActiveId,
        previewTabs,
        diffTabs,
        tabPaneOverrides,
        // Handlers
        openEditor,
        closeEditor,
        handleTabClose,
        handleTabActivate,
        handleTabCloseOthers,
        handleTabCloseAll,
        handleTabTogglePin,
        handleTabTogglePreview,
        handleTabToggleDiff,
        handleTabEditSource,
        revealInExplorer,
    };
}

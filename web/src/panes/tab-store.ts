// @ts-nocheck
/**
 * tab-store.ts — Tab state management for the pane system.
 *
 * Manages open tabs, active tab, dirty flags, MRU order, and per-tab
 * view state cache. Framework-agnostic — emits change events so the
 * UI layer can re-render.
 */

/** State of a single open tab — identity, dirty flag, pin status, and cached view state. */
export interface TabState {
    /** Unique tab id (usually the file path). */
    id: string;
    /** Display label (filename). */
    label: string;
    /** Full file path. */
    path: string;
    /** Whether the tab has unsaved changes. */
    dirty: boolean;
    /** Whether the tab is pinned (won't close with "close others"). */
    pinned: boolean;
    /** Cached view state (cursor position, scroll offset, etc.). */
    viewState?: TabViewState;
}

/** Saved editor view state for restoring cursor position and scroll on tab switch. */
export interface TabViewState {
    /** Cursor line number. */
    cursorLine?: number;
    /** Cursor column. */
    cursorCol?: number;
    /** Scroll top offset. */
    scrollTop?: number;
}

type TabChangeListener = (tabs: TabState[], activeId: string | null) => void;

class TabStoreImpl {
    private tabs: Map<string, TabState> = new Map();
    private activeId: string | null = null;
    /** Most-recently-used order (most recent first). */
    private mruOrder: string[] = [];
    private listeners: Set<TabChangeListener> = new Set();

    /** Subscribe to tab changes. Returns unsubscribe function. */
    onChange(listener: TabChangeListener): () => void {
        this.listeners.add(listener);
        return () => this.listeners.delete(listener);
    }

    private notify() {
        const tabs = this.getTabs();
        const activeId = this.activeId;
        for (const listener of this.listeners) {
            try { listener(tabs, activeId); } catch (err) { console.warn('[tab-store] Change listener failed:', err); }
        }
    }

    /** Open a tab (or activate if already open). */
    open(path: string, label?: string): TabState {
        let tab = this.tabs.get(path);
        if (!tab) {
            tab = {
                id: path,
                label: label || path.split('/').pop() || path,
                path,
                dirty: false,
                pinned: false,
            };
            this.tabs.set(path, tab);
        }
        this.activate(path);
        return tab;
    }

    /** Activate a tab by id. */
    activate(id: string): void {
        if (!this.tabs.has(id)) return;
        this.activeId = id;
        // Move to front of MRU
        this.mruOrder = [id, ...this.mruOrder.filter(x => x !== id)];
        this.notify();
    }

    /** Close a tab by id. Returns true if closed. */
    close(id: string): boolean {
        const tab = this.tabs.get(id);
        if (!tab) return false;
        this.tabs.delete(id);
        this.mruOrder = this.mruOrder.filter(x => x !== id);
        // If we closed the active tab, activate MRU
        if (this.activeId === id) {
            this.activeId = this.mruOrder[0] || null;
        }
        this.notify();
        return true;
    }

    /** Close all tabs except the given id. Respects pinned tabs. */
    closeOthers(keepId: string): void {
        for (const [id, tab] of this.tabs) {
            if (id !== keepId && !tab.pinned) {
                this.tabs.delete(id);
                this.mruOrder = this.mruOrder.filter(x => x !== id);
            }
        }
        if (this.activeId && !this.tabs.has(this.activeId)) {
            this.activeId = keepId;
        }
        this.notify();
    }

    /** Close all unpinned tabs. */
    closeAll(): void {
        for (const [id, tab] of this.tabs) {
            if (!tab.pinned) {
                this.tabs.delete(id);
                this.mruOrder = this.mruOrder.filter(x => x !== id);
            }
        }
        if (this.activeId && !this.tabs.has(this.activeId)) {
            this.activeId = this.mruOrder[0] || null;
        }
        this.notify();
    }

    /** Set dirty flag for a tab. */
    setDirty(id: string, dirty: boolean): void {
        const tab = this.tabs.get(id);
        if (!tab || tab.dirty === dirty) return;
        tab.dirty = dirty;
        this.notify();
    }

    /** Toggle pin state for a tab. */
    togglePin(id: string): void {
        const tab = this.tabs.get(id);
        if (!tab) return;
        tab.pinned = !tab.pinned;
        this.notify();
    }

    /** Save view state for a tab (cursor, scroll). */
    saveViewState(id: string, viewState: TabViewState): void {
        const tab = this.tabs.get(id);
        if (tab) tab.viewState = viewState;
    }

    /** Get view state for a tab. */
    getViewState(id: string): TabViewState | undefined {
        return this.tabs.get(id)?.viewState;
    }

    /** Rename a tab (after file rename/move). */
    rename(oldId: string, newPath: string, newLabel?: string): void {
        const tab = this.tabs.get(oldId);
        if (!tab) return;
        this.tabs.delete(oldId);
        tab.id = newPath;
        tab.path = newPath;
        tab.label = newLabel || newPath.split('/').pop() || newPath;
        this.tabs.set(newPath, tab);
        this.mruOrder = this.mruOrder.map(x => x === oldId ? newPath : x);
        if (this.activeId === oldId) this.activeId = newPath;
        this.notify();
    }

    /** Get all tabs in insertion order. */
    getTabs(): TabState[] {
        return Array.from(this.tabs.values());
    }

    /** Get active tab id. */
    getActiveId(): string | null {
        return this.activeId;
    }

    /** Get active tab. */
    getActive(): TabState | null {
        return this.activeId ? this.tabs.get(this.activeId) || null : null;
    }

    /** Get a tab by id. */
    get(id: string): TabState | undefined {
        return this.tabs.get(id);
    }

    /** Get tab count. */
    get size(): number {
        return this.tabs.size;
    }

    /** Whether any tab has unsaved changes. */
    hasUnsaved(): boolean {
        for (const tab of this.tabs.values()) {
            if (tab.dirty) return true;
        }
        return false;
    }

    /** Get dirty tabs. */
    getDirtyTabs(): TabState[] {
        return Array.from(this.tabs.values()).filter(t => t.dirty);
    }

    /** Cycle to next tab. Wraps around. */
    nextTab(): void {
        const tabs = this.getTabs();
        if (tabs.length <= 1) return;
        const idx = tabs.findIndex(t => t.id === this.activeId);
        const next = tabs[(idx + 1) % tabs.length];
        this.activate(next.id);
    }

    /** Cycle to previous tab. Wraps around. */
    prevTab(): void {
        const tabs = this.getTabs();
        if (tabs.length <= 1) return;
        const idx = tabs.findIndex(t => t.id === this.activeId);
        const prev = tabs[(idx - 1 + tabs.length) % tabs.length];
        this.activate(prev.id);
    }

    /** Switch to most recent tab (MRU[1] if available). */
    mruSwitch(): void {
        if (this.mruOrder.length > 1) {
            this.activate(this.mruOrder[1]);
        }
    }
}

/** Global tab store singleton. */
export const tabStore = new TabStoreImpl();

/** Public type alias for the TabStore singleton. */
export type TabStore = TabStoreImpl;

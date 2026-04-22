// @ts-nocheck
/**
 * pane-registry.ts — Singleton registry for WebPaneExtension instances.
 *
 * Manages registration, resolution (file-type routing), and listing
 * of pane extensions. Used by the app shell to find the right pane
 * for a given file/context.
 */

import type { WebPaneExtension, PaneContext } from './pane-types.js';

/** Singleton pane registry. */
class PaneRegistryImpl {
    private extensions: Map<string, WebPaneExtension> = new Map();

    /** Register a pane extension. Overwrites if id already exists. */
    register(ext: WebPaneExtension): void {
        this.extensions.set(ext.id, ext);
    }

    /** Remove a pane extension by id. */
    unregister(id: string): void {
        this.extensions.delete(id);
    }

    /**
     * Resolve the best "tabs" pane for a given context.
     * Calls canHandle() on each registered tabs pane, picks highest priority.
     * Returns undefined if no pane can handle the context.
     */
    resolve(context: PaneContext): WebPaneExtension | undefined {
        let best: WebPaneExtension | undefined;
        let bestPriority = -Infinity;

        for (const ext of this.extensions.values()) {
            if (ext.placement !== 'tabs') continue;
            if (!ext.canHandle) continue;

            try {
                const result = ext.canHandle(context);
                if (result === false || result === 0) continue;

                const priority = result === true ? 0 : (typeof result === 'number' ? result : 0);
                if (priority > bestPriority) {
                    bestPriority = priority;
                    best = ext;
                }
            } catch (err) {
                // canHandle() threw — skip this extension silently
                console.warn(`[PaneRegistry] canHandle() error for "${ext.id}":`, err);
            }
        }

        return best;
    }

    /** List all registered pane extensions. */
    list(): WebPaneExtension[] {
        return Array.from(this.extensions.values());
    }

    /** Get all registered dock panes. */
    getDockPanes(): WebPaneExtension[] {
        return Array.from(this.extensions.values()).filter(ext => ext.placement === 'dock');
    }

    /** Get all registered tab panes. */
    getTabPanes(): WebPaneExtension[] {
        return Array.from(this.extensions.values()).filter(ext => ext.placement === 'tabs');
    }

    /** Get a specific extension by id. */
    get(id: string): WebPaneExtension | undefined {
        return this.extensions.get(id);
    }

    /** Number of registered extensions. */
    get size(): number {
        return this.extensions.size;
    }
}

/** The global pane registry singleton. */
export const paneRegistry = new PaneRegistryImpl();

/** Public type alias for the PaneRegistry singleton. */
export type PaneRegistry = PaneRegistryImpl;

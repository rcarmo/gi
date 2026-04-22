// @ts-nocheck
/**
 * markdown-preview-pane.ts — Example readonly markdown preview pane.
 *
 * Demonstrates how to create a custom WebPaneExtension that previews
 * .md files as rendered HTML. Overrides the default editor for markdown
 * files with a higher priority (10 vs editor's 1).
 *
 * To use: import and register with paneRegistry in app.ts.
 *
 * @example
 * ```typescript
 * import { paneRegistry } from './panes/index.js';
 * import { markdownPreviewPaneExtension } from './panes/examples/markdown-preview-pane.js';
 * paneRegistry.register(markdownPreviewPaneExtension);
 * ```
 */

import type { WebPaneExtension, PaneContext, PaneInstance, PaneCapability } from '../pane-types.js';

/** Renders markdown to HTML using the global marked library (loaded in index.html). */
function renderMarkdownToHtml(md: string): string {
    // @ts-ignore — marked is loaded globally from vendor
    if (typeof globalThis.marked?.parse === 'function') {
        return globalThis.marked.parse(md);
    }
    // Fallback: escape HTML and wrap in <pre>
    const escaped = md.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return `<pre>${escaped}</pre>`;
}

/**
 * Mounted instance of the markdown preview pane.
 * Renders markdown content as HTML in a scrollable container.
 */
class MarkdownPreviewInstance implements PaneInstance {
    private el: HTMLElement;
    private disposed = false;

    constructor(container: HTMLElement, context: PaneContext) {
        this.el = document.createElement('div');
        this.el.className = 'markdown-preview-pane';
        this.el.setAttribute('tabindex', '0');
        this.el.style.cssText = 'padding: 16px 24px; overflow: auto; height: 100%; font-size: 14px; line-height: 1.6;';
        this.el.innerHTML = renderMarkdownToHtml(context.content || '');
        container.appendChild(this.el);
    }

    /** Readonly pane — no editable content. */
    getContent(): string | undefined {
        return undefined;
    }

    /** Readonly pane — never dirty. */
    isDirty(): boolean {
        return false;
    }

    /** Re-render when file is reloaded from disk. */
    setContent(content: string, _mtime: string): void {
        if (this.disposed) return;
        this.el.innerHTML = renderMarkdownToHtml(content);
    }

    /** Focus the preview container. */
    focus(): void {
        this.el?.focus();
    }

    /** Clean up DOM. */
    dispose(): void {
        if (this.disposed) return;
        this.disposed = true;
        this.el?.remove();
    }
}

/**
 * Markdown preview pane extension.
 *
 * Handles `.md` and `.markdown` files with priority 10 (overrides the
 * default editor at priority 1). Readonly — renders markdown as HTML.
 */
export const markdownPreviewPaneExtension: WebPaneExtension = {
    id: 'markdown-preview',
    label: 'Markdown Preview',
    icon: 'preview',
    capabilities: ['preview', 'readonly'] as PaneCapability[],
    placement: 'tabs',

    /**
     * Handle .md and .markdown files with higher priority than the editor.
     * @param context - The pane context with file path.
     * @returns 10 for markdown files, false otherwise.
     */
    canHandle(context: PaneContext): boolean | number {
        if (!context.path) return false;
        const lower = context.path.toLowerCase();
        return (lower.endsWith('.md') || lower.endsWith('.markdown')) ? 10 : false;
    },

    /**
     * Mount the markdown preview into the container.
     * @param container - DOM element to render into.
     * @param context - File context with path and content.
     * @returns A PaneInstance for host communication.
     */
    mount(container: HTMLElement, context: PaneContext): PaneInstance {
        return new MarkdownPreviewInstance(container, context);
    },
};

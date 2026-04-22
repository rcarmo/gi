// @ts-nocheck
/**
 * drawio-pane.ts — WebPaneExtension for editing .drawio diagrams.
 *
 * The pane uses the shared /drawio/edit.html wrapper so the in-app tab and the
 * standalone draw.io route stay on the exact same load/save/export codepath.
 */

import type { PaneCapability, PaneContext, PaneInstance, WebPaneExtension } from './pane-types.js';

/** Also match .drawio.xml and .drawio.svg by checking full filename. */
function isDrawioFile(filePath?: string): boolean {
    if (!filePath) return false;
    const lower = filePath.toLowerCase();
    return lower.endsWith('.drawio')
        || lower.endsWith('.drawio.xml')
        || lower.endsWith('.drawio.svg')
        || lower.endsWith('.drawio.png');
}

function esc(s: string): string {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

export function buildDrawioEditorUrl(filePath: string): string {
    return `/drawio/edit.html?path=${encodeURIComponent(filePath || '')}`;
}

// ── Preview card (workspace browser) ────────────────────────────

class DrawioPreviewCard implements PaneInstance {
    private container: HTMLElement;
    private disposed = false;

    constructor(container: HTMLElement, context: PaneContext) {
        this.container = container;
        const filePath = context.path || '';
        const name = filePath.split('/').pop() || 'diagram.drawio';

        const wrapper = document.createElement('div');
        wrapper.style.cssText = 'width:100%;height:100%;display:flex;align-items:center;justify-content:center;background:var(--bg-primary,#1a1a1a);';
        wrapper.innerHTML = `
            <div style="text-align:center;max-width:360px;padding:24px;">
                <div style="font-size:56px;margin-bottom:12px;">📐</div>
                <div style="font-size:14px;font-weight:600;color:var(--text-primary,#e0e0e0);margin-bottom:4px;word-break:break-word;">${esc(name)}</div>
                <div style="font-size:11px;color:var(--text-secondary,#888);margin-bottom:20px;">Draw.io Diagram</div>
                <button id="drawio-open-tab" style="padding:8px 20px;background:var(--accent-color,#1d9bf0);color:var(--accent-contrast-text,#fff);
                    border:none;border-radius:5px;font-size:13px;font-weight:500;cursor:pointer;
                    transition:background 0.15s;"
                    onmouseenter="this.style.background='var(--accent-hover,#1a8cd8)'"
                    onmouseleave="this.style.background='var(--accent-color,#1d9bf0)'">
                    Edit in Tab
                </button>
            </div>
        `;
        container.appendChild(wrapper);

        const btn = wrapper.querySelector('#drawio-open-tab') as HTMLButtonElement;
        if (btn) {
            btn.addEventListener('click', () => {
                const evt = new CustomEvent('drawio:open-tab', {
                    bubbles: true,
                    detail: { path: filePath },
                });
                container.dispatchEvent(evt);
            });
        }
    }

    getContent(): string | undefined { return undefined; }
    isDirty(): boolean { return false; }
    focus(): void {}
    resize(): void {}
    dispose(): void {
        if (this.disposed) return;
        this.disposed = true;
        this.container.innerHTML = '';
    }
}

// ── Full editor (editor tab) ────────────────────────────────────

class DrawioEditorInstance implements PaneInstance {
    private container: HTMLElement;
    private iframe: HTMLIFrameElement | null = null;
    private disposed = false;

    constructor(container: HTMLElement, context: PaneContext) {
        this.container = container;

        const iframe = document.createElement('iframe');
        iframe.src = buildDrawioEditorUrl(context.path || '');
        iframe.style.cssText = 'width:100%;height:100%;border:none;background:#1e1e1e;';
        iframe.setAttribute('title', 'Draw.io editor');
        this.iframe = iframe;
        container.appendChild(iframe);
    }

    getContent(): string | undefined { return undefined; }
    isDirty(): boolean { return false; }
    focus(): void { this.iframe?.focus(); }
    resize(): void {}
    dispose(): void {
        if (this.disposed) return;
        this.disposed = true;
        if (this.iframe) {
            this.iframe.src = 'about:blank';
            this.iframe = null;
        }
        this.container.innerHTML = '';
    }
}

// ── Extension ───────────────────────────────────────────────────

export const drawioPaneExtension: WebPaneExtension = {
    id: 'drawio-editor',
    label: 'Draw.io Editor',
    icon: 'git-merge',
    capabilities: ['edit', 'preview'] as PaneCapability[],
    placement: 'tabs',

    canHandle(context: PaneContext): boolean | number {
        if (!isDrawioFile(context?.path)) return false;
        return 60;
    },

    mount(container: HTMLElement, context: PaneContext): PaneInstance {
        if (context?.mode === 'view') {
            return new DrawioPreviewCard(container, context);
        }
        return new DrawioEditorInstance(container, context);
    },
};

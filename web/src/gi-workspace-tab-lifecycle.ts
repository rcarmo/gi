import type { PaneContext, PaneInstance, WebPaneExtension } from './panes/pane-types.js';

type HostOptions = {
    close?: () => void;
    observeResize?: (container: HTMLElement, resized: () => void) => () => void;
};

// Container changes include drawer/layout changes, not just window resizes.
// Older browsers get a window fallback. Both paths are detached on disposal.
export function observeWorkspaceTab(container: HTMLElement, resized: () => void): () => void {
    const view = container.ownerDocument?.defaultView;
    if (!view) return () => {};
    if (view.ResizeObserver) {
        const observer = new view.ResizeObserver(resized);
        observer.observe(container);
        return () => observer.disconnect();
    }
    view.addEventListener('resize', resized);
    return () => view.removeEventListener('resize', resized);
}

// Each mount owns its request, callbacks and instance. Refresh deliberately
// remounts: setContent cannot carry the full binary/truncated preview metadata.
// Internal host entrypoint: WorkspaceTab owns cleanup-before-remount and passes
// its Preact state setter/tab-close handler. Host callbacks are trusted (must
// not throw); only extension lifecycle hooks are isolated here.
export function mountWorkspaceTab(container: HTMLElement, path: string, changed: (state: any) => void,
    read: (path: string, limit: number) => Promise<any>,
    registry: { resolve(context: PaneContext): WebPaneExtension | null }, options: HostOptions = {}) {
    let live = true, instance: PaneInstance | null = null, unobserve: (() => void) | null = null;
    const stop = () => {
        if (!live) return;
        live = false; // Fence callbacks even if disposal re-enters or throws.
        const mounted = instance, disconnect = unobserve;
        instance = null;
        unobserve = null;
        try { disconnect?.(); } catch { /* Continue disposing a broken pane. */ }
        try { mounted?.dispose(); } catch { /* Never leave its host occupied. */ }
        container.replaceChildren();
    };
    const fail = (error: unknown) => {
        if (!live) return;
        stop();
        changed({ loading: false, error: error instanceof Error ? error.message : 'Preview failed.' });
    };
    changed({ loading: true, error: '' });
    void (async () => {
        try {
            const preview = await read(path, 20000);
            if (!live) return;
            const context: PaneContext = { path, mode: 'view', preview };
            if (preview?.kind === 'text' && typeof preview.text === 'string') context.content = preview.text;
            if (typeof preview?.mtime === 'string') context.mtime = preview.mtime;
            if (Number.isFinite(preview?.size) && preview.size >= 0) context.size = preview.size;
            const extension = registry.resolve(context);
            if (!extension) throw new Error('No read-only preview available.');
            // Do not silently mount editors/terminals or mixed-capability panes:
            // Gi has no save/dirty/dock/transfer host contract yet.
            if (extension.placement !== 'tabs' || !extension.capabilities?.length ||
                extension.capabilities.some(capability => capability !== 'readonly' && capability !== 'preview')) {
                throw new Error('This pane requires capabilities outside Gi’s read-only preview host.');
            }
            instance = extension.mount(container, context);
            instance.onClose?.(() => {
                if (!live) return;
                stop();
                options.close?.();
            });
            if (!live) return; // Registration itself may request closure.
            const resized = () => {
                if (!live) return;
                try { instance?.resize?.(); } catch (error) { fail(error); }
            };
            const disconnect = (options.observeResize ?? observeWorkspaceTab)(container, resized);
            if (!live) { disconnect(); return; }
            unobserve = disconnect;
            resized();
            // Never focus after an asynchronous read: the user may now own
            // composer/Settings focus. Existing DOM focus remains authoritative.
            if (live) changed({ loading: false, error: '' });
        } catch (error) { fail(error); }
    })();
    return stop;
}

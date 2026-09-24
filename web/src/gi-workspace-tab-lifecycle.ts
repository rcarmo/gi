// Each path owns its request and supplied pane instance. Switching/closing
// disposes immediately; a late response can never mount into the next tab.
export function mountWorkspaceTab(container: HTMLElement, path: string, changed: (state: any) => void,
    read: (path: string, limit: number) => Promise<any>, registry: any) {
    let live = true, instance: any = null;
    changed({ loading: true, error: '' });
    void read(path, 20000).then(preview => {
        if (!live) return;
        try {
            const context = { path, mode: 'view' as const, preview };
            const extension = registry.resolve(context);
            if (!extension) throw new Error('No read-only preview available.');
            instance = extension.mount(container, context);
            changed({ loading: false, error: '' });
        } catch (error) {
            container.replaceChildren();
            changed({ loading: false, error: error.message || 'Preview failed.' });
        }
    }, error => {
        if (live) changed({ loading: false, error: error.message || 'Preview failed.' });
    });
    return () => {
        if (!live) return;
        live = false;
        instance?.dispose();
        container.replaceChildren();
    };
}


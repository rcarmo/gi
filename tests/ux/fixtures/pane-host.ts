// Test-only bundle. Not imported by the production build or served by Gi.
import { html, render } from '../../../web/src/vendor/preact-htm.js';
import { WorkspaceTab } from '../../../web/src/gi-workspace-tab.js';
import { paneRegistry } from '../../../web/src/panes/pane-registry.js';
import type { PaneContext } from '../../../web/src/panes/pane-types.js';

const events: any[] = [], callbacks: (() => void)[] = [];
let root: HTMLDivElement;
let capabilities: any[] = ['readonly', 'preview'];
paneRegistry.register({
    id: 'conformance', label: 'Conformance', placement: 'tabs',
    get capabilities() { return capabilities; }, canHandle: () => 100,
    mount(container, context: PaneContext) {
        const occurrence = callbacks.length;
        events.push({type:'mount', occurrence, context});
        container.innerHTML = '<pre data-content></pre><button data-pane-close>Pane requests close</button>';
        container.querySelector('[data-content]')!.textContent = context.content ?? 'No text content';
        let closed = () => {};
        container.querySelector('[data-pane-close]')!.addEventListener('click', () => closed());
        return {
            getContent: () => undefined, isDirty: () => false,
            focus() { events.push({type:'focus', occurrence}); },
            resize() { events.push({type:'resize', occurrence, width:container.getBoundingClientRect().width}); },
            onClose(cb) { closed = cb; callbacks.push(cb); events.push({type:'bind-close', occurrence}); },
            dispose() { events.push({type:'dispose', occurrence}); container.replaceChildren(); },
        };
    },
});
(window as any).__paneHarness = {
    events,
    mount(path: string, requestedCapabilities = ['readonly', 'preview']) {
        if (root) render(null, root);
        root?.remove();
        root = document.createElement('div');
        root.dataset.conformance = '';
        root.style.cssText = 'position:fixed;z-index:10000;top:80px;left:20px;width:280px;height:220px;background:white;color:black';
        document.body.append(root);
        capabilities = requestedCapabilities;
        render(html`<${WorkspaceTab} path=${path} onClose=${() => {
            events.push({type:'closed'}); render(null, root);
        }} />`, root);
    },
    resize(width: number) { root.style.width = `${width}px`; },
    closeOld(index: number) { callbacks[index](); },
    stop() { render(null, root); },
};

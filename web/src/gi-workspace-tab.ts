import { html, useState, useLayoutEffect, useRef } from './vendor/preact-htm.js';
import { getWorkspaceFile } from './api.js';
import { paneRegistry } from './panes/pane-registry.js';

import { mountWorkspaceTab } from './gi-workspace-tab-lifecycle.js';

export function WorkspaceTab({ path, onClose }) {
    const host = useRef<HTMLElement>(null);
    const [state, setState] = useState({ loading: true, error: '' });
    const [attempt, setAttempt] = useState(0);
    useLayoutEffect(() => mountWorkspaceTab(host.current!, path, setState, getWorkspaceFile, paneRegistry), [path, attempt]);
    return html`<section class="gi-workspace-tab editor-pane" role="region" aria-label=${`Read-only preview: ${path}`}>
        <div class="gi-workspace-tab-toolbar"><span>Read-only preview</span>
            <button onClick=${() => setAttempt(value => value + 1)} disabled=${state.loading}>Refresh preview</button>
            <button onClick=${onClose}>Close preview</button></div>
        ${state.loading && html`<p role="status">Loading preview…</p>`}
        ${state.error && html`<p role="alert">${state.error} <button onClick=${() => setAttempt(value => value + 1)}>Retry preview</button></p>`}
        <div class="workspace-preview-body gi-workspace-tab-body" ref=${host}></div>
    </section>`;
}

import { html, useState, useEffect } from './vendor/preact-htm.js';

const loaders = {
    models: () => import('./gi-settings-models.js').then(module => module.Models),
    appearance: () => import('./gi-settings-appearance.js').then(module => module.Appearance),
    compaction: () => import('./gi-settings-compaction.js').then(module => module.GiSettingsCompaction),
    providers: () => import('./gi-settings-providers.js').then(module => module.GiSettingsProviders),
};
const labels = { models: 'Models', appearance: 'Appearance', compaction: 'Compaction', providers: 'Providers' };
const components = new Map<string, any>();
const pending = new Map<string, Promise<any>>();

function load(section: string) {
    if (components.has(section)) return Promise.resolve(components.get(section));
    if (pending.has(section)) return pending.get(section)!;
    const promise = loaders[section]().then(component => {
        components.set(section, component); pending.delete(section); return component;
    }, error => { pending.delete(section); throw error; });
    pending.set(section, promise);
    return promise;
}

// Mount keyed by section: late imports may populate the code cache but cannot
// render into another section. Never cache form drafts or native API snapshots.
export function LazySettingsPane({ section, chatJid, filter, onMutationStart, onMutationEnd, onApplied }) {
    const [component, setComponent] = useState(() => components.get(section) || null);
    const [error, setError] = useState(false);
    useEffect(() => {
        let live = true;
        if (!component) load(section).then(value => { if (live) setComponent(() => value); }).catch(() => { if (live) setError(true); });
        return () => { live = false; };
    }, []);
    if (error) return html`<div role="alert">Unable to load ${labels[section]}. Close Settings and try again. If the app was updated, save your work and reload the page.</div>`;
    if (!component) return html`<div role="status" class="settings-loading-pane">Loading ${labels[section]} pane…</div>`;
    return html`<${component} key=${chatJid} chatJid=${chatJid} filter=${filter} onMutationStart=${onMutationStart} onMutationEnd=${onMutationEnd} onApplied=${onApplied} />`;
}

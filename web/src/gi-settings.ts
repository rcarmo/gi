// Gi-owned adaptation of Piclaw 70d33bc93 settings-dialog.ts (MIT).
// Native capabilities only; no Piclaw settings service calls.
import { html, useState, useEffect, useRef } from './vendor/preact-htm.js';
import { BodyPortal } from './components/body-portal.js';
import { getGiSettingsSnapshot, getGiIdentity, saveGiIdentity, getAgentModels, selectAgentModel } from './api.js';
import { modelContextBlocked } from './gi-context-usage.js';
import { appearancePresets, currentAppearance, persistAppearance, subscribeAppearance } from './gi-appearance.js';
import { defaultAppearance } from './gi-appearance-state.js';
import { GiSettingsCompaction } from './gi-settings-compaction.js';
import { GiSettingsProviders } from './gi-settings-providers.js';

let generalCache: any = null;

function Identity() {
    const [snapshot, setSnapshot] = useState<any>(null);
    const [draft, setDraft] = useState({ assistant_name: '', user_name: '' });
    const [error, setError] = useState('');
    const [notice, setNotice] = useState('');
    const [busy, setBusy] = useState(false);
    const [attempt, setAttempt] = useState(0);
    const live = useRef(false);
    const saving = useRef(false);
    useEffect(() => { live.current = true; return () => { live.current = false; }; }, []);
    useEffect(() => {
        let active = true; setError(''); setNotice(''); setSnapshot(null);
        getGiIdentity().then(value => {
            if (active) { setSnapshot(value); setDraft({ assistant_name: value.saved.assistant_name, user_name: value.saved.user_name }); }
        }).catch(err => { if (active) setError(err.message); });
        return () => { active = false; };
    }, [attempt]);
    async function save() {
        if (!snapshot || saving.current) return;
        setError(''); setNotice('');
        if ([draft.assistant_name, draft.user_name].some(name => !name.trim() || [...name].length > 128 || /[\u0000-\u001f\u007f-\u009f]/.test(name))) {
            setError('Names must contain 1–128 characters without control characters.'); return;
        }
        saving.current = true; setBusy(true);
        try {
            const value = await saveGiIdentity({ ...draft, revision: snapshot.saved.revision });
            if (live.current) {
                setSnapshot(value); setDraft({ assistant_name: value.saved.assistant_name, user_name: value.saved.user_name });
                setNotice(value.restart_required ? 'Names saved. Restart Gi manually to activate them.' : 'Names saved. Active names already match.');
            }
        } catch (err) { if (live.current) setError(err.message); }
        finally { saving.current = false; if (live.current) setBusy(false); }
    }
    return html`<section aria-label="Saved display names">
        <h3>Saved display names</h3>
        <p>Instance-wide · saved to .piclaw/config.json. Active names above stay unchanged until you restart Gi manually.</p>
        ${!snapshot && !error && html`<p role="status">Loading saved names…</p>`}
        ${snapshot && html`<label>Assistant display name<input aria-label="Assistant display name" type="text" value=${draft.assistant_name} disabled=${busy} onInput=${e => { setDraft(d => ({ ...d, assistant_name: e.target.value })); setNotice(''); }} /></label>
            <label>User display name<input aria-label="User display name" type="text" value=${draft.user_name} disabled=${busy} onInput=${e => { setDraft(d => ({ ...d, user_name: e.target.value })); setNotice(''); }} /></label>
            ${snapshot.restart_required && html`<p data-testid="identity-restart-required">Restart required to activate the saved names.</p>`}
            <button disabled=${busy} onClick=${save}>${busy ? 'Saving names…' : 'Save names'}</button>`}
        <button disabled=${busy} onClick=${() => setAttempt(n => n + 1)}>Reload saved names</button>
        ${error && html`<p role="alert">${error}</p>`}
        ${notice && html`<p role="status">${notice}</p>`}
    </section>`;
}

function General() {
    const [data, setData] = useState(generalCache);
    const [error, setError] = useState('');
    const [attempt, setAttempt] = useState(0);
    useEffect(() => {
        let live = true;
        setError('');
        getGiSettingsSnapshot().then(snapshot => {
            if (live) { generalCache = snapshot; setData(snapshot); }
        }).catch(error => { if (live) setError(error.message); });
        return () => { live = false; };
    }, [attempt]);
    return html`<section aria-labelledby="gi-general-title">
        <h2 id="gi-general-title">General</h2>
        <p>Active instance settings · read-only</p>
        <p>Loaded at startup from <code>.piclaw/config.json</code> and <code>.pi/settings.json</code>. Edit the files and restart Gi to change these defaults.</p>
        ${error && html`<div role="alert">${error} <button onClick=${() => setAttempt(x => x + 1)}>Retry</button></div>`}
        ${!data && !error && html`<p role="status">Loading settings…</p>`}
        ${data && html`<dl class="gi-settings-values">
            <dt>Assistant</dt><dd>${data.assistant_name}</dd>
            <dt>User</dt><dd>${data.user_name}</dd>
            <dt>Workspace</dt><dd>${data.workspace_root}</dd>
            <dt>Default model</dt><dd>${data.current || data.default_model}</dd>
            <dt>Default thinking</dt><dd>${data.default_thinking_level || 'Unknown'}</dd>
            <dt>Build</dt><dd>${data.version || 'Unknown'}</dd>
        </dl>`}
        <${Identity} />
    </section>`;
}

function Models({ chatJid, onMutationStart, onMutationEnd, onApplied }) {
    const [data, setData] = useState<any>(null);
    const [chosen, setChosen] = useState('');
    const [filter, setFilter] = useState('');
    const [error, setError] = useState('');
    const [notice, setNotice] = useState('');
    const [busy, setBusy] = useState(false);
    const [attempt, setAttempt] = useState(0);
    const mounted = useRef(false);
    const saving = useRef(false);
    const searchRef = useRef<HTMLInputElement>(null);
    useEffect(() => {
        mounted.current = true;
        searchRef.current?.focus();
        return () => { mounted.current = false; };
    }, []);
    useEffect(() => {
        let live = true;
        setData(null); setError(''); setNotice('');
        getAgentModels(chatJid).then(snapshot => {
            if (live) { setData(snapshot); setChosen(snapshot.current); }
        }).catch(error => { if (live) setError(error.message); });
        return () => { live = false; };
    }, [chatJid, attempt]);
    const options = data?.model_options || data?.models || [];
    const matching = options.filter(option => `${option.label || option.id} ${option.provider || ''}`.toLowerCase().includes(filter.trim().toLowerCase()));
    const selected = options.find(option => (option.label || option.id) === chosen);
    const blocked = modelContextBlocked({ contextWindow: selected?.context_window ?? selected?.contextWindow }, data?.context_usage);
    async function apply() {
        if (saving.current || !data || !selected || blocked || chosen === data.current) return;
        saving.current = true; setBusy(true); setError(''); setNotice('');
        const token = onMutationStart();
        try {
            const result = await selectAgentModel(chatJid, chosen);
            if (mounted.current) {
                setData(previous => ({ ...previous, ...result })); setChosen(result.current);
                setNotice('Model applied to this session.'); onApplied(result, token);
            }
        } catch (error) {
            if (mounted.current) setError(error.message);
        } finally {
            onMutationEnd(token); saving.current = false;
            if (mounted.current) setBusy(false);
        }
    }
    return html`<section aria-labelledby="gi-models-title">
        <h2 id="gi-models-title">Models</h2>
        <p>Session settings · <code>${chatJid}</code></p>
        <p>Changes affect this session only. Instance defaults and other sessions are unchanged.</p>
        <label>Filter models<input ref=${searchRef} type="search" aria-label="Filter models" disabled=${busy} value=${filter} onInput=${e => { setFilter(e.target.value); setChosen(''); setNotice(''); }} /></label>
        ${!data && !error && html`<p role="status">Loading models…</p>`}
        ${error && html`<div role="alert">${error}${!data && html` <button onClick=${() => setAttempt(x => x + 1)}>Retry</button>`}</div>`}
        ${data && html`
            <dl class="gi-settings-values"><dt>Current model</dt><dd data-testid="settings-current-model">${data.current}</dd>
            <dt>Thinking (read-only)</dt><dd>${data.thinking_level || 'Unknown'}</dd>
            <dt>Context capacity</dt><dd data-testid="settings-context-capacity">${Number.isFinite(data.context_window) && data.context_window > 0 ? data.context_window : 'Unknown'}</dd></dl>
            <label>Session model<select aria-label="Session model" value=${chosen} disabled=${busy} onChange=${e => { setChosen(e.target.value); setNotice(''); }}>
                <option value="" disabled>Choose a model</option>
                ${matching.slice(0, 50).map(option => html`<option value=${option.label || option.id}>${option.label || option.id}</option>`)}
            </select></label>
            ${matching.length > 50 && html`<p>Showing 50 of ${matching.length} models. Refine the filter.</p>`}
            ${matching.length === 0 && html`<p>No matching models.</p>`}
            ${blocked && html`<p role="status">This model cannot fit the measured context. Compact the session before changing models.</p>`}
            <button disabled=${busy || !selected || blocked || chosen === data.current} onClick=${apply}>${busy ? 'Applying…' : 'Apply model'}</button>
            ${notice && html`<p role="status">${notice}</p>`}
        `}
    </section>`;
}

function Appearance() {
    const [draft, setDraft] = useState(currentAppearance);
    const [error, setError] = useState('');
    const [notice, setNotice] = useState('');
    const dirty = useRef(false);
    const ownSave = useRef(false);
    useEffect(() => subscribeAppearance(value => {
        if (ownSave.current) return;
        if (dirty.current) setNotice('Appearance changed in another tab. Your unsaved fields are unchanged. Save to overwrite or reset to defaults.');
        else { setDraft(value); setNotice('Appearance updated from another tab.'); }
    }), []);
    const update = patch => {
        dirty.current = true; setDraft(previous => ({ ...previous, ...patch })); setError(''); setNotice('');
    };
    const save = value => {
        setError(''); setNotice(''); ownSave.current = true;
        try {
            const saved = persistAppearance(value);
            dirty.current = false; setDraft(saved); setNotice('Appearance saved in this browser.');
        } catch (error) {
            setError(`Appearance was not saved: ${error.message}`);
        } finally { ownSave.current = false; }
    };
    return html`<section aria-labelledby="gi-appearance-title">
        <h2 id="gi-appearance-title">Appearance</h2>
        <p>Browser settings · this origin, across all sessions</p>
        <p>Only this browser profile changes. Server configuration, other devices and the terminal theme are unchanged. Default follows your system colour mode.</p>
        <label>Theme preset<select aria-label="Theme preset" value=${draft.theme} onChange=${e => update({ theme: e.target.value, tint: '' })}>
            ${appearancePresets.map(theme => html`<option value=${theme}>${theme}</option>`)}
        </select></label>
        <label>Custom tint<input aria-label="Custom tint" type="text" placeholder="#RRGGBB" maxLength="7" disabled=${draft.theme !== 'default'} value=${draft.tint} onInput=${e => update({ tint: e.target.value })} /></label>
        <p>Default theme only. Use #RGB or #RRGGBB, or leave empty for no tint.</p>
        <button onClick=${() => save(draft)}>Save appearance</button>
        <button onClick=${() => save(defaultAppearance)}>Reset appearance</button>
        ${error && html`<p role="alert">${error}</p>`}
        ${notice && html`<p role="status">${notice}</p>`}
    </section>`;
}

function Dialog({ chatJid, onClose, onMutationStart, onMutationEnd, onApplied }) {
    const [section, setSection] = useState('general');
    const dialog = useRef<HTMLDivElement>(null);
    useEffect(() => {
        const app = document.getElementById('app');
        const previousInert = app?.inert;
        if (app) app.inert = true;
        const previousOverflow = document.body.style.overflow;
        document.body.style.overflow = 'hidden';
        dialog.current?.querySelector<HTMLButtonElement>('.settings-dialog-close')?.focus();
        const key = (event: KeyboardEvent) => {
            if (event.key === 'Escape') {
                event.preventDefault(); event.stopImmediatePropagation(); onClose(); return;
            }
            if (event.key === 'Tab') {
                const nodes = [...(dialog.current?.querySelectorAll<HTMLElement>('button:not(:disabled), input:not(:disabled), select:not(:disabled), [tabindex="0"]') || [])].filter(node => node.getClientRects().length);
                const first = nodes[0], last = nodes.at(-1);
                if (event.shiftKey && (document.activeElement === first || !dialog.current?.contains(document.activeElement))) {
                    event.preventDefault(); last?.focus();
                } else if (!event.shiftKey && (document.activeElement === last || !dialog.current?.contains(document.activeElement))) {
                    event.preventDefault(); first?.focus();
                }
            }
            // Background popups register document-capture keys; inert alone does
            // not disable those listeners. Native input/select defaults still run.
            event.stopPropagation();
        };
        // Capture Escape even if a pending write disabled the focused button.
        window.addEventListener('keydown', key, true);
        return () => {
            window.removeEventListener('keydown', key, true);
            if (app) app.inert = previousInert || false;
            document.body.style.overflow = previousOverflow;
        };
    }, []);
    return html`<div class="settings-dialog-backdrop" onClick=${e => { if (e.target === e.currentTarget) onClose(); }}>
        <div ref=${dialog} class="settings-dialog" role="dialog" aria-modal="true" aria-labelledby="gi-settings-title" onKeyDown=${e => e.stopPropagation()}>
            <header class="settings-dialog-header"><span class="settings-dialog-title" id="gi-settings-title">Gi Settings</span>
                <button class="settings-dialog-close" aria-label="Close settings" onClick=${onClose}>✕</button></header>
            <div class="settings-dialog-body"><nav class="settings-nav" aria-label="Settings sections">
                ${['general', 'models', 'appearance', 'compaction', 'providers'].map(id => html`<button class=${`settings-nav-item ${section === id ? 'active' : ''}`} aria-current=${section === id ? 'page' : undefined} onClick=${() => setSection(id)}>${{ general: 'General', models: 'Models', appearance: 'Appearance', compaction: 'Compaction', providers: 'Providers' }[id]}</button>`)}
            </nav><main class="settings-content">
                ${section === 'general' ? html`<${General} />` : section === 'appearance' ? html`<${Appearance} />` : section === 'providers' ? html`<${GiSettingsProviders} />` : section === 'compaction' ? html`<${GiSettingsCompaction} key=${chatJid} chatJid=${chatJid} />` : html`<${Models} key=${chatJid} chatJid=${chatJid} onMutationStart=${onMutationStart} onMutationEnd=${onMutationEnd} onApplied=${onApplied} />`}
            </main></div>
        </div>
    </div>`;
}

export function GiSettings({ chatJid, onMutationStart, onMutationEnd, onApplied }) {
    const [open, setOpen] = useState(false);
    const opener = useRef<HTMLElement>(null);
    const isOpen = useRef(false);
    const close = () => {
        isOpen.current = false; setOpen(false);
        requestAnimationFrame(() => {
            if (isOpen.current) return;
            const original = opener.current;
            const target = original?.isConnected && original !== document.body && !original.closest('[inert]')
                ? original : document.querySelector<HTMLElement>('.compose-box textarea');
            target?.focus({ preventScroll: true });
        });
    };
    useEffect(() => {
        const show = () => {
            if (!isOpen.current) opener.current = document.activeElement as HTMLElement;
            isOpen.current = true; setOpen(true);
        };
        const shortcut = (event: KeyboardEvent) => {
            if (event.key === ',' && (event.ctrlKey || event.metaKey || event.altKey)) {
                event.preventDefault(); event.stopImmediatePropagation(); show();
            }
        };
        window.addEventListener('piclaw:open-settings', show);
        window.addEventListener('keydown', shortcut, true);
        return () => { window.removeEventListener('piclaw:open-settings', show); window.removeEventListener('keydown', shortcut, true); };
    }, []);
    return open && html`<${BodyPortal} className="settings-portal"><${Dialog} chatJid=${chatJid} onClose=${close} onMutationStart=${onMutationStart} onMutationEnd=${onMutationEnd} onApplied=${onApplied} /><//>`;
}

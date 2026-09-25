// Gi-owned adaptation of Piclaw 70d33bc93 settings-dialog.ts (MIT).
// Native capabilities only; no Piclaw settings service calls.
import { html, useState, useEffect, useLayoutEffect, useRef, useMemo } from './vendor/preact-htm.js';
import { BodyPortal } from './components/body-portal.js';
import { getGiSettingsSnapshot, getGiIdentity, saveGiIdentity } from './api.js';
import { LazySettingsPane } from './gi-settings-lazy.js';

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

function Dialog({ chatJid, initialSection = 'general', onClose, onMutationStart, onMutationEnd, onApplied }) {
    const [section, setSection] = useState(initialSection);
    const dialog = useRef<HTMLDivElement>(null);
    const filterRef = useRef<HTMLInputElement>(null);
    const [filter, setFilter] = useState('');
    const [busyScope, setBusyScope] = useState<object | null>(null);
    const searchScope = useMemo(() => ({}), [section, chatJid]);
    const [layoutMode, setLayoutMode] = useState({ compact: false, narrow: false });
    // Match the pinned shell's element-width thresholds (not viewport guesses).
    useLayoutEffect(() => {
        const element = dialog.current;
        if (!element) return;
        const update = () => {
            const width = element.clientWidth || 0;
            setLayoutMode(previous => {
                const next = { compact: width > 0 && width <= 860, narrow: width > 0 && width <= 720 };
                return previous.compact === next.compact && previous.narrow === next.narrow ? previous : next;
            });
        };
        update();
        if (typeof ResizeObserver === 'function') {
            const observer = new ResizeObserver(update); observer.observe(element);
            return () => observer.disconnect();
        }
        window.addEventListener('resize', update);
        return () => window.removeEventListener('resize', update);
    }, []);
    useLayoutEffect(() => {
        setFilter('');
        if (section === 'models') filterRef.current?.focus();
    }, [section, chatJid]);
    // Make the first painted shell modal, including focus and Escape handling.
    useLayoutEffect(() => {
        const app = document.getElementById('app');
        const previousInert = app?.inert;
        if (app) app.inert = true;
        const previousOverflow = document.body.style.overflow;
        document.body.style.overflow = 'hidden';
        dialog.current?.querySelector<HTMLButtonElement>('.settings-dialog-close')?.focus();
        const key = (event: KeyboardEvent) => {
            if (event.isComposing) return;
            if (event.key === 'Escape') {
                event.preventDefault(); event.stopImmediatePropagation();
                const auth = dialog.current?.querySelector('[data-auth-escape="true"]');
                if (auth) auth.dispatchEvent(new Event('gi-auth-escape')); else onClose();
                return;
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
            // Trapped keys stay modal; ordinary keys reach the target and its
            // handlers. Background popup listeners independently stand down.
            if (event.defaultPrevented) event.stopImmediatePropagation();
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
        <div ref=${dialog} class=${`settings-dialog${layoutMode.compact ? ' settings-dialog-compact' : ''}${layoutMode.narrow ? ' settings-dialog-narrow' : ''}`} role="dialog" aria-modal="true" aria-labelledby="gi-settings-title" onKeyDown=${e => e.stopPropagation()}>
            <header class="settings-dialog-header"><span class="settings-dialog-title" id="gi-settings-title">Gi Settings</span>
                ${section === 'models' && html`<input ref=${filterRef} type="search" class="settings-header-filter" aria-label="Filter models" placeholder="Filter models…" value=${filter} disabled=${busyScope === searchScope} onInput=${e => setFilter(e.target.value)} />`}
                <button class="settings-dialog-close" aria-label="Close settings" onClick=${onClose}>✕</button></header>
            <div class="settings-dialog-body"><nav class="settings-nav" aria-label="Settings sections">
                ${['general', 'models', 'appearance', 'compaction', 'providers', 'authentication'].map(id => html`<button class=${`settings-nav-item ${section === id ? 'active' : ''}`} aria-current=${section === id ? 'page' : undefined} onClick=${() => setSection(id)}>${{ general: 'General', models: 'Models', appearance: 'Appearance', compaction: 'Compaction', providers: 'Providers', authentication: 'Authentication' }[id]}</button>`)}
            </nav><main class="settings-content">
                ${section === 'general' ? html`<${General} />` : html`<${LazySettingsPane} key=${section} section=${section} chatJid=${chatJid} filter=${filter} onMutationStart=${() => { setBusyScope(searchScope); return onMutationStart(); }} onMutationEnd=${token => { setBusyScope(previous => previous === searchScope ? null : previous); onMutationEnd(token); }} onApplied=${onApplied} />`}
            </main></div>
        </div>
    </div>`;
}

export function GiSettings({ chatJid, onMutationStart, onMutationEnd, onApplied }) {
    const [open, setOpen] = useState(false);
    const [initialSection, setInitialSection] = useState('general');
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
    // Install opening controls before first paint, including the auth gate's
    // asynchronous application mount. A visible composer must be interactive.
    useLayoutEffect(() => {
        const show = (event?: Event) => {
            if (isOpen.current) return;
            const detail = event instanceof CustomEvent ? event.detail : null;
            const requested = detail?.section;
            setInitialSection(['general','models','appearance','compaction','providers','authentication'].includes(requested) ? requested : 'general');
            opener.current = detail?.opener instanceof HTMLElement && detail.opener.isConnected ? detail.opener : document.activeElement as HTMLElement;
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
    return open && html`<${BodyPortal} className="settings-portal"><${Dialog} chatJid=${chatJid} initialSection=${initialSection} onClose=${close} onMutationStart=${onMutationStart} onMutationEnd=${onMutationEnd} onApplied=${onApplied} /><//>`;
}

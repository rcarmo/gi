// Gi-owned lazy Models pane; native data is fetched on each mount.
import { html, useState, useEffect, useLayoutEffect, useRef } from "./vendor/preact-htm.js";
import { getAgentModels, selectAgentModel } from "./api.js";
import { modelContextBlocked } from "./gi-context-usage.js";

export function Models({ chatJid, filter = '', onMutationStart, onMutationEnd, onApplied }) {
    const [data, setData] = useState<any>(null);
    const [chosen, setChosen] = useState('');
    const [error, setError] = useState('');
    const [notice, setNotice] = useState('');
    const [busy, setBusy] = useState(false);
    const [attempt, setAttempt] = useState(0);
    const mounted = useRef(false);
    const saving = useRef(false);
    const previousFilter = useRef(filter);
    useEffect(() => {
        mounted.current = true;
        return () => { mounted.current = false; };
    }, []);
    useLayoutEffect(() => {
        if (previousFilter.current !== filter) {
            previousFilter.current = filter; setChosen(''); setNotice('');
        }
    }, [filter]);
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


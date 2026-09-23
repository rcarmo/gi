import { html, useState, useEffect, useRef } from './vendor/preact-htm.js';
import { getGiCompactionPolicy, saveGiCompactionPolicy } from './api.js';

const fields = [
    ['context_window', 'Saved context window'], ['reserve_tokens', 'Saved reserved tokens'],
    ['keep_recent_tokens', 'Saved keep recent tokens'], ['threshold_tokens', 'Saved trigger threshold'],
];
const toDraft = policy => ({ enabled: policy.enabled, ...Object.fromEntries(fields.map(([key]) => [key, String(policy[key])])) });

export function GiSettingsCompactionPolicy() {
    const [snapshot, setSnapshot] = useState<any>(null);
    const [draft, setDraft] = useState<any>({});
    const [error, setError] = useState('');
    const [notice, setNotice] = useState('');
    const [busy, setBusy] = useState(false);
    const [attempt, setAttempt] = useState(0);
    const alive = useRef(false), saving = useRef(false);
    useEffect(() => { alive.current = true; return () => { alive.current = false; }; }, []);
    useEffect(() => {
        let live = true; setSnapshot(null); setError(''); setNotice('');
        getGiCompactionPolicy().then(value => { if (live) { setSnapshot(value); setDraft(toDraft(value.saved.policy)); } })
            .catch(err => { if (live) setError(err.message); });
        return () => { live = false; };
    }, [attempt]);
    async function save() {
        if (!snapshot || saving.current) return;
        setError(''); setNotice('');
        const numeric = Object.fromEntries(fields.map(([key]) => [key, Number(draft[key])]));
        const { context_window: window, reserve_tokens: reserve, keep_recent_tokens: keep, threshold_tokens: threshold } = numeric;
        if (fields.some(([key]) => !/^\d+$/.test(draft[key])) || Object.values(numeric).some(n => !Number.isSafeInteger(n) || n < 1)
            || window > 16777216 || reserve >= window || keep > threshold || threshold > window - reserve) {
            setError('Use positive whole-number budgets: context at most 16777216; reserve below context; keep recent ≤ threshold ≤ context minus reserve.'); return;
        }
        saving.current = true; setBusy(true);
        try {
            const result = await saveGiCompactionPolicy({ revision: snapshot.saved.revision, enabled: draft.enabled, ...numeric });
            if (alive.current) { setSnapshot(result); setDraft(toDraft(result.saved.policy)); setNotice(result.restart_required ? 'Policy saved. Restart Gi manually to activate it.' : 'Policy saved. Active policy already matches.'); }
        } catch (err) { if (alive.current) setError(err.message); }
        finally { saving.current = false; if (alive.current) setBusy(false); }
    }
    return html`<section aria-label="Saved automatic policy">
        <h3>Saved automatic policy</h3>
        <p>Instance-wide · .pi/settings.json. Saving changes the next startup only; manual actions still use the active engine policy above.</p>
        ${!snapshot && !error && html`<p role="status">Loading saved policy…</p>`}
        ${snapshot && html`
            <label><input class="gi-policy-checkbox" type="checkbox" aria-label="Saved automatic compaction" checked=${draft.enabled} disabled=${busy} onChange=${e => { setDraft(d => ({ ...d, enabled: e.target.checked })); setNotice(''); }} />Automatic compaction after restart</label>
            ${fields.map(([key, label]) => html`<label>${label}<input type="number" min="1" max="16777216" step="1" aria-label=${label} disabled=${busy} value=${draft[key]} onInput=${e => { setDraft(d => ({ ...d, [key]: e.target.value })); setNotice(''); }} /></label>`)}
            <p>Strategy label is preserved: ${snapshot.saved.policy.strategy || 'default'}. No provider model or remote compaction setting is changed.</p>
            ${snapshot.restart_required && html`<p data-testid="compaction-policy-restart">Restart required to activate the saved policy.</p>`}
            <button disabled=${busy} onClick=${save}>${busy ? 'Saving policy…' : 'Save policy'}</button>
        `}
        <button disabled=${busy} onClick=${() => setAttempt(x => x + 1)}>Reload saved policy</button>
        ${error && html`<p role="alert">${error}</p>`}
        ${notice && html`<p role="status">${notice}</p>`}
    </section>`;
}

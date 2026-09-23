// Gi-owned session actions, distinct from startup-only policy editing.
import { html, useState, useEffect, useRef } from './vendor/preact-htm.js';
import { getSessionCompaction, getAgentStatus, compactSession, cancelSessionRun } from './api.js';
import { compactionNotice, compactionElapsed } from './gi-compaction-state.js';

export function GiSettingsCompaction({ chatJid }) {
    const [snapshot, setSnapshot] = useState<any>(null);
    const [error, setError] = useState('');
    const [notice, setNotice] = useState('');
    const [busy, setBusy] = useState(false);
    const [loading, setLoading] = useState(true);
    const [fresh, setFresh] = useState(false);
    const [now, setNow] = useState(Date.now());
    const alive = useRef(false);
    const reading = useRef(false);
    const mutating = useRef(false);
    const revision = useRef(0);
    const readError = useRef(false);
    const [acceptedRun, setAcceptedRun] = useState('');

    async function refresh() {
        if (reading.current || !alive.current) return;
        const version = revision.current;
        reading.current = true;
        try {
            const [capability, activity] = await Promise.all([getSessionCompaction(chatJid), getAgentStatus('', chatJid)]);
            if (alive.current && version === revision.current) {
                setSnapshot({ capability, activity }); setFresh(!mutating.current);
                if (readError.current) { readError.current = false; setError(''); }
            }
        } catch (err) {
            if (alive.current && version === revision.current) { readError.current = true; setError(`Refresh failed: ${err.message}`); setFresh(false); }
        } finally { reading.current = false; if (alive.current) setLoading(false); }
    }
    useEffect(() => {
        alive.current = true;
        void refresh();
        const timer = setInterval(() => { setNow(Date.now()); void refresh(); }, 1000);
        return () => { alive.current = false; ++revision.current; clearInterval(timer); };
    }, []);

    const capability = snapshot?.capability;
    const activity = snapshot?.activity;
    const policy = capability?.policy;
    const occurrence = activity?.compaction?.turn_id === activity?.turn_id ? activity?.compaction : null;
    const noticeState = compactionNotice(activity, now);
    const progress = occurrence ? (occurrence.active && noticeState ? {
        active: true, title: noticeState.title, elapsed: compactionElapsed(noticeState, now),
    } : {
        active: false,
        label: ({ 'compaction.completed': 'Context compacted', 'compaction.cancelled': 'Compaction cancelled', 'compaction.failed': 'Compaction failed', 'compaction.suppressed': 'Compaction temporarily suppressed' })[occurrence.event_type] || 'Compaction inactive',
        detail: occurrence.detail || '',
    }) : null;
    const active = fresh && progress?.active && activity?.turn_id && ['running', 'cancelling'].includes(activity.status);
    const available = fresh && capability?.available && activity?.status === 'idle' && !busy;

    async function act(kind: 'compact' | 'stop') {
        if (mutating.current || !alive.current || !fresh) return;
        if (kind === 'compact' ? !available : !active || activity.status === 'cancelling') return;
        const token = capability.token, turn = activity.turn_id;
        mutating.current = true; ++revision.current; readError.current = false; setBusy(true); setFresh(false); setError(''); setNotice('');
        try {
            if (kind === 'compact') {
                const result = await compactSession(chatJid, token);
                if (alive.current) { setAcceptedRun(result.turn_id); setNotice(`Compaction accepted for ${result.turn_id}; waiting for authoritative progress.`); }
            } else {
                await cancelSessionRun(chatJid, turn);
                if (alive.current) { setAcceptedRun(turn); setNotice(`Cancellation requested for ${turn}; waiting for authoritative state.`); }
            }
        } catch (err) {
            if (alive.current) { readError.current = false; setError(`${kind === 'compact' ? 'Compact' : 'Stop'} failed: ${err.message}`); }
        } finally {
            mutating.current = false; ++revision.current;
            if (alive.current) { setBusy(false); void refresh(); }
        }
    }

    return html`<section aria-labelledby="gi-compaction-title">
        <h2 id="gi-compaction-title">Compaction</h2>
        <p>Session actions · <code>${chatJid}</code></p>
        <p>Automatic policy is read-only and loaded at startup from .pi/settings.json. Edit the file and restart Gi to change it.</p>
        ${policy && html`<dl class="gi-settings-values" data-testid="compaction-policy">
            <dt>Automatic compaction</dt><dd>${policy.enabled ? 'Enabled' : 'Disabled'}</dd>
            <dt>Context window</dt><dd>${policy.context_window}</dd>
            <dt>Trigger threshold</dt><dd>${policy.threshold_tokens}</dd>
            <dt>Reserved tokens</dt><dd>${policy.reserve_tokens}</dd>
            <dt>Keep recent tokens</dt><dd>${policy.keep_recent_tokens}</dd>
            <dt>Strategy label</dt><dd>${policy.strategy || 'Unspecified'}</dd>
        </dl>`}
        <p>Compact now runs Gi's local context compactor (including configured hooks). It keeps timeline messages and does not submit your draft. Stop turn cancels the displayed turn, including any response after automatic compaction.</p>
        ${loading && html`<p role="status">Loading compaction…</p>`}
        ${snapshot && html`<p data-testid="compaction-capability">${!fresh ? 'State unavailable; refresh before acting.' : capability.available ? 'Manual compaction available.' : capability.reason || 'Manual compaction unavailable.'}</p>`}
        ${progress && html`<p data-testid="settings-compaction-progress">${progress.active ? `${progress.title} — ${progress.elapsed}` : progress.label}${activity?.compaction?.turn_id ? ` · ${activity.compaction.turn_id}` : ''}${progress.detail ? ` · ${progress.detail}` : ''}</p>`}
        ${acceptedRun && html`<p>Action turn: <code>${acceptedRun}</code></p>`}
        ${notice && html`<p role="status">${acceptedRun && fresh && activity?.compaction?.turn_id === acceptedRun && progress && !progress.active ? `Authoritative state: ${progress.label}.` : notice}</p>`}
        ${error && html`<p role="alert">${error}</p>`}
        <button disabled=${busy || reading.current} onClick=${() => { readError.current = false; setError(''); setNotice(''); setFresh(false); void refresh(); }}>Refresh compaction</button>
        <button disabled=${!available} onClick=${() => act('compact')}>${busy ? 'Working…' : 'Compact now'}</button>
        ${active && html`<button disabled=${busy || activity.status === 'cancelling'} onClick=${() => act('stop')}>Stop turn</button>`}
    </section>`;
}

import { html, useEffect, useLayoutEffect, useRef, useState } from './vendor/preact-htm.js';
import { authJSON, passkeyUnavailable, runPasskey } from './gi-passkeys.js';
import { parseAuthPolicy } from './gi-auth-policy.js';

export function GiSettingsAuthentication() {
    const [policy, setPolicy] = useState(null);
    const [proof, setProof] = useState(null);
    const [keys, setKeys] = useState(null);
    const [fresh, setFresh] = useState(false);
    const [busy, setBusy] = useState('');
    const [error, setError] = useState('');
    const [notice, setNotice] = useState('');
    const [name, setName] = useState('');
    const [code, setCode] = useState('');
    const [editing, setEditing] = useState(null);
    const [removing, setRemoving] = useState(null);
    const [now, setNow] = useState(Date.now());
    useEffect(() => { const timer = setInterval(() => setNow(Date.now()), 1000); return () => clearInterval(timer); }, []);
    const live = useRef(true), flight = useRef<AbortController | null>(null), root = useRef(null), opener = useRef<HTMLElement | null>(null);
    const restore = () => requestAnimationFrame(() => { if (live.current) (opener.current?.isConnected ? opener.current : root.current?.querySelector('button'))?.focus(); });
    const cancel = () => {
        if (flight.current) { flight.current.abort(); return; }
        setEditing(null); setRemoving(null); restore();
    };
    useLayoutEffect(() => {
        const el = root.current; el.addEventListener('gi-auth-escape', cancel);
        return () => el.removeEventListener('gi-auth-escape', cancel);
    });
    const refresh = async (signal?: AbortSignal) => {
        setFresh(false);
        const p = parseAuthPolicy(await authJSON('/api/auth/status', undefined, signal));
        if (!live.current) return;
        setPolicy(p);
        if (!p.enrolled || !p.passkeys_enabled) { setProof(null); setKeys(null); return; }
        const [nextProof, list] = await Promise.all([authJSON('/api/auth/session/proof', undefined, signal), authJSON('/api/auth/passkeys', undefined, signal)]);
        if (!Array.isArray(list.passkeys) || typeof nextProof.reauth_required !== 'boolean') throw new Error('Invalid authentication response');
        if (live.current) { setProof(nextProof); setKeys(list.passkeys); setFresh(true); }
    };
    const work = async (label: string, action: (signal: AbortSignal) => Promise<void>, message = '') => {
        if (flight.current) return;
        const controller = new AbortController(); flight.current = controller;
        const trigger = document.activeElement as HTMLElement;
        setBusy(label); setError(''); setNotice('');
        try { await action(controller.signal); if (live.current) { await refresh(controller.signal); if (message && live.current) setNotice(message); } }
        catch (e) { if (live.current) { setFresh(false); setError(e.message || 'Request failed. Refresh before retrying.'); } }
        finally { if (flight.current === controller) { flight.current = null; if (live.current) { setBusy(''); requestAnimationFrame(() => {
            if (!live.current || !root.current?.closest('.settings-dialog')) return;
            const target = trigger?.isConnected && !trigger.hasAttribute('disabled') ? trigger : root.current?.querySelector('button:not(:disabled)');
            if (target && (root.current.contains(document.activeElement) || document.activeElement === document.body)) target.focus({preventScroll:true});
        }); } } }
    };
    useEffect(() => { void work('Loading passkeys…', async () => {}); return () => { live.current = false; flight.current?.abort(); flight.current = null; }; }, []);
    const recentlyVerified = fresh && proof && !proof.reauth_required && Date.parse(proof.fresh_until) > now;
    const unavailable = !policy?.enrolled ? 'Authentication must be configured first.'
        : !policy.passkeys_enabled ? 'Passkeys are disabled by policy or are not configured for this origin.' : passkeyUnavailable();
    const add = () => work('Waiting for passkey creation…', async signal => {
        await runPasskey('register', signal, name); if (live.current) setName('');
    }, 'Passkey registered.');
    const mutation = (kind: 'rename' | 'remove', row, newName?: string) => work(kind === 'rename' ? 'Saving name…' : 'Removing passkey…', async signal => {
        const result = await authJSON(`/api/auth/passkeys/${kind}`, { id: row.id, ...(kind === 'rename' ? { name: newName } : {}) }, signal);
        if (result.ok !== true) throw new Error('Change could not be confirmed. Refresh before trying again.');
        if (live.current) { setEditing(null); setRemoving(null); restore(); }
    }, kind === 'rename' ? 'Name saved.' : 'Passkey removed. Existing login sessions are not signed out.');
    const date = value => !value || value.startsWith('0001-') ? 'Never used' : new Date(value).toLocaleString();
    return html`<section ref=${root} class="gi-authentication-pane" aria-labelledby="gi-authentication-heading" data-auth-escape=${busy || editing || removing ? 'true' : undefined}>
        <h2 id="gi-authentication-heading">Authentication</h2>
        <p>Manage passkeys for this instance owner. Each row is one credential, not an inventory of devices.</p>
        <h3>Passkeys</h3>
        ${busy && html`<p role="status">${busy}</p>`}
        ${error && html`<p role="alert">${error}</p>`}
        ${notice && html`<p role="status">${notice}</p>`}
        <button disabled=${!!busy} onClick=${() => work('Refreshing passkeys…', async () => {})}>Refresh passkeys</button>
        ${busy && html`<button onClick=${cancel}>Cancel pending operation</button>`}
        ${policy && unavailable && html`<p>${unavailable}</p>`}
        ${keys && !fresh && html`<p role="status">The displayed list is the last confirmed snapshot. Refresh before making changes.</p>`}
        ${policy?.enrolled && policy?.passkeys_enabled && html`<div class="gi-passkey-proof">
            <p>${recentlyVerified ? 'Recently authenticated for credential changes.' : 'Verify an accepted factor before changing passkeys.'}</p>
            ${policy.totp_login_available && html`<label>Authentication code<input aria-label="Reauthentication code" type="text" inputMode="numeric" autoComplete="one-time-code" value=${code} disabled=${!!busy} onInput=${e => setCode(e.target.value)} /></label>
                <button disabled=${!!busy || !/^\d{6}$/.test(code)} onClick=${() => work('Verifying authentication…', async signal => { await authJSON('/api/auth/session/reauth/totp', {code}, signal); if (live.current) setCode(''); }, 'Authentication verified.')}>Verify code</button>`}
            ${policy.passkey_login_available && html`<button disabled=${!!busy || !!passkeyUnavailable()} onClick=${() => work('Waiting for passkey verification…', signal => runPasskey('reauth', signal), 'Authentication verified.')}>Verify with passkey</button>`}
        </div>`}
        <div class="gi-passkey-add"><label>Passkey name<input aria-label="New passkey name" value=${name} disabled=${!!busy || !!unavailable} onInput=${e => setName(e.target.value)} /></label>
            <button disabled=${!!busy || !!unavailable || !recentlyVerified || !name.trim()} onClick=${add}>Add passkey</button></div>
        ${keys && keys.length === 0 && fresh && html`<p>No passkeys registered.</p>`}
        ${keys && html`<ul class="gi-passkey-list">${keys.map(row => html`<li key=${row.id} class="gi-passkey-row" data-credential-id=${row.id}>
            <strong>${row.name || 'Unnamed passkey'}</strong><small>Identifier: ${row.id}</small>
            <small>Created: ${date(row.created_at)}</small><small>Last used: ${date(row.last_used_at)}</small>
            <div><button disabled=${!!busy || !recentlyVerified} onClick=${e => { opener.current=e.currentTarget; setRemoving(null); setEditing({id:row.id,name:row.name}); }}>Rename ${row.name || 'passkey'}</button>
                <button disabled=${!!busy || !recentlyVerified} onClick=${e => { opener.current=e.currentTarget; setEditing(null); setRemoving(row.id); }}>Remove ${row.name || 'passkey'}</button></div>
            ${editing?.id === row.id && html`<div class="gi-passkey-edit"><label>New name<input aria-label="Rename passkey" value=${editing.name} disabled=${!!busy} onInput=${e => setEditing({...editing,name:e.target.value})} /></label>
                <button disabled=${!!busy || !recentlyVerified} onClick=${() => mutation('rename',row,editing.name)}>Save passkey name</button><button disabled=${!!busy} onClick=${cancel}>Cancel rename</button></div>`}
            ${removing === row.id && html`<div class="gi-passkey-confirm" role="group" aria-label="Confirm passkey removal"><p>Remove ${row.name || 'passkey'} (${row.id})? This blocks future sign-ins with this credential. Existing login sessions are not signed out.</p>
                <button disabled=${!!busy || !recentlyVerified} onClick=${() => mutation('remove',row)}>Confirm removal</button><button disabled=${!!busy} onClick=${cancel}>Cancel removal</button></div>`}
        </li>`)}</ul>`}
    </section>`;
}

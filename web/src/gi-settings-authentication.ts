import { html, useEffect, useLayoutEffect, useRef, useState } from './vendor/preact-htm.js';
import { authJSON, passkeyUnavailable, runPasskey } from './gi-passkeys.js';
import { parseAuthPolicy } from './gi-auth-policy.js';

const removalReasons = new Map([
    ['other-rp', 'The other passkey cannot sign in here because it is registered for another relying party.'],
    ['policy-excludes-totp', 'No other sign-in method is accepted by the current policy. Configured TOTP is not accepted in passkey-only mode.'],
    ['totp-not-enabled', 'TOTP is not enabled and cannot be used for sign-in. An unverified setup is not a sign-in method.'],
    ['sessions-not-factors', 'An active session is not a future sign-in method. Add another passkey before removing this key.'],
    ['no-other-method', 'Add another sign-in method before removing this key. No other accepted factor is configured.'],
]);

export function GiSettingsAuthentication() {
    const [policy, setPolicy] = useState(null);
    const [proof, setProof] = useState(null);
    const [loginPolicy, setLoginPolicy] = useState(null);
    const [policyChoice, setPolicyChoice] = useState('either');
    const [confirmPolicy, setConfirmPolicy] = useState(false);
    const [confirmLogout, setConfirmLogout] = useState(false);
    const [logoutUncertain, setLogoutUncertain] = useState(false);
    const [keys, setKeys] = useState(null);
    const [fresh, setFresh] = useState(false);
    const [busy, setBusy] = useState('');
    const [error, setError] = useState('');
    const [notice, setNotice] = useState('');
    const [removalDetail, setRemovalDetail] = useState('');
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
        setEditing(null); setRemoving(null); setConfirmPolicy(false); setConfirmLogout(false); restore();
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
        if (!p.enrolled) { setProof(null); setKeys(null); setLoginPolicy(null); return; }
        const [nextProof, settings] = await Promise.all([
            authJSON('/api/auth/session/proof', undefined, signal), authJSON('/api/auth/policy', undefined, signal),
        ]);
        // Keep the current-RP inventory visible in TOTP-only mode. Policy
        // disables mutations/ceremonies, not metadata for the signed-in owner.
        const list = settings.passkey_configured ? await authJSON('/api/auth/passkeys', undefined, signal) : null;
        if ((list && !Array.isArray(list.passkeys)) || typeof nextProof.reauth_required !== 'boolean'
            || typeof settings.revision !== 'string' || !['either','totp-only','passkey-only'].includes(settings.policy)) throw new Error('Invalid authentication response');
        if (live.current) { setProof(nextProof); setLoginPolicy(settings); setPolicyChoice(settings.policy); setConfirmPolicy(false); setKeys(list?.passkeys ?? null); setFresh(true); }
    };
    const work = async (label: string, action: (signal: AbortSignal) => Promise<void | string>, message = '') => {
        if (flight.current) return;
        const controller = new AbortController(); flight.current = controller;
        const trigger = document.activeElement as HTMLElement;
        const triggerAction = trigger?.getAttribute('data-auth-action');
        setBusy(label); setError(''); setNotice(''); setRemovalDetail('');
        try { const detail = await action(controller.signal); if (live.current) { await refresh(controller.signal); if (live.current) { if (message) setNotice(message); if (typeof detail === 'string') setRemovalDetail(detail); } } }
        catch (e) { if (live.current) { setFresh(false); const reason = label === 'Removing passkey…' ? removalReasons.get(e.reason) : ''; setError([e.message || 'Request failed. Refresh before retrying.', reason].filter(Boolean).join(' ')); } }
        finally { if (flight.current === controller) { flight.current = null; if (live.current) { setBusy(''); requestAnimationFrame(() => {
            if (!live.current || !root.current?.closest('.settings-dialog')) return;
            // Conditional progress/error rows can replace the proof controls. Resolve
            // the same action in this pane instead of falling back to Refresh.
            const replacement = triggerAction ? root.current.querySelector(`[data-auth-action="${CSS.escape(triggerAction)}"]:not(:disabled)`) : null;
            const target = trigger?.isConnected && !trigger.hasAttribute('disabled') ? trigger : replacement || root.current.querySelector('button:not(:disabled)');
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
        if (kind === 'remove') {
            if (result.remaining_method === 'totp') return 'TOTP remains available for sign-in.';
            if (result.remaining_method === 'passkey') return 'Another passkey remains available for sign-in.';
        }
    }, kind === 'rename' ? 'Name saved.' : 'Passkey removed. Existing login sessions are not signed out.');
    const policyAllowed = loginPolicy && ((policyChoice !== 'passkey-only' && loginPolicy.totp_configured)
        || (policyChoice !== 'totp-only' && loginPolicy.passkey_usable));
    const savePolicy = () => work('Saving sign-in policy…', async signal => {
        const result = await authJSON('/api/auth/policy', {policy:policyChoice,revision:loginPolicy.revision}, signal);
        if (result.policy !== policyChoice || typeof result.revision !== 'string') throw new Error('Policy change could not be confirmed. Refresh before trying again.');
        if (live.current) { setConfirmPolicy(false); restore(); }
    }, 'Sign-in policy saved. Existing login sessions are unchanged.');
    // Logout must remain usable with stale/removed-factor proof. An uncertain
    // POST is never replayed: the separate check reads status only.
    const logout = async (checkOnly = false) => {
        if (flight.current) return;
        const controller = new AbortController(); flight.current = controller;
        setBusy(checkOnly ? 'Checking sign-in status…' : 'Signing out…'); setError(''); setNotice(''); setRemovalDetail('');
        try {
            if (!checkOnly) {
                const result = await authJSON('/api/auth/session/logout', {}, controller.signal);
                if (result.ok !== true) throw new Error('Sign-out could not be confirmed. Check sign-in status before retrying.');
            }
            const status = parseAuthPolicy(await authJSON('/api/auth/status', undefined, controller.signal));
            if (!live.current) return;
            if (status.authenticated) throw new Error('This browser is still signed in. Retry sign-out explicitly when ready.');
            setConfirmLogout(false); setLogoutUncertain(false);
            window.dispatchEvent(new Event('gi-auth-status-changed'));
        } catch (e) { if (live.current) { setLogoutUncertain(true); setError(e.message || 'Sign-out could not be confirmed. Check sign-in status before retrying.'); } }
        finally { if (flight.current === controller) { flight.current = null; if (live.current) setBusy(''); } }
    };
    const date = value => !value || value.startsWith('0001-') ? 'Never used' : new Date(value).toLocaleString();
    return html`<section ref=${root} class="gi-authentication-pane" aria-labelledby="gi-authentication-heading" data-auth-escape=${busy || editing || removing || confirmPolicy || confirmLogout ? 'true' : undefined}>
        <h2 id="gi-authentication-heading">Authentication</h2>
        <p>Manage passkeys for this instance owner. Each row is one credential, not an inventory of devices.</p>
        ${proof && html`<button disabled=${!!busy} onClick=${e => {opener.current=e.currentTarget;setConfirmLogout(true);setEditing(null);setRemoving(null);setConfirmPolicy(false);}}>Sign out this browser</button>`}
        ${confirmLogout && html`<div role="group" aria-label="Confirm browser sign-out"><p>Sign out this browser? Other login sessions and registered credentials stay unchanged. Local drafts and attachments are kept.</p>
            <button disabled=${!!busy} onClick=${() => logout()}>Confirm sign out</button><button disabled=${!!busy} onClick=${cancel}>Cancel sign out</button></div>`}
        ${logoutUncertain && html`<button disabled=${!!busy} onClick=${() => logout(true)}>Check sign-in status</button>`}
        <h3>Passkeys</h3>
        ${busy && html`<p role="status">${busy}</p>`}
        ${error && html`<p role="alert">${error}</p>`}
        ${notice && html`<p role="status">${notice}</p>`}
        ${removalDetail && html`<p role="status">${removalDetail}</p>`}
        <button disabled=${!!busy} onClick=${() => work('Refreshing passkeys…', async () => {})}>Refresh passkeys</button>
        ${busy && html`<button onClick=${cancel}>Cancel pending operation</button>`}
        ${policy && unavailable && html`<p>${unavailable}</p>`}
        ${keys && !fresh && html`<p role="status">The displayed list is the last confirmed snapshot. Refresh before making changes.</p>`}
        ${policy?.enrolled && html`<div class="gi-passkey-proof">
            <p>${recentlyVerified ? 'Recently authenticated for credential changes.' : 'Verify an accepted factor before changing passkeys.'}</p>
            ${policy.totp_login_available && html`<label>Authentication code<input aria-label="Reauthentication code" type="text" inputMode="numeric" autoComplete="one-time-code" value=${code} disabled=${!!busy} onInput=${e => setCode(e.target.value)} /></label>
                <button data-auth-action="verify-totp" disabled=${!!busy || !/^\d{6}$/.test(code)} onClick=${() => work('Verifying authentication…', async signal => { await authJSON('/api/auth/session/reauth/totp', {code}, signal); if (live.current) setCode(''); }, 'Authentication verified.')}>Verify code</button>`}
            ${policy.passkey_login_available && html`<button data-auth-action="verify-passkey" disabled=${!!busy || !!passkeyUnavailable()} onClick=${() => work('Waiting for passkey verification…', signal => runPasskey('reauth', signal), 'Authentication verified.')}>Verify with passkey</button>`}
        </div>`}
        ${loginPolicy && html`<div class="gi-signin-policy"><h3>Sign-in policy</h3>
            <p>Current policy: ${loginPolicy.policy}. Changing accepted factors does not remove credentials or sign out existing sessions.</p>
            <label>Accepted sign-in methods<select aria-label="Accepted sign-in methods" value=${policyChoice} disabled=${!!busy || !fresh} onChange=${e => { setPolicyChoice(e.target.value); setConfirmPolicy(false); }}>
                <option value="either">TOTP or passkey</option><option value="totp-only">TOTP only</option><option value="passkey-only">Passkeys only</option></select></label>
            ${!policyAllowed && html`<p>The selected policy needs a configured, usable sign-in method. Enrol a passkey for this origin or keep verified TOTP enabled.</p>`}
            <button disabled=${!!busy || !recentlyVerified || !policyAllowed || policyChoice === loginPolicy.policy} onClick=${e => {opener.current=e.currentTarget;setConfirmPolicy(true);setRemoving(null);setEditing(null);}}>Change sign-in policy</button>
            ${confirmPolicy && html`<div role="group" aria-label="Confirm sign-in policy"><p>Use ${policyChoice} for future sign-ins? Existing sessions and stored factors will not be removed.</p>
                <button disabled=${!!busy || !recentlyVerified || !policyAllowed} onClick=${savePolicy}>Confirm policy change</button><button disabled=${!!busy} onClick=${cancel}>Cancel policy change</button></div>`}
        </div>`}
        <div class="gi-passkey-add"><label>Passkey name<input aria-label="New passkey name" value=${name} disabled=${!!busy || !!unavailable} onInput=${e => setName(e.target.value)} /></label>
            <button disabled=${!!busy || !!unavailable || !recentlyVerified || !name.trim()} onClick=${add}>Add passkey</button></div>
        ${keys && keys.length === 0 && fresh && html`<p>No passkeys registered.</p>`}
        ${keys && html`<ul class="gi-passkey-list">${keys.map(row => html`<li key=${row.id} class="gi-passkey-row" data-credential-id=${row.id}>
            <strong>${row.name || 'Unnamed passkey'}</strong><small>Identifier: ${row.id}</small>
            <small>Created: ${date(row.created_at)}</small><small>Last used: ${date(row.last_used_at)}</small>
            <div><button disabled=${!!busy || !recentlyVerified || !!unavailable} onClick=${e => { opener.current=e.currentTarget; setRemoving(null); setEditing({id:row.id,name:row.name}); }}>Rename ${row.name || 'passkey'}</button>
                <button disabled=${!!busy || !recentlyVerified || !!unavailable} onClick=${e => { opener.current=e.currentTarget; setEditing(null); setRemoving(row.id); }}>Remove ${row.name || 'passkey'}</button></div>
            ${editing?.id === row.id && html`<div class="gi-passkey-edit"><label>New name<input aria-label="Rename passkey" value=${editing.name} disabled=${!!busy} onInput=${e => setEditing({...editing,name:e.target.value})} /></label>
                <button disabled=${!!busy || !recentlyVerified || !!unavailable} onClick=${() => mutation('rename',row,editing.name)}>Save passkey name</button><button disabled=${!!busy} onClick=${cancel}>Cancel rename</button></div>`}
            ${removing === row.id && html`<div class="gi-passkey-confirm" role="group" aria-label="Confirm passkey removal"><p>Remove ${row.name || 'passkey'} (${row.id})? This blocks future sign-ins with this credential. Existing login sessions are not signed out.</p>
                <button disabled=${!!busy || !recentlyVerified || !!unavailable} onClick=${() => mutation('remove',row)}>Confirm removal</button><button disabled=${!!busy} onClick=${cancel}>Cancel removal</button></div>`}
        </li>`)}</ul>`}
    </section>`;
}

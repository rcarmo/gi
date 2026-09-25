import { html, useEffect, useLayoutEffect, useRef, useState } from './vendor/preact-htm.js';
import { authJSON } from './gi-passkeys.js';
import { parseAuthPolicy } from './gi-auth-policy.js';

// Secrets stay in this mounted pane only. No clipboard, links, QR service,
// browser storage or chat transport is involved. Server TTL is authoritative.
export function GiSettingsSetup({ available, disabled, onComplete, onBusy }) {
    const [pending, setPending] = useState(null);
    const [code, setCode] = useState('');
    const [busy, setBusy] = useState('');
    const [uncertain, setUncertain] = useState(false);
    const [message, setMessage] = useState('');
    const root = useRef(null), live = useRef(true), flight = useRef<AbortController | null>(null);
    const clear = () => { setPending(null); setCode(''); };
    const focusAction = () => requestAnimationFrame(() => {
        if (live.current && (root.current?.contains(document.activeElement) || document.activeElement === document.body))
            (root.current?.querySelector('input:not(:disabled)') || root.current?.querySelector('button:not(:disabled)'))?.focus({preventScroll:true});
    });
    const request = async (operation: 'start' | 'finish' | 'cancel' | 'check') => {
        if (flight.current || (disabled && operation !== 'cancel') || (operation === 'start' && !available)) return;
        const controller = new AbortController(); flight.current = controller;
        // Account for time spent waiting for a start response; a late reply
        // must not extend the displayed secret's server-side lifetime.
        const started = Date.now(), submittedCode = code;
        clear(); setMessage(''); setBusy(operation); onBusy(true);
        try {
            if (operation === 'start') {
                const p = await authJSON('/api/auth/setup/start', {}, controller.signal);
                if (p.username !== 'admin' || typeof p.secret !== 'string' || !/^[A-Z2-7]{32}$/.test(p.secret)
                    || p.expires_in_seconds !== 600) throw new Error('Invalid setup');
                if (live.current && !controller.signal.aborted) {
                    if (Date.now() >= started + 600000) throw new Error('Expired setup');
                    setPending({secret:p.secret,expires:started + 600000}); setUncertain(false);
                    requestAnimationFrame(() => { if (live.current && (root.current?.contains(document.activeElement) || document.activeElement === document.body)) root.current?.querySelector('input')?.focus(); });
                }
                return;
            }
            if (operation !== 'check') {
                const result = await authJSON('/api/auth/setup/' + operation, operation === 'finish' ? {code:submittedCode} : {}, controller.signal);
                if (result.ok !== true) throw new Error('Unconfirmed setup');
            }
            const status = parseAuthPolicy(await authJSON('/api/auth/status', undefined, controller.signal));
            if (!live.current || controller.signal.aborted) return;
            if (status.enrolled) {
                setUncertain(false);
                if (status.authenticated) onComplete();
                else window.dispatchEvent(new Event('gi-auth-status-changed'));
            } else if (operation === 'finish') {
                throw new Error('Unconfirmed owner');
            } else {
                setUncertain(false);
                setMessage(operation === 'cancel' ? 'Setup cancelled. No owner was created.' : 'No owner is configured. Start again to get a new setup key.');
            }
        } catch {
            if (live.current && flight.current === controller) {
                clear(); setUncertain(true);
                // Never interpolate server bodies, codes or secrets into errors.
                setMessage('Setup could not be confirmed. Check setup status before trying again. If setup completed, sign in with your saved authenticator.');
            }
        } finally {
            if (flight.current === controller) {
                flight.current = null;
                if (live.current) { setBusy(''); onBusy(false); focusAction(); }
            }
        }
    };
    const cancel = () => {
        clear();
        if (flight.current) {
            const controller = flight.current; flight.current = null; controller.abort();
            setBusy(''); onBusy(false); setUncertain(true);
            setMessage('Setup could not be confirmed. Check setup status before trying again. If setup completed, sign in with your saved authenticator.');
            focusAction(); return;
        }
        if (pending) void request('cancel');
    };
    useLayoutEffect(() => {
        const node = root.current;
        const escape = (event: Event) => { event.stopPropagation(); cancel(); };
        node.addEventListener('gi-auth-escape', escape);
        return () => node.removeEventListener('gi-auth-escape', escape);
    });
    useEffect(() => {
        if (!pending) return;
        const timer = setTimeout(() => {
            clear(); setUncertain(true); setMessage('Setup key expired. Check setup status before starting again.');
        }, Math.max(0, pending.expires - Date.now()));
        return () => clearTimeout(timer);
    }, [pending]);
    useEffect(() => () => {
        live.current = false; flight.current?.abort(); flight.current = null;
        // Do not race an unobserved cancel POST against a finishing owner write.
        // Server pending setup expires, or the next explicit start replaces it.
        onBusy(false);
    }, []);
    return html`<section ref=${root} aria-labelledby="gi-setup-heading" data-auth-escape=${pending || busy ? 'true' : undefined}>
        <h3 id="gi-setup-heading">Set up instance owner</h3>
        <p>Set up an authenticator for the owner account, admin. This protects future browser access. Keep the authenticator entry; it is needed if a setup response is lost.</p>
        ${!available && html`<p>Open Gi on the server's localhost or loopback address to set up its owner. Remote HTTPS cannot start setup.</p>`}
        ${message && html`<p role=${uncertain ? 'alert' : 'status'}>${message}</p>`}
        ${busy && html`<p role="status">${busy === 'check' ? 'Checking setup status…' : busy === 'start' ? 'Starting owner setup…' : busy === 'finish' ? 'Verifying owner setup…' : 'Cancelling owner setup…'}</p>`}
        ${!pending && !uncertain && html`<button data-setup-action="start" disabled=${!available || disabled || !!busy} onClick=${() => request('start')}>Set up owner</button>`}
        ${pending && html`<p>In your authenticator, add a time-based key for gi / admin: six digits, SHA1, 30 seconds. This key is shown only in this pane for up to ten minutes. Save it in your authenticator before verifying. Leaving the pane hides it; a new setup replaces it.</p>
            <pre aria-label="Authenticator setup key" style="white-space:pre-wrap;overflow-wrap:anywhere;word-break:break-all">${pending.secret}</pre>
            <label>Setup verification code<input aria-label="Setup verification code" type="text" inputMode="numeric" autoComplete="one-time-code" value=${code} disabled=${disabled || !!busy} onInput=${e => setCode(e.target.value)} /></label>
            <button disabled=${disabled || !!busy || !/^\d{6}$/.test(code)} onClick=${() => request('finish')}>Verify and enable authentication</button>`}
        ${uncertain && html`<button disabled=${disabled || !!busy} onClick=${() => request('check')}>Check setup status</button>`}
        ${(pending || busy) && html`<button onClick=${cancel}>Cancel owner setup</button>`}
    </section>`;
}

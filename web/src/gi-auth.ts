import { html, useEffect, useRef, useState } from './vendor/preact-htm.js';

import { parseAuthPolicy, type AuthPolicy } from "./gi-auth-policy.js";
import { passkeyUnavailable, runPasskey } from './gi-passkeys.js';

async function policyRequest(signal: AbortSignal): Promise<AuthPolicy> {
    const response = await fetch('/api/auth/status', { cache: 'no-store', credentials: 'same-origin', signal });
    if (!response.ok) throw new Error('Cannot load sign-in options.');
    return parseAuthPolicy(await response.json());
}

// The application is not mounted until policy and cookie authority are known.
// No token, code, or policy is cached in browser storage.
export function GiAuthGate({ children }) {
    const [policy, setPolicy] = useState<AuthPolicy | null>(null);
    const [error, setError] = useState('');
    const [code, setCode] = useState('');
    const [busy, setBusy] = useState(false);
    const [attempt, setAttempt] = useState(0);
    const flight = useRef<AbortController | null>(null);
    const input = useRef<HTMLInputElement | null>(null);
    const passkeyFlight = useRef<AbortController | null>(null);
    const passkeyButton = useRef<HTMLButtonElement | null>(null);
    useEffect(() => () => { passkeyFlight.current?.abort(); passkeyFlight.current = null; }, []);

    useEffect(() => {
        const controller = new AbortController();
        flight.current = controller;
        const timeout = setTimeout(() => controller.abort(), 15000);
        let disposed = false;
        setPolicy(null);
        setError('');
        policyRequest(controller.signal).then(value => {
            if (!disposed) setPolicy(value);
        }).catch(() => {
            if (!disposed) setError('Cannot load sign-in options. Check your connection and try again.');
        }).finally(() => {
            clearTimeout(timeout);
            if (flight.current === controller) flight.current = null;
        });
        return () => { disposed = true; clearTimeout(timeout); controller.abort(); flight.current?.abort(); flight.current = null; };
    }, [attempt]);

    useEffect(() => { if (policy?.enrolled && !policy.authenticated) input.current?.focus(); }, [policy]);

    const submit = async (event: Event) => {
        event.preventDefault();
        if (flight.current || passkeyFlight.current || !policy?.totp_login_available || !policy?.enrolled || !policy.browser_login_available || !/^\d{6}$/.test(code)) return;
        const controller = new AbortController();
        flight.current = controller;
        const timeout = setTimeout(() => controller.abort(), 15000);
        setBusy(true);
        setError('');
        try {
            const response = await fetch('/api/auth/session', {
                method: 'POST', headers: { 'Content-Type': 'application/json' }, credentials: 'same-origin',
                body: JSON.stringify({ code }), signal: controller.signal,
            });
            if (!response.ok) throw new Error(response.status === 401 ? 'Invalid authentication code. Try again.' : 'Sign-in failed. Try again.');
            // A 200 without a usable cookie is not a successful browser sign-in.
            const confirmed = await policyRequest(controller.signal);
            if (!confirmed.authenticated) throw new Error('Sign-in could not be confirmed. Check that cookies are enabled and try again.');
            if (flight.current !== controller) return;
            setCode('');
            setPolicy(confirmed);
        } catch (failure) {
            if (flight.current !== controller) return;
            setError(failure instanceof TypeError || controller.signal.aborted
                ? 'Unable to reach the server. Check your connection and try again.'
                : failure.message || 'Sign-in failed. Try again.');
        } finally {
            clearTimeout(timeout);
            if (flight.current === controller) { flight.current = null; setBusy(false); }
        }
    };

    const passkeyLogin = async () => {
        if (flight.current || passkeyFlight.current || !policy?.passkey_login_available) return;
        const controller = new AbortController(); passkeyFlight.current = controller;
        setBusy(true); setError('');
        try {
            await runPasskey('login', controller.signal);
            const confirmed = await policyRequest(AbortSignal.any([controller.signal, AbortSignal.timeout(15000)]));
            if (passkeyFlight.current !== controller) return;
            if (!confirmed.authenticated) throw new Error('Sign-in could not be confirmed. Refresh status before trying again.');
            setPolicy(confirmed);
        } catch (failure) { if (passkeyFlight.current === controller) setError(failure.message || 'Passkey sign-in failed.'); }
        finally { if (passkeyFlight.current === controller) { passkeyFlight.current = null; setBusy(false); requestAnimationFrame(() => passkeyButton.current?.focus()); } }
    };

    if (policy && (!policy.enrolled || policy.authenticated)) return children;
    return html`<main class="gi-auth" aria-labelledby="gi-auth-title"><section class="gi-auth-card">
        <h1 id="gi-auth-title">Sign in to Gi</h1>
        ${error && html`<p role="alert">${error}</p>`}
        ${!policy ? error
            ? html`<button type="button" onClick=${() => setAttempt(n => n + 1)}>Retry</button>`
            : html`<p role="status">Loading sign-in options…</p>`
        : !policy.browser_login_available
            ? html`<p>Open Gi over HTTPS or localhost to sign in.</p>`
            : html`<form onSubmit=${submit} aria-busy=${busy}>
                ${policy.totp_login_available && html`<label for="gi-auth-code">Authentication code</label>
                <input ref=${input} id="gi-auth-code" type="text" inputMode="numeric" pattern="[0-9]{6}" maxLength="6"
                    autoComplete="one-time-code" required value=${code} disabled=${busy}
                    onInput=${(event) => setCode(event.currentTarget.value)} />
                <button type="submit" disabled=${busy || !/^\d{6}$/.test(code)}>${busy ? 'Signing in…' : 'Sign in'}</button>`}
                ${policy.passkey_login_available && html`<button ref=${passkeyButton} type="button" disabled=${busy || !!passkeyUnavailable()} onClick=${passkeyLogin}>Sign in with passkey</button>${passkeyUnavailable() && html`<p>${passkeyUnavailable()}</p>`}`}
                ${passkeyFlight.current && html`<p role="status">Waiting for passkey sign-in…</p><button type="button" onClick=${() => passkeyFlight.current?.abort()}>Cancel passkey prompt</button>`}
                ${!policy.totp_login_available && !policy.passkey_login_available && html`<p role="alert">No sign-in method is available for this origin. Check the instance authentication configuration.</p>`}
                ${error && html`<button type="button" disabled=${busy} onClick=${() => setAttempt(n => n + 1)}>Refresh sign-in status</button>`}
            </form>`}
    </section></main>`;
}

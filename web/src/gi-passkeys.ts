// Native browser ceremonies only. Nothing is cached in storage or sent to chat.
export class AuthAPIError extends Error {
    constructor(message: string, public status = 0, public reason = '') { super(message); }
}
export async function authJSON(path: string, body?: unknown, signal?: AbortSignal) {
    const timeout = AbortSignal.timeout(15000);
    const response = await fetch(path, { method: body === undefined ? 'GET' : 'POST',
        credentials: 'same-origin', cache: 'no-store', signal: signal ? AbortSignal.any([signal, timeout]) : timeout,
        ...(body === undefined ? {} : { headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) }) });
    const data = await response.json().catch(() => null);
    if (!response.ok) throw new AuthAPIError(data?.error || 'Authentication request failed', response.status, typeof data?.reason === 'string' ? data.reason : '');
    if (!data || typeof data !== 'object') throw new AuthAPIError('Invalid authentication response');
    return data;
}
export function passkeyUnavailable(): string {
    if (!window.isSecureContext) return 'Passkeys require a secure HTTPS or localhost origin.';
    if (!window.PublicKeyCredential || !navigator.credentials?.create || !navigator.credentials?.get)
        return 'This browser cannot use passkeys.';
    return '';
}
function decode(value: string): Uint8Array {
    return Uint8Array.from(atob(value.replace(/-/g, '+').replace(/_/g, '/')), c => c.charCodeAt(0));
}
function encode(value: ArrayBuffer | null): string | null {
    if (value === null) return null;
    return btoa(Array.from(new Uint8Array(value), b => String.fromCharCode(b)).join('')).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}
export async function runPasskey(operation: 'register' | 'login' | 'reauth', signal: AbortSignal, name?: string) {
    const unavailable = passkeyUnavailable(); if (unavailable) throw new Error(unavailable);
    let created = false, finishing = false;
    try {
        const start = await authJSON(`/api/auth/passkeys/${operation}/start`, name === undefined ? {} : { name }, signal);
        const pub = start.options?.publicKey;
        if (typeof start.ceremony_id !== 'string' || !pub?.challenge) throw new Error('Invalid passkey options');
        pub.challenge = decode(pub.challenge);
        if (operation === 'register') {
            pub.user.id = decode(pub.user.id);
            pub.excludeCredentials = (pub.excludeCredentials || []).map(c => ({ ...c, id: decode(c.id) }));
        } else pub.allowCredentials = (pub.allowCredentials || []).map(c => ({ ...c, id: decode(c.id) }));
        const credential: any = await (operation === 'register'
            ? navigator.credentials.create({ publicKey: pub, signal })
            : navigator.credentials.get({ publicKey: pub, signal }));
        if (!credential) throw new Error('No passkey response');
        created = operation === 'register';
        // Portable serialisation; older browsers lack PublicKeyCredential.toJSON.
        const response = credential.response;
        const data = { id: credential.id, rawId: encode(credential.rawId), type: credential.type,
            authenticatorAttachment: credential.authenticatorAttachment,
            clientExtensionResults: credential.getClientExtensionResults(),
            response: operation === 'register'
                ? { clientDataJSON: encode(response.clientDataJSON), attestationObject: encode(response.attestationObject), transports: response.getTransports?.() || [] }
                : { clientDataJSON: encode(response.clientDataJSON), authenticatorData: encode(response.authenticatorData), signature: encode(response.signature), userHandle: encode(response.userHandle) } };
        finishing = true;
        const result = await authJSON(`/api/auth/passkeys/${operation}/finish`, { ceremony_id: start.ceremony_id, credential: data }, signal);
        if (result.ok !== true) throw new Error('Completion could not be confirmed');
    } catch (error) {
        if (!finishing && (error.name === 'AbortError' || error.name === 'NotAllowedError'))
            throw new Error('Passkey prompt cancelled or timed out. Retry explicitly when ready.');
        if (created) {
            if (error instanceof AuthAPIError && error.status >= 400 && error.status < 500)
                throw new Error(`${error.message}. Not registered on the server. A local credential may remain in your authenticator or password manager; Gi has not removed it.`);
            throw new Error('Registration result could not be confirmed. Refresh the list before trying again. A local credential may remain in your authenticator or password manager.');
        }
        if (finishing) throw new Error('Passkey sign-in result could not be confirmed. Refresh status before trying again.');
        throw error;
    }
}

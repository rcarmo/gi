export interface AuthPolicy {
    enrolled: boolean;
    authenticated: boolean;
    totp_enabled: boolean;
    mode: 'single-user';
    browser_login_available: boolean;
    setup_available: boolean;
    totp_login_available: boolean;
    passkeys_enabled: boolean;
    passkey_login_available: boolean;
}

export function parseAuthPolicy(value: unknown): AuthPolicy {
    const p = value as AuthPolicy;
    if (!p || p.mode !== 'single-user' || typeof p.enrolled !== 'boolean'
        || typeof p.authenticated !== 'boolean' || typeof p.totp_enabled !== 'boolean'
        || typeof p.browser_login_available !== 'boolean'
        || (p.totp_login_available !== undefined && typeof p.totp_login_available !== 'boolean')
        || (p.passkeys_enabled !== undefined && typeof p.passkeys_enabled !== 'boolean')
        || (p.passkey_login_available !== undefined && typeof p.passkey_login_available !== 'boolean')
        || (p.setup_available !== undefined && typeof p.setup_available !== 'boolean')
        || (p.setup_available === true && p.enrolled)
        || (p.passkey_login_available === true && p.passkeys_enabled !== true)
        || (p.enrolled && !p.totp_enabled && p.passkeys_enabled === undefined)) {
        throw new Error('Invalid authentication policy');
    }
    return { ...p, setup_available: p.setup_available ?? false, totp_login_available: p.totp_login_available ?? p.totp_enabled,
        passkeys_enabled: p.passkeys_enabled ?? false, passkey_login_available: p.passkey_login_available ?? false };
}

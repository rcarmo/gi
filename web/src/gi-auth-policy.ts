export interface AuthPolicy {
    enrolled: boolean;
    authenticated: boolean;
    totp_enabled: boolean;
    mode: 'single-user';
    browser_login_available: boolean;
}

export function parseAuthPolicy(value: unknown): AuthPolicy {
    const p = value as AuthPolicy;
    if (!p || p.mode !== 'single-user' || typeof p.enrolled !== 'boolean'
        || typeof p.authenticated !== 'boolean' || typeof p.totp_enabled !== 'boolean'
        || typeof p.browser_login_available !== 'boolean' || (p.enrolled && !p.totp_enabled)) {
        throw new Error('Invalid authentication policy');
    }
    return p;
}

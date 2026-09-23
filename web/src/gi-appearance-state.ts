export const APPEARANCE_KEY = 'gi_browser_appearance_v1';
export type Appearance = { version: 1; theme: string; tint: string };
export const defaultAppearance: Appearance = { version: 1, theme: 'default', tint: '' };

export function validateAppearance(value: any, presets: readonly string[]): Appearance {
    if (!value || value.version !== 1 || typeof value.theme !== 'string' || !presets.includes(value.theme)) {
        throw new Error('Choose a supported theme preset.');
    }
    if (typeof value.tint !== 'string') throw new Error('Tint must be a hex colour.');
    let tint = value.tint.trim().toLowerCase();
    if (tint && !/^#[0-9a-f]{3}([0-9a-f]{3})?$/.test(tint)) throw new Error('Use #RGB or #RRGGBB for the tint, or leave it empty.');
    if (tint.length === 4) tint = '#' + [...tint.slice(1)].map(c => c + c).join('');
    return { version: 1, theme: value.theme, tint: value.theme === 'default' ? tint : '' };
}

export function readAppearance(storage: Pick<Storage, 'getItem'>, presets: readonly string[]): Appearance | null {
    try {
        const raw = storage.getItem(APPEARANCE_KEY);
        return raw ? validateAppearance(JSON.parse(raw), presets) : null;
    } catch { return null; }
}

export function saveAppearance(storage: Pick<Storage, 'setItem'>, value: any, presets: readonly string[]): Appearance {
    const next = validateAppearance(value, presets);
    // One write: do not apply the visual state until persistence succeeds.
    storage.setItem(APPEARANCE_KEY, JSON.stringify(next));
    return next;
}

// Catalogue/renderer exports are added by gi-appearance-renderer in build.js.
// The supplied theme module itself is never modified.
import { giThemePresets, giApplyThemeState } from './ui/theme.js';
import { APPEARANCE_KEY, defaultAppearance, readAppearance, saveAppearance, type Appearance } from './gi-appearance-state.js';
export const appearancePresets = Object.keys(giThemePresets);
const changeEvent = 'gi:appearance-changed';

function readStoredAppearance() {
    try { return readAppearance(window.localStorage, appearancePresets); } catch { return null; }
}

export function currentAppearance(): Appearance {
    const stored = readStoredAppearance();
    if (stored) return stored;
    const theme = document.documentElement.dataset.colorTheme || 'default';
    return { version: 1, theme: appearancePresets.includes(theme) ? theme : 'default', tint: document.documentElement.dataset.tint || '' };
}

function render(value: Appearance) {
    giApplyThemeState(value, { persist: false });
    window.dispatchEvent(new CustomEvent(changeEvent, { detail: value }));
}

export function persistAppearance(value: Appearance) {
    const saved = saveAppearance(window.localStorage, value, appearancePresets);
    render(saved);
    return saved;
}

export function subscribeAppearance(onChange: (value: Appearance) => void) {
    const listener = (event: CustomEvent) => onChange(event.detail);
    window.addEventListener(changeEvent, listener);
    return () => window.removeEventListener(changeEvent, listener);
}

export function initGiAppearance() {
    // The explicit Gi preference overrides legacy URL/chat theme lookup without
    // changing those old keys. No preference means legacy startup is unchanged.
    const saved = readStoredAppearance();
    if (saved) render(saved);
    const storage = (event: StorageEvent) => {
        if (event.storageArea !== window.localStorage || (event.key !== APPEARANCE_KEY && event.key !== null)) return;
        const next = readStoredAppearance();
        if (next) render(next);
        else if (event.newValue === null) render(defaultAppearance);
    };
    window.addEventListener('storage', storage);
    return () => window.removeEventListener('storage', storage);
}

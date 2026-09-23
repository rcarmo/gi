// Gi-owned lazy Appearance pane; storage remains browser-local.
import { html, useState, useEffect, useRef } from "./vendor/preact-htm.js";
import { appearancePresets, currentAppearance, persistAppearance, subscribeAppearance } from "./gi-appearance.js";
import { defaultAppearance } from "./gi-appearance-state.js";

export function Appearance() {
    const [draft, setDraft] = useState(currentAppearance);
    const [error, setError] = useState('');
    const [notice, setNotice] = useState('');
    const dirty = useRef(false);
    const ownSave = useRef(false);
    useEffect(() => subscribeAppearance(value => {
        if (ownSave.current) return;
        if (dirty.current) setNotice('Appearance changed in another tab. Your unsaved fields are unchanged. Save to overwrite or reset to defaults.');
        else { setDraft(value); setNotice('Appearance updated from another tab.'); }
    }), []);
    const update = patch => {
        dirty.current = true; setDraft(previous => ({ ...previous, ...patch })); setError(''); setNotice('');
    };
    const save = value => {
        setError(''); setNotice(''); ownSave.current = true;
        try {
            const saved = persistAppearance(value);
            dirty.current = false; setDraft(saved); setNotice('Appearance saved in this browser.');
        } catch (error) {
            setError(`Appearance was not saved: ${error.message}`);
        } finally { ownSave.current = false; }
    };
    return html`<section aria-labelledby="gi-appearance-title">
        <h2 id="gi-appearance-title">Appearance</h2>
        <p>Browser settings · this origin, across all sessions</p>
        <p>Only this browser profile changes. Server configuration, other devices and the terminal theme are unchanged. Default follows your system colour mode.</p>
        <label>Theme preset<select aria-label="Theme preset" value=${draft.theme} onChange=${e => update({ theme: e.target.value, tint: '' })}>
            ${appearancePresets.map(theme => html`<option value=${theme}>${theme}</option>`)}
        </select></label>
        <label>Custom tint<input aria-label="Custom tint" type="text" placeholder="#RRGGBB" maxLength="7" disabled=${draft.theme !== 'default'} value=${draft.tint} onInput=${e => update({ tint: e.target.value })} /></label>
        <p>Default theme only. Use #RGB or #RRGGBB, or leave empty for no tint.</p>
        <button onClick=${() => save(draft)}>Save appearance</button>
        <button onClick=${() => save(defaultAppearance)}>Reset appearance</button>
        ${error && html`<p role="alert">${error}</p>`}
        ${notice && html`<p role="status">${notice}</p>`}
    </section>`;
}


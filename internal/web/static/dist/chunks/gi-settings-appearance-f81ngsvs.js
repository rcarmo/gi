import {
  F_,
  K_,
  Q_,
  fe,
  defaultAppearance,
  appearancePresets,
  currentAppearance,
  persistAppearance,
  subscribeAppearance
} from "./app-42jecs31.js";

// web/src/gi-settings-appearance.ts
function Appearance() {
  const [draft, setDraft] = F_(currentAppearance);
  const [error, setError] = F_("");
  const [notice, setNotice] = F_("");
  const dirty = Q_(false);
  const ownSave = Q_(false);
  K_(() => subscribeAppearance((value) => {
    if (ownSave.current)
      return;
    if (dirty.current)
      setNotice("Appearance changed in another tab. Your unsaved fields are unchanged. Save to overwrite or reset to defaults.");
    else {
      setDraft(value);
      setNotice("Appearance updated from another tab.");
    }
  }), []);
  const update = (patch) => {
    dirty.current = true;
    setDraft((previous) => ({ ...previous, ...patch }));
    setError("");
    setNotice("");
  };
  const save = (value) => {
    setError("");
    setNotice("");
    ownSave.current = true;
    try {
      const saved = persistAppearance(value);
      dirty.current = false;
      setDraft(saved);
      setNotice("Appearance saved in this browser.");
    } catch (error) {
      setError(`Appearance was not saved: ${error.message}`);
    } finally {
      ownSave.current = false;
    }
  };
  return fe`<section aria-labelledby="gi-appearance-title">
        <h2 id="gi-appearance-title">Appearance</h2>
        <p>Browser settings · this origin, across all sessions</p>
        <p>Only this browser profile changes. Server configuration, other devices and the terminal theme are unchanged. Default follows your system colour mode.</p>
        <label>Theme preset<select aria-label="Theme preset" value=${draft.theme} onChange=${(e) => update({ theme: e.target.value, tint: "" })}>
            ${appearancePresets.map((theme) => fe`<option value=${theme}>${theme}</option>`)}
        </select></label>
        <label>Custom tint<input aria-label="Custom tint" type="text" placeholder="#RRGGBB" maxLength="7" disabled=${draft.theme !== "default"} value=${draft.tint} onInput=${(e) => update({ tint: e.target.value })} /></label>
        <p>Default theme only. Use #RGB or #RRGGBB, or leave empty for no tint.</p>
        <button onClick=${() => save(draft)}>Save appearance</button>
        <button onClick=${() => save(defaultAppearance)}>Reset appearance</button>
        ${error && fe`<p role="alert">${error}</p>`}
        ${notice && fe`<p role="status">${notice}</p>`}
    </section>`;
}
export {
  Appearance
};

//# debugId=458DF4723E18513B64756E2164756E21
//# sourceMappingURL=gi-settings-appearance-f81ngsvs.js.map

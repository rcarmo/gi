import {
  F_,
  K_,
  Q_,
  fe,
  getGiProviders,
  saveGiProviderKey,
  removeGiProviderKey
} from "./app-r5wpse0h.js";

// web/src/gi-settings-providers.ts
function GiSettingsProviders() {
  const [snapshot, setSnapshot] = F_(null);
  const [error, setError] = F_("");
  const [notice, setNotice] = F_("");
  const [filter, setFilter] = F_("");
  const [busy, setBusy] = F_(false);
  const [attempt, setAttempt] = F_(0);
  const [confirm, setConfirm] = F_("");
  const alive = Q_(false), mutating = Q_(false);
  const inputs = Q_({});
  const clearSecrets = () => {
    Object.values(inputs.current).forEach((input) => {
      if (input)
        input.value = "";
    });
  };
  K_(() => {
    alive.current = true;
    return () => {
      alive.current = false;
      clearSecrets();
    };
  }, []);
  K_(() => {
    let live = true;
    clearSecrets();
    setSnapshot(null);
    setError("");
    setNotice("");
    setConfirm("");
    getGiProviders().then((value) => {
      if (live)
        setSnapshot(value);
    }).catch(() => {
      if (live)
        setError("Cannot read providers. Repair the native credential store or retry.");
    });
    return () => {
      live = false;
    };
  }, [attempt]);
  async function mutate(provider, remove = false) {
    if (!snapshot?.can_write || mutating.current)
      return;
    const row = snapshot.providers.find((p) => p.id === provider);
    if (!row?.editable)
      return;
    if (!remove && !inputs.current[provider]?.value) {
      setError("Enter an API key before saving.");
      return;
    }
    mutating.current = true;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const value = remove ? await removeGiProviderKey(provider, snapshot.revision) : await saveGiProviderKey(provider, snapshot.revision, inputs.current[provider].value);
      if (alive.current) {
        clearSecrets();
        setSnapshot(value);
        setConfirm("");
        setNotice(remove ? "API-key entry removed. Other providers are unchanged." : "API key stored, not verified with the provider. New requests can use it; model selection is unchanged.");
      }
    } catch (err) {
      if (alive.current) {
        clearSecrets();
        setError(err.status === 409 ? "Credential store changed or is busy. Refresh providers, re-enter the key and retry explicitly." : "Credential change failed. Refresh providers to verify; re-enter the key to retry.");
      }
    } finally {
      mutating.current = false;
      if (alive.current)
        setBusy(false);
    }
  }
  const rows = (snapshot?.providers || []).filter((row) => `${row.id} ${row.name}`.toLowerCase().includes(filter.toLowerCase().trim()));
  return fe`<section aria-labelledby="gi-providers-title">
        <h2 id="gi-providers-title">Providers</h2>
        <p>Service-user credentials · shared across this user's Gi processes and workspaces</p>
        <p>Stored means present in ~/.pi/agent/auth.json, not verified remotely. This native file is private plaintext, not an encrypted keychain. No network probe runs here.</p>
        <p>OpenAI and Anthropic API keys can be saved here. OAuth/token and custom entries are read-only: complete setup with native tooling. Existing requests are not revoked by removal.</p>
        <label>Filter providers<input type="search" aria-label="Filter providers" value=${filter} disabled=${busy} onInput=${(e) => {
    clearSecrets();
    setFilter(e.target.value);
    setConfirm("");
  }} /></label>
        ${!snapshot && !error && fe`<p role="status">Loading providers…</p>`}
        ${snapshot && !snapshot.can_write && fe`<p>Use HTTPS or a loopback connection to edit credentials. Forwarded headers are not trusted.</p>`}
        ${rows.slice(0, 50).map((row) => fe`<section key=${row.id} class="gi-provider-entry" aria-label=${`Provider ${row.id}`}>
            <h3>${row.name}</h3><p><code>${row.id}</code> · ${row.stored ? `Stored ${row.kind}${row.has_material ? "" : " (no usable material)"}` : "Not stored"}</p>
            ${row.editable && snapshot.can_write ? fe`
                <label>API key for ${row.id}<input ref=${(el) => {
    if (el)
      inputs.current[row.id] = el;
    else
      delete inputs.current[row.id];
  }} type="password" aria-label=${`API key for ${row.id}`} autocomplete="off" spellcheck="false" disabled=${busy} /></label>
                <button disabled=${busy} onClick=${() => mutate(row.id)}>Save key</button>
                ${row.stored && (confirm === row.id ? fe`<p>Remove this stored API-key entry?</p><button disabled=${busy} onClick=${() => mutate(row.id, true)}>Confirm removal</button><button disabled=${busy} onClick=${() => setConfirm("")}>Cancel removal</button>` : fe`<button disabled=${busy} onClick=${() => setConfirm(row.id)}>Remove key</button>`)}
            ` : fe`<p>Read-only credential metadata.</p>`}
        </section>`)}
        ${rows.length > 50 && fe`<p>Showing 50 providers. Refine the filter.</p>`}
        <button disabled=${busy} onClick=${() => {
    clearSecrets();
    setSnapshot(null);
    setError("");
    setConfirm("");
    setAttempt((n) => n + 1);
  }}>Refresh providers</button>
        ${error && fe`<p role="alert">${error}</p>`}
        ${notice && fe`<p role="status">${notice}</p>`}
    </section>`;
}
export {
  GiSettingsProviders
};

//# debugId=45C15057B85C7F6C64756E2164756E21
//# sourceMappingURL=gi-settings-providers-9hsjxa4v.js.map

import {
  F_,
  K_,
  W_,
  Q_,
  fe,
  getAgentModels,
  selectAgentModel,
  modelContextBlocked
} from "./app-ks2qgcdv.js";

// web/src/gi-settings-models.ts
function Models({ chatJid, filter = "", onMutationStart, onMutationEnd, onApplied }) {
  const [data, setData] = F_(null);
  const [chosen, setChosen] = F_("");
  const [error, setError] = F_("");
  const [notice, setNotice] = F_("");
  const [busy, setBusy] = F_(false);
  const [attempt, setAttempt] = F_(0);
  const mounted = Q_(false);
  const saving = Q_(false);
  const previousFilter = Q_(filter);
  K_(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);
  W_(() => {
    if (previousFilter.current !== filter) {
      previousFilter.current = filter;
      setChosen("");
      setNotice("");
    }
  }, [filter]);
  K_(() => {
    let live = true;
    setData(null);
    setError("");
    setNotice("");
    getAgentModels(chatJid).then((snapshot) => {
      if (live) {
        setData(snapshot);
        setChosen(snapshot.current);
      }
    }).catch((error) => {
      if (live)
        setError(error.message);
    });
    return () => {
      live = false;
    };
  }, [chatJid, attempt]);
  const options = data?.model_options || data?.models || [];
  const matching = options.filter((option) => `${option.label || option.id} ${option.provider || ""}`.toLowerCase().includes(filter.trim().toLowerCase()));
  const selected = options.find((option) => (option.label || option.id) === chosen);
  const blocked = modelContextBlocked({ contextWindow: selected?.context_window ?? selected?.contextWindow }, data?.context_usage);
  async function apply() {
    if (saving.current || !data || !selected || blocked || chosen === data.current)
      return;
    saving.current = true;
    setBusy(true);
    setError("");
    setNotice("");
    const token = onMutationStart();
    try {
      const result = await selectAgentModel(chatJid, chosen);
      if (mounted.current) {
        setData((previous) => ({ ...previous, ...result }));
        setChosen(result.current);
        setNotice("Model applied to this session.");
        onApplied(result, token);
      }
    } catch (error) {
      if (mounted.current)
        setError(error.message);
    } finally {
      onMutationEnd(token);
      saving.current = false;
      if (mounted.current)
        setBusy(false);
    }
  }
  return fe`<section aria-labelledby="gi-models-title">
        <h2 id="gi-models-title">Models</h2>
        <p>Session settings · <code>${chatJid}</code></p>
        <p>Changes affect this session only. Instance defaults and other sessions are unchanged.</p>
        ${!data && !error && fe`<p role="status">Loading models…</p>`}
        ${error && fe`<div role="alert">${error}${!data && fe` <button onClick=${() => setAttempt((x) => x + 1)}>Retry</button>`}</div>`}
        ${data && fe`
            <dl class="gi-settings-values"><dt>Current model</dt><dd data-testid="settings-current-model">${data.current}</dd>
            <dt>Thinking (read-only)</dt><dd>${data.thinking_level || "Unknown"}</dd>
            <dt>Context capacity</dt><dd data-testid="settings-context-capacity">${Number.isFinite(data.context_window) && data.context_window > 0 ? data.context_window : "Unknown"}</dd></dl>
            <label>Session model<select aria-label="Session model" value=${chosen} disabled=${busy} onChange=${(e) => {
    setChosen(e.target.value);
    setNotice("");
  }}>
                <option value="" disabled>Choose a model</option>
                ${matching.slice(0, 50).map((option) => fe`<option value=${option.label || option.id}>${option.label || option.id}</option>`)}
            </select></label>
            ${matching.length > 50 && fe`<p>Showing 50 of ${matching.length} models. Refine the filter.</p>`}
            ${matching.length === 0 && fe`<p>No matching models.</p>`}
            ${blocked && fe`<p role="status">This model cannot fit the measured context. Compact the session before changing models.</p>`}
            <button disabled=${busy || !selected || blocked || chosen === data.current} onClick=${apply}>${busy ? "Applying…" : "Apply model"}</button>
            ${notice && fe`<p role="status">${notice}</p>`}
        `}
    </section>`;
}
export {
  Models
};

//# debugId=2C01FA557241079164756E2164756E21
//# sourceMappingURL=gi-settings-models-1fcfs9ee.js.map

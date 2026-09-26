import {
  F_,
  K_,
  W_,
  Q_,
  fe,
  subscribeModelSettlement,
  getAgentModels,
  selectAgentModel,
  modelContextBlocked
} from "./app-mbtajdwk.js";

// web/src/gi-settings-models.ts
function Models({ chatJid, filter = "", onMutationStart, onMutationEnd, onApplied }) {
  const [data, setData] = F_(null);
  const [chosen, setChosen] = F_("");
  const [error, setError] = F_("");
  const [notice, setNotice] = F_("");
  const [busy, setBusy] = F_(false);
  const [attempt, setAttempt] = F_(0);
  const [reading, setReading] = F_(true);
  const [readError, setReadError] = F_("");
  const generation = Q_(0);
  const readPending = Q_(true);
  const dirty = Q_(false);
  const mounted = Q_(false);
  const saving = Q_(false);
  const previousFilter = Q_(filter);
  W_(() => {
    mounted.current = true;
    const unsubscribe = subscribeModelSettlement(chatJid, () => {
      generation.current++;
      readPending.current = true;
      setReading(true);
      setAttempt((value) => value + 1);
    });
    return () => {
      mounted.current = false;
      generation.current++;
      unsubscribe();
    };
  }, [chatJid]);
  W_(() => {
    if (previousFilter.current !== filter) {
      previousFilter.current = filter;
      dirty.current = true;
      setChosen("");
      setNotice("");
    }
  }, [filter]);
  K_(() => {
    if (saving.current)
      return;
    const request = ++generation.current;
    readPending.current = true;
    setReading(true);
    setReadError("");
    getAgentModels(chatJid).then((snapshot) => {
      if (mounted.current && request === generation.current) {
        setData(snapshot);
        if (!dirty.current)
          setChosen(snapshot.current);
        readPending.current = false;
        setReading(false);
      }
    }).catch((error) => {
      if (mounted.current && request === generation.current) {
        setReadError(error.message);
        setReading(false);
      }
    });
    return () => {
      if (request === generation.current)
        generation.current++;
    };
  }, [chatJid, attempt, busy]);
  function refresh() {
    generation.current++;
    readPending.current = true;
    setReading(true);
    setAttempt((value) => value + 1);
  }
  const options = data?.model_options || data?.models || [];
  const matching = options.filter((option) => `${option.label || option.id} ${option.provider || ""}`.toLowerCase().includes(filter.trim().toLowerCase()));
  const selected = options.find((option) => (option.label || option.id) === chosen);
  const blocked = modelContextBlocked({ contextWindow: selected?.context_window ?? selected?.contextWindow }, data?.context_usage);
  async function apply() {
    if (saving.current || readPending.current || !data || !selected || blocked || chosen === data.current)
      return;
    saving.current = true;
    setBusy(true);
    setError("");
    setNotice("");
    const token = onMutationStart();
    try {
      const result = await selectAgentModel(chatJid, chosen);
      if (mounted.current) {
        dirty.current = false;
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
        ${reading && fe`<p role="status">${data ? "Refreshing models…" : "Loading models…"}</p>`}
        ${readError && fe`<div role="alert">${readError} <button disabled=${busy || reading} onClick=${refresh}>Retry</button></div>`}
        ${error && fe`<div role="alert">${error}</div>`}
        <button disabled=${busy || reading} onClick=${refresh}>Refresh models</button>
        ${data && fe`
            <dl class="gi-settings-values"><dt>Current model</dt><dd data-testid="settings-current-model">${data.current}</dd>
            <dt>Thinking (read-only)</dt><dd>${data.thinking_level || "Unknown"}</dd>
            <dt>Context capacity</dt><dd data-testid="settings-context-capacity">${Number.isFinite(data.context_window) && data.context_window > 0 ? data.context_window : "Unknown"}</dd></dl>
            <label>Session model<select aria-label="Session model" value=${chosen} disabled=${busy} onChange=${(e) => {
    dirty.current = true;
    setChosen(e.target.value);
    setNotice("");
  }}>
                <option value="" disabled>Choose a model</option>
                ${matching.slice(0, 50).map((option) => fe`<option value=${option.label || option.id}>${option.label || option.id}</option>`)}
            </select></label>
            ${matching.length > 50 && fe`<p>Showing 50 of ${matching.length} models. Refine the filter.</p>`}
            ${matching.length === 0 && fe`<p>No matching models.</p>`}
            ${blocked && fe`<p role="status">This model cannot fit the measured context. Compact the session before changing models.</p>`}
            <button disabled=${busy || reading || !!readError || !selected || blocked || chosen === data.current} onClick=${apply}>${busy ? "Applying…" : "Apply model"}</button>
            ${notice && fe`<p role="status">${notice}</p>`}
        `}
    </section>`;
}
export {
  Models
};

//# debugId=1F115AFE0558409064756E2164756E21
//# sourceMappingURL=gi-settings-models-hdggfpb1.js.map

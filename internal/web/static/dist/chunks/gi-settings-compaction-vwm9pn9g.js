import {
  F_,
  K_,
  Q_,
  fe,
  getAgentStatus,
  getSessionCompaction,
  compactSession,
  cancelSessionRun,
  getGiCompactionPolicy,
  saveGiCompactionPolicy,
  compactionNotice,
  compactionElapsed
} from "./app-pzycpm5c.js";

// web/src/gi-settings-compaction-policy.ts
var fields = [
  ["context_window", "Saved context window"],
  ["reserve_tokens", "Saved reserved tokens"],
  ["keep_recent_tokens", "Saved keep recent tokens"],
  ["threshold_tokens", "Saved trigger threshold"]
];
var toDraft = (policy) => ({ enabled: policy.enabled, ...Object.fromEntries(fields.map(([key]) => [key, String(policy[key])])) });
function GiSettingsCompactionPolicy() {
  const [snapshot, setSnapshot] = F_(null);
  const [draft, setDraft] = F_({});
  const [error, setError] = F_("");
  const [notice, setNotice] = F_("");
  const [busy, setBusy] = F_(false);
  const [attempt, setAttempt] = F_(0);
  const alive = Q_(false), saving = Q_(false);
  K_(() => {
    alive.current = true;
    return () => {
      alive.current = false;
    };
  }, []);
  K_(() => {
    let live = true;
    setSnapshot(null);
    setError("");
    setNotice("");
    getGiCompactionPolicy().then((value) => {
      if (live) {
        setSnapshot(value);
        setDraft(toDraft(value.saved.policy));
      }
    }).catch((err) => {
      if (live)
        setError(err.message);
    });
    return () => {
      live = false;
    };
  }, [attempt]);
  async function save() {
    if (!snapshot || saving.current)
      return;
    setError("");
    setNotice("");
    const numeric = Object.fromEntries(fields.map(([key]) => [key, Number(draft[key])]));
    const { context_window: window, reserve_tokens: reserve, keep_recent_tokens: keep, threshold_tokens: threshold } = numeric;
    if (fields.some(([key]) => !/^\d+$/.test(draft[key])) || Object.values(numeric).some((n) => !Number.isSafeInteger(n) || n < 1) || window > 16777216 || reserve >= window || keep > threshold || threshold > window - reserve) {
      setError("Use positive whole-number budgets: context at most 16777216; reserve below context; keep recent ≤ threshold ≤ context minus reserve.");
      return;
    }
    saving.current = true;
    setBusy(true);
    try {
      const result = await saveGiCompactionPolicy({ revision: snapshot.saved.revision, enabled: draft.enabled, ...numeric });
      if (alive.current) {
        setSnapshot(result);
        setDraft(toDraft(result.saved.policy));
        setNotice(result.restart_required ? "Policy saved. Restart Gi manually to activate it." : "Policy saved. Active policy already matches.");
      }
    } catch (err) {
      if (alive.current)
        setError(err.message);
    } finally {
      saving.current = false;
      if (alive.current)
        setBusy(false);
    }
  }
  return fe`<section aria-label="Saved automatic policy">
        <h3>Saved automatic policy</h3>
        <p>Instance-wide · .pi/settings.json. Saving changes the next startup only; manual actions still use the active engine policy above.</p>
        ${!snapshot && !error && fe`<p role="status">Loading saved policy…</p>`}
        ${snapshot && fe`
            <label><input class="gi-policy-checkbox" type="checkbox" aria-label="Saved automatic compaction" checked=${draft.enabled} disabled=${busy} onChange=${(e) => {
    setDraft((d) => ({ ...d, enabled: e.target.checked }));
    setNotice("");
  }} />Automatic compaction after restart</label>
            ${fields.map(([key, label]) => fe`<label>${label}<input type="number" min="1" max="16777216" step="1" aria-label=${label} disabled=${busy} value=${draft[key]} onInput=${(e) => {
    setDraft((d) => ({ ...d, [key]: e.target.value }));
    setNotice("");
  }} /></label>`)}
            <p>Strategy label is preserved: ${snapshot.saved.policy.strategy || "default"}. No provider model or remote compaction setting is changed.</p>
            ${snapshot.restart_required && fe`<p data-testid="compaction-policy-restart">Restart required to activate the saved policy.</p>`}
            <button disabled=${busy} onClick=${save}>${busy ? "Saving policy…" : "Save policy"}</button>
        `}
        <button disabled=${busy} onClick=${() => setAttempt((x) => x + 1)}>Reload saved policy</button>
        ${error && fe`<p role="alert">${error}</p>`}
        ${notice && fe`<p role="status">${notice}</p>`}
    </section>`;
}

// web/src/gi-settings-compaction.ts
function GiSettingsCompaction({ chatJid }) {
  const [snapshot, setSnapshot] = F_(null);
  const [error, setError] = F_("");
  const [notice, setNotice] = F_("");
  const [busy, setBusy] = F_(false);
  const [loading, setLoading] = F_(true);
  const [fresh, setFresh] = F_(false);
  const [now, setNow] = F_(Date.now());
  const alive = Q_(false);
  const reading = Q_(false);
  const mutating = Q_(false);
  const revision = Q_(0);
  const readError = Q_(false);
  const [acceptedRun, setAcceptedRun] = F_("");
  async function refresh() {
    if (reading.current || !alive.current)
      return;
    const version = revision.current;
    reading.current = true;
    try {
      const [capability, activity] = await Promise.all([getSessionCompaction(chatJid), getAgentStatus("", chatJid)]);
      if (alive.current && version === revision.current) {
        setSnapshot({ capability, activity });
        setFresh(!mutating.current);
        if (readError.current) {
          readError.current = false;
          setError("");
        }
      }
    } catch (err) {
      if (alive.current && version === revision.current) {
        readError.current = true;
        setError(`Refresh failed: ${err.message}`);
        setFresh(false);
      }
    } finally {
      reading.current = false;
      if (alive.current)
        setLoading(false);
    }
  }
  K_(() => {
    alive.current = true;
    refresh();
    const timer = setInterval(() => {
      setNow(Date.now());
      refresh();
    }, 1000);
    return () => {
      alive.current = false;
      ++revision.current;
      clearInterval(timer);
    };
  }, []);
  const capability = snapshot?.capability;
  const activity = snapshot?.activity;
  const policy = capability?.policy;
  const occurrence = activity?.compaction?.turn_id === activity?.turn_id ? activity?.compaction : null;
  const noticeState = compactionNotice(activity, now);
  const progress = occurrence ? occurrence.active && noticeState ? {
    active: true,
    title: noticeState.title,
    elapsed: compactionElapsed(noticeState, now)
  } : {
    active: false,
    label: { "compaction.completed": "Context compacted", "compaction.cancelled": "Compaction cancelled", "compaction.failed": "Compaction failed", "compaction.suppressed": "Compaction temporarily suppressed" }[occurrence.event_type] || "Compaction inactive",
    detail: occurrence.detail || ""
  } : null;
  const active = fresh && progress?.active && activity?.turn_id && ["running", "cancelling"].includes(activity.status);
  const available = fresh && capability?.available && activity?.status === "idle" && !busy;
  async function act(kind) {
    if (mutating.current || !alive.current || !fresh)
      return;
    if (kind === "compact" ? !available : !active || activity.status === "cancelling")
      return;
    const token = capability.token, turn = activity.turn_id;
    mutating.current = true;
    ++revision.current;
    readError.current = false;
    setBusy(true);
    setFresh(false);
    setError("");
    setNotice("");
    try {
      if (kind === "compact") {
        const result = await compactSession(chatJid, token);
        if (alive.current) {
          setAcceptedRun(result.turn_id);
          setNotice(`Compaction accepted for ${result.turn_id}; waiting for authoritative progress.`);
        }
      } else {
        await cancelSessionRun(chatJid, turn);
        if (alive.current) {
          setAcceptedRun(turn);
          setNotice(`Cancellation requested for ${turn}; waiting for authoritative state.`);
        }
      }
    } catch (err) {
      if (alive.current) {
        readError.current = false;
        setError(`${kind === "compact" ? "Compact" : "Stop"} failed: ${err.message}`);
      }
    } finally {
      mutating.current = false;
      ++revision.current;
      if (alive.current) {
        setBusy(false);
        refresh();
      }
    }
  }
  return fe`<section aria-labelledby="gi-compaction-title">
        <h2 id="gi-compaction-title">Compaction</h2>
        <p>Session actions · <code>${chatJid}</code></p>
        <p>Active automatic policy is read-only and loaded at startup from .pi/settings.json. Saved policy below takes effect only after a manual restart.</p>
        ${policy && fe`<dl class="gi-settings-values" data-testid="compaction-policy">
            <dt>Automatic compaction</dt><dd>${policy.enabled ? "Enabled" : "Disabled"}</dd>
            <dt>Context window</dt><dd>${policy.context_window}</dd>
            <dt>Trigger threshold</dt><dd>${policy.threshold_tokens}</dd>
            <dt>Reserved tokens</dt><dd>${policy.reserve_tokens}</dd>
            <dt>Keep recent tokens</dt><dd>${policy.keep_recent_tokens}</dd>
            <dt>Strategy label</dt><dd>${policy.strategy || "Unspecified"}</dd>
        </dl>`}
        <p>Compact now runs Gi's local context compactor (including configured hooks). It keeps timeline messages and does not submit your draft. Stop turn cancels the displayed turn, including any response after automatic compaction.</p>
        ${loading && fe`<p role="status">Loading compaction…</p>`}
        ${snapshot && fe`<p data-testid="compaction-capability">${!fresh ? "State unavailable; refresh before acting." : capability.available ? "Manual compaction available." : capability.reason || "Manual compaction unavailable."}</p>`}
        ${progress && fe`<p data-testid="settings-compaction-progress">${progress.active ? `${progress.title} — ${progress.elapsed}` : progress.label}${activity?.compaction?.turn_id ? ` · ${activity.compaction.turn_id}` : ""}${progress.detail ? ` · ${progress.detail}` : ""}</p>`}
        ${acceptedRun && fe`<p>Action turn: <code>${acceptedRun}</code></p>`}
        ${notice && fe`<p role="status">${acceptedRun && fresh && activity?.compaction?.turn_id === acceptedRun && progress && !progress.active ? `Authoritative state: ${progress.label}.` : notice}</p>`}
        ${error && fe`<p role="alert">${error}</p>`}
        <button disabled=${busy || reading.current} onClick=${() => {
    readError.current = false;
    setError("");
    setNotice("");
    setFresh(false);
    refresh();
  }}>Refresh compaction</button>
        <button disabled=${!available} onClick=${() => act("compact")}>${busy ? "Working…" : "Compact now"}</button>
        ${active && fe`<button disabled=${busy || activity.status === "cancelling"} onClick=${() => act("stop")}>Stop turn</button>`}
        <${GiSettingsCompactionPolicy} />
    </section>`;
}
export {
  GiSettingsCompaction
};

//# debugId=9EE789BB5CB57A2664756E2164756E21
//# sourceMappingURL=gi-settings-compaction-vwm9pn9g.js.map

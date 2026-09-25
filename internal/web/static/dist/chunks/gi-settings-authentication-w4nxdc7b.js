import {
  F_,
  K_,
  W_,
  Q_,
  fe,
  authJSON,
  passkeyUnavailable,
  runPasskey,
  parseAuthPolicy
} from "./app-14c0x7hx.js";

// web/src/gi-settings-authentication.ts
var removalReasons = new Map([
  ["other-rp", "The other passkey cannot sign in here because it is registered for another relying party."],
  ["policy-excludes-totp", "No other sign-in method is accepted by the current policy. Configured TOTP is not accepted in passkey-only mode."],
  ["totp-not-enabled", "TOTP is not enabled and cannot be used for sign-in. An unverified setup is not a sign-in method."],
  ["sessions-not-factors", "An active session is not a future sign-in method. Add another passkey before removing this key."],
  ["no-other-method", "Add another sign-in method before removing this key. No other accepted factor is configured."]
]);
function GiSettingsAuthentication() {
  const [policy, setPolicy] = F_(null);
  const [proof, setProof] = F_(null);
  const [loginPolicy, setLoginPolicy] = F_(null);
  const [policyChoice, setPolicyChoice] = F_("either");
  const [confirmPolicy, setConfirmPolicy] = F_(false);
  const [keys, setKeys] = F_(null);
  const [fresh, setFresh] = F_(false);
  const [busy, setBusy] = F_("");
  const [error, setError] = F_("");
  const [notice, setNotice] = F_("");
  const [removalDetail, setRemovalDetail] = F_("");
  const [name, setName] = F_("");
  const [code, setCode] = F_("");
  const [editing, setEditing] = F_(null);
  const [removing, setRemoving] = F_(null);
  const [now, setNow] = F_(Date.now());
  K_(() => {
    const timer = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(timer);
  }, []);
  const live = Q_(true), flight = Q_(null), root = Q_(null), opener = Q_(null);
  const restore = () => requestAnimationFrame(() => {
    if (live.current)
      (opener.current?.isConnected ? opener.current : root.current?.querySelector("button"))?.focus();
  });
  const cancel = () => {
    if (flight.current) {
      flight.current.abort();
      return;
    }
    setEditing(null);
    setRemoving(null);
    setConfirmPolicy(false);
    restore();
  };
  W_(() => {
    const el = root.current;
    el.addEventListener("gi-auth-escape", cancel);
    return () => el.removeEventListener("gi-auth-escape", cancel);
  });
  const refresh = async (signal) => {
    setFresh(false);
    const p = parseAuthPolicy(await authJSON("/api/auth/status", undefined, signal));
    if (!live.current)
      return;
    setPolicy(p);
    if (!p.enrolled) {
      setProof(null);
      setKeys(null);
      setLoginPolicy(null);
      return;
    }
    const [nextProof, settings] = await Promise.all([
      authJSON("/api/auth/session/proof", undefined, signal),
      authJSON("/api/auth/policy", undefined, signal)
    ]);
    const list = settings.passkey_configured ? await authJSON("/api/auth/passkeys", undefined, signal) : null;
    if (list && !Array.isArray(list.passkeys) || typeof nextProof.reauth_required !== "boolean" || typeof settings.revision !== "string" || !["either", "totp-only", "passkey-only"].includes(settings.policy))
      throw new Error("Invalid authentication response");
    if (live.current) {
      setProof(nextProof);
      setLoginPolicy(settings);
      setPolicyChoice(settings.policy);
      setConfirmPolicy(false);
      setKeys(list?.passkeys ?? null);
      setFresh(true);
    }
  };
  const work = async (label, action, message = "") => {
    if (flight.current)
      return;
    const controller = new AbortController;
    flight.current = controller;
    const trigger = document.activeElement;
    const triggerAction = trigger?.getAttribute("data-auth-action");
    setBusy(label);
    setError("");
    setNotice("");
    setRemovalDetail("");
    try {
      const detail = await action(controller.signal);
      if (live.current) {
        await refresh(controller.signal);
        if (live.current) {
          if (message)
            setNotice(message);
          if (typeof detail === "string")
            setRemovalDetail(detail);
        }
      }
    } catch (e) {
      if (live.current) {
        setFresh(false);
        const reason = label === "Removing passkey…" ? removalReasons.get(e.reason) : "";
        setError([e.message || "Request failed. Refresh before retrying.", reason].filter(Boolean).join(" "));
      }
    } finally {
      if (flight.current === controller) {
        flight.current = null;
        if (live.current) {
          setBusy("");
          requestAnimationFrame(() => {
            if (!live.current || !root.current?.closest(".settings-dialog"))
              return;
            const replacement = triggerAction ? root.current.querySelector(`[data-auth-action="${CSS.escape(triggerAction)}"]:not(:disabled)`) : null;
            const target = trigger?.isConnected && !trigger.hasAttribute("disabled") ? trigger : replacement || root.current.querySelector("button:not(:disabled)");
            if (target && (root.current.contains(document.activeElement) || document.activeElement === document.body))
              target.focus({ preventScroll: true });
          });
        }
      }
    }
  };
  K_(() => {
    work("Loading passkeys…", async () => {});
    return () => {
      live.current = false;
      flight.current?.abort();
      flight.current = null;
    };
  }, []);
  const recentlyVerified = fresh && proof && !proof.reauth_required && Date.parse(proof.fresh_until) > now;
  const unavailable = !policy?.enrolled ? "Authentication must be configured first." : !policy.passkeys_enabled ? "Passkeys are disabled by policy or are not configured for this origin." : passkeyUnavailable();
  const add = () => work("Waiting for passkey creation…", async (signal) => {
    await runPasskey("register", signal, name);
    if (live.current)
      setName("");
  }, "Passkey registered.");
  const mutation = (kind, row, newName) => work(kind === "rename" ? "Saving name…" : "Removing passkey…", async (signal) => {
    const result = await authJSON(`/api/auth/passkeys/${kind}`, { id: row.id, ...kind === "rename" ? { name: newName } : {} }, signal);
    if (result.ok !== true)
      throw new Error("Change could not be confirmed. Refresh before trying again.");
    if (live.current) {
      setEditing(null);
      setRemoving(null);
      restore();
    }
    if (kind === "remove") {
      if (result.remaining_method === "totp")
        return "TOTP remains available for sign-in.";
      if (result.remaining_method === "passkey")
        return "Another passkey remains available for sign-in.";
    }
  }, kind === "rename" ? "Name saved." : "Passkey removed. Existing login sessions are not signed out.");
  const policyAllowed = loginPolicy && (policyChoice !== "passkey-only" && loginPolicy.totp_configured || policyChoice !== "totp-only" && loginPolicy.passkey_usable);
  const savePolicy = () => work("Saving sign-in policy…", async (signal) => {
    const result = await authJSON("/api/auth/policy", { policy: policyChoice, revision: loginPolicy.revision }, signal);
    if (result.policy !== policyChoice || typeof result.revision !== "string")
      throw new Error("Policy change could not be confirmed. Refresh before trying again.");
    if (live.current) {
      setConfirmPolicy(false);
      restore();
    }
  }, "Sign-in policy saved. Existing login sessions are unchanged.");
  const date = (value) => !value || value.startsWith("0001-") ? "Never used" : new Date(value).toLocaleString();
  return fe`<section ref=${root} class="gi-authentication-pane" aria-labelledby="gi-authentication-heading" data-auth-escape=${busy || editing || removing || confirmPolicy ? "true" : undefined}>
        <h2 id="gi-authentication-heading">Authentication</h2>
        <p>Manage passkeys for this instance owner. Each row is one credential, not an inventory of devices.</p>
        <h3>Passkeys</h3>
        ${busy && fe`<p role="status">${busy}</p>`}
        ${error && fe`<p role="alert">${error}</p>`}
        ${notice && fe`<p role="status">${notice}</p>`}
        ${removalDetail && fe`<p role="status">${removalDetail}</p>`}
        <button disabled=${!!busy} onClick=${() => work("Refreshing passkeys…", async () => {})}>Refresh passkeys</button>
        ${busy && fe`<button onClick=${cancel}>Cancel pending operation</button>`}
        ${policy && unavailable && fe`<p>${unavailable}</p>`}
        ${keys && !fresh && fe`<p role="status">The displayed list is the last confirmed snapshot. Refresh before making changes.</p>`}
        ${policy?.enrolled && fe`<div class="gi-passkey-proof">
            <p>${recentlyVerified ? "Recently authenticated for credential changes." : "Verify an accepted factor before changing passkeys."}</p>
            ${policy.totp_login_available && fe`<label>Authentication code<input aria-label="Reauthentication code" type="text" inputMode="numeric" autoComplete="one-time-code" value=${code} disabled=${!!busy} onInput=${(e) => setCode(e.target.value)} /></label>
                <button data-auth-action="verify-totp" disabled=${!!busy || !/^\d{6}$/.test(code)} onClick=${() => work("Verifying authentication…", async (signal) => {
    await authJSON("/api/auth/session/reauth/totp", { code }, signal);
    if (live.current)
      setCode("");
  }, "Authentication verified.")}>Verify code</button>`}
            ${policy.passkey_login_available && fe`<button data-auth-action="verify-passkey" disabled=${!!busy || !!passkeyUnavailable()} onClick=${() => work("Waiting for passkey verification…", (signal) => runPasskey("reauth", signal), "Authentication verified.")}>Verify with passkey</button>`}
        </div>`}
        ${loginPolicy && fe`<div class="gi-signin-policy"><h3>Sign-in policy</h3>
            <p>Current policy: ${loginPolicy.policy}. Changing accepted factors does not remove credentials or sign out existing sessions.</p>
            <label>Accepted sign-in methods<select aria-label="Accepted sign-in methods" value=${policyChoice} disabled=${!!busy || !fresh} onChange=${(e) => {
    setPolicyChoice(e.target.value);
    setConfirmPolicy(false);
  }}>
                <option value="either">TOTP or passkey</option><option value="totp-only">TOTP only</option><option value="passkey-only">Passkeys only</option></select></label>
            ${!policyAllowed && fe`<p>The selected policy needs a configured, usable sign-in method. Enrol a passkey for this origin or keep verified TOTP enabled.</p>`}
            <button disabled=${!!busy || !recentlyVerified || !policyAllowed || policyChoice === loginPolicy.policy} onClick=${(e) => {
    opener.current = e.currentTarget;
    setConfirmPolicy(true);
    setRemoving(null);
    setEditing(null);
  }}>Change sign-in policy</button>
            ${confirmPolicy && fe`<div role="group" aria-label="Confirm sign-in policy"><p>Use ${policyChoice} for future sign-ins? Existing sessions and stored factors will not be removed.</p>
                <button disabled=${!!busy || !recentlyVerified || !policyAllowed} onClick=${savePolicy}>Confirm policy change</button><button disabled=${!!busy} onClick=${cancel}>Cancel policy change</button></div>`}
        </div>`}
        <div class="gi-passkey-add"><label>Passkey name<input aria-label="New passkey name" value=${name} disabled=${!!busy || !!unavailable} onInput=${(e) => setName(e.target.value)} /></label>
            <button disabled=${!!busy || !!unavailable || !recentlyVerified || !name.trim()} onClick=${add}>Add passkey</button></div>
        ${keys && keys.length === 0 && fresh && fe`<p>No passkeys registered.</p>`}
        ${keys && fe`<ul class="gi-passkey-list">${keys.map((row) => fe`<li key=${row.id} class="gi-passkey-row" data-credential-id=${row.id}>
            <strong>${row.name || "Unnamed passkey"}</strong><small>Identifier: ${row.id}</small>
            <small>Created: ${date(row.created_at)}</small><small>Last used: ${date(row.last_used_at)}</small>
            <div><button disabled=${!!busy || !recentlyVerified || !!unavailable} onClick=${(e) => {
    opener.current = e.currentTarget;
    setRemoving(null);
    setEditing({ id: row.id, name: row.name });
  }}>Rename ${row.name || "passkey"}</button>
                <button disabled=${!!busy || !recentlyVerified || !!unavailable} onClick=${(e) => {
    opener.current = e.currentTarget;
    setEditing(null);
    setRemoving(row.id);
  }}>Remove ${row.name || "passkey"}</button></div>
            ${editing?.id === row.id && fe`<div class="gi-passkey-edit"><label>New name<input aria-label="Rename passkey" value=${editing.name} disabled=${!!busy} onInput=${(e) => setEditing({ ...editing, name: e.target.value })} /></label>
                <button disabled=${!!busy || !recentlyVerified || !!unavailable} onClick=${() => mutation("rename", row, editing.name)}>Save passkey name</button><button disabled=${!!busy} onClick=${cancel}>Cancel rename</button></div>`}
            ${removing === row.id && fe`<div class="gi-passkey-confirm" role="group" aria-label="Confirm passkey removal"><p>Remove ${row.name || "passkey"} (${row.id})? This blocks future sign-ins with this credential. Existing login sessions are not signed out.</p>
                <button disabled=${!!busy || !recentlyVerified || !!unavailable} onClick=${() => mutation("remove", row)}>Confirm removal</button><button disabled=${!!busy} onClick=${cancel}>Cancel removal</button></div>`}
        </li>`)}</ul>`}
    </section>`;
}
export {
  GiSettingsAuthentication
};

//# debugId=D693D4B8B060CA2E64756E2164756E21
//# sourceMappingURL=gi-settings-authentication-w4nxdc7b.js.map

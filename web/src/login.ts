// @ts-nocheck
/**
 * login.ts – Web login page behavior (TOTP + optional passkey login).
 *
 * This module is bundled to /static/dist/login.bundle.js and loaded by
 * web/static/login.html. It intentionally has no framework/runtime deps.
 */

import { probePasskeyCapabilityBestEffort, readJsonBodyBestEffort, runPasskeyAttemptBestEffort } from './login-safety.js';

const form = document.getElementById("login-form");
const codeInput = document.getElementById("code");
const errorEl = document.getElementById("error");

if (!form || !codeInput || !errorEl) {
  throw new Error("Login form markup is missing required elements.");
}

const base64UrlToBuffer = (value: string): Uint8Array => {
  const pad = "=".repeat((4 - (value.length % 4)) % 4);
  const base64 = (value + pad).replace(/-/g, "+").replace(/_/g, "/");
  const raw = atob(base64);
  const buffer = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i += 1) buffer[i] = raw.charCodeAt(i);
  return buffer;
};

const bufferToBase64Url = (buffer: ArrayBufferLike): string => {
  const bytes = new Uint8Array(buffer);
  let binary = "";
  for (const b of bytes) binary += String.fromCharCode(b);
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
};

const credentialToJSON = (cred: PublicKeyCredential) => ({
  id: cred.id,
  rawId: bufferToBase64Url(cred.rawId),
  type: cred.type,
  response: {
    clientDataJSON: bufferToBase64Url((cred.response as AuthenticatorAssertionResponse).clientDataJSON),
    authenticatorData: bufferToBase64Url((cred.response as AuthenticatorAssertionResponse).authenticatorData),
    signature: bufferToBase64Url((cred.response as AuthenticatorAssertionResponse).signature),
    userHandle: (cred.response as AuthenticatorAssertionResponse).userHandle
      ? bufferToBase64Url((cred.response as AuthenticatorAssertionResponse).userHandle as ArrayBuffer)
      : null,
  },
});

type LoginAllowCredential = {
  id: string;
  type?: string;
  transports?: string[];
};

type LoginOptionsPayload = {
  challenge: string;
  allowCredentials?: LoginAllowCredential[];
  [key: string]: unknown;
};

const parseLoginOptions = (options: LoginOptionsPayload) => ({
  ...options,
  challenge: base64UrlToBuffer(options.challenge),
  allowCredentials: (options.allowCredentials || []).map((cred: LoginAllowCredential) => ({
    ...cred,
    id: base64UrlToBuffer(cred.id),
  })),
});

let passkeyInFlight = false;
let passkeySucceeded = false;

const attemptPasskey = async ({ conditional }: { conditional: boolean }) => {
  if (!window.PublicKeyCredential || !navigator.credentials) return;
  if (passkeyInFlight || passkeySucceeded) return;
  passkeyInFlight = true;
  const succeeded = await runPasskeyAttemptBestEffort(async () => {
    const res = await fetch("/auth/webauthn/login/start", { method: "POST" });
    if (!res.ok) {
      passkeyInFlight = false;
      return false;
    }

    const payload = await res.json();
    const publicKey = parseLoginOptions(payload.options);
    const cred = (await navigator.credentials.get({
      publicKey,
      mediation: conditional ? "conditional" : "required",
    })) as PublicKeyCredential | null;

    if (!cred) {
      passkeyInFlight = false;
      return false;
    }

    const finish = await fetch("/auth/webauthn/login/finish", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token: payload.token, credential: credentialToJSON(cred) }),
    });

    if (finish.ok) {
      passkeySucceeded = true;
      window.location.href = "/";
      return true;
    }
    return false;
  });
  if (succeeded) return;
  passkeyInFlight = false;
};

if (window.PublicKeyCredential && typeof PublicKeyCredential.isConditionalMediationAvailable === "function") {
  void probePasskeyCapabilityBestEffort(() => PublicKeyCredential.isConditionalMediationAvailable())
    .then((available) => {
      if (available) attemptPasskey({ conditional: true });
    });
}

codeInput.addEventListener("focus", () => {
  void attemptPasskey({ conditional: false });
});

const submitCode = async () => {
  const code = (codeInput as HTMLInputElement).value?.trim() || "";
  errorEl.textContent = "";

  if (!code) {
    errorEl.textContent = "Please enter your code.";
    return;
  }

  const res = await fetch("/auth/verify", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ code }),
  });

  if (res.ok) {
    window.location.href = "/";
    return;
  }

  const payload = await readJsonBodyBestEffort(res, {} as Record<string, any>);
  errorEl.textContent = payload.error || "Invalid code. Try again.";
};

form.addEventListener("submit", (e) => {
  e.preventDefault();
  void submitCode();
});

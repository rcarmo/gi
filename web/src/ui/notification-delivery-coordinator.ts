const DEVICE_ID_KEY = 'piclaw.notifications.deviceId';
const CLIENT_ID_KEY = 'piclaw.notifications.clientId';
const PRESENCE_KEY_PREFIX = 'piclaw.notifications.presence.';
export const LOCAL_NOTIFICATION_PRESENCE_TTL_MS = 120000;

function safeStorageGet(storage, key) {
  if (!storage || !key) return null;
  try {
    return storage.getItem(key);
  } catch {
    return null;
  }
}

function safeStorageSet(storage, key, value) {
  if (!storage || !key) return;
  try {
    storage.setItem(key, value);
  } catch {
    return;
  }
}

function safeStorageRemove(storage, key) {
  if (!storage || !key) return;
  try {
    storage.removeItem(key);
  } catch {
    return;
  }
}

function createRandomId(prefix = 'piclaw') {
  try {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
      return `${prefix}-${crypto.randomUUID()}`;
    }
  } catch (error) {
    console.debug('[notification-delivery-coordinator] crypto.randomUUID threw; falling back to Math.random-based id.', error);
  }
  return `${prefix}-${Math.random().toString(36).slice(2)}-${Date.now().toString(36)}`;
}

export function getOrCreateNotificationDeviceId(runtimeWindow = typeof window !== 'undefined' ? window : null) {
  const storage = runtimeWindow?.localStorage;
  const existing = safeStorageGet(storage, DEVICE_ID_KEY);
  if (existing) return existing;
  const created = createRandomId('device');
  safeStorageSet(storage, DEVICE_ID_KEY, created);
  return safeStorageGet(storage, DEVICE_ID_KEY) || created;
}

export function getOrCreateNotificationClientId(runtimeWindow = typeof window !== 'undefined' ? window : null) {
  const sessionStorage = runtimeWindow?.sessionStorage;
  const existing = safeStorageGet(sessionStorage, CLIENT_ID_KEY);
  if (existing) return existing;
  const fallbackExisting = runtimeWindow?.__PICLAW_NOTIFICATION_CLIENT_ID__;
  if (typeof fallbackExisting === 'string' && fallbackExisting.trim()) return fallbackExisting.trim();
  const created = createRandomId('client');
  safeStorageSet(sessionStorage, CLIENT_ID_KEY, created);
  if (runtimeWindow) {
    runtimeWindow.__PICLAW_NOTIFICATION_CLIENT_ID__ = safeStorageGet(sessionStorage, CLIENT_ID_KEY) || created;
  }
  return runtimeWindow?.__PICLAW_NOTIFICATION_CLIENT_ID__ || created;
}

function getPresenceStorageKey(deviceId, clientId) {
  return `${PRESENCE_KEY_PREFIX}${String(deviceId || '').trim()}:${String(clientId || '').trim()}`;
}

export function createLocalNotificationPresenceSnapshot(options = {}) {
  const runtimeWindow = options.runtimeWindow ?? (typeof window !== 'undefined' ? window : null);
  const runtimeDocument = options.runtimeDocument ?? (typeof document !== 'undefined' ? document : null);
  const chatJid = typeof options.chatJid === 'string' && options.chatJid.trim() ? options.chatJid.trim() : '';
  const deviceId = typeof options.deviceId === 'string' && options.deviceId.trim()
    ? options.deviceId.trim()
    : getOrCreateNotificationDeviceId(runtimeWindow);
  const clientId = typeof options.clientId === 'string' && options.clientId.trim()
    ? options.clientId.trim()
    : getOrCreateNotificationClientId(runtimeWindow);
  const updatedAtMs = Number.isFinite(options.updatedAtMs) ? Number(options.updatedAtMs) : Date.now();
  const hasFocus = Boolean(typeof runtimeDocument?.hasFocus === 'function' ? runtimeDocument.hasFocus() : true);
  const rawVisibility = String(runtimeDocument?.visibilityState || '').trim().toLowerCase();
  const visibilityState = rawVisibility === 'hidden' ? 'hidden' : 'visible';

  return {
    deviceId,
    clientId,
    chatJid,
    visibilityState,
    hasFocus,
    updatedAtMs,
  };
}

export function publishLocalNotificationPresence(snapshot, runtimeWindow = typeof window !== 'undefined' ? window : null) {
  const storage = runtimeWindow?.localStorage;
  const deviceId = typeof snapshot?.deviceId === 'string' ? snapshot.deviceId.trim() : '';
  const clientId = typeof snapshot?.clientId === 'string' ? snapshot.clientId.trim() : '';
  const chatJid = typeof snapshot?.chatJid === 'string' ? snapshot.chatJid.trim() : '';
  if (!storage || !deviceId || !clientId || !chatJid) return false;
  safeStorageSet(storage, getPresenceStorageKey(deviceId, clientId), JSON.stringify({
    deviceId,
    clientId,
    chatJid,
    visibilityState: snapshot.visibilityState === 'hidden' ? 'hidden' : 'visible',
    hasFocus: Boolean(snapshot.hasFocus),
    updatedAtMs: Number.isFinite(snapshot.updatedAtMs) ? Number(snapshot.updatedAtMs) : Date.now(),
  }));
  return true;
}

export function withdrawLocalNotificationPresence(value, runtimeWindow = typeof window !== 'undefined' ? window : null) {
  const storage = runtimeWindow?.localStorage;
  const deviceId = typeof value?.deviceId === 'string' ? value.deviceId.trim() : '';
  const clientId = typeof value?.clientId === 'string' ? value.clientId.trim() : '';
  if (!storage || !deviceId || !clientId) return false;
  safeStorageRemove(storage, getPresenceStorageKey(deviceId, clientId));
  return true;
}

export function listLiveLocalNotificationPresence(options = {}) {
  const runtimeWindow = options.runtimeWindow ?? (typeof window !== 'undefined' ? window : null);
  const storage = runtimeWindow?.localStorage;
  const deviceId = typeof options.deviceId === 'string' && options.deviceId.trim()
    ? options.deviceId.trim()
    : getOrCreateNotificationDeviceId(runtimeWindow);
  const nowMs = Number.isFinite(options.nowMs) ? Number(options.nowMs) : Date.now();
  const ttlMs = Number.isFinite(options.ttlMs) ? Number(options.ttlMs) : LOCAL_NOTIFICATION_PRESENCE_TTL_MS;
  if (!storage || !deviceId) return [];

  const keyPrefix = `${PRESENCE_KEY_PREFIX}${deviceId}:`;
  const live = [];
  const staleKeys = [];
  for (let index = 0; index < storage.length; index += 1) {
    const key = storage.key(index);
    if (!key || !key.startsWith(keyPrefix)) continue;
    const raw = safeStorageGet(storage, key);
    if (!raw) {
      staleKeys.push(key);
      continue;
    }
    try {
      const parsed = JSON.parse(raw);
      const updatedAtMs = Number(parsed?.updatedAtMs);
      if (!Number.isFinite(updatedAtMs) || nowMs - updatedAtMs > ttlMs) {
        staleKeys.push(key);
        continue;
      }
      const chatJid = typeof parsed?.chatJid === 'string' ? parsed.chatJid.trim() : '';
      const clientId = typeof parsed?.clientId === 'string' ? parsed.clientId.trim() : '';
      if (!chatJid || !clientId) {
        staleKeys.push(key);
        continue;
      }
      live.push({
        deviceId,
        clientId,
        chatJid,
        visibilityState: parsed?.visibilityState === 'hidden' ? 'hidden' : 'visible',
        hasFocus: Boolean(parsed?.hasFocus),
        updatedAtMs,
      });
    } catch {
      staleKeys.push(key);
    }
  }

  staleKeys.forEach((key) => safeStorageRemove(storage, key));
  return live.sort((left, right) => left.clientId.localeCompare(right.clientId));
}

export function shouldNotifyLocallyForChat(options = {}) {
  const snapshot = createLocalNotificationPresenceSnapshot(options);
  const chatJid = snapshot.chatJid;
  if (!chatJid) return false;
  const entries = listLiveLocalNotificationPresence({
    runtimeWindow: options.runtimeWindow,
    deviceId: snapshot.deviceId,
    nowMs: snapshot.updatedAtMs,
    ttlMs: options.ttlMs,
  }).filter((entry) => entry.chatJid === chatJid && entry.clientId !== snapshot.clientId);
  const candidates = [snapshot, ...entries];
  if (candidates.some((entry) => entry.visibilityState === 'visible')) {
    return false;
  }
  const leader = [...candidates].sort((left, right) => left.clientId.localeCompare(right.clientId))[0] || null;
  return Boolean(leader && leader.clientId === snapshot.clientId);
}

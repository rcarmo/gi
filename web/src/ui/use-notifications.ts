import { deleteWebPushSubscription, getWebPushPublicKey, saveWebPushSubscription } from '../api.js';
import { useCallback, useEffect, useRef, useState } from '../vendor/preact-htm.js';
import { getLocalStorageBoolean, setLocalStorageItem } from '../utils/storage.js';
import { focusWindowBestEffort } from './notification-focus.js';
import {
  createLocalNotificationPresenceSnapshot,
  getOrCreateNotificationClientId,
  getOrCreateNotificationDeviceId,
  publishLocalNotificationPresence,
  shouldNotifyLocallyForChat,
  withdrawLocalNotificationPresence,
} from './notification-delivery-coordinator.ts';

export const LOCAL_NOTIFICATION_SOURCE_LABEL = 'Local';
export const WEB_PUSH_NOTIFICATION_SOURCE_LABEL = 'Web Push';

export function shouldShowNotificationSourceLabels(runtimeWindow = typeof window !== 'undefined' ? window : null) {
  return Boolean(runtimeWindow?.__PICLAW_NOTIFICATION_SOURCE_LABELS_ENABLED__);
}

export function formatNotificationTitle(title, sourceLabel = '', runtimeWindow = typeof window !== 'undefined' ? window : null) {
  const normalizedTitle = typeof title === 'string' && title.trim() ? title.trim() : 'PiClaw';
  const normalizedSource = shouldShowNotificationSourceLabels(runtimeWindow) && typeof sourceLabel === 'string'
    ? sourceLabel.trim()
    : '';
  return normalizedSource ? `${normalizedTitle} [${normalizedSource}]` : normalizedTitle;
}

function decodeBase64UrlToUint8Array(value) {
  const normalized = String(value || '').replace(/-/g, '+').replace(/_/g, '/');
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=');
  const binary = atob(padded);
  const bytes = new Uint8Array(binary.length);
  for (let index = 0; index < binary.length; index += 1) {
    bytes[index] = binary.charCodeAt(index);
  }
  return bytes;
}

function supportsWebPush(runtimeWindow = typeof window !== 'undefined' ? window : null) {
  if (!runtimeWindow || !runtimeWindow.isSecureContext) return false;
  return 'serviceWorker' in runtimeWindow.navigator && 'PushManager' in runtimeWindow;
}

async function ensureServiceWorkerReady(runtimeWindow) {
  await runtimeWindow.navigator.serviceWorker.register('/sw.js', { updateViaCache: 'none' });
  return await runtimeWindow.navigator.serviceWorker.ready;
}

async function ensureStoredWebPushSubscription(runtimeWindow) {
  const registration = await ensureServiceWorkerReady(runtimeWindow);
  const existing = await registration.pushManager.getSubscription();
  if (existing) return existing;
  const payload = await getWebPushPublicKey();
  const publicKey = typeof payload?.publicKey === 'string' ? payload.publicKey.trim() : '';
  if (!publicKey) {
    throw new Error('Missing web push public key.');
  }
  return registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: decodeBase64UrlToUint8Array(publicKey),
  });
}

async function syncWebPushSubscription(runtimeWindow, deviceId) {
  if (!supportsWebPush(runtimeWindow)) return false;
  const subscription = await ensureStoredWebPushSubscription(runtimeWindow);
  await saveWebPushSubscription(subscription.toJSON ? subscription.toJSON() : subscription, { deviceId });
  return true;
}

async function disableWebPushSubscription(runtimeWindow, deviceId) {
  if (!supportsWebPush(runtimeWindow)) return false;
  const registration = await ensureServiceWorkerReady(runtimeWindow);
  const subscription = await registration.pushManager.getSubscription();
  if (!subscription) return false;
  const serialized = subscription.toJSON ? subscription.toJSON() : subscription;
  try {
    await deleteWebPushSubscription(serialized, { deviceId });
  } catch (error) {
    console.warn('Failed to remove web push subscription from the server:', error);
  }
  try {
    await subscription.unsubscribe();
  } catch (error) {
    console.warn('Failed to unsubscribe from web push notifications:', error);
  }
  return true;
}

function postWebPushPresence(payload, options = {}) {
  const runtimeWindow = options.runtimeWindow ?? (typeof window !== 'undefined' ? window : null);
  if (!runtimeWindow?.fetch) return Promise.resolve(false);
  return runtimeWindow.fetch('/agent/push/presence', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
    keepalive: Boolean(options.keepalive),
  }).then(() => true).catch(() => false);
}

function sendWebPushPresenceBeacon(payload, runtimeWindow = typeof window !== 'undefined' ? window : null) {
  try {
    if (runtimeWindow?.navigator?.sendBeacon) {
      const blob = new Blob([JSON.stringify(payload)], { type: 'application/json' });
      return runtimeWindow.navigator.sendBeacon('/agent/push/presence', blob);
    }
  } catch (error) {
    console.debug('[use-notifications] Ignoring sendBeacon failure for best-effort notification presence teardown.', error, {
      hasNavigator: Boolean(runtimeWindow?.navigator),
    });
  }
  return false;
}

export function useNotifications(options = {}) {
  const currentChatJid = typeof options?.chatJid === 'string' && options.chatJid.trim() ? options.chatJid.trim() : 'web:default';
  const [notificationsEnabled, setNotificationsEnabled] = useState(false);
  const [notificationPermission, setNotificationPermission] = useState('default');
  const notificationsEnabledRef = useRef(false);
  const deviceIdRef = useRef(null);
  const clientIdRef = useRef(null);

  useEffect(() => {
    const enabled = getLocalStorageBoolean('notificationsEnabled', false);
    notificationsEnabledRef.current = enabled;
    setNotificationsEnabled(enabled);
    if (typeof window !== 'undefined') {
      deviceIdRef.current = getOrCreateNotificationDeviceId(window);
      clientIdRef.current = getOrCreateNotificationClientId(window);
    }
    if (typeof Notification === 'undefined') {
      return;
    }

    const permission = Notification.permission;
    setNotificationPermission(permission);

    if (permission === 'denied' && enabled) {
      notificationsEnabledRef.current = false;
      setNotificationsEnabled(false);
      setLocalStorageItem('notificationsEnabled', 'false');
      return;
    }

    if (permission === 'granted' && enabled && typeof window !== 'undefined' && supportsWebPush(window)) {
      void syncWebPushSubscription(window, deviceIdRef.current || getOrCreateNotificationDeviceId(window)).catch((error) => {
        console.warn('Failed to refresh stored web push subscription:', error);
      });
    }
  }, []);

  useEffect(() => {
    notificationsEnabledRef.current = notificationsEnabled;
  }, [notificationsEnabled]);

  useEffect(() => {
    if (typeof window === 'undefined' || typeof document === 'undefined') return;
    const deviceId = deviceIdRef.current || getOrCreateNotificationDeviceId(window);
    const clientId = clientIdRef.current || getOrCreateNotificationClientId(window);
    deviceIdRef.current = deviceId;
    clientIdRef.current = clientId;

    const publishPresence = (active = true, transport = 'fetch') => {
      const snapshot = createLocalNotificationPresenceSnapshot({
        chatJid: currentChatJid,
        runtimeWindow: window,
        runtimeDocument: document,
        deviceId,
        clientId,
      });
      if (active) {
        publishLocalNotificationPresence(snapshot, window);
      } else {
        withdrawLocalNotificationPresence({ deviceId, clientId }, window);
      }
      const payload = {
        device_id: deviceId,
        client_id: clientId,
        chat_jid: currentChatJid,
        visibility_state: snapshot.visibilityState,
        has_focus: snapshot.hasFocus,
        active,
      };
      if (!active && transport === 'beacon' && sendWebPushPresenceBeacon(payload, window)) {
        return;
      }
      void postWebPushPresence(payload, { runtimeWindow: window, keepalive: !active || transport === 'keepalive' });
    };

    const handleStateChange = () => publishPresence(true);
    const handlePageHide = () => publishPresence(false, 'beacon');

    publishPresence(true);
    const interval = setInterval(() => publishPresence(true), 15000);
    document.addEventListener('visibilitychange', handleStateChange);
    window.addEventListener('focus', handleStateChange);
    window.addEventListener('blur', handleStateChange);
    window.addEventListener('pageshow', handleStateChange);
    window.addEventListener('pagehide', handlePageHide);
    window.addEventListener('beforeunload', handlePageHide);

    return () => {
      clearInterval(interval);
      document.removeEventListener('visibilitychange', handleStateChange);
      window.removeEventListener('focus', handleStateChange);
      window.removeEventListener('blur', handleStateChange);
      window.removeEventListener('pageshow', handleStateChange);
      window.removeEventListener('pagehide', handlePageHide);
      window.removeEventListener('beforeunload', handlePageHide);
      publishPresence(false, 'beacon');
    };
  }, [currentChatJid]);

  const requestNotificationPermission = useCallback(() => {
    if (typeof Notification === 'undefined') return Promise.resolve('denied');
    try {
      const result = Notification.requestPermission();
      if (result && typeof result.then === 'function') {
        return result;
      }
      return Promise.resolve(result);
    } catch (error) {
      console.debug('[use-notifications] Notification permission request threw; returning default permission state.', error);
      return Promise.resolve('default');
    }
  }, []);

  const toggleNotifications = useCallback(async () => {
    if (typeof window === 'undefined' || typeof Notification === 'undefined') return;
    if (!window.isSecureContext) {
      alert('Notifications require a secure context (HTTPS or installed app).');
      return;
    }
    if (Notification.permission === 'denied') {
      setNotificationPermission('denied');
      alert('Browser notifications are blocked. Enable them in your browser settings.');
      return;
    }
    if (Notification.permission === 'default') {
      const result = await requestNotificationPermission();
      setNotificationPermission(result || 'default');
      if (result !== 'granted') {
        notificationsEnabledRef.current = false;
        setNotificationsEnabled(false);
        setLocalStorageItem('notificationsEnabled', 'false');
        return;
      }
    }
    const next = !notificationsEnabledRef.current;
    notificationsEnabledRef.current = next;
    setNotificationsEnabled(next);
    setLocalStorageItem('notificationsEnabled', String(next));

    const deviceId = deviceIdRef.current || getOrCreateNotificationDeviceId(window);
    deviceIdRef.current = deviceId;

    if (supportsWebPush(window)) {
      try {
        if (next) {
          await syncWebPushSubscription(window, deviceId);
        } else {
          await disableWebPushSubscription(window, deviceId);
        }
      } catch (error) {
        console.warn('Failed to sync web push notifications:', error);
        if (next) {
          alert('Notifications were enabled, but web push setup failed. If you are on iPhone or iPad, reopen PiClaw from the Home Screen and try again.');
        }
      }
    }
  }, [requestNotificationPermission]);

  const notify = useCallback((title, body, options = {}) => {
    if (!notificationsEnabledRef.current) return false;
    if (typeof Notification === 'undefined') return false;
    if (Notification.permission !== 'granted') return false;
    try {
      const notification = new Notification(formatNotificationTitle(title, options?.sourceLabel || '', window), { body });
      notification.onclick = () => {
        focusWindowBestEffort(window);
      };
      return true;
    } catch (error) {
      console.debug('[use-notifications] Local notification creation failed.', error, {
        title: typeof title === 'string' ? title : null,
      });
      return false;
    }
  }, []);

  const shouldNotifyLocallyForChatId = useCallback((chatJid) => {
    if (typeof window === 'undefined' || typeof document === 'undefined') return false;
    return shouldNotifyLocallyForChat({
      chatJid: typeof chatJid === 'string' && chatJid.trim() ? chatJid.trim() : currentChatJid,
      runtimeWindow: window,
      runtimeDocument: document,
      deviceId: deviceIdRef.current || getOrCreateNotificationDeviceId(window),
      clientId: clientIdRef.current || getOrCreateNotificationClientId(window),
    });
  }, [currentChatJid]);

  return {
    notificationsEnabled,
    notificationPermission,
    toggleNotifications,
    notify,
    shouldNotifyLocallyForChat: shouldNotifyLocallyForChatId,
  };
}

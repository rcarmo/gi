import { useCallback, useEffect, useRef } from '../vendor/preact-htm.js';
import { setLocalStorageItem } from '../utils/storage.js';
import {
  DESKTOP_WORKSPACE_LAYOUT_MEDIA_QUERY,
  persistWorkspaceOpenPreference,
  readStoredWorkspaceOpenPreference,
  resolveWorkspaceLayoutBucket,
} from './workspace-visibility.js';
import { initTheme } from './theme.js';
import { useTimestampRefresh } from './app-helpers.js';
import { watchReturnToApp, watchStandaloneWebAppMode } from './app-resume.js';
import { installStandaloneMobileViewportFix } from './mobile-viewport.js';
import { BTW_SESSION_KEY } from './app-shell-state.js';

interface RefBox<T> {
  current: T;
}

export const RESUME_LAYOUT_SETTLING_CLASS = 'resume-layout-settling';
export const RESUME_LAYOUT_SETTLING_MS = 220;

const resumeLayoutSettlingTimers = new WeakMap<HTMLElement, ReturnType<typeof setTimeout>>();

export interface UseAppShellEnvironmentEffectsOptions {
  isRenameBranchFormOpen: boolean;
  renameBranchNameInputRef: RefBox<any>;
  appShellRef: RefBox<HTMLElement | null>;
  setIsWebAppMode: (next: boolean) => void;
  workspaceOpen: boolean;
  setWorkspaceOpen: (next: boolean) => void;
  btwSession: any;
  agents: Record<string, unknown> | null | undefined;
  agentsRef: RefBox<Record<string, unknown>>;
  currentChatJid: string;
  activeChatJidRef: RefBox<string>;
  userProfile: any;
  userProfileRef: RefBox<any>;
  brandingRef: RefBox<{ title: string | null; avatarBase: string | null }>;
  panePopoutMode?: boolean;
}

export function persistBtwSession(btwSession: any): void {
  if (!btwSession) {
    setLocalStorageItem(BTW_SESSION_KEY, '');
    return;
  }

  setLocalStorageItem(BTW_SESSION_KEY, JSON.stringify({
    question: btwSession.question || '',
    answer: btwSession.answer || '',
    thinking: btwSession.thinking || '',
    error: btwSession.error || null,
    status: btwSession.status || 'success',
  }));
}

export function shouldApplyBrandingDocumentTitle(options: {
  panePopoutMode?: boolean;
  search?: string | null;
} = {}): boolean {
  if (options.panePopoutMode) return false;
  const search = typeof options.search === 'string' ? options.search : '';
  return !/(?:^|[?&])pane_popout=1(?:&|$)/.test(search);
}

export function scheduleResumeLayoutSettling(
  element: HTMLElement | null | undefined,
  options: {
    durationMs?: number;
    scheduleTimeout?: typeof setTimeout;
    clearScheduledTimeout?: typeof clearTimeout;
  } = {},
): () => void {
  if (!element) return () => {};

  const {
    durationMs = RESUME_LAYOUT_SETTLING_MS,
    scheduleTimeout = setTimeout,
    clearScheduledTimeout = clearTimeout,
  } = options;

  const previous = resumeLayoutSettlingTimers.get(element);
  if (previous) {
    clearScheduledTimeout(previous);
  }

  element.classList.add(RESUME_LAYOUT_SETTLING_CLASS);
  const timer = scheduleTimeout(() => {
    if (resumeLayoutSettlingTimers.get(element) === timer) {
      element.classList.remove(RESUME_LAYOUT_SETTLING_CLASS);
      resumeLayoutSettlingTimers.delete(element);
    }
  }, durationMs);
  resumeLayoutSettlingTimers.set(element, timer);

  return () => {
    const current = resumeLayoutSettlingTimers.get(element);
    if (current) {
      clearScheduledTimeout(current);
      resumeLayoutSettlingTimers.delete(element);
    }
    element.classList.remove(RESUME_LAYOUT_SETTLING_CLASS);
  };
}

export function useAppShellEnvironmentEffects(options: UseAppShellEnvironmentEffectsOptions) {
  const {
    isRenameBranchFormOpen,
    renameBranchNameInputRef,
    appShellRef,
    setIsWebAppMode,
    workspaceOpen,
    setWorkspaceOpen,
    btwSession,
    agents,
    agentsRef,
    currentChatJid,
    activeChatJidRef,
    userProfile,
    userProfileRef,
    brandingRef,
    panePopoutMode = false,
  } = options;

  useTimestampRefresh(30000);

  useEffect(() => {
    if (!isRenameBranchFormOpen) return;
    requestAnimationFrame(() => {
      if (isRenameBranchFormOpen) {
        renameBranchNameInputRef.current?.focus?.();
        renameBranchNameInputRef.current?.select?.();
      }
    });
  }, [isRenameBranchFormOpen, renameBranchNameInputRef]);

  useEffect(() => initTheme(), []);

  useEffect(() => watchStandaloneWebAppMode(setIsWebAppMode), [setIsWebAppMode]);

  useEffect(() => {
    let disposeSettling = () => {};
    const stopWatching = watchReturnToApp(() => {
      disposeSettling();
      disposeSettling = scheduleResumeLayoutSettling(appShellRef.current);
    });

    return () => {
      stopWatching();
      disposeSettling();
    };
  }, [appShellRef]);

  const workspaceLayoutBucketRef = useRef(resolveWorkspaceLayoutBucket());

  useEffect(() => {
    persistWorkspaceOpenPreference(workspaceOpen, {
      bucket: workspaceLayoutBucketRef.current,
    });
  }, [workspaceOpen]);

  useEffect(() => {
    if (typeof window === 'undefined' || !window.matchMedia) return undefined;

    const media = window.matchMedia(DESKTOP_WORKSPACE_LAYOUT_MEDIA_QUERY);
    const applyLayoutPreference = () => {
      const nextBucket = resolveWorkspaceLayoutBucket(window);
      if (workspaceLayoutBucketRef.current === nextBucket) return;
      workspaceLayoutBucketRef.current = nextBucket;
      setWorkspaceOpen(readStoredWorkspaceOpenPreference({
        bucket: nextBucket,
        defaultValue: false,
      }));
    };

    if (media.addEventListener) media.addEventListener('change', applyLayoutPreference);
    else if (media.addListener) media.addListener(applyLayoutPreference);

    return () => {
      if (media.removeEventListener) media.removeEventListener('change', applyLayoutPreference);
      else if (media.removeListener) media.removeListener(applyLayoutPreference);
    };
  }, [setWorkspaceOpen]);

  useEffect(() => installStandaloneMobileViewportFix(), []);

  useEffect(() => {
    persistBtwSession(btwSession);
  }, [btwSession]);

  useEffect(() => {
    agentsRef.current = agents || {};
  }, [agents, agentsRef]);

  useEffect(() => {
    activeChatJidRef.current = currentChatJid;
  }, [activeChatJidRef, currentChatJid]);

  useEffect(() => {
    userProfileRef.current = userProfile || { name: 'You', avatar_url: null, avatar_background: null };
  }, [userProfile, userProfileRef]);

  const applyBranding = useCallback((name: string, avatarUrl: string | null, avatarVersion: string | null = null) => {
    if (typeof document === 'undefined') return;

    const title = (name || '').trim() || 'PiClaw';
    if (brandingRef.current.title !== title) {
      if (shouldApplyBrandingDocumentTitle({
        panePopoutMode,
        search: typeof window !== 'undefined' ? window.location.search : '',
      })) {
        document.title = title;
        const titleMeta = document.querySelector('meta[name="apple-mobile-web-app-title"]');
        if (titleMeta && titleMeta.getAttribute('content') !== title) {
          titleMeta.setAttribute('content', title);
        }
      }
      brandingRef.current.title = title;
    }

    const favicon = document.getElementById('dynamic-favicon') as HTMLLinkElement | null;
    if (!favicon) return;

    // The server already serves /favicon.ico as the agent avatar PNG (with
    // WebP→PNG conversion).  We only need to cache-bust the URL so the
    // browser refetches when the avatar changes.
    const avatarKey = avatarUrl ? `${avatarUrl}|${avatarVersion || ''}` : '';
    if (brandingRef.current.avatarBase !== avatarKey) {
      brandingRef.current.avatarBase = avatarKey;
      const buster = avatarVersion || Date.now();
      favicon.href = `/favicon.ico?v=${buster}`;
    }
  }, [brandingRef]);

  return {
    applyBranding,
  };
}

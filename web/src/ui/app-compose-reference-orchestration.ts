import { useCallback, useEffect } from '../vendor/preact-htm.js';
import { resolveFilePillOpenAction } from './file-pill-open.js';
import {
  appendUniqueStringRef,
  normalizeComposeRefs,
  removeStringRef,
} from './app-shell-ref-utils.js';
import { scrollToTimelineMessage } from './app-timeline-actions.js';

type StateSetter<T> = (next: T | ((prev: T) => T)) => void;

interface RefBox<T> {
  current: T;
}

export function resolveComposeSubmitErrorDetail(message: unknown): string {
  if (typeof message === 'string' && message.trim()) {
    return message.trim();
  }
  return 'Could not send your message.';
}

interface UseComposeReferenceOrchestrationOptions {
  setIntentToast: StateSetter<any>;
  intentToastTimerRef: RefBox<ReturnType<typeof setTimeout> | null>;
  editorOpen: boolean;
  openEditor: (path: string) => void;
  resolvePane: (context: Record<string, unknown>) => unknown;
  tabStripActiveId: string | null;
  setFileRefs: StateSetter<string[]>;
  setMessageRefs: StateSetter<string[]>;
  currentChatJid: string;
  getThread: (id: string | number, chatJid: string) => Promise<any>;
  setPosts: StateSetter<any[] | null>;
}

export function useComposeReferenceOrchestration(options: UseComposeReferenceOrchestrationOptions) {
  const {
    setIntentToast,
    intentToastTimerRef,
    editorOpen,
    openEditor,
    resolvePane,
    tabStripActiveId,
    setFileRefs,
    setMessageRefs,
    currentChatJid,
    getThread,
    setPosts,
  } = options;

  const clearIntentToast = useCallback(() => {
    if (intentToastTimerRef.current) {
      clearTimeout(intentToastTimerRef.current);
      intentToastTimerRef.current = null;
    }
    setIntentToast(null);
  }, [intentToastTimerRef, setIntentToast]);

  useEffect(() => {
    return () => {
      clearIntentToast();
    };
  }, [clearIntentToast]);

  const addFileRef = useCallback((path: unknown) => {
    setFileRefs((prev) => appendUniqueStringRef(prev, path));
  }, [setFileRefs]);

  const removeFileRef = useCallback((path: unknown) => {
    setFileRefs((prev) => removeStringRef(prev, path));
  }, [setFileRefs]);

  const clearFileRefs = useCallback(() => {
    setFileRefs([]);
  }, [setFileRefs]);

  const setFileRefsFromCompose = useCallback((next: unknown) => {
    setFileRefs(normalizeComposeRefs(next));
  }, [setFileRefs]);

  const showIntentToast = useCallback((title: string, detail: string | null = null, kind = 'info', durationMs = 3000) => {
    clearIntentToast();
    setIntentToast({ title, detail: detail || null, kind: kind || 'info' });
    intentToastTimerRef.current = setTimeout(() => {
      setIntentToast((current: any) => (current?.title === title ? null : current));
    }, durationMs);
  }, [clearIntentToast, intentToastTimerRef, setIntentToast]);

  const openFileFromPillWithMode = useCallback((rawPath: unknown, { autoOpenEditor = false } = {}) => {
    const result = resolveFilePillOpenAction(rawPath, {
      editorOpen,
      autoOpenEditor,
      resolvePane,
    });

    if (result.kind === 'open') {
      openEditor(result.path);
      return;
    }

    if (result.kind === 'toast') {
      showIntentToast(result.title, result.detail, result.level);
    }
  }, [editorOpen, openEditor, resolvePane, showIntentToast]);

  const openFileFromPill = useCallback((rawPath: unknown) => {
    openFileFromPillWithMode(rawPath, { autoOpenEditor: false });
  }, [openFileFromPillWithMode]);

  const openTimelineFileFromPill = useCallback((rawPath: unknown) => {
    openFileFromPillWithMode(rawPath, { autoOpenEditor: true });
  }, [openFileFromPillWithMode]);

  const attachActiveEditorFile = useCallback(() => {
    const activeId = tabStripActiveId;
    if (activeId) addFileRef(activeId);
  }, [addFileRef, tabStripActiveId]);

  const addMessageRef = useCallback((id: unknown) => {
    setMessageRefs((prev) => appendUniqueStringRef(prev, id));
  }, [setMessageRefs]);

  const scrollToMessage = useCallback(async (id: string | number, targetChatJid: string | null = null) => {
    await scrollToTimelineMessage({
      id,
      targetChatJid,
      currentChatJid,
      getThread,
      setPosts,
    });
  }, [currentChatJid, getThread, setPosts]);

  const removeMessageRef = useCallback((id: unknown) => {
    setMessageRefs((prev) => removeStringRef(prev, id));
  }, [setMessageRefs]);

  const clearMessageRefs = useCallback(() => {
    setMessageRefs([]);
  }, [setMessageRefs]);

  const setMessageRefsFromCompose = useCallback((next: unknown) => {
    setMessageRefs(normalizeComposeRefs(next));
  }, [setMessageRefs]);

  const handleComposeSubmitError = useCallback((message: unknown) => {
    showIntentToast('Compose failed', resolveComposeSubmitErrorDetail(message), 'error', 5000);
  }, [showIntentToast]);

  return {
    clearIntentToast,
    addFileRef,
    removeFileRef,
    clearFileRefs,
    setFileRefsFromCompose,
    showIntentToast,
    openFileFromPill,
    openTimelineFileFromPill,
    attachActiveEditorFile,
    addMessageRef,
    scrollToMessage,
    removeMessageRef,
    clearMessageRefs,
    setMessageRefsFromCompose,
    handleComposeSubmitError,
  };
}

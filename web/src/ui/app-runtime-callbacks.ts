import { useCallback, useEffect } from '../vendor/preact-htm.js';
import { finalizeStalledResponse as finalizeStalledResponseState } from './app-agent-status-orchestration.js';

type StateSetter<T> = (next: T | ((prev: T) => T)) => void;

interface RefBox<T> {
  current: T;
}

export function removeStalledPostFromTimeline(options: {
  stalledPostIdRef: RefBox<string | number | null>;
  setPosts: StateSetter<any[] | null>;
}): void {
  const { stalledPostIdRef, setPosts } = options;
  const stalledId = stalledPostIdRef.current;
  if (!stalledId) return;
  setPosts((prev) => (prev ? prev.filter((post) => post.id !== stalledId) : prev));
  stalledPostIdRef.current = null;
}

export function runFinalizeStalledResponse(options: Record<string, unknown>): void {
  finalizeStalledResponseState(options as any);
}

export function useViewStateRefSync(options: {
  viewStateRef: RefBox<Record<string, unknown> | null | undefined>;
  currentHashtag: string | null;
  searchQuery: string | null;
  searchOpen: boolean;
}): void {
  const {
    viewStateRef,
    currentHashtag,
    searchQuery,
    searchOpen,
  } = options;

  useEffect(() => {
    viewStateRef.current = { currentHashtag, searchQuery, searchOpen };
  }, [currentHashtag, searchOpen, searchQuery, viewStateRef]);
}

export function useAgentRecoveryCallbacks(options: {
  isAgentRunningRef: RefBox<boolean>;
  lastSilenceNoticeRef: RefBox<number | null>;
  lastAgentEventRef: RefBox<number | null>;
  currentTurnIdRef: RefBox<string | null>;
  thoughtExpandedRef: RefBox<boolean>;
  draftExpandedRef: RefBox<boolean>;
  draftBufferRef: RefBox<string>;
  thoughtBufferRef: RefBox<string>;
  pendingRequestRef: RefBox<any>;
  lastAgentResponseRef: RefBox<any>;
  agentStatusRef: RefBox<any>;
  stalledPostIdRef: RefBox<string | number | null>;
  scrollToBottomRef: RefBox<(() => void) | null>;
  setCurrentTurnId: (value: string | null) => void;
  setAgentDraft: (value: any) => void;
  setAgentPlan: (value: any) => void;
  setAgentThought: (value: any) => void;
  setPendingRequest: (value: any) => void;
  setAgentStatus: (value: any) => void;
  setPosts: StateSetter<any[] | null>;
  dedupePosts: (posts: any[]) => any[];
}) {
  const {
    stalledPostIdRef,
    setPosts,
    isAgentRunningRef,
    lastSilenceNoticeRef,
    lastAgentEventRef,
    currentTurnIdRef,
    thoughtExpandedRef,
    draftExpandedRef,
    draftBufferRef,
    thoughtBufferRef,
    pendingRequestRef,
    lastAgentResponseRef,
    agentStatusRef,
    scrollToBottomRef,
    setCurrentTurnId,
    setAgentDraft,
    setAgentPlan,
    setAgentThought,
    setPendingRequest,
    setAgentStatus,
    dedupePosts,
  } = options;

  const removeStalledPost = useCallback(() => {
    removeStalledPostFromTimeline({
      stalledPostIdRef,
      setPosts,
    });
  }, [setPosts, stalledPostIdRef]);

  const finalizeStalledResponse = useCallback(() => {
    runFinalizeStalledResponse({
      isAgentRunningRef,
      lastSilenceNoticeRef,
      lastAgentEventRef,
      currentTurnIdRef,
      thoughtExpandedRef,
      draftExpandedRef,
      draftBufferRef,
      thoughtBufferRef,
      pendingRequestRef,
      lastAgentResponseRef,
      agentStatusRef,
      stalledPostIdRef,
      scrollToBottomRef,
      setCurrentTurnId,
      setAgentDraft,
      setAgentPlan,
      setAgentThought,
      setPendingRequest,
      setAgentStatus,
      setPosts,
      dedupePosts,
    });
  }, [
    currentTurnIdRef,
    dedupePosts,
    draftBufferRef,
    draftExpandedRef,
    isAgentRunningRef,
    lastAgentEventRef,
    lastAgentResponseRef,
    agentStatusRef,
    lastSilenceNoticeRef,
    pendingRequestRef,
    scrollToBottomRef,
    setAgentDraft,
    setAgentPlan,
    setAgentStatus,
    setAgentThought,
    setCurrentTurnId,
    setPendingRequest,
    setPosts,
    stalledPostIdRef,
    thoughtBufferRef,
    thoughtExpandedRef,
  ]);

  return {
    removeStalledPost,
    finalizeStalledResponse,
  };
}

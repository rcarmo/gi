import { useCallback } from '../vendor/preact-htm.js';
import {
  applyChatPaneStateSnapshot,
  captureChatPaneStateSnapshot,
} from './app-chat-pane-state.js';
import { shouldClearQueuedSteerState } from './queue-state.js';

type StateSetter<T> = (next: T | ((prev: T) => T)) => void;

interface RefBox<T> {
  current: T;
}

export function shouldResetSteerQueue(options: {
  remainingQueueCount: number;
  isAgentTurnActive: boolean;
  steerQueuedTurnId: string | null;
  currentTurnId: string | null;
}): boolean {
  return shouldClearQueuedSteerState(options);
}

interface UseChatPaneRuntimeOrchestrationOptions {
  isAgentTurnActive: boolean;
  steerQueuedTurnId: string | null;
  currentTurnId: string | null;
  steerQueuedTurnIdRef: RefBox<string | null>;
  setSteerQueuedTurnId: (turnId: string | null) => void;

  agentStatus: any;
  agentDraft: any;
  agentPlan: any;
  agentThought: any;
  pendingRequest: any;
  pendingRequestRef: RefBox<any>;
  followupQueueItems: any[];
  activeModel: string | null;
  activeThinkingLevel: string | null;
  supportsThinking: boolean;
  activeModelUsage: any;
  contextUsage: any;

  isAgentRunningRef: RefBox<boolean>;
  wasAgentActiveRef: RefBox<boolean>;
  draftBufferRef: RefBox<string>;
  thoughtBufferRef: RefBox<string>;
  lastAgentEventRef: RefBox<number | null>;
  lastSilenceNoticeRef: RefBox<number>;
  lastAgentResponseRef: RefBox<any>;
  currentTurnIdRef: RefBox<string | null>;
  thoughtExpandedRef: RefBox<boolean>;
  draftExpandedRef: RefBox<boolean>;
  agentStatusRef: RefBox<any>;
  silentRecoveryRef: RefBox<{ inFlight: boolean; lastAttemptAt: number; turnId: string | null }>;

  clearLastActivityTimer: () => void;

  setIsAgentTurnActive: StateSetter<boolean>;
  setAgentStatus: StateSetter<any>;
  setAgentDraft: StateSetter<any>;
  setAgentPlan: StateSetter<any>;
  setAgentThought: StateSetter<any>;
  setPendingRequest: StateSetter<any>;
  setCurrentTurnId: (turnId: string | null) => void;
  setFollowupQueueItems: StateSetter<any[]>;
  setActiveModel: StateSetter<string | null>;
  setActiveThinkingLevel: StateSetter<string | null>;
  setSupportsThinking: StateSetter<boolean>;
  setActiveModelUsage: StateSetter<any>;
  setContextUsage: StateSetter<any>;

  lastNotifiedIdRef: RefBox<string | number | null>;
  agentsRef: RefBox<Record<string, any>>;
  notify: (title: string, body: string) => void;
  shouldNotifyLocallyForChat: (chatJid: string) => boolean;
}

export function formatAgentReplyNotificationBody(post: any): string {
  const content = String(post?.data?.content || '').trim();
  if (content) {
    return content.replace(/\s+/g, ' ').slice(0, 200);
  }
  const contentBlocks = Array.isArray(post?.data?.content_blocks)
    ? post.data.content_blocks
    : [];
  if (contentBlocks.length > 0) {
    return 'Sent an attachment.';
  }
  return '';
}

export function useChatPaneRuntimeOrchestration(options: UseChatPaneRuntimeOrchestrationOptions) {
  const {
    isAgentTurnActive,
    steerQueuedTurnId,
    currentTurnId,
    steerQueuedTurnIdRef,
    setSteerQueuedTurnId,
    agentStatus,
    agentDraft,
    agentPlan,
    agentThought,
    pendingRequest,
    pendingRequestRef,
    followupQueueItems,
    activeModel,
    activeThinkingLevel,
    supportsThinking,
    activeModelUsage,
    contextUsage,
    isAgentRunningRef,
    wasAgentActiveRef,
    draftBufferRef,
    thoughtBufferRef,
    lastAgentEventRef,
    lastSilenceNoticeRef,
    lastAgentResponseRef,
    currentTurnIdRef,
    thoughtExpandedRef,
    draftExpandedRef,
    agentStatusRef,
    silentRecoveryRef,
    clearLastActivityTimer,
    setIsAgentTurnActive,
    setAgentStatus,
    setAgentDraft,
    setAgentPlan,
    setAgentThought,
    setPendingRequest,
    setCurrentTurnId,
    setFollowupQueueItems,
    setActiveModel,
    setActiveThinkingLevel,
    setSupportsThinking,
    setActiveModelUsage,
    setContextUsage,
    lastNotifiedIdRef,
    agentsRef,
    notify,
    shouldNotifyLocallyForChat,
  } = options;

  const clearQueuedSteerStateIfStale = useCallback((remainingQueueCount: number) => {
    if (!shouldResetSteerQueue({
      remainingQueueCount,
      steerQueuedTurnId: steerQueuedTurnIdRef.current,
      currentTurnId: currentTurnIdRef.current,
      isAgentTurnActive,
    })) {
      return;
    }
    steerQueuedTurnIdRef.current = null;
    setSteerQueuedTurnId(null);
  }, [isAgentTurnActive, currentTurnIdRef, setSteerQueuedTurnId, steerQueuedTurnIdRef]);

  const snapshotCurrentChatPaneState = useCallback(() => captureChatPaneStateSnapshot({
    agentStatus,
    agentDraft,
    agentPlan,
    agentThought,
    pendingRequest,
    currentTurnId,
    steerQueuedTurnId,
    isAgentTurnActive,
    followupQueueItems,
    activeModel,
    activeThinkingLevel,
    supportsThinking,
    activeModelUsage,
    contextUsage,
    isAgentRunning: isAgentRunningRef.current,
    wasAgentActive: wasAgentActiveRef.current,
    draftBuffer: draftBufferRef.current,
    thoughtBuffer: thoughtBufferRef.current,
    lastAgentEvent: lastAgentEventRef.current,
    lastSilenceNotice: lastSilenceNoticeRef.current,
    lastAgentResponse: lastAgentResponseRef.current,
    currentTurnIdRef: currentTurnIdRef.current,
    steerQueuedTurnIdRef: steerQueuedTurnIdRef.current,
    thoughtExpanded: thoughtExpandedRef.current,
    draftExpanded: draftExpandedRef.current,
    agentStatusRef: agentStatusRef.current,
    silentRecovery: silentRecoveryRef.current,
  }), [
    activeModel,
    activeModelUsage,
    activeThinkingLevel,
    agentDraft,
    agentPlan,
    agentStatus,
    agentThought,
    contextUsage,
    currentTurnId,
    followupQueueItems,
    isAgentTurnActive,
    pendingRequest,
    steerQueuedTurnId,
    supportsThinking,
    isAgentRunningRef,
    wasAgentActiveRef,
    draftBufferRef,
    thoughtBufferRef,
    lastAgentEventRef,
    lastSilenceNoticeRef,
    lastAgentResponseRef,
    currentTurnIdRef,
    steerQueuedTurnIdRef,
    thoughtExpandedRef,
    draftExpandedRef,
    agentStatusRef,
    silentRecoveryRef,
  ]);

  const restoreChatPaneState = useCallback((snapshot: unknown) => {
    applyChatPaneStateSnapshot({
      snapshot,
      clearLastActivityTimer,
      refs: {
        isAgentRunningRef,
        wasAgentActiveRef,
        lastAgentEventRef,
        lastSilenceNoticeRef,
        draftBufferRef,
        thoughtBufferRef,
        pendingRequestRef,
        lastAgentResponseRef,
        currentTurnIdRef,
        steerQueuedTurnIdRef,
        agentStatusRef,
        silentRecoveryRef,
        thoughtExpandedRef,
        draftExpandedRef,
      },
      setters: {
        setIsAgentTurnActive,
        setAgentStatus,
        setAgentDraft,
        setAgentPlan,
        setAgentThought,
        setPendingRequest,
        setCurrentTurnId,
        setSteerQueuedTurnId,
        setFollowupQueueItems,
        setActiveModel,
        setActiveThinkingLevel,
        setSupportsThinking,
        setActiveModelUsage,
        setContextUsage,
      },
    });
  }, [
    agentStatusRef,
    clearLastActivityTimer,
    currentTurnIdRef,
    draftBufferRef,
    draftExpandedRef,
    isAgentRunningRef,
    lastAgentEventRef,
    lastAgentResponseRef,
    lastSilenceNoticeRef,
    pendingRequestRef,
    setActiveModel,
    setActiveModelUsage,
    setActiveThinkingLevel,
    setAgentDraft,
    setAgentPlan,
    setAgentStatus,
    setAgentThought,
    setContextUsage,
    setCurrentTurnId,
    setFollowupQueueItems,
    setIsAgentTurnActive,
    setPendingRequest,
    setSteerQueuedTurnId,
    setSupportsThinking,
    silentRecoveryRef,
    steerQueuedTurnIdRef,
    thoughtBufferRef,
    thoughtExpandedRef,
    wasAgentActiveRef,
  ]);

  const setActiveTurn = useCallback((turnId: string | null | undefined) => {
    if (!turnId) return;
    if (currentTurnIdRef.current === turnId) return;
    currentTurnIdRef.current = turnId;
    silentRecoveryRef.current = { inFlight: false, lastAttemptAt: 0, turnId };
    setCurrentTurnId(turnId);
    steerQueuedTurnIdRef.current = null;
    setSteerQueuedTurnId(null);
    draftBufferRef.current = '';
    thoughtBufferRef.current = '';
    setAgentDraft({ text: '', totalLines: 0 });
    setAgentPlan('');
    setAgentThought({ text: '', totalLines: 0 });
    setPendingRequest(null);
    pendingRequestRef.current = null;
    lastAgentResponseRef.current = null;
    thoughtExpandedRef.current = false;
    draftExpandedRef.current = false;
  }, [currentTurnIdRef, draftBufferRef, draftExpandedRef, lastAgentResponseRef, pendingRequestRef, setAgentDraft, setAgentPlan, setAgentThought, setCurrentTurnId, setPendingRequest, setSteerQueuedTurnId, silentRecoveryRef, steerQueuedTurnIdRef, thoughtBufferRef, thoughtExpandedRef]);

  const notifyForFinalResponse = useCallback((turnId: string | null | undefined) => {
    const entry = lastAgentResponseRef.current;
    if (!entry || !entry.post) return;
    if (turnId && entry.turnId && entry.turnId !== turnId) return;
    const post = entry.post;
    const chatJid = typeof post?.chat_jid === 'string' && post.chat_jid.trim() ? post.chat_jid.trim() : '';
    if (!chatJid || !shouldNotifyLocallyForChat(chatJid)) return;
    if (post.id && lastNotifiedIdRef.current === post.id) return;
    const body = formatAgentReplyNotificationBody(post);
    if (!body) return;
    lastNotifiedIdRef.current = post.id || lastNotifiedIdRef.current;
    lastAgentResponseRef.current = null;
    const agentsMap = agentsRef.current || {};
    const agent = post?.data?.agent_id ? agentsMap[post.data.agent_id] : null;
    const title = agent?.name || 'Pi';
    notify(title, body, { sourceLabel: 'Local' });
  }, [agentsRef, lastAgentResponseRef, lastNotifiedIdRef, notify, shouldNotifyLocallyForChat]);

  return {
    clearQueuedSteerStateIfStale,
    snapshotCurrentChatPaneState,
    restoreChatPaneState,
    setActiveTurn,
    notifyForFinalResponse,
  };
}

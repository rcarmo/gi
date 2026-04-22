import { useCallback, useEffect } from '../vendor/preact-htm.js';
import { handleAgentPanelToggle } from './app-agent-panel-toggle.js';

type StateSetter<T> = (next: T | ((prev: T) => T)) => void;

interface RefBox<T> {
  current: T;
}

export function clearLastActivityFlagState(previous: any): any {
  if (!previous) return previous;
  if (!(previous.last_activity || previous.lastActivity)) return previous;
  const { last_activity, lastActivity, ...rest } = previous;
  return rest;
}

export function buildLastActivityStatus(payload: any): any {
  if (!payload || typeof payload !== 'object') {
    return { type: 'active', last_activity: true };
  }
  return {
    ...payload,
    last_activity: true,
  };
}

interface UseAgentActivityOrchestrationOptions {
  lastActivityTtlMs: number;
  lastActivityTimerRef: RefBox<ReturnType<typeof setTimeout> | null>;
  lastActivityTokenRef: RefBox<number>;
  lastAgentEventRef: RefBox<number | null>;
  lastSilenceNoticeRef: RefBox<number>;
  isAgentRunningRef: RefBox<boolean>;
  setIsAgentTurnActive: StateSetter<boolean>;
  setAgentStatus: StateSetter<any>;

  draftBufferRef: RefBox<string>;
  thoughtBufferRef: RefBox<string>;
  pendingRequestRef: RefBox<any>;
  lastAgentResponseRef: RefBox<any>;
  currentTurnIdRef: RefBox<string | null>;
  steerQueuedTurnIdRef: RefBox<string | null>;
  agentStatusRef: RefBox<any>;
  silentRecoveryRef: RefBox<{ inFlight: boolean; lastAttemptAt: number; turnId: string | null }>;
  thoughtExpandedRef: RefBox<boolean>;
  draftExpandedRef: RefBox<boolean>;
  setCurrentTurnId: (turnId: string | null) => void;
  setSteerQueuedTurnId: (turnId: string | null) => void;

  currentTurnIdRefForPanel: RefBox<string | null>;
  setAgentThoughtVisibility: (turnId: string, visible: boolean) => Promise<any>;
  getAgentThought: (turnId: string) => Promise<any>;
  setAgentThought: StateSetter<any>;
  setAgentDraft: StateSetter<any>;
}

export function useAgentActivityOrchestration(options: UseAgentActivityOrchestrationOptions) {
  const {
    lastActivityTtlMs,
    lastActivityTimerRef,
    lastActivityTokenRef,
    lastAgentEventRef,
    lastSilenceNoticeRef,
    isAgentRunningRef,
    setIsAgentTurnActive,
    setAgentStatus,
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
    setCurrentTurnId,
    setSteerQueuedTurnId,
    currentTurnIdRefForPanel,
    setAgentThoughtVisibility,
    getAgentThought,
    setAgentThought,
    setAgentDraft,
  } = options;

  const noteAgentActivity = useCallback((activityOptions: Record<string, unknown> = {}) => {
    const activityAtRaw = Number(activityOptions.atMs);
    const now = Number.isFinite(activityAtRaw) && activityAtRaw > 0 ? activityAtRaw : Date.now();
    lastAgentEventRef.current = now;
    if (activityOptions.running) {
      isAgentRunningRef.current = true;
      setIsAgentTurnActive((prev) => (prev ? prev : true));
    }
    if (activityOptions.clearSilence) {
      lastSilenceNoticeRef.current = 0;
    }
  }, [isAgentRunningRef, lastAgentEventRef, lastSilenceNoticeRef, setIsAgentTurnActive]);

  const clearLastActivityTimer = useCallback(() => {
    if (lastActivityTimerRef.current) {
      clearTimeout(lastActivityTimerRef.current);
      lastActivityTimerRef.current = null;
    }
    lastActivityTokenRef.current = 0;
  }, [lastActivityTimerRef, lastActivityTokenRef]);

  useEffect(() => () => {
    clearLastActivityTimer();
  }, [clearLastActivityTimer]);

  const clearLastActivityFlag = useCallback(() => {
    clearLastActivityTimer();
    setAgentStatus((prev: any) => clearLastActivityFlagState(prev));
  }, [clearLastActivityTimer, setAgentStatus]);

  const showLastActivity = useCallback((payload: any) => {
    if (!payload) return;
    clearLastActivityTimer();
    const token = Date.now();
    lastActivityTokenRef.current = token;
    setAgentStatus(buildLastActivityStatus(payload));
    lastActivityTimerRef.current = setTimeout(() => {
      if (lastActivityTokenRef.current !== token) return;
      setAgentStatus((prev: any) => {
        if (!prev || !(prev.last_activity || prev.lastActivity)) return prev;
        return null;
      });
    }, lastActivityTtlMs);
  }, [clearLastActivityTimer, lastActivityTimerRef, lastActivityTokenRef, lastActivityTtlMs, setAgentStatus]);

  const clearAgentRunState = useCallback(() => {
    isAgentRunningRef.current = false;
    setIsAgentTurnActive(false);
    lastAgentEventRef.current = null;
    lastSilenceNoticeRef.current = 0;
    draftBufferRef.current = '';
    thoughtBufferRef.current = '';
    pendingRequestRef.current = null;
    lastAgentResponseRef.current = null;
    currentTurnIdRef.current = null;
    steerQueuedTurnIdRef.current = null;
    agentStatusRef.current = null;
    silentRecoveryRef.current = { inFlight: false, lastAttemptAt: 0, turnId: null };
    clearLastActivityTimer();
    setCurrentTurnId(null);
    setSteerQueuedTurnId(null);
    thoughtExpandedRef.current = false;
    draftExpandedRef.current = false;
  }, [agentStatusRef, clearLastActivityTimer, currentTurnIdRef, draftBufferRef, draftExpandedRef, isAgentRunningRef, lastAgentEventRef, lastAgentResponseRef, lastSilenceNoticeRef, pendingRequestRef, setCurrentTurnId, setIsAgentTurnActive, setSteerQueuedTurnId, silentRecoveryRef, steerQueuedTurnIdRef, thoughtBufferRef, thoughtExpandedRef]);

  const handlePanelToggle = useCallback(async (panelKey: string, expanded: boolean) => {
    await handleAgentPanelToggle({
      panelKey,
      expanded,
      currentTurnIdRef: currentTurnIdRefForPanel,
      thoughtExpandedRef,
      draftExpandedRef,
      setAgentThoughtVisibility,
      getAgentThought,
      thoughtBufferRef,
      draftBufferRef,
      setAgentThought,
      setAgentDraft,
    });
  }, [currentTurnIdRefForPanel, draftBufferRef, draftExpandedRef, getAgentThought, setAgentDraft, setAgentThought, setAgentThoughtVisibility, thoughtBufferRef, thoughtExpandedRef]);

  return {
    noteAgentActivity,
    clearLastActivityTimer,
    clearLastActivityFlag,
    showLastActivity,
    clearAgentRunState,
    handlePanelToggle,
  };
}

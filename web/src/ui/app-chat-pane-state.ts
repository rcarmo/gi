export interface AppTextPreviewState {
  text: string;
  totalLines: number;
}

export interface AppSilentRecoveryState {
  inFlight: boolean;
  lastAttemptAt: number;
  turnId: string | null;
}

export interface ChatPaneStateSnapshot {
  agentStatus: unknown;
  agentDraft: AppTextPreviewState;
  agentPlan: string;
  agentThought: AppTextPreviewState;
  pendingRequest: unknown;
  currentTurnId: string | null;
  steerQueuedTurnId: string | null;
  isAgentTurnActive: boolean;
  followupQueueItems: Array<Record<string, unknown>>;
  activeModel: unknown;
  activeThinkingLevel: unknown;
  supportsThinking: boolean;
  activeModelUsage: unknown;
  contextUsage: unknown;
  isAgentRunning: boolean;
  wasAgentActive: boolean;
  draftBuffer: string;
  thoughtBuffer: string;
  lastAgentEvent: unknown;
  lastSilenceNotice: number;
  lastAgentResponse: unknown;
  currentTurnIdRef: string | null;
  steerQueuedTurnIdRef: string | null;
  thoughtExpanded: boolean;
  draftExpanded: boolean;
  agentStatusRef: unknown;
  silentRecovery: AppSilentRecoveryState;
}

export interface ChatPaneStateSnapshotSource {
  agentStatus: unknown;
  agentDraft: AppTextPreviewState | null | undefined;
  agentPlan: string | null | undefined;
  agentThought: AppTextPreviewState | null | undefined;
  pendingRequest: unknown;
  currentTurnId: string | null | undefined;
  steerQueuedTurnId: string | null | undefined;
  isAgentTurnActive: boolean;
  followupQueueItems: Array<Record<string, unknown>> | null | undefined;
  activeModel: unknown;
  activeThinkingLevel: unknown;
  supportsThinking: boolean;
  activeModelUsage: unknown;
  contextUsage: unknown;
  isAgentRunning: boolean;
  wasAgentActive: boolean;
  draftBuffer: string | null | undefined;
  thoughtBuffer: string | null | undefined;
  lastAgentEvent: unknown;
  lastSilenceNotice: number | null | undefined;
  lastAgentResponse: unknown;
  currentTurnIdRef: string | null | undefined;
  steerQueuedTurnIdRef: string | null | undefined;
  thoughtExpanded: boolean;
  draftExpanded: boolean;
  agentStatusRef: unknown;
  silentRecovery: Partial<AppSilentRecoveryState> | null | undefined;
}

export interface MutableRefLike<T> {
  current: T;
}

export interface ChatPaneStateRestoreRefs {
  isAgentRunningRef: MutableRefLike<boolean>;
  wasAgentActiveRef: MutableRefLike<boolean>;
  lastAgentEventRef: MutableRefLike<unknown>;
  lastSilenceNoticeRef: MutableRefLike<number>;
  draftBufferRef: MutableRefLike<string>;
  thoughtBufferRef: MutableRefLike<string>;
  pendingRequestRef: MutableRefLike<unknown>;
  lastAgentResponseRef: MutableRefLike<unknown>;
  currentTurnIdRef: MutableRefLike<string | null>;
  steerQueuedTurnIdRef: MutableRefLike<string | null>;
  agentStatusRef: MutableRefLike<unknown>;
  silentRecoveryRef: MutableRefLike<AppSilentRecoveryState>;
  thoughtExpandedRef: MutableRefLike<boolean>;
  draftExpandedRef: MutableRefLike<boolean>;
}

export interface ChatPaneStateRestoreSetters {
  setIsAgentTurnActive(value: boolean): void;
  setAgentStatus(value: unknown): void;
  setAgentDraft(value: AppTextPreviewState): void;
  setAgentPlan(value: string): void;
  setAgentThought(value: AppTextPreviewState): void;
  setPendingRequest(value: unknown): void;
  setCurrentTurnId(value: string | null): void;
  setSteerQueuedTurnId(value: string | null): void;
  setFollowupQueueItems(value: Array<Record<string, unknown>>): void;
  setActiveModel(value: unknown): void;
  setActiveThinkingLevel(value: unknown): void;
  setSupportsThinking(value: boolean): void;
  setActiveModelUsage(value: unknown): void;
  setContextUsage(value: unknown): void;
}

export interface ChatPaneStateRestoreOptions {
  snapshot: ChatPaneStateSnapshot | null | undefined;
  clearLastActivityTimer?(): void;
  refs: ChatPaneStateRestoreRefs;
  setters: ChatPaneStateRestoreSetters;
}

function cloneTextPreviewState(value: AppTextPreviewState | null | undefined): AppTextPreviewState {
  return value ? { ...value } : { text: '', totalLines: 0 };
}

function cloneFollowupQueueItems(items: Array<Record<string, unknown>> | null | undefined): Array<Record<string, unknown>> {
  return Array.isArray(items) ? items.map((item) => ({ ...item })) : [];
}

function normalizeSilentRecoveryState(value: Partial<AppSilentRecoveryState> | null | undefined): AppSilentRecoveryState {
  return {
    inFlight: Boolean(value?.inFlight),
    lastAttemptAt: Number(value?.lastAttemptAt || 0),
    turnId: typeof value?.turnId === 'string' ? value.turnId : null,
  };
}

export function createEmptyChatPaneState(): ChatPaneStateSnapshot {
  return {
    agentStatus: null,
    agentDraft: { text: '', totalLines: 0 },
    agentPlan: '',
    agentThought: { text: '', totalLines: 0 },
    pendingRequest: null,
    currentTurnId: null,
    steerQueuedTurnId: null,
    isAgentTurnActive: false,
    followupQueueItems: [],
    activeModel: null,
    activeThinkingLevel: null,
    supportsThinking: false,
    activeModelUsage: null,
    contextUsage: null,
    isAgentRunning: false,
    wasAgentActive: false,
    draftBuffer: '',
    thoughtBuffer: '',
    lastAgentEvent: null,
    lastSilenceNotice: 0,
    lastAgentResponse: null,
    currentTurnIdRef: null,
    steerQueuedTurnIdRef: null,
    thoughtExpanded: false,
    draftExpanded: false,
    agentStatusRef: null,
    silentRecovery: { inFlight: false, lastAttemptAt: 0, turnId: null },
  };
}

export function captureChatPaneStateSnapshot(source: ChatPaneStateSnapshotSource): ChatPaneStateSnapshot {
  return {
    agentStatus: source.agentStatus,
    agentDraft: cloneTextPreviewState(source.agentDraft),
    agentPlan: source.agentPlan || '',
    agentThought: cloneTextPreviewState(source.agentThought),
    pendingRequest: source.pendingRequest,
    currentTurnId: source.currentTurnId || null,
    steerQueuedTurnId: source.steerQueuedTurnId || null,
    isAgentTurnActive: Boolean(source.isAgentTurnActive),
    followupQueueItems: cloneFollowupQueueItems(source.followupQueueItems),
    activeModel: source.activeModel,
    activeThinkingLevel: source.activeThinkingLevel,
    supportsThinking: Boolean(source.supportsThinking),
    activeModelUsage: source.activeModelUsage,
    contextUsage: source.contextUsage,
    isAgentRunning: Boolean(source.isAgentRunning),
    wasAgentActive: Boolean(source.wasAgentActive),
    draftBuffer: source.draftBuffer || '',
    thoughtBuffer: source.thoughtBuffer || '',
    lastAgentEvent: source.lastAgentEvent || null,
    lastSilenceNotice: Number(source.lastSilenceNotice || 0),
    lastAgentResponse: source.lastAgentResponse || null,
    currentTurnIdRef: source.currentTurnIdRef || null,
    steerQueuedTurnIdRef: source.steerQueuedTurnIdRef || null,
    thoughtExpanded: Boolean(source.thoughtExpanded),
    draftExpanded: Boolean(source.draftExpanded),
    agentStatusRef: source.agentStatusRef || null,
    silentRecovery: normalizeSilentRecoveryState(source.silentRecovery),
  };
}

export function applyChatPaneStateSnapshot(options: ChatPaneStateRestoreOptions): ChatPaneStateSnapshot {
  const next = options.snapshot || createEmptyChatPaneState();
  const { refs, setters } = options;

  options.clearLastActivityTimer?.();
  refs.isAgentRunningRef.current = Boolean(next.isAgentRunning);
  refs.wasAgentActiveRef.current = Boolean(next.wasAgentActive);
  setters.setIsAgentTurnActive(Boolean(next.isAgentTurnActive));
  refs.lastAgentEventRef.current = next.lastAgentEvent || null;
  refs.lastSilenceNoticeRef.current = Number(next.lastSilenceNotice || 0);
  refs.draftBufferRef.current = next.draftBuffer || '';
  refs.thoughtBufferRef.current = next.thoughtBuffer || '';
  refs.pendingRequestRef.current = next.pendingRequest || null;
  refs.lastAgentResponseRef.current = next.lastAgentResponse || null;
  refs.currentTurnIdRef.current = next.currentTurnIdRef || null;
  refs.steerQueuedTurnIdRef.current = next.steerQueuedTurnIdRef || null;
  refs.agentStatusRef.current = next.agentStatusRef || null;
  refs.silentRecoveryRef.current = next.silentRecovery || { inFlight: false, lastAttemptAt: 0, turnId: null };
  refs.thoughtExpandedRef.current = Boolean(next.thoughtExpanded);
  refs.draftExpandedRef.current = Boolean(next.draftExpanded);
  setters.setAgentStatus(next.agentStatus || null);
  setters.setAgentDraft(cloneTextPreviewState(next.agentDraft));
  setters.setAgentPlan(next.agentPlan || '');
  setters.setAgentThought(cloneTextPreviewState(next.agentThought));
  setters.setPendingRequest(next.pendingRequest || null);
  setters.setCurrentTurnId(next.currentTurnId || null);
  setters.setSteerQueuedTurnId(next.steerQueuedTurnId || null);
  setters.setFollowupQueueItems(cloneFollowupQueueItems(next.followupQueueItems));
  setters.setActiveModel(next.activeModel || null);
  setters.setActiveThinkingLevel(next.activeThinkingLevel || null);
  setters.setSupportsThinking(Boolean(next.supportsThinking));
  setters.setActiveModelUsage(next.activeModelUsage ?? null);
  setters.setContextUsage(next.contextUsage ?? null);
  return next;
}

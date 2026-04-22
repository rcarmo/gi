export interface SseEventRoutingContext {
  turnId: unknown;
  eventChatJid: string | null;
  isGlobalUiEvent: boolean;
  isCurrentChatEvent: boolean;
}

export function resolveSseEventRoutingContext(
  eventType: string | null | undefined,
  payload: Record<string, unknown> | null | undefined,
  currentChatJid: string | null | undefined,
): SseEventRoutingContext {
  const turnId = payload?.turn_id;
  const rawChatJid = payload?.chat_jid;
  const eventChatJid = typeof rawChatJid === 'string' && rawChatJid.trim() ? rawChatJid.trim() : null;
  const isGlobalUiEvent = eventType === 'connected' || eventType === 'workspace_update';
  const isCurrentChatEvent = eventChatJid ? eventChatJid === currentChatJid : isGlobalUiEvent;

  return {
    turnId,
    eventChatJid,
    isGlobalUiEvent,
    isCurrentChatEvent,
  };
}

export function isNoisyAgentSseEvent(eventType: string | null | undefined): boolean {
  return eventType === 'agent_draft_delta'
    || eventType === 'agent_thought_delta'
    || eventType === 'agent_draft'
    || eventType === 'agent_thought';
}

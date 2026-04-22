export interface AgentPreviewStateLike {
  text?: unknown;
  totalLines?: unknown;
  total_lines?: unknown;
  [key: string]: unknown;
}

export function readAgentTurnId(payload: Record<string, unknown> | null | undefined): unknown {
  return payload?.turn_id || payload?.turnId || null;
}

export function resolveAgentPreviewRestoreState(
  payload: AgentPreviewStateLike | null | undefined,
): { text: string; totalLines: number } | null {
  if (typeof payload?.text !== 'string' || !payload.text) return null;
  const totalLines = Number.isFinite(payload?.totalLines as number)
    ? Number(payload.totalLines)
    : (Number.isFinite(payload?.total_lines as number) ? Number(payload.total_lines) : 0);
  return {
    text: payload.text,
    totalLines,
  };
}

export function shouldKeepExistingPreview(
  previous: AgentPreviewStateLike | null | undefined,
  incomingText: string,
): boolean {
  return typeof previous?.text === 'string' && previous.text.length >= incomingText.length;
}

import { estimatePreviewLines } from './app-helpers.js';

export interface AgentPreviewState {
  text: string;
  totalLines: number;
  fullText?: string;
}

export function inferAgentPreviewTotalLines(text: string, explicitTotalLines: unknown): number {
  return Number.isFinite(explicitTotalLines as number)
    ? Number(explicitTotalLines)
    : (text ? text.replace(/\r\n/g, '\n').split('\n').length : 0);
}

export function buildCollapsedAgentPreviewState(text: string, explicitTotalLines: unknown): AgentPreviewState {
  return {
    text,
    totalLines: inferAgentPreviewTotalLines(text, explicitTotalLines),
  };
}

export function buildExpandedAgentPreviewState(fullText: string, previousState: AgentPreviewState | null | undefined): AgentPreviewState {
  return {
    text: previousState?.text || '',
    totalLines: estimatePreviewLines(fullText),
    fullText,
  };
}

export function resolveAgentPlanText(currentPlan: string | null | undefined, text: string, mode: unknown): string {
  return mode === 'replace' ? text : `${currentPlan || ''}${text}`;
}

export function applyDraftDeltaBuffer(currentBuffer: string | null | undefined, data: Record<string, any> | null | undefined): string {
  let next = currentBuffer || '';
  if (data?.reset) {
    next = '';
  }
  if (data?.delta) {
    next += String(data.delta);
  }
  return next;
}

export function applyThoughtDeltaBuffer(currentBuffer: string | null | undefined, data: Record<string, any> | null | undefined): string {
  let next = currentBuffer || '';
  if (data?.reset) {
    next = '';
  }
  if (typeof data?.delta === 'string') {
    next += data.delta;
  }
  return next;
}

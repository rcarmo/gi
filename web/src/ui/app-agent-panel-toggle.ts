export interface AgentPreviewState {
  text?: string;
  totalLines?: number;
  fullText?: string;
}

export interface HandleAgentPanelToggleOptions {
  panelKey: unknown;
  expanded: boolean;
  currentTurnIdRef: { current: string | null };
  thoughtExpandedRef: { current: boolean };
  draftExpandedRef: { current: boolean };
  setAgentThoughtVisibility: (turnId: string, panel: 'thought' | 'draft', expanded: boolean) => Promise<unknown>;
  getAgentThought: (turnId: string, panel: 'thought' | 'draft') => Promise<any>;
  thoughtBufferRef: { current: string };
  draftBufferRef: { current: string };
  setAgentThought: (next: AgentPreviewState | ((prev: AgentPreviewState | null | undefined) => AgentPreviewState)) => void;
  setAgentDraft: (next: AgentPreviewState | ((prev: AgentPreviewState | null | undefined) => AgentPreviewState)) => void;
}

export async function handleAgentPanelToggle(options: HandleAgentPanelToggleOptions): Promise<void> {
  const {
    panelKey,
    expanded,
    currentTurnIdRef,
    thoughtExpandedRef,
    draftExpandedRef,
    setAgentThoughtVisibility,
    getAgentThought,
    thoughtBufferRef,
    draftBufferRef,
    setAgentThought,
    setAgentDraft,
  } = options;

  if (panelKey !== 'thought' && panelKey !== 'draft') return;

  const turnId = currentTurnIdRef.current;
  if (panelKey === 'thought') {
    thoughtExpandedRef.current = expanded;
    if (turnId) {
      try {
        await setAgentThoughtVisibility(turnId, 'thought', expanded);
      } catch (error) {
        console.warn('Failed to update thought visibility:', error);
      }
    }
    if (!expanded) return;
    try {
      const data = turnId ? await getAgentThought(turnId, 'thought') : null;
      if (data?.text) {
        thoughtBufferRef.current = data.text;
      }
      setAgentThought((prev) => ({
        ...(prev || { text: '', totalLines: 0 }),
        fullText: thoughtBufferRef.current || prev?.fullText || '',
        totalLines: Number.isFinite(data?.total_lines) ? data.total_lines : prev?.totalLines || 0,
      }));
    } catch (error) {
      console.warn('Failed to fetch full thought:', error);
    }
    return;
  }

  draftExpandedRef.current = expanded;
  if (turnId) {
    try {
      await setAgentThoughtVisibility(turnId, 'draft', expanded);
    } catch (error) {
      console.warn('Failed to update draft visibility:', error);
    }
  }
  if (!expanded) return;

  try {
    const data = turnId ? await getAgentThought(turnId, 'draft') : null;
    if (data?.text) {
      draftBufferRef.current = data.text;
    }
    setAgentDraft((prev) => ({
      ...(prev || { text: '', totalLines: 0 }),
      fullText: draftBufferRef.current || prev?.fullText || '',
      totalLines: Number.isFinite(data?.total_lines) ? data.total_lines : prev?.totalLines || 0,
    }));
  } catch (error) {
    console.warn('Failed to fetch full draft:', error);
  }
}

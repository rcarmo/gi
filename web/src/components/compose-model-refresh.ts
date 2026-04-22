export async function refreshAgentModelStateBestEffort(
  getAgentModels: ((chatJid: string) => Promise<unknown>) | null | undefined,
  chatJid: string,
  emitModelState: (state: unknown) => void,
): Promise<boolean> {
  if (typeof getAgentModels !== 'function') return false;
  try {
    const latest = await getAgentModels(chatJid);
    if (!latest) return false;
    emitModelState(latest);
    return true;
  } catch (_error) {
    return false;
  }
}

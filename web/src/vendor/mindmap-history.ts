export interface MindmapHistoryEntry<T> {
  document: T;
  label: string;
}

export interface MindmapHistoryState<T> {
  undoStack: MindmapHistoryEntry<T>[];
  redoStack: MindmapHistoryEntry<T>[];
}

export function cloneMindmapHistoryDocument<T>(document: T): T {
  if (typeof structuredClone === 'function') {
    return structuredClone(document);
  }
  return JSON.parse(JSON.stringify(document)) as T;
}

export function recordMindmapHistory<T>(
  state: MindmapHistoryState<T>,
  currentDocument: T,
  label: string,
): MindmapHistoryState<T> {
  return {
    undoStack: [...state.undoStack, { document: cloneMindmapHistoryDocument(currentDocument), label }],
    redoStack: [],
  };
}

export function undoMindmapHistory<T>(
  state: MindmapHistoryState<T>,
  currentDocument: T,
): (MindmapHistoryState<T> & { restored: T | null; label: string | null }) {
  if (state.undoStack.length === 0) {
    return { ...state, restored: null, label: null };
  }
  const restoredEntry = state.undoStack[state.undoStack.length - 1];
  return {
    undoStack: state.undoStack.slice(0, -1),
    redoStack: [...state.redoStack, { document: cloneMindmapHistoryDocument(currentDocument), label: restoredEntry.label }],
    restored: cloneMindmapHistoryDocument(restoredEntry.document),
    label: restoredEntry.label,
  };
}

export function redoMindmapHistory<T>(
  state: MindmapHistoryState<T>,
  currentDocument: T,
): (MindmapHistoryState<T> & { restored: T | null; label: string | null }) {
  if (state.redoStack.length === 0) {
    return { ...state, restored: null, label: null };
  }
  const restoredEntry = state.redoStack[state.redoStack.length - 1];
  return {
    undoStack: [...state.undoStack, { document: cloneMindmapHistoryDocument(currentDocument), label: restoredEntry.label }],
    redoStack: state.redoStack.slice(0, -1),
    restored: cloneMindmapHistoryDocument(restoredEntry.document),
    label: restoredEntry.label,
  };
}

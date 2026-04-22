export interface LayoutPosition {
  x: number;
  y: number;
}

export interface LayoutPositionNode {
  position?: LayoutPosition;
  children?: LayoutPositionNode[];
}

export function clearStoredNodePositions(node: LayoutPositionNode | null | undefined): number {
  if (!node || typeof node !== 'object') return 0;

  let cleared = 0;
  if (node.position) {
    delete node.position;
    cleared += 1;
  }

  for (const child of node.children ?? []) {
    cleared += clearStoredNodePositions(child);
  }

  return cleared;
}

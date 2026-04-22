// @ts-nocheck
import { html, useEffect, useState } from '../vendor/preact-htm-entry.ts';
import { getWorkspaceTree, getWorkspaceFile } from '../api.ts';

function flatten(node, expanded, depth = 0, rows = []) {
  if (!node) return rows;
  rows.push({ node, depth });
  if (node.type === 'dir' && expanded.has(node.path) && Array.isArray(node.children)) {
    for (const child of node.children) flatten(child, expanded, depth + 1, rows);
  }
  return rows;
}

export function WorkspaceBrowser({ onOpenFile }) {
  const [tree, setTree] = useState(null);
  const [expanded, setExpanded] = useState(() => new Set(['.']));
  useEffect(() => { getWorkspaceTree().then(setTree).catch((err) => console.error('[workspace] tree load failed', err)); }, []);
  const rows = flatten(tree, expanded);
  return html`<aside class="workspace-sidebar gi-workspace-sidebar" style=${{display:'flex'}}><div class="panel-title">Workspace</div><div class="workspace-tree">${rows.map(({ node, depth }) => html`<button class="workspace-row" style=${{paddingLeft:`${8 + depth*14}px`}} onClick=${async () => {
    if (node.type === 'dir') {
      const next = new Set(expanded); if (next.has(node.path)) next.delete(node.path); else next.add(node.path); setExpanded(next); return;
    }
    const file = await getWorkspaceFile(node.path); onOpenFile?.(file.path, file.content);
  }}><span class="workspace-row-icon">${node.type === 'dir' ? (expanded.has(node.path) ? '▾' : '▸') : '•'}</span><span class="workspace-row-label">${node.name}</span></button>`)}</div></aside>`;
}

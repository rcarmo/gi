// Exact guarded adapter for the supplied status renderer; upstream stays verbatim.
export function patchStatusPreview(source) {
  const patches = [
    ["// @ts-nocheck\n", "// @ts-nocheck\nimport { usePreviewOverflow } from '../gi-preview-overflow.js';\n"],
    ["    const [expandedPanels, setExpandedPanels] = useState(new Set());", "    const [expandedPanels, setExpandedPanels] = useState(new Set());\n    const previewOverflow = usePreviewOverflow(draft, thought, expandedPanels);"],
    ["        const bodyClass = `agent-thinking-body", "        const canDisclose = truncated.omitted > 0 || previewOverflow.overflow[panelKey] === true;\n        const bodyClass = `agent-thinking-body"],
    ["                    class=${bodyClass}\n", "                    ref=${previewOverflow.refs[panelKey]}\n                    class=${bodyClass}\n"],
    ["${!isExpanded && truncated.omitted > 0 && html`", "${!isExpanded && canDisclose && html`"],
    ["                        ▸ ${truncated.omitted} more lines", "                        ${truncated.omitted > 0 ? `▸ ${truncated.omitted} more lines` : '▸ Show more'}"],
    ["${isExpanded && truncated.omitted > 0 && html`", "${isExpanded && canDisclose && html`"],
  ];
  for (const [from, to] of patches) {
    if (source.split(from).length !== 2) throw new Error(`Status preview adapter anchor changed: ${from}`);
    source = source.replace(from, to);
  }
  return source;
}

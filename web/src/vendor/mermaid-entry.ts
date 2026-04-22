import {
  renderMermaidSVG,
  renderMermaidSVGAsync,
  renderMermaid,
  renderMermaidSync,
  THEMES,
  DEFAULTS,
  fromShikiTheme,
  parseMermaid,
} from 'beautiful-mermaid';

(globalThis as any).beautifulMermaid = {
  renderMermaid,
  renderMermaidSync,
  renderMermaidSVG,
  renderMermaidSVGAsync,
  THEMES,
  DEFAULTS,
  fromShikiTheme,
  parseMermaid,
};

/**
 * Browser entry point for beautiful-mermaid vendor bundle.
 *
 * Exposes the library as `globalThis.beautifulMermaid` (IIFE-style)
 * so markdown.ts can access it via `window.beautifulMermaid`.
 *
 * Build with:
 *   bun run build:vendor:mermaid
 *
 * Rebuild or upgrade reproducibly with:
 *   bun run update:vendor:mermaid
 *   bun run update:vendor:mermaid --version 1.1.3
 */

import {
  renderMermaidSVG,
  renderMermaidSVGAsync,
  renderMermaid,
  renderMermaidSync,
  THEMES,
  DEFAULTS,
  fromShikiTheme,
  parseMermaid,
} from "beautiful-mermaid";

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
